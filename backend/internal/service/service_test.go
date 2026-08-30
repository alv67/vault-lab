package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/amelamela/vault-lab/internal/cache"
	"github.com/amelamela/vault-lab/internal/geo"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/repository"
)

type fakePortfolioRepo struct {
	portfolio *model.Portfolio
	holdings  []*model.Holding
}

func (f *fakePortfolioRepo) Create(ctx context.Context, p *model.Portfolio) (*model.Portfolio, error) {
	return p, nil
}
func (f *fakePortfolioRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Portfolio, error) {
	return f.portfolio, nil
}
func (f *fakePortfolioRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.Portfolio, error) {
	return nil, nil
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
	regions map[string][]model.ExposureRow
	sectors map[string][]model.ExposureRow
}

func (f *fakeExposureRepo) FindRegions(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error) {
	return f.regions[assetID.String()], nil
}
func (f *fakeExposureRepo) FindSectors(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error) {
	return f.sectors[assetID.String()], nil
}
func (f *fakeExposureRepo) ReplaceRegions(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	return nil
}
func (f *fakeExposureRepo) ReplaceSectors(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	return nil
}
func (f *fakeExposureRepo) FindRegionsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error) {
	return f.regions, nil
}
func (f *fakeExposureRepo) FindSectorsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error) {
	return f.sectors, nil
}

type fakeFXRepo struct {
	rates map[string]decimal.Decimal
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

func newTestService(t *testing.T, p *fakePortfolioRepo, e *fakeExposureRepo, f *fakeFXRepo) *Service {
	t.Helper()
	repos := &repository.Repository{
		Portfolio: p,
		Exposure:  e,
		FX:        f,
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
	if !equalDecimal(got.Regions[2].Value, wantVal.Mul(decimal.NewFromInt(40)).Div(decimal.NewFromInt(100))) {
		t.Fatalf("Europe Developed value = %v, want 400", got.Regions[2].Value)
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
			holding(assetID.String(), "USD", "", "", model.AssetTypeBond, decimal.NewFromInt(10), decimal.NewFromInt(100)),
		},
	}
	// Bond with no stored rows and (for geography) an empty country -> region
	// default is "Other / Not Classified" for the geography, but for the sector
	// NormalizeSector("") returns "" so no default and sum is 0 -> Other.
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
}
