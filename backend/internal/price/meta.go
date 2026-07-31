package price

import (
	"context"
	"strings"
)

type AssetMeta struct {
	Ticker   string `json:"ticker"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
	Exchange string `json:"exchange"`
	Country  string `json:"country,omitempty"`
}

// FetchMeta resolves metadata (currency, name, type, country) for a ticker via
// the Yahoo chart endpoint. Country is only derived for equities, since ETFs
// usually span multiple countries.
func (f *YahooFetcher) FetchMeta(ctx context.Context, ticker string) (*AssetMeta, error) {
	chart, err := f.fetchChart(ctx, ticker)
	if err != nil {
		return nil, err
	}

	meta := chart.Meta
	name := meta.LongName
	if name == "" {
		name = meta.ShortName
	}

	assetType := mapType(meta.InstrumentType)
	assetMeta := &AssetMeta{
		Ticker:   meta.Symbol,
		Name:     name,
		Type:     assetType,
		Currency: meta.Currency,
		Exchange: meta.ExchangeName,
	}

	// Country only makes sense for a single equity; ETFs/other types span
	// multiple countries.
	if assetType == "stock" {
		assetMeta.Country = exchangeCountry(meta.ExchangeName)
	}

	return assetMeta, nil
}

// exchangeCountry maps a Yahoo exchange code to an ISO country code for equities.
func exchangeCountry(exchange string) string {
	return exchangeCountries[strings.ToUpper(exchange)]
}

var exchangeCountries = map[string]string{
	// US
	"NMS": "US", "NGM": "US", "NYS": "US", "ASE": "US", "PCX": "US",
	"NYE": "US", "PNK": "US", "OQB": "US", "OQX": "US", "OTC": "US",
	// Canada
	"TOR": "CA", "VAN": "CA",
	// Germany
	"FRA": "DE", "BER": "DE", "HAM": "DE", "MUN": "DE", "DUS": "DE",
	"STU": "DE", "XETRA": "DE", "HAN": "DE", "DRF": "DE", "GER": "DE",
	// Italy
	"MIL": "IT",
	// UK
	"LSE": "GB",
	// France
	"PAR": "FR",
	// Switzerland
	"EBS": "CH", "EBM": "CH",
	// Netherlands
	"AMS": "NL",
	// Spain
	"MCE": "ES", "BME": "ES",
	// Nordic
	"STO": "SE", "CPH": "DK", "OSL": "NO", "HEL": "FI",
	// Benelux / Austria / Portugal / Ireland
	"BRU": "BE", "VIE": "AT", "ELI": "PT", "ISE": "IE",
	// Japan / Hong Kong / China
	"TYO": "JP", "FUK": "JP", "HKG": "HK", "SHA": "CN", "SHE": "CN",
	// APAC
	"ASX": "AU", "SES": "SG", "KSC": "KR", "TPE": "TW",
	// Americas
	"SAO": "BR", "BUE": "AR", "MEX": "MX",
	// India / South Africa
	"BSE": "IN", "NSE": "IN", "JNB": "ZA",
}
