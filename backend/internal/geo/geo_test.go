package geo

import "testing"

func TestClassifyAssetClass(t *testing.T) {
	cases := []struct {
		name      string
		assetType string
		assetName string
		category  string
		want      string
	}{
		{"stock defaults to equity", "stock", "Apple Inc.", "", "equity"},
		{"bond type", "bond", "Some Bond", "", "bond"},
		{"commodity type", "commodity", "Gold", "", "commodity"},
		{"crypto type", "crypto", "Bitcoin", "", "crypto"},
		{"cash type", "cash", "Cash account", "", "currency"},
		{"bond ETF by name", "etf", "Vanguard Total Bond Market ETF", "", "bond"},
		{"bond ETF by category", "etf", "iShares Core US Agg", "Intermediate Core Bond", "bond"},
		{"government bond ETF", "etf", "Vanguard Short Treasury ETF", "Short Government", "bond"},
		{"money market lands on bond", "etf", "Money Market Fund", "", "bond"},
		{"commodity ETF by category", "etf", "SPDR Gold Shares", "Commodities Focused", "commodity"},
		{"commodity ETF by name", "etf", "iShares Silver Trust", "", "commodity"},
		{"real estate ETF", "etf", "Vanguard Real Estate Index Fund", "Real Estate", "real_estate"},
		{"REIT ETF", "etf", "iShares U.S. Real Estate ETF", "", "real_estate"},
		{"crypto ETF by name", "etf", "ProShares Bitcoin Strategy ETF", "", "crypto"},
		{"currency ETF", "etf", "Invesco CurrencyShares Euro Trust", "Currency", "currency"},
		{"balanced fund", "mutual_fund", "Vanguard Balanced Fund", "Moderate Allocation", "mixed"},
		{"all-world equity ETF defaults to equity", "etf", "Vanguard FTSE All-World UCITS ETF USD Acc", "", "equity"},
		{"index ETF defaults to equity", "etf", "SPDR S&P 500 ETF", "Large Blend", "equity"},
		{"stock via etf name with oil in category word is avoided", "etf", "Global X Oil Equities", "Equity", "equity"},
	}

	for _, tc := range cases {
		got := ClassifyAssetClass(tc.assetType, tc.assetName, tc.category)
		if got != tc.want {
			t.Errorf("%s: ClassifyAssetClass(%q, %q, %q) = %q, want %q",
				tc.name, tc.assetType, tc.assetName, tc.category, got, tc.want)
		}
	}
}

func TestFundClassifiable(t *testing.T) {
	cases := map[string]bool{
		"stock":       false,
		"bond":        false,
		"commodity":   false,
		"crypto":      false,
		"cash":        false,
		"etf":         true,
		"mutual_fund": true,
		"other":       true,
		"":            true,
	}
	for typ, want := range cases {
		if got := FundClassifiable(typ); got != want {
			t.Errorf("FundClassifiable(%q) = %v, want %v", typ, got, want)
		}
	}
}

func TestRegionForCountry(t *testing.T) {
	cases := []struct {
		name    string
		country string
		want    string
	}{
		{"US", "US", "North America"},
		{"us lowercase", "us", "North America"},
		{"Canada", "CA", "North America"},
		{"Italy", "IT", "Europe Developed"},
		{"Italy lowercase", "it", "Europe Developed"},
		{"Japan", "JP", "Asia Developed"},
		{"China", "CN", "Asia Emerging"},
		{"empty", "", "Other / Not Classified"},
		{"unknown", "XYZ", "Other / Not Classified"},
		{"spaces", "  US  ", "North America"},
	}
	for _, tc := range cases {
		if got := RegionForCountry(tc.country); got != tc.want {
			t.Errorf("%s: RegionForCountry(%q) = %q, want %q", tc.name, tc.country, got, tc.want)
		}
	}
}

func TestNormalizeCountry(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"full name", "United States", "US"},
		{"full name lowercase", "united kingdom", "GB"},
		{"already a code", "US", "US"},
		{"code lowercase", "it", "IT"},
		{"full name with spaces", "  Switzerland ", "CH"},
		{"full name multi-word", "Hong Kong", "HK"},
		{"unknown", "XYZ", ""},
		{"unknown name", "Atlantis", ""},
		{"empty", "", ""},
		{"one letter is unknown", "U", ""},
		{"usa alias resolved via map", "USA", "US"},
		{"justetf name", "South Korea", "KR"},
		{"justetf name", "Saudi Arabia", "SA"},
		{"justetf name", "United Arab Emirates", "AE"},
		{"justetf name", "Thailand", "TH"},
		{"justetf name", "Malaysia", "MY"},
	}
	for _, tc := range cases {
		if got := NormalizeCountry(tc.in); got != tc.want {
			t.Errorf("%s: NormalizeCountry(%q) = %q, want %q", tc.name, tc.in, got, tc.want)
		}
	}
}

func TestNormalizeSectorJustETF(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Technology", "Information Technology"},
		{"Finance", "Financials"},
		{"Non-Energy Materials", "Materials"},
		{"Consumer Non-Cyclicals", "Consumer Staples"},
		{"Consumer Cyclicals", "Consumer Discretionary"},
		{"Consumer Services", "Consumer Discretionary"},
		{"Telecommunication", "Communication Services"},
		{"Healthcare", "Health Care"},
		{"Industrials", "Industrials"},
		{"Energy", "Energy"},
		{"Utilities", "Utilities"},
	}
	for _, tc := range cases {
		if got := NormalizeSector(tc.in); got != tc.want {
			t.Errorf("NormalizeSector(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
