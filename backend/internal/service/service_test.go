package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"

	"github.com/amelamela/vault-lab/internal/cache"
	"github.com/amelamela/vault-lab/internal/geo"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/price"
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
	// provenance mirrors the asset_exposure_provenance table, keyed by asset id
	// then by dimension ("countries"/"regions"/"sectors").
	provenance map[string]map[string]model.ExposureProvenance
	// replace*Calls record every dimension write so tests can assert which
	// operations persist (PUT/save) and which stay read-only previews.
	replaceRegionsCalls   int
	replaceSectorsCalls   int
	replaceCountriesCalls int
	setProvenanceCalls    int
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
	f.replaceRegionsCalls++
	if f.regions == nil {
		f.regions = map[string][]model.ExposureRow{}
	}
	f.regions[assetID.String()] = rows
	return nil
}
func (f *fakeExposureRepo) ReplaceSectors(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	f.replaceSectorsCalls++
	if f.sectors == nil {
		f.sectors = map[string][]model.ExposureRow{}
	}
	f.sectors[assetID.String()] = rows
	return nil
}
func (f *fakeExposureRepo) ReplaceCountries(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	f.replaceCountriesCalls++
	if f.countries == nil {
		f.countries = map[string][]model.ExposureRow{}
	}
	f.countries[assetID.String()] = rows
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
func (f *fakeExposureRepo) FindProvenance(ctx context.Context, assetID uuid.UUID) (map[string]model.ExposureProvenance, error) {
	if f.provenance == nil {
		return nil, nil
	}
	return f.provenance[assetID.String()], nil
}
func (f *fakeExposureRepo) SetProvenance(ctx context.Context, assetID uuid.UUID, dimension, source string) error {
	f.setProvenanceCalls++
	if f.provenance == nil {
		f.provenance = map[string]map[string]model.ExposureProvenance{}
	}
	key := assetID.String()
	if f.provenance[key] == nil {
		f.provenance[key] = map[string]model.ExposureProvenance{}
	}
	f.provenance[key][dimension] = model.ExposureProvenance{Source: source, UpdatedAt: time.Now().UTC()}
	return nil
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
	// updateCalls and lastUpdate record asset writes so tests can assert the
	// exposure fetches persist nothing but the auto-resolved ISIN.
	updateCalls int
	lastUpdate  *model.Asset
}

func (f *fakeAssetRepo) Create(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	return asset, nil
}
func (f *fakeAssetRepo) Update(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	f.updateCalls++
	f.lastUpdate = asset
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

// fakeLookupRepo is an in-memory stand-in for the Redis-backed
// repository.LookupRepository. It counts reads and writes so the provider
// exposure cache tests can assert hit/miss/refresh behavior; misses return
// redis.Nil like the real repository.
type fakeLookupRepo struct {
	data     map[string][]byte
	getCalls int
	setCalls int
}

func (f *fakeLookupRepo) Get(ctx context.Context, key string) ([]byte, error) {
	f.getCalls++
	v, ok := f.data[key]
	if !ok {
		return nil, redis.Nil
	}
	return v, nil
}

func (f *fakeLookupRepo) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	f.setCalls++
	if f.data == nil {
		f.data = map[string][]byte{}
	}
	f.data[key] = data
	return nil
}

func newTestService(t *testing.T, p *fakePortfolioRepo, e *fakeExposureRepo, f *fakeFXRepo) *Service {
	t.Helper()
	repos := &repository.Repository{
		Asset:     &fakeAssetRepo{},
		Portfolio: p,
		Exposure:  e,
		FX:        f,
	}
	return New(repos, nil, nil, nil, time.Minute, time.Hour, cache.New(nil), 0, 0, nil)
}

func newTestServiceWithAsset(t *testing.T, a *fakeAssetRepo) *Service {
	t.Helper()
	repos := &repository.Repository{
		Asset:     a,
		Portfolio: &fakePortfolioRepo{},
		Exposure:  &fakeExposureRepo{},
		FX:        &fakeFXRepo{},
	}
	return New(repos, nil, nil, nil, time.Minute, time.Hour, cache.New(nil), 0, 0, nil)
}

// fakeYahooFetcher stubs the yahooFetcher seam; only the profile calls carry
// canned data, the rest are inert zero-value stubs.
type fakeYahooFetcher struct {
	sector   string
	industry string
	country  string
	// weightings is what the provider reports as sectorWeightings; gotISIN
	// style counters record the calls so tests can assert the fetch ran.
	weightings           []model.ExposureRow
	profileExtendedCalls int
}

func (f *fakeYahooFetcher) FetchAssetProfile(ctx context.Context, ticker string) (string, string, string, error) {
	return f.sector, f.industry, f.country, nil
}
func (f *fakeYahooFetcher) FetchAssetProfileExtended(ctx context.Context, ticker string) (string, string, string, []model.ExposureRow, error) {
	f.profileExtendedCalls++
	return f.sector, f.industry, f.country, f.weightings, nil
}
func (f *fakeYahooFetcher) FetchMeta(ctx context.Context, ticker string) (*price.AssetMeta, error) {
	return nil, nil
}
func (f *fakeYahooFetcher) FetchFXRate(ctx context.Context, quote string) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (f *fakeYahooFetcher) RefreshFX(ctx context.Context) ([]price.FetchIssue, error) {
	return nil, nil
}
func (f *fakeYahooFetcher) RefreshStale(ctx context.Context, assets []*model.Asset) (price.RefreshReport, error) {
	return price.RefreshReport{}, nil
}
func (f *fakeYahooFetcher) RefreshStaleForPortfolio(ctx context.Context, portfolioID uuid.UUID) (price.RefreshReport, error) {
	return price.RefreshReport{}, nil
}
func (f *fakeYahooFetcher) EnsureHistory(ctx context.Context, assets []price.HistoryAsset) error {
	return nil
}
func (f *fakeYahooFetcher) EnsureSplits(ctx context.Context, assets []*model.Asset) error {
	return nil
}

// fakeETFFetcher stubs price.ETFFetcher and records the ISIN each fetch ran
// with so tests can assert the ticker→ISIN auto-resolution result is used. The
// *Calls counters record how often the provider was actually reached, which
// lets the cache tests verify that a hit skips the fetch entirely.
type fakeETFFetcher struct {
	exposure        *model.AssetExposure
	morningstar     *model.AssetExposure
	search          []price.EtfSearchResult
	exposureISIN    string
	morningstarISIN string
	searchTicker    string
	exposureCalls   int
	morningstarCall int
}

func (f *fakeETFFetcher) FetchExposure(ctx context.Context, isin string) (*model.AssetExposure, error) {
	f.exposureISIN = isin
	f.exposureCalls++
	return f.exposure, nil
}
func (f *fakeETFFetcher) FetchMorningstarExposure(ctx context.Context, isin string) (*model.AssetExposure, error) {
	f.morningstarISIN = isin
	f.morningstarCall++
	return f.morningstar, nil
}
func (f *fakeETFFetcher) SearchTicker(ctx context.Context, query string) ([]price.EtfSearchResult, error) {
	f.searchTicker = query
	return f.search, nil
}

func newFetchTestService(t *testing.T, a *fakeAssetRepo, e *fakeExposureRepo, lk *fakeLookupRepo, yf yahooFetcher, etf price.ETFFetcher) *Service {
	t.Helper()
	if lk == nil {
		lk = &fakeLookupRepo{}
	}
	repos := &repository.Repository{
		Asset:     a,
		Portfolio: &fakePortfolioRepo{},
		Exposure:  e,
		FX:        &fakeFXRepo{},
		Lookup:    lk,
	}
	return New(repos, nil, yf, etf, time.Minute, time.Hour, cache.New(nil), 0, 0, nil)
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

func exposureRowsByName(rows []model.ExposureRow) map[string]decimal.Decimal {
	byName := map[string]decimal.Decimal{}
	for _, r := range rows {
		byName[r.Name] = r.Weight
	}
	return byName
}

func TestSaveExposureDimensions_CountriesDoNotTouchRegions(t *testing.T) {
	assetID := uuid.New()
	key := assetID.String()
	ex := &fakeExposureRepo{
		regions: map[string][]model.ExposureRow{
			key: {{Name: "North America", Weight: decimal.NewFromInt(100)}},
		},
	}
	repos := &repository.Repository{Exposure: ex}

	err := saveExposureDimensions(context.Background(), repos, assetID, &model.AssetExposure{
		Countries: []model.ExposureRow{
			{Name: "United States", Weight: decimal.NewFromInt(60)},
			{Name: "JP", Weight: decimal.NewFromInt(35)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ex.replaceRegionsCalls != 0 {
		t.Fatalf("ReplaceRegions calls = %d, want 0 (saving countries must not rewrite the regions dimension)", ex.replaceRegionsCalls)
	}
	gotRegions := ex.regions[key]
	if len(gotRegions) != 1 || gotRegions[0].Name != "North America" || !gotRegions[0].Weight.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("stored regions = %+v, want the pre-existing [North America 100]", gotRegions)
	}
	gotCountries := exposureRowsByName(ex.countries[key])
	if len(gotCountries) != 2 || !gotCountries["US"].Equal(decimal.NewFromInt(60)) || !gotCountries["JP"].Equal(decimal.NewFromInt(35)) {
		t.Fatalf("stored countries = %+v, want normalized US 60 + JP 35", ex.countries[key])
	}
	if _, ok := ex.sectors[key]; ok {
		t.Fatalf("sectors written for an absent dimension: %+v", ex.sectors[key])
	}
}

func TestSaveExposureDimensions_ExplicitRegionsStillSaved(t *testing.T) {
	assetID := uuid.New()
	key := assetID.String()
	ex := &fakeExposureRepo{
		regions: map[string][]model.ExposureRow{
			key: {{Name: "North America", Weight: decimal.NewFromInt(100)}},
		},
	}
	repos := &repository.Repository{Exposure: ex}

	err := saveExposureDimensions(context.Background(), repos, assetID, &model.AssetExposure{
		Regions: []model.ExposureRow{
			{Name: "North America", Weight: decimal.NewFromInt(60)},
			{Name: "United Kingdom", Weight: decimal.NewFromInt(25)},
			{Name: "Japan", Weight: decimal.NewFromInt(10)},
		},
		Countries: []model.ExposureRow{
			{Name: "US", Weight: decimal.NewFromInt(60)},
			{Name: "GB", Weight: decimal.NewFromInt(25)},
			{Name: "JP", Weight: decimal.NewFromInt(10)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ex.replaceRegionsCalls != 1 {
		t.Fatalf("ReplaceRegions calls = %d, want 1 (explicit regions must persist)", ex.replaceRegionsCalls)
	}
	gotRegions := exposureRowsByName(ex.regions[key])
	if !gotRegions["North America"].Equal(decimal.NewFromInt(60)) || !gotRegions["United Kingdom"].Equal(decimal.NewFromInt(25)) ||
		!gotRegions["Japan"].Equal(decimal.NewFromInt(10)) {
		t.Fatalf("stored regions = %+v, want the explicit 60/25/10", ex.regions[key])
	}
	if !gotRegions[geo.OtherRegion].Equal(decimal.NewFromInt(5)) {
		t.Fatalf("residual %q = %v, want 5 (prepareRegions invariant)", geo.OtherRegion, gotRegions[geo.OtherRegion])
	}
	if gotCountries := exposureRowsByName(ex.countries[key]); len(gotCountries) != 3 || !gotCountries["GB"].Equal(decimal.NewFromInt(25)) {
		t.Fatalf("stored countries = %+v, want US/GB/JP 60/25/10", ex.countries[key])
	}
}

func TestSaveExposureDimensions_EmptyPayloadWritesNothing(t *testing.T) {
	assetID := uuid.New()
	key := assetID.String()
	ex := &fakeExposureRepo{
		regions:   map[string][]model.ExposureRow{key: {{Name: "Japan", Weight: decimal.NewFromInt(100)}}},
		sectors:   map[string][]model.ExposureRow{key: {{Name: "Health Care", Weight: decimal.NewFromInt(100)}}},
		countries: map[string][]model.ExposureRow{key: {{Name: "JP", Weight: decimal.NewFromInt(100)}}},
	}
	repos := &repository.Repository{Exposure: ex}

	if err := saveExposureDimensions(context.Background(), repos, assetID, &model.AssetExposure{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ex.replaceRegionsCalls != 0 {
		t.Fatalf("ReplaceRegions calls = %d, want 0", ex.replaceRegionsCalls)
	}
	if got := exposureRowsByName(ex.regions[key]); !got["Japan"].Equal(decimal.NewFromInt(100)) || len(got) != 1 {
		t.Fatalf("stored regions = %+v, want the pre-existing [Japan 100]", ex.regions[key])
	}
	if got := exposureRowsByName(ex.countries[key]); !got["JP"].Equal(decimal.NewFromInt(100)) || len(got) != 1 {
		t.Fatalf("stored countries = %+v, want the pre-existing [JP 100]", ex.countries[key])
	}
	if ex.setProvenanceCalls != 0 {
		t.Fatalf("SetProvenance calls = %d, want 0 (an empty payload must not write provenance)", ex.setProvenanceCalls)
	}
}

func TestSaveExposureDimensions_RecordsProvenance(t *testing.T) {
	assetID := uuid.New()
	key := assetID.String()
	ex := &fakeExposureRepo{
		provenance: map[string]map[string]model.ExposureProvenance{
			key: {
				model.ExposureDimensionRegions: {Source: "yahoo", UpdatedAt: time.Now().Add(-24 * time.Hour).UTC()},
			},
		},
	}
	repos := &repository.Repository{Exposure: ex}

	err := saveExposureDimensions(context.Background(), repos, assetID, &model.AssetExposure{
		Countries:       []model.ExposureRow{{Name: "US", Weight: decimal.NewFromInt(100)}},
		CountriesSource: "morningstar",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ex.setProvenanceCalls != 1 {
		t.Fatalf("SetProvenance calls = %d, want 1 (only the saved dimension)", ex.setProvenanceCalls)
	}
	got := ex.provenance[key]
	if got[model.ExposureDimensionCountries].Source != "morningstar" {
		t.Fatalf("countries provenance = %+v, want source morningstar", got[model.ExposureDimensionCountries])
	}
	if got[model.ExposureDimensionCountries].UpdatedAt.IsZero() {
		t.Fatalf("countries provenance updated_at is zero, want a timestamp")
	}
	if got[model.ExposureDimensionRegions].Source != "yahoo" {
		t.Fatalf("regions provenance = %+v, want the pre-existing yahoo (saving countries must not touch it)", got[model.ExposureDimensionRegions])
	}
	if _, ok := got[model.ExposureDimensionSectors]; ok {
		t.Fatalf("sectors provenance written for an absent dimension: %+v", got[model.ExposureDimensionSectors])
	}
}

func TestSaveExposureDimensions_DefaultsManualSource(t *testing.T) {
	assetID := uuid.New()
	key := assetID.String()
	ex := &fakeExposureRepo{}
	repos := &repository.Repository{Exposure: ex}

	err := saveExposureDimensions(context.Background(), repos, assetID, &model.AssetExposure{
		Regions: []model.ExposureRow{{Name: "North America", Weight: decimal.NewFromInt(100)}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ex.provenance[key][model.ExposureDimensionRegions].Source; got != "manual" {
		t.Fatalf("regions provenance = %q, want manual (absent source means a manual edit)", got)
	}

	err = saveExposureDimensions(context.Background(), repos, assetID, &model.AssetExposure{
		Sectors:       []model.ExposureRow{{Name: "Technology", Weight: decimal.NewFromInt(100)}},
		SectorsSource: "justetf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ex.provenance[key][model.ExposureDimensionSectors].Source; got != "justetf" {
		t.Fatalf("sectors provenance = %q, want justetf (explicit source must win)", got)
	}
}

func TestGetAssetExposure_ReturnsStoredProvenance(t *testing.T) {
	assetID := uuid.New()
	key := assetID.String()
	updatedAt := time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "SXR8.DE", Type: model.AssetTypeETF, ISIN: "IE00B4L5Y983"}}
	ex := &fakeExposureRepo{
		countries: map[string][]model.ExposureRow{key: {{Name: "US", Weight: decimal.NewFromInt(100)}}},
		provenance: map[string]map[string]model.ExposureProvenance{
			key: {
				model.ExposureDimensionCountries: {Source: "morningstar", UpdatedAt: updatedAt},
			},
		},
	}
	svc := newFetchTestService(t, a, ex, nil, nil, nil)

	got, err := svc.GetAssetExposure(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Provenance) != 1 {
		t.Fatalf("provenance = %+v, want only the stored countries entry", got.Provenance)
	}
	prov, ok := got.Provenance[model.ExposureDimensionCountries]
	if !ok || prov.Source != "morningstar" || !prov.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("countries provenance = %+v, want morningstar @ %s", prov, updatedAt)
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

func TestValidateExposureWeights(t *testing.T) {
	cases := []struct {
		name    string
		sum     string
		rule    weightSumRule
		wantErr bool
	}{
		{"exact: 100 is valid", "100", weightSumExact100, false},
		{"exact: 100.4 is within tolerance", "100.4", weightSumExact100, false},
		{"exact: 99.5 is within tolerance", "99.5", weightSumExact100, false},
		{"exact: 100.6 is rejected", "100.6", weightSumExact100, true},
		{"exact: 99.4 is rejected", "99.4", weightSumExact100, true},
		{"exact: 103 is rejected", "103", weightSumExact100, true},
		{"max: 100 is valid", "100", weightSumMax100, false},
		{"max: 100.5 is the accepted boundary", "100.5", weightSumMax100, false},
		{"max: 100.6 is rejected", "100.6", weightSumMax100, true},
		{"max: 103 is rejected", "103", weightSumMax100, true},
		{"max: a partial 95 is valid", "95", weightSumMax100, false},
		{"max: zero rows are valid", "0", weightSumMax100, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows := []model.ExposureRow{{Name: "X", Weight: decimal.RequireFromString(tc.sum)}}
			err := validateExposureWeights(rows, tc.rule)
			if tc.wantErr {
				if !errors.Is(err, ErrInvalidWeights) {
					t.Fatalf("err = %v, want ErrInvalidWeights", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPrepareRegions(t *testing.T) {
	cases := []struct {
		name    string
		in      []model.ExposureRow
		wantErr bool
		want    map[string]string
		wantSum string
	}{
		{
			name: "sum of 95 injects the residual into Other so the stored total is 100",
			in: []model.ExposureRow{
				{Name: "North America", Weight: decimal.NewFromInt(60)},
				{Name: "United Kingdom", Weight: decimal.NewFromInt(25)},
				{Name: "Japan", Weight: decimal.NewFromInt(10)},
			},
			want: map[string]string{
				"North America":          "60",
				"United Kingdom":         "25",
				"Japan":                  "10",
				"Other / Not Classified": "5",
			},
			wantSum: "100",
		},
		{
			name: "sum of 103 is rejected",
			in: []model.ExposureRow{
				{Name: "North America", Weight: decimal.NewFromInt(103)},
			},
			wantErr: true,
		},
		{
			name: "exact 100 without an Other row needs no injection",
			in: []model.ExposureRow{
				{Name: "North America", Weight: decimal.NewFromInt(95)},
				{Name: "Japan", Weight: decimal.NewFromInt(5)},
			},
			want: map[string]string{
				"North America": "95",
				"Japan":         "5",
			},
			wantSum: "100",
		},
		{
			name: "explicit zero Other row is dropped and not re-added",
			in: []model.ExposureRow{
				{Name: "North America", Weight: decimal.NewFromInt(100)},
				{Name: "Other / Not Classified", Weight: decimal.Zero},
			},
			want: map[string]string{
				"North America": "100",
			},
			wantSum: "100",
		},
		{
			name: "explicit positive Other row is kept as sent",
			in: []model.ExposureRow{
				{Name: "North America", Weight: decimal.NewFromInt(90)},
				{Name: geo.OtherRegion, Weight: decimal.NewFromInt(8)},
			},
			want: map[string]string{
				"North America":          "90",
				"Other / Not Classified": "8",
			},
			wantSum: "98",
		},
		{
			name: "residual within the 0.5 rounding tolerance is not injected",
			in: []model.ExposureRow{
				{Name: "North America", Weight: decimal.NewFromFloat(99.7)},
			},
			want: map[string]string{
				"North America": "99.7",
			},
			wantSum: "99.7",
		},
		{
			name:    "empty dimension closes entirely into Other",
			in:      []model.ExposureRow{},
			want:    map[string]string{"Other / Not Classified": "100"},
			wantSum: "100",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := prepareRegions(tc.in)
			if tc.wantErr {
				if !errors.Is(err, ErrInvalidWeights) {
					t.Fatalf("err = %v, want ErrInvalidWeights", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			byName := map[string]decimal.Decimal{}
			for _, r := range got {
				byName[r.Name] = r.Weight
			}
			if len(got) != len(tc.want) {
				t.Fatalf("rows = %+v, want %d rows", got, len(tc.want))
			}
			for name, w := range tc.want {
				if !byName[name].Equal(decimal.RequireFromString(w)) {
					t.Fatalf("region %q = %v, want %s", name, byName[name], w)
				}
			}
			if tc.wantSum != "" && !exposureWeightsSum(got).Equal(decimal.RequireFromString(tc.wantSum)) {
				t.Fatalf("stored sum = %v, want %s", exposureWeightsSum(got), tc.wantSum)
			}
		})
	}
}

func TestPrepareCountries(t *testing.T) {
	cases := []struct {
		name    string
		in      []model.ExposureRow
		wantErr bool
		want    []model.ExposureRow
	}{
		{
			name: "sum of 120 is rejected",
			in: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(70)},
				{Name: "JP", Weight: decimal.NewFromInt(50)},
			},
			wantErr: true,
		},
		{
			name: "sum of 95 is accepted with no lower bound",
			in: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(60)},
				{Name: "GB", Weight: decimal.NewFromInt(25)},
				{Name: "JP", Weight: decimal.NewFromInt(10)},
			},
			want: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(60)},
				{Name: "GB", Weight: decimal.NewFromInt(25)},
				{Name: "JP", Weight: decimal.NewFromInt(10)},
			},
		},
		{
			name: "non-canonical rows are dropped before the upper-bound check",
			in: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(60)},
				{Name: "Atlantis", Weight: decimal.NewFromInt(60)},
			},
			want: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(60)},
			},
		},
		{
			name: "names normalize to canonical ISO codes",
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
			name: "empty and non-positive rows are dropped",
			in: []model.ExposureRow{
				{Name: "", Weight: decimal.NewFromInt(50)},
				{Name: "JP", Weight: decimal.Zero},
				{Name: "US", Weight: decimal.NewFromInt(100)},
			},
			want: []model.ExposureRow{
				{Name: "US", Weight: decimal.NewFromInt(100)},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := prepareCountries(tc.in)
			if tc.wantErr {
				if !errors.Is(err, ErrInvalidWeights) {
					t.Fatalf("err = %v, want ErrInvalidWeights", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("rows = %+v, want %+v", got, tc.want)
			}
			for i := range got {
				if got[i].Name != tc.want[i].Name || !got[i].Weight.Equal(tc.want[i].Weight) {
					t.Fatalf("row[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestPrepareSectorsKeepsExactRule(t *testing.T) {
	sectors := normalizeExposureRows([]model.ExposureRow{
		{Name: "Technology", Weight: decimal.NewFromInt(103)},
	})
	if err := validateExposureWeights(sectors, weightSumExact100); !errors.Is(err, ErrInvalidWeights) {
		t.Fatalf("103%% sectors err = %v, want ErrInvalidWeights", err)
	}
	sectors = normalizeExposureRows([]model.ExposureRow{
		{Name: "Technology", Weight: decimal.NewFromInt(100)},
	})
	if err := validateExposureWeights(sectors, weightSumExact100); err != nil {
		t.Fatalf("100%% sectors err = %v, want nil", err)
	}
	sectors = normalizeExposureRows([]model.ExposureRow{
		{Name: "Technology", Weight: decimal.NewFromInt(95)},
	})
	if err := validateExposureWeights(sectors, weightSumExact100); !errors.Is(err, ErrInvalidWeights) {
		t.Fatalf("95%% sectors err = %v, want ErrInvalidWeights (sectors keep the exact-100 rule)", err)
	}
}

// assertNoExposureWrites checks that a preview fetch wrote no exposure
// dimension and no provenance at all.
func assertNoExposureWrites(t *testing.T, e *fakeExposureRepo, what string) {
	t.Helper()
	if e.replaceCountriesCalls != 0 || e.replaceRegionsCalls != 0 || e.replaceSectorsCalls != 0 {
		t.Fatalf("%s persisted exposure: ReplaceCountries=%d ReplaceRegions=%d ReplaceSectors=%d, want all 0",
			what, e.replaceCountriesCalls, e.replaceRegionsCalls, e.replaceSectorsCalls)
	}
	if e.setProvenanceCalls != 0 {
		t.Fatalf("%s persisted provenance: SetProvenance=%d, want 0", what, e.setProvenanceCalls)
	}
}

func TestFetchETFExposure_PreviewDoesNotPersist(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "SXR8.DE", Type: model.AssetTypeETF, ISIN: "IE00B4L5Y983"}}
	e := &fakeExposureRepo{}
	etf := &fakeETFFetcher{exposure: &model.AssetExposure{
		Countries: []model.ExposureRow{
			{Name: "United States", Weight: decimal.NewFromInt(60)},
			{Name: "Germany", Weight: decimal.NewFromInt(40)},
		},
		Sectors: []model.ExposureRow{
			{Name: "Technology", Weight: decimal.NewFromInt(50)},
			{Name: "Healthcare", Weight: decimal.NewFromInt(50)},
		},
	}}
	svc := newFetchTestService(t, a, e, nil, nil, etf)

	got, err := svc.FetchETFExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNoExposureWrites(t, e, "FetchETFExposure")
	if a.updateCalls != 0 {
		t.Fatalf("asset.Update calls = %d, want 0 (an asset with an ISIN must not be touched)", a.updateCalls)
	}
	if etf.exposureISIN != "IE00B4L5Y983" {
		t.Fatalf("provider called with ISIN %q, want IE00B4L5Y983", etf.exposureISIN)
	}
	if got.ISIN != "IE00B4L5Y983" {
		t.Fatalf("response ISIN = %q, want IE00B4L5Y983", got.ISIN)
	}
	if len(got.Countries) != len(geo.Countries) || len(got.Regions) != len(geo.Regions) || len(got.Sectors) != len(geo.GICSSectors) {
		t.Fatalf("preview not canonical: countries=%d regions=%d sectors=%d", len(got.Countries), len(got.Regions), len(got.Sectors))
	}
	countries := exposureRowsByName(got.Countries)
	if !countries["US"].Equal(decimal.NewFromInt(60)) || !countries["DE"].Equal(decimal.NewFromInt(40)) {
		t.Fatalf("preview countries = %+v, want normalized US 60 + DE 40", got.Countries)
	}
	regions := exposureRowsByName(got.Regions)
	if !regions["North America"].Equal(decimal.NewFromInt(60)) || !regions["Europe Developed"].Equal(decimal.NewFromInt(40)) {
		t.Fatalf("preview regions = %+v, want aggregated North America 60 + Europe Developed 40", got.Regions)
	}
	sectors := exposureRowsByName(got.Sectors)
	if !sectors["Information Technology"].Equal(decimal.NewFromInt(50)) || !sectors["Health Care"].Equal(decimal.NewFromInt(50)) {
		t.Fatalf("preview sectors = %+v, want aggregated Info Tech 50 + Health Care 50", got.Sectors)
	}
}

func TestFetchETFExposure_PreviewPersistsOnlyResolvedISIN(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "CW9", Name: "iShares Core MSCI World UCITS ETF USD (Acc)", Type: model.AssetTypeETF}}
	e := &fakeExposureRepo{}
	etf := &fakeETFFetcher{
		search: []price.EtfSearchResult{{ISIN: "IE00B4L5Y983", Ticker: "CW9", Name: "iShares Core MSCI World UCITS ETF USD (Acc)"}},
		exposure: &model.AssetExposure{
			Countries: []model.ExposureRow{{Name: "US", Weight: decimal.NewFromInt(100)}},
			Sectors:   []model.ExposureRow{{Name: "Technology", Weight: decimal.NewFromInt(100)}},
		},
	}
	svc := newFetchTestService(t, a, e, nil, nil, etf)

	got, err := svc.FetchETFExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.updateCalls != 1 {
		t.Fatalf("asset.Update calls = %d, want 1 (only the resolved ISIN is persisted)", a.updateCalls)
	}
	if a.lastUpdate == nil || a.lastUpdate.ISIN != "IE00B4L5Y983" {
		t.Fatalf("updated asset = %+v, want ISIN IE00B4L5Y983", a.lastUpdate)
	}
	if etf.exposureISIN != "IE00B4L5Y983" {
		t.Fatalf("provider called with ISIN %q, want the resolved IE00B4L5Y983", etf.exposureISIN)
	}
	assertNoExposureWrites(t, e, "FetchETFExposure")
	if got.ISIN != "IE00B4L5Y983" {
		t.Fatalf("response ISIN = %q, want IE00B4L5Y983", got.ISIN)
	}
}

func TestFetchMorningstarExposure_PreviewDoesNotPersist(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "SXR8.DE", Type: model.AssetTypeETF, ISIN: "IE00B4L5Y983"}}
	e := &fakeExposureRepo{}
	etf := &fakeETFFetcher{morningstar: &model.AssetExposure{
		Countries: []model.ExposureRow{
			{Name: "United States", Weight: decimal.NewFromInt(60)},
			{Name: "Germany", Weight: decimal.NewFromInt(40)},
		},
		Regions: []model.ExposureRow{
			{Name: "North America", Weight: decimal.NewFromInt(60)},
			{Name: "Europe Developed", Weight: decimal.NewFromInt(40)},
		},
		Sectors: []model.ExposureRow{
			{Name: "Technology", Weight: decimal.NewFromInt(100)},
		},
	}}
	svc := newFetchTestService(t, a, e, nil, nil, etf)

	got, err := svc.FetchMorningstarExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNoExposureWrites(t, e, "FetchMorningstarExposure")
	if a.updateCalls != 0 {
		t.Fatalf("asset.Update calls = %d, want 0 (an asset with an ISIN must not be touched)", a.updateCalls)
	}
	if etf.morningstarISIN != "IE00B4L5Y983" {
		t.Fatalf("provider called with ISIN %q, want IE00B4L5Y983", etf.morningstarISIN)
	}
	countries := exposureRowsByName(got.Countries)
	if !countries["US"].Equal(decimal.NewFromInt(60)) || !countries["DE"].Equal(decimal.NewFromInt(40)) {
		t.Fatalf("preview countries = %+v, want normalized US 60 + DE 40", got.Countries)
	}
	regions := exposureRowsByName(got.Regions)
	if !regions["North America"].Equal(decimal.NewFromInt(60)) || !regions["Europe Developed"].Equal(decimal.NewFromInt(40)) {
		t.Fatalf("preview regions = %+v, want official North America 60 + Europe Developed 40", got.Regions)
	}
	sectors := exposureRowsByName(got.Sectors)
	if !sectors["Information Technology"].Equal(decimal.NewFromInt(100)) {
		t.Fatalf("preview sectors = %+v, want aggregated Information Technology 100", got.Sectors)
	}
	if len(got.Regions) != len(geo.Regions) || len(got.Sectors) != len(geo.GICSSectors) {
		t.Fatalf("preview not canonical: regions=%d sectors=%d", len(got.Regions), len(got.Sectors))
	}
}

func TestFetchMorningstarExposure_PreviewPersistsOnlyResolvedISIN(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "CW9", Name: "iShares Core MSCI World UCITS ETF USD (Acc)", Type: model.AssetTypeETF}}
	e := &fakeExposureRepo{}
	etf := &fakeETFFetcher{
		search: []price.EtfSearchResult{{ISIN: "IE00B4L5Y983", Ticker: "CW9", Name: "iShares Core MSCI World UCITS ETF USD (Acc)"}},
		morningstar: &model.AssetExposure{
			Countries: []model.ExposureRow{{Name: "US", Weight: decimal.NewFromInt(100)}},
		},
	}
	svc := newFetchTestService(t, a, e, nil, nil, etf)

	got, err := svc.FetchMorningstarExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.updateCalls != 1 {
		t.Fatalf("asset.Update calls = %d, want 1 (only the resolved ISIN is persisted)", a.updateCalls)
	}
	if a.lastUpdate == nil || a.lastUpdate.ISIN != "IE00B4L5Y983" {
		t.Fatalf("updated asset = %+v, want ISIN IE00B4L5Y983", a.lastUpdate)
	}
	if etf.morningstarISIN != "IE00B4L5Y983" {
		t.Fatalf("provider called with ISIN %q, want the resolved IE00B4L5Y983", etf.morningstarISIN)
	}
	assertNoExposureWrites(t, e, "FetchMorningstarExposure")
	if got.ISIN != "IE00B4L5Y983" {
		t.Fatalf("response ISIN = %q, want IE00B4L5Y983", got.ISIN)
	}
}

func TestFetchAssetExposure_PreviewDoesNotPersist(t *testing.T) {
	assetID := uuid.New()
	key := assetID.String()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "AAPL", Type: model.AssetTypeStock, Country: "US", Sector: "Utilities"}}
	e := &fakeExposureRepo{
		regions:   map[string][]model.ExposureRow{key: {{Name: "North America", Weight: decimal.NewFromInt(100)}}},
		countries: map[string][]model.ExposureRow{key: {{Name: "US", Weight: decimal.NewFromInt(100)}}},
	}
	yf := &fakeYahooFetcher{
		sector:   "Technology",
		industry: "Consumer Electronics",
		country:  "United States",
		weightings: []model.ExposureRow{
			{Name: "Information Technology", Weight: decimal.NewFromInt(60)},
			{Name: "Health Care", Weight: decimal.NewFromInt(40)},
		},
	}
	svc := newFetchTestService(t, a, e, nil, yf, nil)

	got, err := svc.FetchAssetExposure(context.Background(), assetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if yf.profileExtendedCalls != 1 {
		t.Fatalf("provider calls = %d, want 1", yf.profileExtendedCalls)
	}
	assertNoExposureWrites(t, e, "FetchAssetExposure")
	if a.updateCalls != 0 {
		t.Fatalf("asset.Update calls = %d, want 0 (profile fields are preview-only)", a.updateCalls)
	}
	if a.asset.Sector != "Utilities" || a.asset.Country != "US" {
		t.Fatalf("asset profile mutated by preview: sector=%q country=%q", a.asset.Sector, a.asset.Country)
	}
	sectors := exposureRowsByName(got.Sectors)
	if !sectors["Information Technology"].Equal(decimal.NewFromInt(60)) || !sectors["Health Care"].Equal(decimal.NewFromInt(40)) {
		t.Fatalf("preview sectors = %+v, want the provider's Info Tech 60 + Health Care 40", got.Sectors)
	}
	if len(got.Sectors) != len(geo.GICSSectors) {
		t.Fatalf("preview sectors len = %d, want %d (canonical)", len(got.Sectors), len(geo.GICSSectors))
	}
	// The stored regions/countries dimensions are read back into the preview.
	regions := exposureRowsByName(got.Regions)
	if !regions["North America"].Equal(decimal.NewFromInt(100)) {
		t.Fatalf("preview regions = %+v, want the stored North America 100", got.Regions)
	}
	countries := exposureRowsByName(got.Countries)
	if !countries["US"].Equal(decimal.NewFromInt(100)) {
		t.Fatalf("preview countries = %+v, want the stored US 100", got.Countries)
	}
}

func TestFetchETFExposure_CachesFetchOnMissAndServesFromCache(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "SXR8.DE", Type: model.AssetTypeETF, ISIN: "IE00B4L5Y983"}}
	e := &fakeExposureRepo{}
	lk := &fakeLookupRepo{}
	etf := &fakeETFFetcher{exposure: &model.AssetExposure{
		Countries: []model.ExposureRow{
			{Name: "United States", Weight: decimal.NewFromInt(60)},
			{Name: "Germany", Weight: decimal.NewFromInt(40)},
		},
		Sectors: []model.ExposureRow{{Name: "Technology", Weight: decimal.NewFromInt(100)}},
	}}
	svc := newFetchTestService(t, a, e, lk, nil, etf)

	first, err := svc.FetchETFExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if etf.exposureCalls != 1 {
		t.Fatalf("provider calls = %d, want 1 on cache miss", etf.exposureCalls)
	}
	if lk.setCalls != 1 {
		t.Fatalf("cache Set calls = %d, want 1 after a miss", lk.setCalls)
	}
	if _, ok := lk.data[exposureCacheKey("justetf", "IE00B4L5Y983")]; !ok {
		t.Fatalf("payload not stored under the justetf cache key: %v", lk.data)
	}

	// A different provider payload must not surface: the second call is served
	// from the cached countries, so the fetcher is never reached again.
	etf.exposure = &model.AssetExposure{
		Countries: []model.ExposureRow{{Name: "Japan", Weight: decimal.NewFromInt(100)}},
	}
	second, err := svc.FetchETFExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if etf.exposureCalls != 1 {
		t.Fatalf("provider calls = %d, want 1 (cache hit must skip the fetcher)", etf.exposureCalls)
	}
	if lk.setCalls != 1 {
		t.Fatalf("cache Set calls = %d, want 1 (a hit must not rewrite the cache)", lk.setCalls)
	}
	countries := exposureRowsByName(second.Countries)
	if !countries["US"].Equal(decimal.NewFromInt(60)) || !countries["DE"].Equal(decimal.NewFromInt(40)) {
		t.Fatalf("cached preview countries = %+v, want US 60 + DE 40 from the first fetch", second.Countries)
	}
	if !countries["JP"].IsZero() {
		t.Fatalf("cached preview leaked the fresh provider payload: %+v", second.Countries)
	}
	if len(first.Countries) != len(geo.Countries) {
		t.Fatalf("first preview not canonical: countries=%d", len(first.Countries))
	}
}

func TestFetchETFExposure_RefreshBypassesCache(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "SXR8.DE", Type: model.AssetTypeETF, ISIN: "IE00B4L5Y983"}}
	lk := &fakeLookupRepo{}
	etf := &fakeETFFetcher{exposure: &model.AssetExposure{
		Countries: []model.ExposureRow{{Name: "Germany", Weight: decimal.NewFromInt(100)}},
	}}
	svc := newFetchTestService(t, a, &fakeExposureRepo{}, lk, nil, etf)

	if _, err := svc.FetchETFExposure(context.Background(), assetID, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	getsAfterPrime := lk.getCalls

	etf.exposure = &model.AssetExposure{
		Countries: []model.ExposureRow{{Name: "United States", Weight: decimal.NewFromInt(100)}},
	}
	got, err := svc.FetchETFExposure(context.Background(), assetID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if etf.exposureCalls != 2 {
		t.Fatalf("provider calls = %d, want 2 (refresh=true must re-fetch)", etf.exposureCalls)
	}
	if lk.getCalls != getsAfterPrime {
		t.Fatalf("cache Get calls = %d, want %d (refresh=true must not read the cache)", lk.getCalls, getsAfterPrime)
	}
	if lk.setCalls != 2 {
		t.Fatalf("cache Set calls = %d, want 2 (refresh=true must rewrite the cache)", lk.setCalls)
	}
	countries := exposureRowsByName(got.Countries)
	if !countries["US"].Equal(decimal.NewFromInt(100)) {
		t.Fatalf("preview countries = %+v, want the refreshed US 100 payload", got.Countries)
	}

	// The rewritten payload is what subsequent non-refresh calls serve.
	if _, err := svc.FetchETFExposure(context.Background(), assetID, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if etf.exposureCalls != 2 {
		t.Fatalf("provider calls = %d, want 2 (the refreshed payload must be cached)", etf.exposureCalls)
	}
}

func TestFetchETFExposure_EmptyResultNotCached(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "SXR8.DE", Type: model.AssetTypeETF, ISIN: "IE00B4L5Y983"}}
	lk := &fakeLookupRepo{}
	etf := &fakeETFFetcher{exposure: &model.AssetExposure{}}
	svc := newFetchTestService(t, a, &fakeExposureRepo{}, lk, nil, etf)

	if _, err := svc.FetchETFExposure(context.Background(), assetID, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lk.setCalls != 0 {
		t.Fatalf("cache Set calls = %d, want 0 (a country-less payload must not be cached)", lk.setCalls)
	}
	if len(lk.data) != 0 {
		t.Fatalf("cache data = %v, want empty", lk.data)
	}
	if _, err := svc.FetchETFExposure(context.Background(), assetID, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if etf.exposureCalls != 2 {
		t.Fatalf("provider calls = %d, want 2 (nothing was cached, so the fetcher runs again)", etf.exposureCalls)
	}
}

func TestFetchMorningstarExposure_CachesUnderItsOwnKey(t *testing.T) {
	assetID := uuid.New()
	a := &fakeAssetRepo{asset: &model.Asset{ID: assetID, Ticker: "SXR8.DE", Type: model.AssetTypeETF, ISIN: "IE00B4L5Y983"}}
	lk := &fakeLookupRepo{}
	etf := &fakeETFFetcher{
		exposure: &model.AssetExposure{
			Countries: []model.ExposureRow{{Name: "Germany", Weight: decimal.NewFromInt(100)}},
		},
		morningstar: &model.AssetExposure{
			Countries: []model.ExposureRow{{Name: "United States", Weight: decimal.NewFromInt(60)}},
			Regions:   []model.ExposureRow{{Name: "North America", Weight: decimal.NewFromInt(60)}},
			Sectors:   []model.ExposureRow{{Name: "Technology", Weight: decimal.NewFromInt(100)}},
		},
	}
	svc := newFetchTestService(t, a, &fakeExposureRepo{}, lk, nil, etf)

	// Priming the justetf cache must not feed the morningstar source.
	if _, err := svc.FetchETFExposure(context.Background(), assetID, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	first, err := svc.FetchMorningstarExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if etf.morningstarCall != 1 {
		t.Fatalf("morningstar provider calls = %d, want 1 (separate cache key per source)", etf.morningstarCall)
	}
	if _, ok := lk.data[exposureCacheKey("morningstar", "IE00B4L5Y983")]; !ok {
		t.Fatalf("payload not stored under the morningstar cache key: %v", lk.data)
	}

	etf.morningstar = &model.AssetExposure{
		Countries: []model.ExposureRow{{Name: "Japan", Weight: decimal.NewFromInt(100)}},
	}
	second, err := svc.FetchMorningstarExposure(context.Background(), assetID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if etf.morningstarCall != 1 {
		t.Fatalf("morningstar provider calls = %d, want 1 (cache hit must skip the fetcher)", etf.morningstarCall)
	}
	countries := exposureRowsByName(second.Countries)
	if !countries["US"].Equal(decimal.NewFromInt(60)) || !countries["JP"].IsZero() {
		t.Fatalf("cached preview countries = %+v, want US 60 from the first fetch", second.Countries)
	}
	regions := exposureRowsByName(second.Regions)
	if !regions["North America"].Equal(decimal.NewFromInt(60)) {
		t.Fatalf("cached preview regions = %+v, want the cached official North America 60", second.Regions)
	}
	if len(first.Regions) != len(geo.Regions) {
		t.Fatalf("first preview not canonical: regions=%d", len(first.Regions))
	}
}
