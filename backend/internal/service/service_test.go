package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/amelamela/vault-lab/internal/cache"
	"github.com/amelamela/vault-lab/internal/geo"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/repository"
)

type fakePortfolioRepo struct {
	portfolio  *model.Portfolio
	portfolios []*model.Portfolio
	holdings   []*model.Holding
}

func (f *fakePortfolioRepo) Create(ctx context.Context, p *model.Portfolio) (*model.Portfolio, error) {
	return p, nil
}
func (f *fakePortfolioRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Portfolio, error) {
	return f.portfolio, nil
}
func (f *fakePortfolioRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.Portfolio, error) {
	return f.portfolios, nil
}
func (f *fakePortfolioRepo) Update(ctx context.Context, p *model.Portfolio) error { return nil }
func (f *fakePortfolioRepo) Delete(ctx context.Context, id uuid.UUID) error       { return nil }
func (f *fakePortfolioRepo) HeldAssets(ctx context.Context, portfolioID uuid.UUID) ([]*model.Asset, error) {
	return nil, nil
}
func (f *fakePortfolioRepo) HoldingsDetailed(ctx context.Context, portfolioIDs []uuid.UUID) ([]*model.Holding, error) {
	return f.holdings, nil
}
func (f *fakePortfolioRepo) FindAll(ctx context.Context) ([]uuid.UUID, error) { return nil, nil }

type fakeExposureRepo struct {
	regions   map[string][]model.ExposureRow
	sectors   map[string][]model.ExposureRow
	countries map[string][]model.ExposureRow
}

func (f *fakeExposureRepo) FindRegions(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error) {
	return f.regions[assetID.String()], nil
}
func (f *fakeExposureRepo) FindSectors(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error) {
	return f.sectors[assetID.String()], nil
}
func (f *fakeExposureRepo) FindCountries(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error) {
	return f.countries[assetID.String()], nil
}
func (f *fakeExposureRepo) ReplaceRegions(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	return nil
}
func (f *fakeExposureRepo) ReplaceSectors(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	return nil
}
func (f *fakeExposureRepo) ReplaceCountries(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	return nil
}
func (f *fakeExposureRepo) FindRegionsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error) {
	return f.regions, nil
}
func (f *fakeExposureRepo) FindSectorsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error) {
	return f.sectors, nil
}
func (f *fakeExposureRepo) FindCountriesByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error) {
	return f.countries, nil
}

type fakeFXRepo struct {
	rates   map[string]decimal.Decimal
	history map[string][]model.FXRatePoint
}

