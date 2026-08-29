package price

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/amelamela/vault-lab/internal/geo"
	"github.com/amelamela/vault-lab/internal/model"
)

// ETFFetcher resolves the geographic and sector exposure of an ETF from its ISIN.
type ETFFetcher interface {
	FetchExposure(ctx context.Context, isin string) (*model.AssetExposure, error)
}

// JustETFFetcher fetches ETF exposure from the python-service microservice.
type JustETFFetcher struct {
	baseURL string
	client  *http.Client
}

const (
	etfMaxRetries = 2
	etfRetryDelay = 400 * time.Millisecond
)

type etfExposureRow struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

type etfExposureResponse struct {
	Countries []etfExposureRow `json:"countries"`
	Sectors   []etfExposureRow `json:"sectors"`
}

type etfErrorResponse struct {
	Detail string `json:"detail"`
}

// NewJustETFFetcher builds a fetcher that talks to the python-service base URL.
func NewJustETFFetcher(baseURL string) *JustETFFetcher {
	return &JustETFFetcher{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// FetchExposure calls the python-service exposure endpoint for an ISIN. Non-2xx
// responses are returned without retrying; network/timeout failures are retried
// up to etfMaxRetries times with a small backoff.
func (f *JustETFFetcher) FetchExposure(ctx context.Context, isin string) (*model.AssetExposure, error) {
	url := fmt.Sprintf("%s/api/v1/etf/%s/exposure", f.baseURL, isin)

	var err error
	for attempt := 0; attempt <= etfMaxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, etfRetryDelay); err != nil {
				return nil, err
			}
		}
		var exposure *model.AssetExposure
		exposure, err = f.fetchOnce(ctx, url)
		if err == nil {
			return exposure, nil
		}
		var transportErr *etfTransportError
		if !errors.As(err, &transportErr) {
			return nil, err
		}
	}
	return nil, err
}

func (f *JustETFFetcher) fetchOnce(ctx context.Context, url string) (*model.AssetExposure, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, &etfTransportError{err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &etfStatusError{status: resp.StatusCode, message: statusMessage(resp.StatusCode, body)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result etfExposureResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}

	return &model.AssetExposure{
		Regions: exposureRows(result.Countries),
		Sectors: exposureRows(result.Sectors),
	}, nil
}

// statusMessage builds a descriptive error for a non-2xx response, including
// the upstream detail when the body is the documented JSON error.
func statusMessage(status int, body []byte) string {
	msg := fmt.Sprintf("python service returned status %d", status)
	var errResp etfErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Detail != "" {
		msg = fmt.Sprintf("%s: %s", msg, errResp.Detail)
	}
	return msg
}

// exposureRows converts the raw python-service rows into model rows.
func exposureRows(rows []etfExposureRow) []model.ExposureRow {
	out := make([]model.ExposureRow, 0, len(rows))
	for _, row := range rows {
		if row.Name == "" || row.Weight <= 0 {
			continue
		}
		out = append(out, model.ExposureRow{Name: row.Name, Weight: decimal.NewFromFloat(row.Weight)})
	}
	return out
}

// AggregateRegions maps raw country names to canonical macro-regions, summing
// the weights of countries in the same region. Country names that resolve to no
// ISO code land in the geo package's fallback region.
func AggregateRegions(rows []model.ExposureRow) []model.ExposureRow {
	weights := map[string]decimal.Decimal{}
	for _, row := range rows {
		if row.Name == "" || !row.Weight.IsPositive() {
			continue
		}
		region := geo.RegionForCountry(geo.NormalizeCountry(row.Name))
		if region == "" {
			region = row.Name
		}
		weights[region] = weights[region].Add(row.Weight)
	}
	return sortedExposureRows(weights)
}

// AggregateSectors maps raw provider sector names to canonical GICS sectors,
// summing the weights of names that normalize to the same sector.
func AggregateSectors(rows []model.ExposureRow) []model.ExposureRow {
	weights := map[string]decimal.Decimal{}
	for _, row := range rows {
		if row.Name == "" || !row.Weight.IsPositive() {
			continue
		}
		sector := geo.NormalizeSector(row.Name)
		if sector == "" {
			continue
		}
		weights[sector] = weights[sector].Add(row.Weight)
	}
	return sortedExposureRows(weights)
}

func sortedExposureRows(weights map[string]decimal.Decimal) []model.ExposureRow {
	rows := make([]model.ExposureRow, 0, len(weights))
	for name, weight := range weights {
		rows = append(rows, model.ExposureRow{Name: name, Weight: weight})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// etfStatusError reports a non-2xx response from the python-service.
type etfStatusError struct {
	status  int
	message string
}

func (e *etfStatusError) Error() string { return e.message }

// etfTransportError reports a network/timeout failure reaching the
// python-service. It is retryable, unlike status and parse errors.
type etfTransportError struct {
	err error
}

func (e *etfTransportError) Error() string { return fmt.Sprintf("python service request: %v", e.err) }
func (e *etfTransportError) Unwrap() error { return e.err }
