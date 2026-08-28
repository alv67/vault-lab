package geo

import "testing"

func TestClassifyAssetClass(t *testing.T) {
	cases := []struct {
		name       string
		assetType  string
		assetName  string
		category   string
		want       string
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