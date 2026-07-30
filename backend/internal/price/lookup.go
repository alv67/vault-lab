package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AssetLookup struct {
	Ticker   string `json:"ticker"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
	Exchange string `json:"exchange"`
	Country  string `json:"country,omitempty"`
}

type yahooSearchResponse struct {
	Quotes []struct {
		Symbol        string `json:"symbol"`
		LongName      string `json:"longname"`
		ShortName     string `json:"shortname"`
		QuoteType     string `json:"quoteType"`
		Currency      string `json:"currency"`
		Exchange      string `json:"exchange"`
	} `json:"quotes"`
}

func LookupAsset(ctx context.Context, query string) ([]AssetLookup, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v1/finance/search?q=%s&lang=en-US&region=US", query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yahoo search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result yahooSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	var assets []AssetLookup
	for _, q := range result.Quotes {
		if q.Symbol == "" {
			continue
		}
		name := q.LongName
		if name == "" {
			name = q.ShortName
		}

		assetType := mapType(q.QuoteType)

		assets = append(assets, AssetLookup{
			Ticker:   q.Symbol,
			Name:     name,
			Type:     assetType,
			Currency: q.Currency,
			Exchange: q.Exchange,
		})
	}

	if assets == nil {
		assets = []AssetLookup{}
	}

	return assets, nil
}

func mapType(quoteType string) string {
	switch quoteType {
	case "EQUITY":
		return "stock"
	case "ETF":
		return "etf"
	case "MUTUALFUND":
		return "mutual_fund"
	case "BOND":
		return "bond"
	case "CRYPTOCURRENCY":
		return "crypto"
	case "COMMODITY":
		return "commodity"
	default:
		return "stock"
	}
}