func (f *fakeFXRepo) Upsert(ctx context.Context, base, quote string, rate decimal.Decimal) error {
	return nil
}
func (f *fakeFXRepo) LatestByQuotes(ctx context.Context, quotes []string) (map[string]decimal.Decimal, error) {
	return f.rates, nil
}
func (f *fakeFXRepo) FetchedAt(ctx context.Context, quote string) (*time.Time, error) {
	return nil, nil
}
func (f *fakeFXRepo) UpsertHistory(ctx context.Context, base, quote string, date time.Time, rate decimal.Decimal, source string) error {
	return nil
}
func (f *fakeFXRepo) MinMaxDate(ctx context.Context, base, quote string) (*time.Time, *time.Time, error) {
	return nil, nil, nil
}
func (f *fakeFXRepo) RateForDate(ctx context.Context, base, quote string, date time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (f *fakeFXRepo) History(ctx context.Context, base, quote string) ([]model.FXRatePoint, error) {
	return f.history[quote], nil
}

type fakeAssetRepo struct {
	asset *model.Asset
	err   error
}

func (f *fakeAssetRepo) Create(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	return asset, nil
}
func (f *fakeAssetRepo) Update(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	return asset, nil
}
func (f *fakeAssetRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.asset == nil || f.asset.ID != id {
		return nil, pgx.ErrNoRows
	}
	return f.asset, nil
}
func (f *fakeAssetRepo) FindByTicker(ctx context.Context, ticker string) (*model.Asset, error) {
	return nil, nil
}
func (f *fakeAssetRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Asset, error) {
	return nil, nil
}
func (f *fakeAssetRepo) Search(ctx context.Context, query string) ([]*model.Asset, error) {
	return nil, nil
}
func (f *fakeAssetRepo) List(ctx context.Context) ([]*model.Asset, error) {
	return nil, nil
}
func (f *fakeAssetRepo) ListYahoo(ctx context.Context) ([]*model.Asset, error) {
	return nil, nil
}
func (f *fakeAssetRepo) AllStocks(ctx context.Context) ([]*model.Asset, error) {
	return nil, nil
}
func (f *fakeAssetRepo) MarkPricesFetched(ctx context.Context, ids []uuid.UUID, at time.Time) error {
	return nil
}
func (f *fakeAssetRepo) MarkHistoryBackfilled(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (f *fakeAssetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (f *fakeAssetRepo) Currencies(ctx context.Context) ([]string, error) {
	return nil, nil
}

func newTestService(t *testing.T, p *fakePortfolioRepo, e *fakeExposureRepo, f *fakeFXRepo) *Service {
	t.Helper()
	repos := &repository.Repository{
		Asset:     &fakeAssetRepo{},
		Portfolio: p,
		Exposure:  e,
		FX:        f,
	}
	return New(repos, nil, nil, nil, time.Minute, cache.New(nil), 0, 0, nil)
}

func newTestServiceWithAsset(t *testing.T, a *fakeAssetRepo) *Service {
	t.Helper()
	repos := &repository.Repository{
		Asset:     a,
		Portfolio: &fakePortfolioRepo{},
		Exposure:  &fakeExposureRepo{},
		FX:        &fakeFXRepo{},
	}
	return New(repos, nil, nil, nil, time.Minute, cache.New(nil), 0, 0, nil)
}

func holding(id, currency, country, sector string, typ model.AssetType, qty, lastClose decimal.Decimal) *model.Holding {
	return &model.Holding{
		AssetID:    id,
		Currency:   currency,
		Country:    country,
		Sector:     sector,
		Type:       typ,
		Qty:        qty,
		LastClose:  lastClose,
		HasPrice:   true,
		AssetClass: "equity",
	}
}

func equalDecimal(a, b decimal.Decimal) bool { return a.Equal(b) }

func TestGetPortfolioGeographyAllocation_ETF(t *testing.T) {
	assetID := uuid.New()
	pf := &fakePortfolioRepo{
		portfolio: &model.Portfolio{Currency: "EUR"},
		holdings: []*model.Holding{
			holding(assetID.String(), "EUR", "", "", model.AssetTypeETF, decimal.NewFromInt(10), decimal.NewFromInt(100)),
		},
	}
	ex := &fakeExposureRepo{
		regions: map[string][]model.ExposureRow{
			assetID.String(): {
				{Name: "North America", Weight: decimal.NewFromInt(60)},
				{Name: "Europe Developed", Weight: decimal.NewFromInt(40)},
			},
		},
	}
	svc := newTestService(t, pf, ex, &fakeFXRepo{})

	got, err := svc.GetPortfolioGeographyAllocation(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Currency != "EUR" {
		t.Fatalf("currency = %q, want EUR", got.Currency)
	}
	if len(got.Regions) != len(geo.Regions) {
		t.Fatalf("regions len = %d, want %d", len(got.Regions), len(geo.Regions))
	}
	for i, r := range got.Regions {
		if r.Region != geo.Regions[i] {
			t.Fatalf("region[%d] = %q, want %q (canonical order)", i, r.Region, geo.Regions[i])
		}
	}
	wantVal := decimal.NewFromInt(1000) // 10 * 100
	if !equalDecimal(got.Regions[0].Value, wantVal.Mul(decimal.NewFromInt(60)).Div(decimal.NewFromInt(100))) {
		t.Fatalf("North America value = %v, want 600", got.Regions[0].Value)
	}
	if !equalDecimal(got.Regions[3].Value, wantVal.Mul(decimal.NewFromInt(40)).Div(decimal.NewFromInt(100))) {
		t.Fatalf("Europe Developed value = %v, want 400", got.Regions[3].Value)
	}
	sum := decimal.Zero
	for _, r := range got.Regions {
		sum = sum.Add(r.Weight)
	}
	if !equalDecimal(sum, decimal.NewFromInt(100)) {
		t.Fatalf("weights sum = %v, want 100", sum)
	}
	for _, r := range got.Regions {
		if r.Region == "Other" {
			t.Fatalf("unexpected Other bucket for fully-mapped ETF")
		}
	}
}

func TestGetPortfolioGeographyAllocation_StockFallback(t *testing.T) {
	assetID := uuid.New()
	pf := &fakePortfolioRepo{
		portfolio: &model.Portfolio{Currency: "USD"},
		holdings: []*model.Holding{
			holding(assetID.String(), "USD", "US", "Technology", model.AssetTypeStock, decimal.NewFromInt(5), decimal.NewFromInt(200)),
		},
	}
	svc := newTestService(t, pf, &fakeExposureRepo{}, &fakeFXRepo{})

	geoAlloc, err := svc.GetPortfolioGeographyAllocation(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	na := geoAlloc.Regions[0]
	if na.Region != "North America" {
		t.Fatalf("region[0] = %q, want North America", na.Region)
	}
	if !equalDecimal(na.Value, decimal.NewFromInt(1000)) {
		t.Fatalf("North America value = %v, want 1000", na.Value)
	}
	if !equalDecimal(na.Weight, decimal.NewFromInt(100)) {
		t.Fatalf("North America weight = %v, want 100", na.Weight)
	}

	secAlloc, err := svc.GetPortfolioSectorAllocation(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secAlloc.Sectors) != len(geo.GICSSectors) {
		t.Fatalf("sectors len = %d, want %d", len(secAlloc.Sectors), len(geo.GICSSectors))
	}
	for i, s := range secAlloc.Sectors {
		if s.Sector != geo.GICSSectors[i] {
			t.Fatalf("sector[%d] = %q, want %q (canonical order)", i, s.Sector, geo.GICSSectors[i])
		}
	}
	if secAlloc.Sectors[7].Sector != "Information Technology" {
		t.Fatalf("sector[7] = %q, want Information Technology", secAlloc.Sectors[7].Sector)
	}
	if !equalDecimal(secAlloc.Sectors[7].Value, decimal.NewFromInt(1000)) {
		t.Fatalf("IT value = %v, want 1000", secAlloc.Sectors[7].Value)
	}
	if !equalDecimal(secAlloc.Sectors[7].Weight, decimal.NewFromInt(100)) {
		t.Fatalf("IT weight = %v, want 100", secAlloc.Sectors[7].Weight)
	}
}

func TestGetPortfolioGeographyAllocation_FXConversion(t *testing.T) {
	assetID := uuid.New()
	pf := &fakePortfolioRepo{
		portfolio: &model.Portfolio{Currency: "EUR"},
		holdings: []*model.Holding{
			holding(assetID.String(), "USD", "", "", model.AssetTypeETF, decimal.NewFromInt(10), decimal.NewFromInt(100)),
		},
	}
	ex := &fakeExposureRepo{
		regions: map[string][]model.ExposureRow{
			assetID.String(): {{Name: "North America", Weight: decimal.NewFromInt(100)}},
		},
	}
	fx := &fakeFXRepo{rates: map[string]decimal.Decimal{"USD": decimal.NewFromInt(1), "EUR": decimal.RequireFromString("1.1")}}
	svc := newTestService(t, pf, ex, fx)

	got, err := svc.GetPortfolioGeographyAllocation(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := decimal.NewFromInt(1000).Mul(decimal.RequireFromString("1.1")) // 1100
	if !equalDecimal(got.Regions[0].Value, want) {
		t.Fatalf("North America value = %v, want %v", got.Regions[0].Value, want)
	}
	if !equalDecimal(got.Regions[0].Weight, decimal.NewFromInt(100)) {
		t.Fatalf("North America weight = %v, want 100", got.Regions[0].Weight)
	}
}

func TestGetPortfolioGeographyAllocation_EmptyPortfolio(t *testing.T) {
	pf := &fakePortfolioRepo{
		portfolio: &model.Portfolio{Currency: "USD"},
		holdings:  []*model.Holding{},
	}
	svc := newTestService(t, pf, &fakeExposureRepo{}, &fakeFXRepo{})

	geoAlloc, err := svc.GetPortfolioGeographyAllocation(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(geoAlloc.Regions) != len(geo.Regions) {
		t.Fatalf("regions len = %d, want %d", len(geoAlloc.Regions), len(geo.Regions))
	}
	for i, r := range geoAlloc.Regions {
		if r.Region != geo.Regions[i] {
			t.Fatalf("region[%d] = %q, want %q", i, r.Region, geo.Regions[i])
		}
		if !r.Value.IsZero() || !r.Weight.IsZero() {
			t.Fatalf("region %q not zero: value=%v weight=%v", r.Region, r.Value, r.Weight)
		}
		if r.Region == "Other" {
			t.Fatalf("unexpected Other bucket in empty portfolio")
		}
	}

	secAlloc, err := svc.GetPortfolioSectorAllocation(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secAlloc.Sectors) != len(geo.GICSSectors) {
		t.Fatalf("sectors len = %d, want %d", len(secAlloc.Sectors), len(geo.GICSSectors))
	}
	for i, s := range secAlloc.Sectors {
		if s.Sector != geo.GICSSectors[i] {
			t.Fatalf("sector[%d] = %q, want %q", i, s.Sector, geo.GICSSectors[i])
		}
		if !s.Value.IsZero() || !s.Weight.IsZero() {
			t.Fatalf("sector %q not zero: value=%v weight=%v", s.Sector, s.Value, s.Weight)
		}
	}
}

func TestGetPortfolioGeographyAllocation_TotalZeroNoDenominator(t *testing.T) {
	assetID := uuid.New()
	pf := &fakePortfolioRepo{
		portfolio: &model.Portfolio{Currency: "EUR"},
		holdings: []*model.Holding{
			holding(assetID.String(), "JPY", "", "", model.AssetTypeETF, decimal.NewFromInt(10), decimal.NewFromInt(100)),
		},
	}
	// No JPY rate (only USD and EUR present): the holding's factor is missing,
	// so total stays zero and no value is assigned to any bucket.
	svc := newTestService(t, pf, &fakeExposureRepo{}, &fakeFXRepo{rates: map[string]decimal.Decimal{"USD": decimal.NewFromInt(1), "EUR": decimal.NewFromInt(1)}})

	got, err := svc.GetPortfolioGeographyAllocation(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range got.Regions {
		if !r.Value.IsZero() {
			t.Fatalf("region %q value = %v, want zero", r.Region, r.Value)
		}
		if !r.Weight.IsZero() {
			t.Fatalf("region %q weight = %v, want zero", r.Region, r.Weight)
		}
	}
}

func TestGetPortfolioGeographyAllocation_OtherBucket(t *testing.T) {
	assetID := uuid.New()
	pf := &fakePortfolioRepo{
		portfolio: &model.Portfolio{Currency: "USD"},
		holdings: []*model.Holding{
			holding(assetID.String(), "USD", "", "", model.AssetTypeETF, decimal.NewFromInt(10), decimal.NewFromInt(100)),
		},
	}
	// Eligible equity ETF (asset_class "equity") with no stored rows and an
	// empty country/sector: for geography the region default is
	// "Other / Not Classified" but (not a stock) the fallback does not map it,
	// and for the sector NormalizeSector("") returns "" so no default and the
	// weight sum is 0 -> Other. The holding is eligible, so it is covered
	// rather than excluded.
	svc := newTestService(t, pf, &fakeExposureRepo{}, &fakeFXRepo{})

	geoAlloc, err := svc.GetPortfolioGeographyAllocation(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(geoAlloc.Regions) != len(geo.Regions)+1 {
		t.Fatalf("regions len = %d, want %d", len(geoAlloc.Regions), len(geo.Regions)+1)
	}
	other := geoAlloc.Regions[len(geoAlloc.Regions)-1]
	if other.Region != "Other" {
		t.Fatalf("last region = %q, want Other", other.Region)
	}
	if !equalDecimal(other.Value, decimal.NewFromInt(1000)) {
		t.Fatalf("Other value = %v, want 1000", other.Value)
	}
	if !equalDecimal(other.Weight, decimal.NewFromInt(100)) {
		t.Fatalf("Other weight = %v, want 100", other.Weight)
	}
	for i := 0; i < len(geoAlloc.Regions)-1; i++ {
		if !geoAlloc.Regions[i].Value.IsZero() {
			t.Fatalf("region %q value = %v, want zero", geoAlloc.Regions[i].Region, geoAlloc.Regions[i].Value)
		}
	}
	sum := decimal.Zero
	for _, r := range geoAlloc.Regions {
		sum = sum.Add(r.Weight)
	}
	if !equalDecimal(sum, decimal.NewFromInt(100)) {
		t.Fatalf("weights sum = %v, want 100", sum)
	}
	if !equalDecimal(geoAlloc.Covered, decimal.NewFromInt(1000)) {
		t.Fatalf("covered = %v, want 1000", geoAlloc.Covered)
	}
	if !geoAlloc.Excluded.IsZero() {
		t.Fatalf("excluded = %v, want zero", geoAlloc.Excluded)
	}

	secAlloc, err := svc.GetPortfolioSectorAllocation(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secAlloc.Sectors) != len(geo.GICSSectors)+1 {
		t.Fatalf("sectors len = %d, want %d", len(secAlloc.Sectors), len(geo.GICSSectors)+1)
	}
	sOther := secAlloc.Sectors[len(secAlloc.Sectors)-1]
	if sOther.Sector != "Other" {
		t.Fatalf("last sector = %q, want Other", sOther.Sector)
	}
	if !equalDecimal(sOther.Value, decimal.NewFromInt(1000)) {
		t.Fatalf("sector Other value = %v, want 1000", sOther.Value)
	}
	if !equalDecimal(secAlloc.Covered, decimal.NewFromInt(1000)) {
		t.Fatalf("covered = %v, want 1000", secAlloc.Covered)
	}
	if !secAlloc.Excluded.IsZero() {
		t.Fatalf("excluded = %v, want zero", secAlloc.Excluded)
	}
}

func regionByName(t *testing.T, regions []*model.RegionAllocation, name string) *model.RegionAllocation {
	t.Helper()
	for _, r := range regions {
		if r.Region == name {
			return r
		}
	}
	t.Fatalf("region %q not found", name)
	return nil
}

func sectorByName(t *testing.T, sectors []*model.SectorAllocation, name string) *model.SectorAllocation {
	t.Helper()
	for _, s := range sectors {
		if s.Sector == name {
			return s
		}
	}
	t.Fatalf("sector %q not found", name)
	return nil
}

func assertDecimalInDelta(t *testing.T, got, want decimal.Decimal, delta, what string) {
	t.Helper()
	if got.Sub(want).Abs().GreaterThan(decimal.RequireFromString(delta)) {
		t.Fatalf("%s = %v, want %v ± %s", what, got, want, delta)
	}
}

func TestGetDashboardAllocation_AggregatesAcrossPortfolios(t *testing.T) {
	etfID := uuid.New()
	stockID := uuid.New()
	pf := &fakePortfolioRepo{
		portfolios: []*model.Portfolio{
			{Currency: "USD"},
			{Currency: "EUR"},
		},
		holdings: []*model.Holding{
			holding(etfID.String(), "EUR", "", "", model.AssetTypeETF, decimal.NewFromInt(1), decimal.NewFromInt(100)),
			holding(stockID.String(), "USD", "IT", "Financials", model.AssetTypeStock, decimal.NewFromInt(1), decimal.NewFromInt(100)),
		},
	}
	ex := &fakeExposureRepo{
		regions: map[string][]model.ExposureRow{
			etfID.String(): {{Name: "North America", Weight: decimal.NewFromInt(100)}},
		},
	}
	fx := &fakeFXRepo{rates: map[string]decimal.Decimal{"EUR": decimal.RequireFromString("0.9")}}
	svc := newTestService(t, pf, ex, fx)

	got, err := svc.GetDashboardAllocation(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Currency != "USD" {
		t.Fatalf("currency = %q, want USD", got.Currency)
	}

	na := regionByName(t, got.Regions, "North America")
	assertDecimalInDelta(t, na.Value, decimal.RequireFromString("111.11"), "0.01", "North America value")
	assertDecimalInDelta(t, na.Weight, decimal.RequireFromString("52.63"), "0.01", "North America weight")

	eu := regionByName(t, got.Regions, "Europe Developed")
	assertDecimalInDelta(t, eu.Value, decimal.NewFromInt(100), "0.01", "Europe value")
	assertDecimalInDelta(t, eu.Weight, decimal.RequireFromString("47.37"), "0.01", "Europe weight")

	weightSum := decimal.Zero
	for _, r := range got.Regions {
		weightSum = weightSum.Add(r.Weight)
	}
	assertDecimalInDelta(t, weightSum, decimal.NewFromInt(100), "0.01", "regions weight sum")

	fin := sectorByName(t, got.Sectors, "Financials")
	if !fin.Value.IsPositive() {
		t.Fatalf("Financials value = %v, want positive", fin.Value)
	}

	// Both holdings (ETF equity + stock) are eligible: fully covered, nothing excluded.
	assertDecimalInDelta(t, got.Covered, decimal.RequireFromString("211.11"), "0.01", "covered value")
	if !got.Excluded.IsZero() {
		t.Fatalf("excluded = %v, want zero", got.Excluded)
	}
}

func TestGetDashboardAllocation_EmptyUser(t *testing.T) {
	svc := newTestService(t, &fakePortfolioRepo{}, &fakeExposureRepo{}, &fakeFXRepo{})

	got, err := svc.GetDashboardAllocation(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Currency != "USD" {
		t.Fatalf("currency = %q, want USD", got.Currency)
	}
	if got.Regions == nil || len(got.Regions) != 0 {
		t.Fatalf("regions = %v, want empty non-nil slice", got.Regions)
	}
	if got.Sectors == nil || len(got.Sectors) != 0 {
		t.Fatalf("sectors = %v, want empty non-nil slice", got.Sectors)
	}
	if !got.Covered.IsZero() {
		t.Fatalf("covered = %v, want zero", got.Covered)
	}
	if !got.Excluded.IsZero() {
		t.Fatalf("excluded = %v, want zero", got.Excluded)
	}
}

func TestGetDashboardAllocation_ExcludesNonEquity(t *testing.T) {
	stockID := uuid.New()
	bondID := uuid.New()
	unclID := uuid.New()
	pf := &fakePortfolioRepo{
		portfolio: &model.Portfolio{Currency: "USD"},
		portfolios: []*model.Portfolio{
			{Currency: "USD"},
		},
		holdings: []*model.Holding{
			holding(stockID.String(), "USD", "IT", "Financials", model.AssetTypeStock, decimal.NewFromInt(1), decimal.NewFromInt(100)),
			holding(bondID.String(), "USD", "", "", model.AssetTypeBond, decimal.NewFromInt(2), decimal.NewFromInt(100)),
			{
				AssetID:    unclID.String(),
				Currency:   "USD",
				Country:    "US",
				Sector:     "Technology",
				Type:       model.AssetTypeETF,
				AssetClass: "other",
				Qty:        decimal.NewFromInt(3),
				LastClose:  decimal.NewFromInt(100),
				HasPrice:   true,
			},
		},
	}
	svc := newTestService(t, pf, &fakeExposureRepo{}, &fakeFXRepo{})

	got, err := svc.GetDashboardAllocation(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the eligible stock (100) participates: everything else (bond 200 +
	// unclassified ETF 300) is excluded and must not flow into buckets/weights.
	if !equalDecimal(got.Covered, decimal.NewFromInt(100)) {
		t.Fatalf("covered = %v, want 100", got.Covered)
	}
	if !equalDecimal(got.Excluded, decimal.NewFromInt(500)) {
		t.Fatalf("excluded = %v, want 500", got.Excluded)
	}

	eu := regionByName(t, got.Regions, "Europe Developed")
	if !equalDecimal(eu.Value, decimal.NewFromInt(100)) {
		t.Fatalf("Europe Developed value = %v, want 100", eu.Value)
	}
	if !equalDecimal(eu.Weight, decimal.NewFromInt(100)) {
		t.Fatalf("Europe Developed weight = %v, want 100", eu.Weight)
	}

	fin := sectorByName(t, got.Sectors, "Financials")
	if !equalDecimal(fin.Value, decimal.NewFromInt(100)) {
		t.Fatalf("Financials value = %v, want 100", fin.Value)
	}
	if !equalDecimal(fin.Weight, decimal.NewFromInt(100)) {
		t.Fatalf("Financials weight = %v, want 100", fin.Weight)
	}

	// The excluded holdings must leave no trace: no literal "Other" bucket and
	// weights still sum to 100 over the eligible universe.
	for _, r := range got.Regions {
		if r.Region == "Other" && !r.Value.IsZero() {
			t.Fatalf("excluded value leaked into Other region: %v", r.Value)
		}
	}
	regionWeightSum := decimal.Zero
	for _, r := range got.Regions {
		regionWeightSum = regionWeightSum.Add(r.Weight)
	}
	if !equalDecimal(regionWeightSum, decimal.NewFromInt(100)) {
		t.Fatalf("regions weight sum = %v, want 100", regionWeightSum)
	}
	for _, s := range got.Sectors {
		if s.Sector == "Other" && !s.Value.IsZero() {
			t.Fatalf("excluded value leaked into Other sector: %v", s.Value)
		}
	}
	sectorWeightSum := decimal.Zero
	for _, s := range got.Sectors {
		sectorWeightSum = sectorWeightSum.Add(s.Weight)
	}
	if !equalDecimal(sectorWeightSum, decimal.NewFromInt(100)) {
		t.Fatalf("sectors weight sum = %v, want 100", sectorWeightSum)
	}
}

func TestNormalizeCountries(t *testing.T) {
	cases := []struct {
		name string
		in   []model.ExposureRow
		want []model.ExposureRow
	}{
		{
			name: "ISO codes and full names normalize to canonical codes",
			in: []model.ExposureRow{
				{Name: "United States", Weight: decimal.NewFromInt(60)},
				{Name: "DE", Weight: decimal.NewFromInt(40)},
			},
			want: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(60)},
				{Name: "DE", Weight: decimal.NewFromInt(40)},
			},
		},
		{
			name: "non-canonical country names are dropped",
			in: []model.ExposureRow{
				{Name: "Atlantis", Weight: decimal.NewFromInt(50)},
				{Name: "IT", Weight: decimal.NewFromInt(50)},
			},
			want: []model.ExposureRow{
				{Name: "IT", Weight: decimal.NewFromInt(50)},
			},
		},
		{
			name: "empty and non-positive rows are dropped",
			in: []model.ExposureRow{
				{Name: "", Weight: decimal.NewFromInt(50)},
				{Name: "JP", Weight: decimal.NewFromInt(0)},
				{Name: "US", Weight: decimal.NewFromInt(100)},
			},
			want: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(100)},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeCountries(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d: %+v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i].Name != tc.want[i].Name || !got[i].Weight.Equal(tc.want[i].Weight) {
					t.Fatalf("row[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestIsCanonicalCountry(t *testing.T) {
	if !isCanonicalCountry("US") {
		t.Fatal("US should be canonical")
	}
	if !isCanonicalCountry("ZA") {
		t.Fatal("ZA should be canonical")
	}
	if isCanonicalCountry("") {
		t.Fatal("empty should not be canonical")
	}
	if isCanonicalCountry("XYZ") {
		t.Fatal("XYZ should not be canonical")
	}
}

func TestDeriveRegions(t *testing.T) {
	assetID := uuid.New()
	svc := newTestServiceWithAsset(t, &fakeAssetRepo{asset: &model.Asset{ID: assetID, Type: model.AssetTypeETF}})

	regions, err := svc.DeriveRegions(context.Background(), assetID, []model.ExposureRow{
		{Name: "US", Weight: decimal.NewFromInt(60)},
		{Name: "GB", Weight: decimal.NewFromInt(25)},
		{Name: "JP", Weight: decimal.NewFromInt(15)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(regions) != len(geo.Regions) {
		t.Fatalf("regions len = %d, want %d", len(regions), len(geo.Regions))
	}
	for i, name := range geo.Regions {
		if regions[i].Name != name {
			t.Fatalf("regions[%d] = %q, want %q (canonical order)", i, regions[i].Name, name)
		}
	}
	want := map[string]string{
		"North America":  "60",
		"United Kingdom": "25",
		"Japan":          "15",
	}
	for _, r := range regions {
		if w, ok := want[r.Name]; ok && !r.Weight.Equal(decimal.RequireFromString(w)) {
			t.Fatalf("region %q weight = %v, want %s", r.Name, r.Weight, w)
		}
	}
}

func TestDeriveRegions_AssetNotFound(t *testing.T) {
	svc := newTestServiceWithAsset(t, &fakeAssetRepo{})
	_, err := svc.DeriveRegions(context.Background(), uuid.New(), nil)
	if !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("err = %v, want ErrAssetNotFound", err)
	}
}

func TestDeriveRegions_EmptyCountries(t *testing.T) {
	assetID := uuid.New()
	svc := newTestServiceWithAsset(t, &fakeAssetRepo{asset: &model.Asset{ID: assetID, Type: model.AssetTypeETF}})

	regions, err := svc.DeriveRegions(context.Background(), assetID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(regions) != len(geo.Regions) {
		t.Fatalf("regions len = %d, want %d", len(regions), len(geo.Regions))
	}
	for _, r := range regions {
		if !r.Weight.IsZero() {
			t.Fatalf("region %q weight = %v, want zero", r.Name, r.Weight)
		}
	}
}

func TestMapMorningstarExposure_UsesOfficialRegions(t *testing.T) {
	raw := &model.AssetExposure{
		Countries: []model.ExposureRow{
			{Name: "United States", Weight: decimal.NewFromInt(60)},
		},
		Regions: []model.ExposureRow{
			{Name: "North America", Weight: decimal.NewFromInt(60)},
			{Name: "United Kingdom", Weight: decimal.NewFromInt(25)},
			{Name: "Japan", Weight: decimal.NewFromInt(15)},
		},
		Sectors: []model.ExposureRow{
			{Name: "Technology", Weight: decimal.NewFromFloat(27.5)},
		},
	}
	mapped := mapMorningstarExposure(raw)

	byName := map[string]decimal.Decimal{}
	if len(mapped.Regions) != len(geo.Regions) {
		t.Fatalf("regions len = %d, want %d", len(mapped.Regions), len(geo.Regions))
	}
	for i, name := range geo.Regions {
		if mapped.Regions[i].Name != name {
			t.Fatalf("regions[%d] = %q, want %q (canonical order)", i, mapped.Regions[i].Name, name)
		}
		byName[name] = mapped.Regions[i].Weight
	}
	if !byName["North America"].Equal(decimal.NewFromInt(60)) {
		t.Fatalf("North America = %v, want 60", byName["North America"])
	}
	if !byName["United Kingdom"].Equal(decimal.NewFromInt(25)) {
		t.Fatalf("United Kingdom = %v, want 25", byName["United Kingdom"])
	}
	if !byName["Japan"].Equal(decimal.NewFromInt(15)) {
		t.Fatalf("Japan = %v, want 15", byName["Japan"])
	}
	if !byName["Asia Emerging"].IsZero() {
		t.Fatalf("Asia Emerging = %v, want zero (zero-filled)", byName["Asia Emerging"])
	}
	// Countries are normalized and sectors aggregated as before.
	if len(mapped.Countries) != 1 || mapped.Countries[0].Name != "US" ||
		!mapped.Countries[0].Weight.Equal(decimal.NewFromInt(60)) {
		t.Fatalf("countries = %+v, want [US 60]", mapped.Countries)
	}
	if len(mapped.Sectors) != 1 || mapped.Sectors[0].Name != "Information Technology" ||
		!mapped.Sectors[0].Weight.Equal(decimal.NewFromFloat(27.5)) {
		t.Fatalf("sectors = %+v, want [Information Technology 27.5]", mapped.Sectors)
	}
}

func TestMapMorningstarExposure_FallsBackToDerivation(t *testing.T) {
	raw := &model.AssetExposure{
		Countries: []model.ExposureRow{
			{Name: "US", Weight: decimal.NewFromInt(60)},
			{Name: "JP", Weight: decimal.NewFromInt(40)},
		},
	}
	mapped := mapMorningstarExposure(raw)

	byName := map[string]decimal.Decimal{}
	if len(mapped.Regions) != len(geo.Regions) {
		t.Fatalf("regions len = %d, want %d", len(mapped.Regions), len(geo.Regions))
	}
	for i, name := range geo.Regions {
		if mapped.Regions[i].Name != name {
			t.Fatalf("regions[%d] = %q, want %q (canonical order)", i, mapped.Regions[i].Name, name)
		}
		byName[name] = mapped.Regions[i].Weight
	}
	if !byName["North America"].Equal(decimal.NewFromInt(60)) {
		t.Fatalf("North America = %v, want 60", byName["North America"])
	}
	if !byName["Japan"].Equal(decimal.NewFromInt(40)) {
		t.Fatalf("Japan = %v, want 40", byName["Japan"])
	}
	if !byName["United Kingdom"].IsZero() {
		t.Fatalf("United Kingdom = %v, want zero", byName["United Kingdom"])
	}
}
