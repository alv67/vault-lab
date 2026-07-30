package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/repository"
)

type YahooFetcher struct {
	repos  *repository.Repository
	client *http.Client
}

func NewYahooFetcher(repos *repository.Repository) *YahooFetcher {
	return &YahooFetcher{
		repos: repos,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type yahooQuoteResult struct {
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow float64 `json:"regularMarketDayLow"`
	RegularMarketOpen float64 `json:"regularMarketOpen"`
	RegularMarketVolume int64 `json:"regularMarketVolume"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
}

type yahooResponse struct {
	QuoteResponse struct {
		Result []struct {
			Symbol string `json:"symbol"`
			RegularMarketPrice float64 `json:"regularMarketPrice"`
			RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
			RegularMarketDayLow float64 `json:"regularMarketDayLow"`
			RegularMarketOpen float64 `json:"regularMarketOpen"`
			RegularMarketVolume int64 `json:"regularMarketVolume"`
			RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"quoteResponse"`
}

func (f *YahooFetcher) FetchAll(ctx context.Context) error {
	assets, err := f.repos.Asset.List(ctx)
	if err != nil {
		return fmt.Errorf("list assets: %w", err)
	}

	if len(assets) == 0 {
		log.Info().Msg("no assets to fetch prices for")
		return nil
	}

	// Process in batches of 10
	batchSize := 10
	for i := 0; i < len(assets); i += batchSize {
		end := i + batchSize
		if end > len(assets) {
			end = len(assets)
		}
		batch := assets[i:end]

		if err := f.fetchBatch(ctx, batch); err != nil {
			log.Warn().Err(err).Int("batch", i/batchSize).Msg("batch fetch failed")
		}

		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (f *YahooFetcher) fetchBatch(ctx context.Context, assets []*model.Asset) error {
	symbols := ""
	for i, a := range assets {
		if i > 0 {
			symbols += ","
		}
		symbols += a.Ticker
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", symbols)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("yahoo request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var result yahooResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}

	if result.QuoteResponse.Error != nil {
		return fmt.Errorf("yahoo error: %v", result.QuoteResponse.Error)
	}

	for _, q := range result.QuoteResponse.Result {
		price := &model.Price{
			Date:   time.Now().Truncate(24 * time.Hour),
			Open:   decimal.NewFromFloat(q.RegularMarketOpen),
			High:   decimal.NewFromFloat(q.RegularMarketDayHigh),
			Low:    decimal.NewFromFloat(q.RegularMarketDayLow),
			Close:  decimal.NewFromFloat(q.RegularMarketPrice),
			Volume: q.RegularMarketVolume,
			Source: "yahoo",
		}

		// Find matching asset
		for _, a := range assets {
			if a.Ticker == q.Symbol {
				price.AssetID = a.ID
				break
			}
		}

		if price.AssetID == [16]byte{} {
			log.Warn().Str("symbol", q.Symbol).Msg("no matching asset found, skipping")
			continue
		}

		if _, err := f.repos.Price.Create(ctx, price); err != nil {
			log.Warn().Err(err).Str("symbol", q.Symbol).Msg("failed to save price")
		}
	}

	return nil
}
