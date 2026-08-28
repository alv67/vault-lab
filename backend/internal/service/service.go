package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/amelamela/vault-lab/internal/auth"
	"github.com/amelamela/vault-lab/internal/cache"
	"github.com/amelamela/vault-lab/internal/geo"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/price"
	"github.com/amelamela/vault-lab/internal/repository"
	"github.com/amelamela/vault-lab/internal/series"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already registered")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrAssetInUse         = errors.New("asset is used in transactions")
	ErrAssetNotFound      = errors.New("asset not found")
	ErrInvalidWeights     = errors.New("weights must sum to 100")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidInput       = errors.New("invalid input")
	ErrCurrencyExists     = errors.New("currency already in whitelist")
	ErrCurrencyInUse      = errors.New("currency is used by assets or portfolios")
	ErrCurrencyProtected  = errors.New("currency cannot be removed")
	ErrCurrencyNotManaged = errors.New("currency conversion not available")
	ErrInvalidAssetClass = errors.New("invalid asset class")
)

// assetClasses is the allowed set for asset_class plus a helper to derive a
// default class from the asset type.
var assetClasses = map[string]bool{
	"equity": true, "bond": true, "commodity": true, "currency": true,
	"crypto": true, "real_estate": true, "mixed": true, "other": true,
}

// defaultAssetClassForType maps an asset type to its default investment class.
func defaultAssetClassForType(t model.AssetType) string {
	switch t {
	case model.AssetTypeStock:
		return "equity"
	case model.AssetTypeBond:
		return "bond"
	case model.AssetTypeCommodity:
		return "commodity"
	case model.AssetTypeCrypto:
		return "crypto"
	case model.AssetTypeCash:
		return "currency"
	default:
		return "other"
	}
}

const (
	cacheTTLStats  = 5 * time.Minute
	cacheTTLPrices = time.Hour
)

type Service struct {
	repos           *repository.Repository
	jwtAuth         *auth.JWTAuth
	fetcher         *price.YahooFetcher
	lookupCacheTTL  time.Duration
	cache           *cache.Cache
	seriesMaxPoints int
	stalePriceDays  int
	Health          *HealthService
}

func New(repos *repository.Repository, jwtAuth *auth.JWTAuth, fetcher *price.YahooFetcher, lookupCacheTTL time.Duration, c *cache.Cache, seriesMaxPoints int, stalePriceDays int, health *HealthService) *Service {
	if seriesMaxPoints <= 0 {
		seriesMaxPoints = 500
	}
	if stalePriceDays <= 0 {
		stalePriceDays = 7
	}
	return &Service{repos: repos, jwtAuth: jwtAuth, fetcher: fetcher, lookupCacheTTL: lookupCacheTTL, cache: c, seriesMaxPoints: seriesMaxPoints, stalePriceDays: stalePriceDays, Health: health}
}

// cached implements the read-through cache pattern: it reads the current data
// revision, returns the payload from Redis on a hit and otherwise computes and
// stores it. When bump is set the revision is advanced after compute so writes
// performed during the computation invalidate previously cached entries.
func cached[T any](c *cache.Cache, ctx context.Context, kind string, id string, ttl time.Duration, bump bool, compute func() (T, error)) (T, error) {
	rev, err := c.Rev(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("cache rev read failed")
	}
	key := fmt.Sprintf("vl:%s:%s:%d", kind, id, rev)
	var v T
	if hit, err := c.GetJSON(ctx, key, &v); err != nil {
		log.Warn().Err(err).Msg("cache read failed")
	} else if hit {
		return v, nil
	}
	result, err := compute()
	if err != nil {
		return result, err
	}
	if bump {
		rev, err = c.Bump(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("cache rev bump failed")
		}
		key = fmt.Sprintf("vl:%s:%s:%d", kind, id, rev)
	}
	if err := c.SetJSON(ctx, key, result, ttl); err != nil {
		log.Warn().Err(err).Msg("cache write failed")
	}
	return result, nil
}

// bumpRev advances the global data revision after any write so cached reads
// keyed on the previous revision are invalidated.
func (s *Service) bumpRev(ctx context.Context) {
	if _, err := s.cache.Bump(ctx); err != nil {
		log.Warn().Err(err).Msg("cache rev bump failed")
	}
}

func (s *Service) Register(ctx context.Context, email, name, password string) (*model.User, error) {
	existing, _ := s.repos.User.FindByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailExists
	}
	return s.repos.User.Create(ctx, email, name, password)
}

func (s *Service) Login(ctx context.Context, email, password string) (*model.User, string, string, error) {
	user, err := s.repos.User.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, err := s.jwtAuth.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.jwtAuth.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, tokenString string) (string, string, error) {
	claims, err := s.jwtAuth.ValidateToken(tokenString)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}
	if claims.TokenType != "refresh" {
		return "", "", ErrInvalidCredentials
	}

	user, err := s.repos.User.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", "", ErrNotFound
	}

	accessToken, err := s.jwtAuth.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwtAuth.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, claims *auth.Claims) (*model.User, error) {
	return s.repos.User.FindByID(ctx, claims.UserID)
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, name, email string) (*model.User, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	if email == "" || !strings.Contains(email, "@") {
		return nil, ErrInvalidInput
	}

	user, err := s.repos.User.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}

	if email != user.Email {
		existing, _ := s.repos.User.FindByEmail(ctx, email)
		if existing != nil && existing.ID != userID {
			return nil, ErrEmailExists
		}
		user.Email = email
	}
	user.Name = name

	if err := s.repos.User.Update(ctx, user); err != nil {
		return nil, err
	}
	return s.repos.User.FindByID(ctx, userID)
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	user, err := s.repos.User.FindByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repos.User.UpdatePassword(ctx, userID, string(hash))
}

func (s *Service) CreateAsset(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	if asset.AssetClass == "" {
		asset.AssetClass = defaultAssetClassForType(asset.Type)
	}
	created, err := s.repos.Asset.Create(ctx, asset)
	if err != nil {
		return nil, err
	}
	s.syncAssetBackground(created.ID)
	return created, nil
}

func (s *Service) GetAsset(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	return s.repos.Asset.FindByID(ctx, id)
}

// UpdateAsset merges the editable asset fields from the patch into the stored
// asset and persists the result. Only fields explicitly present in the patch
// are applied; for the required fields an empty value keeps the current one,
// while the optional string fields can be cleared by sending an empty string.
func (s *Service) UpdateAsset(ctx context.Context, id uuid.UUID, patch *model.AssetPatch) (*model.Asset, error) {
	existing, err := s.repos.Asset.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	if patch.Ticker != nil && *patch.Ticker != "" {
		existing.Ticker = *patch.Ticker
	}
	if patch.ISIN != nil {
		existing.ISIN = *patch.ISIN
	}
	if patch.Name != nil && *patch.Name != "" {
		existing.Name = *patch.Name
	}
	if patch.Type != nil && *patch.Type != "" {
		existing.Type = *patch.Type
	}
	if patch.AssetClass != nil {
		if *patch.AssetClass != "" && !assetClasses[*patch.AssetClass] {
			return nil, ErrInvalidAssetClass
		}
		existing.AssetClass = *patch.AssetClass
	}
	if patch.Country != nil {
		existing.Country = *patch.Country
	}
	if patch.Currency != nil && *patch.Currency != "" {
		existing.Currency = *patch.Currency
	}
	if patch.Exchange != nil {
		existing.Exchange = *patch.Exchange
	}
	if patch.Sector != nil {
		existing.Sector = *patch.Sector
	}
	if patch.Industry != nil {
		existing.Industry = *patch.Industry
	}

	updated, err := s.repos.Asset.Update(ctx, existing)
	if err != nil {
		return nil, err
	}
	s.bumpRev(ctx)
	return updated, nil
}

// GetAssetQuote returns the headline metrics for the asset detail page: latest
// close and the percentage changes vs. 1 day, 1 week, 1 month, 1 year and
// year-to-date reference closes. Changes are zero when the reference price is
// missing.
func (s *Service) GetAssetQuote(ctx context.Context, id uuid.UUID) (*model.AssetQuote, error) {
	asset, err := s.repos.Asset.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	quote := &model.AssetQuote{Currency: asset.Currency}
	latest, err := s.repos.Price.FindLatest(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return quote, nil
		}
		return nil, err
	}

	quote.HasData = true
	quote.LastClose = latest.Close
	quote.LastDate = latest.Date

	now := time.Now()
	refDates := []time.Time{
		series.DayOf(now.AddDate(0, 0, -1)),
		series.DayOf(now.AddDate(0, 0, -7)),
		series.DayOf(now.AddDate(0, -1, 0)),
		series.DayOf(now.AddDate(-1, 0, 0)),
		series.DayOf(time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	refCloses, err := s.repos.Price.ReferenceCloses(ctx, id, refDates)
	if err != nil {
		return nil, err
	}

	changes := []struct {
		ref time.Time
		pct *decimal.Decimal
	}{
		{refDates[0], &quote.Change1D},
		{refDates[1], &quote.Change1W},
		{refDates[2], &quote.Change1M},
		{refDates[3], &quote.Change1Y},
		{refDates[4], &quote.ChangeYTD},
	}
	for _, c := range changes {
		refClose, ok := refCloses[c.ref]
		if !ok || refClose.IsZero() {
			continue
		}
		*c.pct = quote.LastClose.Sub(refClose).Div(refClose).Mul(decimal.NewFromInt(100))
	}
	return quote, nil
}

// FetchAssetProfile resolves sector and industry from Yahoo and persists them
// on the asset. The Yahoo error is propagated to the caller when the profile
// cannot be fetched.
func (s *Service) FetchAssetProfile(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	asset, err := s.repos.Asset.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	sector, industry, err := s.fetcher.FetchAssetProfile(ctx, asset.Ticker)
	if err != nil {
		return nil, err
	}
	asset.Sector = sector
	asset.Industry = industry

	updated, err := s.repos.Asset.Update(ctx, asset)
	if err != nil {
		return nil, err
	}
	s.bumpRev(ctx)
	return updated, nil
}

// GetAssetExposure returns the region and sector weight distribution of an
// asset. The output always contains every canonical region and GICS sector, in
// canonical order, with zero weight when not stored. When no weights are stored
// for a dimension, stocks fall back to a single 100% entry derived from the
// asset country and sector.
func (s *Service) GetAssetExposure(ctx context.Context, id uuid.UUID) (*model.AssetExposure, error) {
	asset, err := s.repos.Asset.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	regions, err := s.repos.Exposure.FindRegions(ctx, id)
	if err != nil {
		return nil, err
	}
	sectors, err := s.repos.Exposure.FindSectors(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.buildExposure(asset, regions, sectors), nil
}

// SaveAssetExposure validates and persists the weight distribution for an
// asset. Each dimension is validated and saved independently: a dimension that
// is absent from the body (nil slice) is left untouched. Present dimensions
// must sum to ~100 (tolerance 0.5); rows with an empty name or a non-positive
// weight are ignored. Returns the complete output built from the stored state.
func (s *Service) SaveAssetExposure(ctx context.Context, id uuid.UUID, exposure *model.AssetExposure) (*model.AssetExposure, error) {
	asset, err := s.repos.Asset.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	var regions, sectors []model.ExposureRow
	err = s.repos.WithTx(ctx, func(rx *repository.Repository) error {
		if exposure.Regions != nil {
			regions = normalizeExposureRows(exposure.Regions)
			if err := validateExposureWeights(regions); err != nil {
				return err
			}
			if err := rx.Exposure.ReplaceRegions(ctx, id, regions); err != nil {
				return err
			}
		}
		if exposure.Sectors != nil {
			sectors = normalizeExposureRows(exposure.Sectors)
			if err := validateExposureWeights(sectors); err != nil {
				return err
			}
			if err := rx.Exposure.ReplaceSectors(ctx, id, sectors); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.bumpRev(ctx)

	// Ricostruisci l'output completo dallo stato persistito (la dimensione non
	// toccata mantiene i valori memorizzati, non quelli del body).
	storedRegions, err := s.repos.Exposure.FindRegions(ctx, id)
	if err != nil {
		return nil, err
	}
	storedSectors, err := s.repos.Exposure.FindSectors(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.buildExposure(asset, storedRegions, storedSectors), nil
}

// FetchAssetExposure fetches sector/industry + sector weights from Yahoo,
// persists them into the exposure tables and returns the complete exposure.
func (s *Service) FetchAssetExposure(ctx context.Context, id uuid.UUID) (*model.AssetExposure, error) {
	asset, err := s.repos.Asset.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	sector, industry, weightings, fundCategory, err := s.fetcher.FetchAssetProfileExtended(ctx, asset.Ticker)
	if err != nil {
		return nil, err
	}

	sectors := make([]model.ExposureRow, 0, len(weightings))
	for _, w := range weightings {
		if w.Name == "" || !w.Weight.IsPositive() {
			continue
		}
		sectors = append(sectors, w)
	}
	// Single stocks have no topHoldings weights: fall back to the assetProfile
	// sector at 100%.
	if len(sectors) == 0 && sector != "" {
		sectors = []model.ExposureRow{{Name: geo.NormalizeSector(sector), Weight: decimal.NewFromInt(100)}}
	}

	// Keep existing non-empty profile fields untouched when Yahoo returns empty.
	if sector != "" {
		asset.Sector = geo.NormalizeSector(sector)
	}
	if industry != "" {
		asset.Industry = industry
	}
	// Apply the detected class only when the current one is 'other' or empty so
	// a manual override wins. Classification is best-effort (keyword heuristics
	// on name + fund category), never authoritative.
	detected := geo.ClassifyAssetClass(string(asset.Type), asset.Name, fundCategory)
	if detected != "" && (asset.AssetClass == "" || asset.AssetClass == "other") {
		asset.AssetClass = detected
	}

	regions, err := s.repos.Exposure.FindRegions(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.repos.WithTx(ctx, func(rx *repository.Repository) error {
		if _, err := rx.Asset.Update(ctx, asset); err != nil {
			return err
		}
		return rx.Exposure.ReplaceSectors(ctx, id, sectors)
	})
	if err != nil {
		return nil, err
	}
	s.bumpRev(ctx)
	return s.buildExposure(asset, regions, sectors), nil
}

// buildExposure assembles the complete exposure output for an asset: every
// canonical region and GICS sector in canonical order, overlaid with the stored
// weights and the stock defaults when nothing is stored.
func (s *Service) buildExposure(asset *model.Asset, regions, sectors []model.ExposureRow) *model.AssetExposure {
	out := &model.AssetExposure{
		Regions: make([]model.ExposureRow, 0, len(geo.Regions)),
		Sectors: make([]model.ExposureRow, 0, len(geo.GICSSectors)),
	}

	regionWeights := make(map[string]decimal.Decimal, len(regions))
	for _, r := range regions {
		regionWeights[r.Name] = r.Weight
	}
	for _, name := range geo.Regions {
		out.Regions = append(out.Regions, model.ExposureRow{Name: name, Weight: regionWeights[name]})
	}

	sectorWeights := make(map[string]decimal.Decimal, len(sectors))
	for _, r := range sectors {
		sectorWeights[r.Name] = r.Weight
	}
	for _, name := range geo.GICSSectors {
		out.Sectors = append(out.Sectors, model.ExposureRow{Name: name, Weight: sectorWeights[name]})
	}

	if asset.Type == model.AssetTypeStock {
		if len(regions) == 0 {
			region := geo.RegionForCountry(asset.Country)
			for i := range out.Regions {
				if out.Regions[i].Name == region {
					out.Regions[i].Weight = decimal.NewFromInt(100)
					break
				}
			}
		}
		if len(sectors) == 0 && asset.Sector != "" {
			sector := geo.NormalizeSector(asset.Sector)
			for i := range out.Sectors {
				if out.Sectors[i].Name == sector {
					out.Sectors[i].Weight = decimal.NewFromInt(100)
					break
				}
			}
		}
	}
	return out
}

// normalizeExposureRows drops rows with an empty name or a non-positive weight.
func normalizeExposureRows(rows []model.ExposureRow) []model.ExposureRow {
	out := make([]model.ExposureRow, 0, len(rows))
	for _, row := range rows {
		if row.Name == "" || !row.Weight.IsPositive() {
			continue
		}
		out = append(out, row)
	}
	return out
}

// validateExposureWeights checks that a dimension sums to ~100 with a 0.5
// tolerance.
func validateExposureWeights(rows []model.ExposureRow) error {
	sum := decimal.Zero
	for _, row := range rows {
		sum = sum.Add(row.Weight)
	}
	if sum.Sub(decimal.NewFromInt(100)).Abs().GreaterThan(decimal.NewFromFloat(0.5)) {
		return ErrInvalidWeights
	}
	return nil
}

func (s *Service) SearchAssets(ctx context.Context, query string) ([]*model.Asset, error) {
	return s.repos.Asset.Search(ctx, query)
}

func (s *Service) GetAssetMeta(ctx context.Context, ticker string) (*price.AssetMeta, error) {
	key := "meta:" + strings.ToUpper(strings.TrimSpace(ticker))

	if cached, err := s.repos.Lookup.Get(ctx, key); err == nil {
		var m price.AssetMeta
		if err := json.Unmarshal(cached, &m); err == nil {
			return &m, nil
		}
	}

	meta, err := s.fetcher.FetchMeta(ctx, ticker)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(meta); err == nil {
		if err := s.repos.Lookup.Set(ctx, key, data, s.lookupCacheTTL); err != nil {
			log.Warn().Err(err).Str("ticker", key).Msg("failed to cache asset meta")
		}
	}

	return meta, nil
}

func (s *Service) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	used, err := s.repos.Transaction.CountByAsset(ctx, id)
	if err != nil {
		return err
	}
	if used > 0 {
		return ErrAssetInUse
	}
	return s.repos.Asset.Delete(ctx, id)
}

func (s *Service) LookupAsset(ctx context.Context, query string) ([]price.AssetLookup, error) {
	key := strings.ToLower(strings.TrimSpace(query))
	if key == "" {
		return []price.AssetLookup{}, nil
	}

	if cached, err := s.repos.Lookup.Get(ctx, key); err == nil {
		var results []price.AssetLookup
		if err := json.Unmarshal(cached, &results); err == nil {
			return results, nil
		}
	}

	results, err := price.LookupAsset(ctx, query)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(results); err == nil {
		if err := s.repos.Lookup.Set(ctx, key, data, s.lookupCacheTTL); err != nil {
			log.Warn().Err(err).Str("query", key).Msg("failed to cache lookup results")
		}
	}

	return results, nil
}

// RefreshPrices refreshes stale stored closes for the given portfolio (or all
// assets when portfolioID is nil) plus the USD->X FX rates, hitting Yahoo
// Finance only when needed.
func (s *Service) RefreshPrices(ctx context.Context, portfolioID *uuid.UUID) (price.RefreshReport, error) {
	report := price.RefreshReport{Refreshed: []string{}, Issues: []price.FetchIssue{}}
	var err error
	if portfolioID != nil {
		report, err = s.fetcher.RefreshStaleForPortfolio(ctx, *portfolioID)
	} else {
		var assets []*model.Asset
		assets, err = s.repos.Asset.List(ctx)
		if err != nil {
			return report, err
		}
		report, err = s.fetcher.RefreshStale(ctx, assets)
	}
	if err != nil {
		return report, err
	}
	fxIssues, err := s.fetcher.RefreshFX(ctx)
	if err != nil {
		return report, err
	}
	report.Issues = append(report.Issues, fxIssues...)
	report.RateLimited = false
	for _, iss := range report.Issues {
		if iss.Code == "rate_limited" {
			report.RateLimited = true
			break
		}
	}
	if err := series.RecomputeAll(ctx, s.repos); err != nil {
		log.Warn().Err(err).Msg("series recompute all failed")
	}
	s.bumpRev(ctx)
	return report, nil
}

func (s *Service) ListAssets(ctx context.Context) ([]*model.Asset, error) {
	return s.repos.Asset.List(ctx)
}

// ListCurrencies returns the enabled whitelisted currencies, ordered.
func (s *Service) ListCurrencies(ctx context.Context) ([]model.Currency, error) {
	return s.repos.Currency.ListEnabled(ctx)
}

// AddCurrency validates that a USD->code conversion is available on Yahoo and
// adds the currency to the enabled whitelist.
func (s *Service) AddCurrency(ctx context.Context, code, name string) (*model.Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, fmt.Errorf("%w: empty currency code", ErrInvalidInput)
	}
	if err := s.ValidateCurrency(ctx, code); err != nil {
		return nil, err
	}

	existing, err := s.repos.Currency.Get(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCurrencyExists
	}

	if name == "" {
		name = defaultCurrencyName(code)
	}

	currencies, err := s.repos.Currency.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	c := &model.Currency{
		Code:    code,
		Name:    name,
		Enabled: true,
		Sort:    nextSort(currencies),
	}
	if err := s.repos.Currency.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// DeleteCurrency removes a currency from the whitelist. Removal is blocked for
// the default USD base and for any currency still referenced by assets or
// portfolios.
func (s *Service) DeleteCurrency(ctx context.Context, code string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return fmt.Errorf("%w: empty currency code", ErrInvalidInput)
	}
	if code == "USD" {
		return ErrCurrencyProtected
	}

	existing, err := s.repos.Currency.Get(ctx, code)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	inUse, err := s.repos.Currency.CountInUse(ctx, code)
	if err != nil {
		return err
	}
	if inUse > 0 {
		return ErrCurrencyInUse
	}
	return s.repos.Currency.Delete(ctx, code)
}

// ValidateCurrency ensures a USD->code FX conversion is available on Yahoo, so
// the currency can actually be managed.
func (s *Service) ValidateCurrency(ctx context.Context, code string) error {
	if _, err := s.fetcher.FetchFXRate(ctx, code); err != nil {
		return fmt.Errorf("%w: USD->%s conversion unavailable", ErrCurrencyNotManaged, code)
	}
	return nil
}

func defaultCurrencyName(code string) string {
	if code == "USD" {
		return "US Dollar"
	}
	if code == "EUR" {
		return "Euro"
	}
	return code
}

func nextSort(currencies []model.Currency) int {
	max := 0
	for _, c := range currencies {
		if c.Sort > max {
			max = c.Sort
		}
	}
	return max + 1
}

func (s *Service) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, currency string) (*model.Portfolio, error) {
	p := &model.Portfolio{
		UserID:      userID,
		Name:        name,
		Description: description,
		Currency:    currency,
	}
	created, err := s.repos.Portfolio.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	s.bumpRev(ctx)
	return created, nil
}

func (s *Service) ListPortfolios(ctx context.Context, userID uuid.UUID) ([]*model.Portfolio, error) {
	return s.repos.Portfolio.FindByUser(ctx, userID)
}

func (s *Service) GetPortfolio(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Portfolio, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.canAccessPortfolio(ctx, p, userID) {
		return nil, ErrForbidden
	}
	return p, nil
}

func (s *Service) UpdatePortfolio(ctx context.Context, p *model.Portfolio) error {
	if err := s.repos.Portfolio.Update(ctx, p); err != nil {
		return err
	}
	if err := series.Recompute(ctx, s.repos, p.ID); err != nil {
		log.Warn().Err(err).Str("portfolio_id", p.ID.String()).Msg("series recompute failed")
	}
	s.bumpRev(ctx)
	return nil
}

func (s *Service) DeletePortfolio(ctx context.Context, id uuid.UUID) error {
	if err := s.repos.Portfolio.Delete(ctx, id); err != nil {
		return err
	}
	s.bumpRev(ctx)
	return nil
}

func (s *Service) AddTransaction(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	tx, err := s.repos.Transaction.Create(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err := series.Recompute(ctx, s.repos, tx.PortfolioID); err != nil {
		log.Warn().Err(err).Str("portfolio_id", tx.PortfolioID.String()).Msg("series recompute failed")
	}
	s.bumpRev(ctx)
	return tx, nil
}

func (s *Service) ListTransactions(ctx context.Context, portfolioID uuid.UUID) ([]model.TransactionWithAsset, error) {
	return s.repos.Transaction.FindByPortfolio(ctx, portfolioID)
}

// UpdateTransaction edits a transaction after verifying the caller owns the
// portfolio the transaction belongs to.
func (s *Service) UpdateTransaction(ctx context.Context, userID uuid.UUID, tx *model.Transaction) error {
	existing, err := s.repos.Transaction.FindByID(ctx, tx.ID)
	if err != nil {
		return ErrNotFound
	}
	p, err := s.repos.Portfolio.FindByID(ctx, existing.PortfolioID)
	if err != nil {
		return err
	}
	if !s.canAccessPortfolio(ctx, p, userID) {
		return ErrForbidden
	}
	tx.PortfolioID = existing.PortfolioID
	if err := s.repos.Transaction.Update(ctx, tx); err != nil {
		return err
	}
	if err := series.Recompute(ctx, s.repos, existing.PortfolioID); err != nil {
		log.Warn().Err(err).Str("portfolio_id", existing.PortfolioID.String()).Msg("series recompute failed")
	}
	s.bumpRev(ctx)
	return nil
}

// DeleteTransaction removes a transaction after verifying the caller owns the
// portfolio it belongs to.
func (s *Service) DeleteTransaction(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	existing, err := s.repos.Transaction.FindByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	p, err := s.repos.Portfolio.FindByID(ctx, existing.PortfolioID)
	if err != nil {
		return err
	}
	if !s.canAccessPortfolio(ctx, p, userID) {
		return ErrForbidden
	}
	if err := s.repos.Transaction.Delete(ctx, id); err != nil {
		return err
	}
	if err := series.Recompute(ctx, s.repos, existing.PortfolioID); err != nil {
		log.Warn().Err(err).Str("portfolio_id", existing.PortfolioID.String()).Msg("series recompute failed")
	}
	s.bumpRev(ctx)
	return nil
}

func (s *Service) GetPortfolioSummary(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioSummary, error) {
	return cached(s.cache, ctx, "summary", portfolioID.String(), cacheTTLStats, false, func() (*model.PortfolioSummary, error) {
		p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
		if err != nil {
			return nil, err
		}
		rates, err := series.LoadRates(ctx, s.repos, holdings, p.Currency)
		if err != nil {
			return nil, err
		}

		summary := &model.PortfolioSummary{
			PortfolioID:   portfolioID.String(),
			PortfolioName: p.Name,
			AssetCount:    len(holdings),
		}
		var totalValue, totalCost, totalRealized decimal.Decimal
		summary.Holdings = make([]model.AssetHolding, 0, len(holdings))
		staleThreshold := time.Duration(s.stalePriceDays) * 24 * time.Hour
		for _, h := range holdings {
			totalRealized = totalRealized.Add(h.Realized)
			if h.Country == "" {
				summary.MissingCountry++
			}
			if h.Sector == "" {
				summary.MissingSector++
			}
			stale := h.PriceFetchedAt == nil || time.Since(*h.PriceFetchedAt) > staleThreshold
			if stale {
				summary.StaleCount++
			}
			ah := model.AssetHolding{
				AssetID:     h.AssetID,
				Ticker:      h.Ticker,
				Name:        h.Name,
				Currency:    h.Currency,
				Qty:         h.Qty,
				Cost:        h.Cost,
				CostCCY:     h.CostCCY,
				Realized:    h.Realized,
				RealizedCCY: h.RealizedCCY,
				Stale:       stale,
				Closed:      !h.Qty.IsPositive(),
			}
			if h.HasPrice && h.Qty.IsPositive() {
				value := h.Qty.Mul(h.LastClose)
				ah.Value = value
				if factor, ok := series.FxFactor(rates, h.Currency, p.Currency); ok {
					ah.ValuePF = value.Mul(factor)
					totalCost = totalCost.Add(h.Cost)
					totalValue = totalValue.Add(value.Mul(factor))
				} else {
					ah.FXMissing = true
					summary.FXMissingCount++
					summary.FXMissingValue = summary.FXMissingValue.Add(value)
				}
				ah.Unrealized = ah.ValuePF.Sub(h.Cost)
				if h.Cost.IsPositive() {
					ah.ROI = ah.Unrealized.Div(h.Cost).Mul(decimal.NewFromInt(100))
				}
			}
			summary.Holdings = append(summary.Holdings, ah)
		}
		summary.TotalCost = totalCost
		summary.TotalValue = totalValue
		summary.GainLoss = totalValue.Sub(totalCost)
		summary.RealizedGL = totalRealized
		summary.UnrealizedGL = summary.GainLoss
		if totalCost.IsPositive() {
			summary.GainLossPct = summary.GainLoss.Div(totalCost).Mul(decimal.NewFromInt(100))
		}
		return summary, nil
	})
}

func (s *Service) GetPortfolioAllocation(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetAllocation, error) {
	return cached(s.cache, ctx, "allocation", portfolioID.String(), cacheTTLStats, false, func() ([]*model.AssetAllocation, error) {
		p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
		if err != nil {
			return nil, err
		}
		rates, err := series.LoadRates(ctx, s.repos, holdings, p.Currency)
		if err != nil {
			return nil, err
		}

		var allocs []*model.AssetAllocation
		var total decimal.Decimal
		for _, h := range holdings {
			if !h.Qty.IsPositive() {
				continue
			}
			if !h.HasPrice {
				continue
			}
			value := h.Qty.Mul(h.LastClose)
			factor, ok := series.FxFactor(rates, h.Currency, p.Currency)
			if ok {
				value = value.Mul(factor)
				total = total.Add(value)
			}
			allocs = append(allocs, &model.AssetAllocation{
				AssetID:   h.AssetID,
				Ticker:    h.Ticker,
				Name:      h.Name,
				Value:     value,
				FXMissing: !ok,
			})
		}
		for _, a := range allocs {
			if total.IsPositive() && !a.FXMissing {
				a.AllocPct = a.Value.Div(total).Mul(decimal.NewFromInt(100))
			}
		}
		return allocs, nil
	})
}

func (s *Service) GetPortfolioClassAllocation(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioClassAllocation, error) {
	return cached(s.cache, ctx, "allocation-class", portfolioID.String(), cacheTTLStats, false, func() (*model.PortfolioClassAllocation, error) {
		p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
		if err != nil {
			return nil, err
		}
		rates, err := series.LoadRates(ctx, s.repos, holdings, p.Currency)
		if err != nil {
			return nil, err
		}

		byClass := map[string]decimal.Decimal{}
		var total decimal.Decimal
		for _, h := range holdings {
			if !h.Qty.IsPositive() || !h.HasPrice {
				continue
			}
			value := h.Qty.Mul(h.LastClose)
			if factor, ok := series.FxFactor(rates, h.Currency, p.Currency); ok {
				value = value.Mul(factor)
			} else {
				continue
			}
			if !value.IsPositive() {
				continue
			}
			class := h.AssetClass
			if class == "" {
				class = "other"
			}
			byClass[class] = byClass[class].Add(value)
			total = total.Add(value)
		}

		classes := make([]*model.ClassAllocation, 0, len(byClass))
		for class, value := range byClass {
			classes = append(classes, &model.ClassAllocation{
				Class: class,
				Value: value,
			})
		}
		sort.Slice(classes, func(i, j int) bool { return classes[i].Value.GreaterThan(classes[j].Value) })
		for _, c := range classes {
			if total.IsPositive() {
				c.Weight = c.Value.Div(total).Mul(decimal.NewFromInt(100))
			}
		}
		return &model.PortfolioClassAllocation{Currency: p.Currency, Classes: classes}, nil
	})
}

func (s *Service) GetPortfolioPerformance(ctx context.Context, portfolioID uuid.UUID) ([]*model.PortfolioPerformance, error) {
	return cached(s.cache, ctx, "performance", portfolioID.String(), cacheTTLStats, false, func() ([]*model.PortfolioPerformance, error) {
		p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		agg, err := s.repos.Series.FindPortfolioAgg(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		perf := make([]*model.PortfolioPerformance, 0, len(agg))
		for _, pt := range agg {
			perf = append(perf, &model.PortfolioPerformance{Date: pt.Date, Value: pt.MarketValue})
		}
		if len(perf) > s.seriesMaxPoints {
			vals := make([]model.PortfolioPerformance, 0, len(perf))
			for _, p := range perf {
				vals = append(vals, *p)
			}
			vals = series.PortfolioPerformance(vals, s.seriesMaxPoints)
			perf = make([]*model.PortfolioPerformance, 0, len(vals))
			for i := range vals {
				perf = append(perf, &vals[i])
			}
		}
		return perf, nil
	})
}

func (s *Service) GetPortfolioROI(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetROI, error) {
	return cached(s.cache, ctx, "roi", portfolioID.String(), cacheTTLStats, false, func() ([]*model.AssetROI, error) {
		p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
		if err != nil {
			return nil, err
		}
		rates, err := series.LoadRates(ctx, s.repos, holdings, p.Currency)
		if err != nil {
			return nil, err
		}

		var rois []*model.AssetROI
		for _, h := range holdings {
			roi := &model.AssetROI{
				AssetID:       h.AssetID,
				Ticker:        h.Ticker,
				Name:          h.Name,
				TotalInvested: h.Cost,
				Realized:      h.Realized,
			}
			if h.HasPrice {
				value := h.Qty.Mul(h.LastClose)
				if factor, ok := series.FxFactor(rates, h.Currency, p.Currency); ok {
					roi.CurrentValue = value.Mul(factor)
				} else {
					roi.FXMissing = true
				}
			}
			if roi.TotalInvested.IsPositive() {
				roi.ROI = roi.CurrentValue.Sub(roi.TotalInvested).Div(roi.TotalInvested).Mul(decimal.NewFromInt(100))
			}
			rois = append(rois, roi)
		}
		return rois, nil
	})
}

// GetPortfolioHistory returns the running AVCO cost basis, market value and
// realized P&L per date for a portfolio and for each asset in it.
func (s *Service) GetPortfolioHistory(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioPositionHistory, error) {
	return cached(s.cache, ctx, "history", portfolioID.String(), cacheTTLStats, true, func() (*model.PortfolioPositionHistory, error) {
		p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		txs, err := s.repos.Transaction.FindByPortfoliosAsc(ctx, []uuid.UUID{portfolioID})
		if err != nil {
			return nil, err
		}

		txByAsset := map[uuid.UUID][]model.TransactionWithAsset{}
		for _, tx := range txs {
			txByAsset[tx.AssetID] = append(txByAsset[tx.AssetID], tx)
		}
		historyAssets := make([]price.HistoryAsset, 0, len(txByAsset))
		assetPtrs := make([]*model.Asset, 0, len(txByAsset))
		for aid, assetTxs := range txByAsset {
			historyAssets = append(historyAssets, price.HistoryAsset{
				ID:     aid,
				Ticker: assetTxs[0].AssetTicker,
				From:   series.DayOf(assetTxs[0].Date),
			})
			assetPtrs = append(assetPtrs, &model.Asset{ID: aid, Ticker: assetTxs[0].AssetTicker})
		}
		if err := s.fetcher.EnsureHistory(ctx, historyAssets); err != nil {
			log.Warn().Err(err).Msg("history ensure failed")
		}
		if err := s.fetcher.EnsureSplits(ctx, assetPtrs); err != nil {
			log.Warn().Err(err).Msg("splits ensure failed")
		}
		if err := series.Recompute(ctx, s.repos, portfolioID); err != nil {
			log.Warn().Err(err).Str("portfolio_id", portfolioID.String()).Msg("series recompute failed")
		}

		agg, err := s.repos.Series.FindPortfolioAgg(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		assets, err := s.repos.Series.FindPortfolio(ctx, portfolioID)
		if err != nil {
			return nil, err
		}
		agg = series.PositionPoints(agg, s.seriesMaxPoints)
		for i := range assets {
			assets[i].Series = series.PositionPoints(assets[i].Series, s.seriesMaxPoints)
		}
		firstTxByAsset := map[uuid.UUID]time.Time{}
		for aid, assetTxs := range txByAsset {
			firstTxByAsset[aid] = series.DayOf(assetTxs[0].Date)
		}
		assetIDs := make([]uuid.UUID, 0, len(assets))
		for _, a := range assets {
			if id, err := uuid.Parse(a.AssetID); err == nil {
				assetIDs = append(assetIDs, id)
			}
		}
		splitRows, err := s.repos.Split.FindByAssets(ctx, assetIDs)
		if err != nil {
			return nil, err
		}
		splitSeen := map[time.Time]bool{}
		aggSplits := []model.SplitInfo{}
		for _, sp := range splitRows {
			d := series.DayOf(sp.Date)
			if firstTx, ok := firstTxByAsset[sp.AssetID]; ok && !d.After(firstTx) {
				continue
			}
			if !splitSeen[d] {
				splitSeen[d] = true
				aggSplits = append(aggSplits, model.SplitInfo{
					Date:  d,
					Ratio: fmt.Sprintf("%s:%s", sp.Numerator.String(), sp.Denominator.String()),
				})
			}
		}
		sort.Slice(aggSplits, func(i, j int) bool { return aggSplits[i].Date.Before(aggSplits[j].Date) })
		return &model.PortfolioPositionHistory{
			PortfolioID:   portfolioID.String(),
			PortfolioName: p.Name,
			Currency:      p.Currency,
			Series:        agg,
			Assets:        assets,
			Splits:        aggSplits,
		}, nil
	})
}

// ExportPortfolio builds the JSON document for a portfolio: portfolio
// metadata, the assets referenced by its transactions and the full transaction
// history. Assets are referenced by ticker so the file is human-readable and
// editable. Historical prices and split events are asset-level market data and
// are not exported (they are re-synced from the data provider).
func (s *Service) ExportPortfolio(ctx context.Context, portfolioID uuid.UUID, userID uuid.UUID) (*model.PortfolioExport, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	if !s.canAccessPortfolio(ctx, p, userID) {
		return nil, ErrForbidden
	}
	txs, err := s.repos.Transaction.FindByPortfoliosAsc(ctx, []uuid.UUID{portfolioID})
	if err != nil {
		return nil, err
	}

	assetIDs := make([]uuid.UUID, 0, len(txs))
	seen := map[uuid.UUID]bool{}
	for _, tx := range txs {
		if !seen[tx.AssetID] {
			seen[tx.AssetID] = true
			assetIDs = append(assetIDs, tx.AssetID)
		}
	}
	assets, err := s.repos.Asset.FindByIDs(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	assetByID := map[uuid.UUID]*model.Asset{}
	for _, a := range assets {
		assetByID[a.ID] = a
	}

	doc := &model.PortfolioExport{
		Version:    1,
		ExportedAt: time.Now().UTC(),
		Portfolio: model.ExportPortfolio{
			Name:        p.Name,
			Description: p.Description,
			Currency:    p.Currency,
		},
		Assets:       make([]model.ExportAsset, 0, len(assets)),
		Transactions: make([]model.ExportTransaction, 0, len(txs)),
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Ticker < assets[j].Ticker })
	for _, a := range assets {
		doc.Assets = append(doc.Assets, model.ExportAsset{
			Ticker:   a.Ticker,
			Name:     a.Name,
			Type:     a.Type,
			Currency: a.Currency,
			ISIN:     a.ISIN,
		})
	}
	for _, tx := range txs {
		a := assetByID[tx.AssetID]
		if a == nil {
			continue
		}
		doc.Transactions = append(doc.Transactions, model.ExportTransaction{
			Date:        tx.Date,
			Type:        tx.Type,
			AssetTicker: a.Ticker,
			Quantity:    tx.Quantity,
			Price:       tx.Price,
			Fees:        tx.Fees,
			Notes:       tx.Notes,
		})
	}
	return doc, nil
}

// ImportPortfolio restores a portfolio from an exported document. Assets are
// matched by ticker and created if missing. In "new" mode a fresh portfolio is
// created with the given name (falling back to the document's). In "overwrite"
// mode the target portfolio is deleted and recreated from the document. Both
// paths run atomically.
func (s *Service) ImportPortfolio(ctx context.Context, userID uuid.UUID, doc *model.PortfolioExport, mode, name string, targetID *uuid.UUID) (*model.Portfolio, error) {
	if doc == nil || doc.Version != 1 {
		return nil, ErrInvalidInput
	}
	if strings.TrimSpace(doc.Portfolio.Name) == "" {
		return nil, ErrInvalidInput
	}
	if mode == "" {
		mode = "new"
	}
	if mode != "new" && mode != "overwrite" {
		return nil, ErrInvalidInput
	}
	if mode == "overwrite" && targetID == nil {
		return nil, ErrInvalidInput
	}

	var created *model.Portfolio
	err := s.repos.WithTx(ctx, func(rx *repository.Repository) error {
		if mode == "overwrite" {
			target, err := rx.Portfolio.FindByID(ctx, *targetID)
			if err != nil {
				return err
			}
			if !s.canAccessPortfolio(ctx, target, userID) {
				return ErrForbidden
			}
			if err := rx.Portfolio.Delete(ctx, *targetID); err != nil {
				return err
			}
		}

		assetByTicker := map[string]*model.Asset{}
		createAsset := func(ticker string) (*model.Asset, error) {
			if a, ok := assetByTicker[ticker]; ok {
				return a, nil
			}
			a, err := rx.Asset.FindByTicker(ctx, ticker)
			if err != nil {
				return nil, err
			}
			if a == nil {
				a = &model.Asset{Ticker: ticker, Name: ticker, Type: model.AssetTypeStock, Currency: "USD", AssetClass: "equity"}
				for i := range doc.Assets {
					if strings.EqualFold(doc.Assets[i].Ticker, ticker) {
						if doc.Assets[i].Name != "" {
							a.Name = doc.Assets[i].Name
						}
						a.ISIN = doc.Assets[i].ISIN
						if doc.Assets[i].Type != "" {
							a.Type = doc.Assets[i].Type
							a.AssetClass = defaultAssetClassForType(a.Type)
						}
						if doc.Assets[i].Currency != "" {
							a.Currency = doc.Assets[i].Currency
						}
						break
					}
				}
				a, err = rx.Asset.Create(ctx, a)
				if err != nil {
					return nil, err
				}
			}
			assetByTicker[ticker] = a
			return a, nil
		}
		for _, ea := range doc.Assets {
			if _, err := createAsset(strings.TrimSpace(ea.Ticker)); err != nil {
				return err
			}
		}

		pname := strings.TrimSpace(name)
		if pname == "" {
			pname = doc.Portfolio.Name
		}
		p, err := rx.Portfolio.Create(ctx, &model.Portfolio{
			UserID:      userID,
			Name:        pname,
			Description: doc.Portfolio.Description,
			Currency:    doc.Portfolio.Currency,
		})
		if err != nil {
			return err
		}
		created = p

		for _, et := range doc.Transactions {
			if strings.TrimSpace(et.AssetTicker) == "" {
				return fmt.Errorf("%w: transaction without asset_ticker", ErrInvalidInput)
			}
			if et.Type != model.TxBuy && et.Type != model.TxSell && et.Type != model.TxDividend && et.Type != model.TxSplit && et.Type != model.TxFee {
				return fmt.Errorf("%w: invalid transaction type %q", ErrInvalidInput, et.Type)
			}
			a, err := createAsset(strings.TrimSpace(et.AssetTicker))
			if err != nil {
				return err
			}
			if _, err := rx.Transaction.Create(ctx, &model.Transaction{
				PortfolioID: p.ID,
				AssetID:     a.ID,
				Type:        et.Type,
				Quantity:    et.Quantity,
				Price:       et.Price,
				Fees:        et.Fees,
				Date:        et.Date,
				Notes:       et.Notes,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := series.Recompute(ctx, s.repos, created.ID); err != nil {
		log.Warn().Err(err).Str("portfolio_id", created.ID.String()).Msg("series recompute failed")
	}
	s.bumpRev(ctx)
	return created, nil
}

// SyncAssetData refreshes asset-level market data for every asset
// independently of any portfolio: split events are re-checked and the price
// history is brought up to date. It is meant to run once per app load.
func (s *Service) SyncAssetData(ctx context.Context) error {
	assets, err := s.repos.Asset.List(ctx)
	if err != nil {
		return err
	}
	ids := make([]uuid.UUID, 0, len(assets))
	for _, a := range assets {
		ids = append(ids, a.ID)
	}
	if err := s.syncAssets(ctx, ids); err != nil {
		return err
	}
	// Newly fetched prices change every series; recompute so charts and the
	// dashboard reflect the fresh data immediately instead of waiting for the
	// next worker tick.
	if err := series.RecomputeAll(ctx, s.repos); err != nil {
		return err
	}
	// Invalidate cached reads so the fresh prices/series are visible right away.
	s.bumpRev(ctx)
	return nil
}

func (s *Service) syncAssets(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	assets, err := s.repos.Asset.FindByIDs(ctx, ids)
	if err != nil {
		return err
	}
	firstDates, err := s.repos.Transaction.MinDateByAsset(ctx, ids)
	if err != nil {
		return err
	}
	historyAssets := make([]price.HistoryAsset, 0, len(assets))
	for _, a := range assets {
		from := series.DayOf(time.Now())
		if fd, ok := firstDates[a.ID]; ok {
			from = series.DayOf(fd)
		}
		historyAssets = append(historyAssets, price.HistoryAsset{ID: a.ID, Ticker: a.Ticker, From: from, Full: !a.HistoryBackfilled})
	}
	if err := s.fetcher.EnsureSplits(ctx, assets); err != nil {
		return err
	}
	return s.fetcher.EnsureHistory(ctx, historyAssets)
}

// BackfillAssetHistory forces a full price-history backfill for a single
// asset, regardless of stored data. Used by the asset detail page.
func (s *Service) BackfillAssetHistory(ctx context.Context, id uuid.UUID) error {
	asset, err := s.repos.Asset.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAssetNotFound
		}
		return err
	}
	if err := s.fetcher.EnsureHistory(ctx, []price.HistoryAsset{{ID: asset.ID, Ticker: asset.Ticker, Full: true}}); err != nil {
		return err
	}
	s.bumpRev(ctx)
	return nil
}

func (s *Service) syncAssetBackground(assetID uuid.UUID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := s.syncAssets(ctx, []uuid.UUID{assetID}); err != nil {
			log.Warn().Err(err).Str("asset_id", assetID.String()).Msg("asset background sync failed")
			return
		}
		s.bumpRev(ctx)
	}()
}

// GetDashboard returns the consolidated dashboard for a user: performance
// grouped by currency, per-portfolio summaries, assets grouped per portfolio
// and per-portfolio historical series.
func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (*model.Dashboard, error) {
	return cached(s.cache, ctx, "dash", userID.String(), cacheTTLStats, false, func() (*model.Dashboard, error) {
		portfolios, err := s.repos.Portfolio.FindByUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		dash := &model.Dashboard{
			ByCurrency: []model.CurrencyPerformance{},
			Portfolios: []model.PortfolioPerformanceSummary{},
			Assets:     []model.PortfolioAssets{},
			History:    []model.PortfolioHistory{},
		}
		if len(portfolios) == 0 {
			return dash, nil
		}

		ids := make([]uuid.UUID, 0, len(portfolios))
		for _, p := range portfolios {
			ids = append(ids, p.ID)
		}
		holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, ids)
		if err != nil {
			return nil, err
		}
		rates, err := series.LoadRates(ctx, s.repos, holdings, "")
		if err != nil {
			return nil, err
		}

		byID := make(map[uuid.UUID]*model.Portfolio, len(portfolios))
		for _, p := range portfolios {
			byID[p.ID] = p
		}

		byCurrency := map[string]*model.CurrencyPerformance{}
		for _, h := range holdings {
			p := byID[mustUUID(h.PortfolioID)]
			if p == nil {
				continue
			}
			cp := byCurrency[h.Currency]
			if cp == nil {
				cp = &model.CurrencyPerformance{Currency: h.Currency}
				byCurrency[h.Currency] = cp
			}
			cp.Invested = cp.Invested.Add(h.CostCCY)
			cp.Realized = cp.Realized.Add(h.RealizedCCY)
			if h.HasPrice {
				cp.Value = cp.Value.Add(h.Qty.Mul(h.LastClose))
			}
		}
		for _, cp := range byCurrency {
			cp.GainLoss = cp.Value.Sub(cp.Invested)
			if cp.Invested.IsPositive() {
				cp.GainLossPct = cp.GainLoss.Div(cp.Invested).Mul(decimal.NewFromInt(100))
			}
			dash.ByCurrency = append(dash.ByCurrency, *cp)
		}

		assetsByPF := map[uuid.UUID][]model.AssetPerformance{}
		for _, h := range holdings {
			pfID := mustUUID(h.PortfolioID)
			p := byID[pfID]
			ap := model.AssetPerformance{
				AssetID:    h.AssetID,
				Ticker:     h.Ticker,
				Name:       h.Name,
				Currency:   h.Currency,
				Qty:        h.Qty,
				Invested:   h.CostCCY,
				Realized:   h.RealizedCCY,
				RealizedPF: h.Realized,
			}
			if h.HasPrice && h.Qty.IsPositive() {
				value := h.Qty.Mul(h.LastClose)
				ap.Value = value
				ap.GainLoss = value.Sub(h.CostCCY)
				if h.CostCCY.IsPositive() {
					ap.ROI = ap.GainLoss.Div(h.CostCCY).Mul(decimal.NewFromInt(100))
				}
				if p != nil {
					if factor, ok := series.FxFactor(rates, h.Currency, p.Currency); ok {
						ap.ValuePF = value.Mul(factor)
					} else {
						ap.FXMissing = true
					}
				}
			}
			assetsByPF[pfID] = append(assetsByPF[pfID], ap)
		}

		for _, p := range portfolios {
			ps := &model.PortfolioPerformanceSummary{
				PortfolioID:   p.ID.String(),
				PortfolioName: p.Name,
				Currency:      p.Currency,
				AssetCount:    len(assetsByPF[p.ID]),
			}
			for _, ap := range assetsByPF[p.ID] {
				ps.Invested = ps.Invested.Add(ap.Invested)
				ps.Value = ps.Value.Add(ap.ValuePF)
				ps.RealizedGL = ps.RealizedGL.Add(ap.RealizedPF)
				if ap.FXMissing {
					ps.FXMissing++
				}
			}
			ps.GainLoss = ps.Value.Sub(ps.Invested)
			if ps.Invested.IsPositive() {
				ps.GainLossPct = ps.GainLoss.Div(ps.Invested).Mul(decimal.NewFromInt(100))
			}
			dash.Portfolios = append(dash.Portfolios, *ps)
			dash.Assets = append(dash.Assets, model.PortfolioAssets{
				PortfolioID:   p.ID.String(),
				PortfolioName: p.Name,
				Currency:      p.Currency,
				Assets:        assetsByPF[p.ID],
			})

			agg, err := s.repos.Series.FindPortfolioAgg(ctx, p.ID)
			if err != nil {
				return nil, err
			}
			seriesVals := make([]model.PortfolioPerformance, 0, len(agg))
			for _, pt := range agg {
				seriesVals = append(seriesVals, model.PortfolioPerformance{Date: pt.Date, Value: pt.MarketValue})
			}
			seriesVals = series.PortfolioPerformance(seriesVals, s.seriesMaxPoints)
			dash.History = append(dash.History, model.PortfolioHistory{
				PortfolioID:   p.ID.String(),
				PortfolioName: p.Name,
				Currency:      p.Currency,
				Series:        seriesVals,
			})
		}

		return dash, nil
	})
}

func mustUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

func (s *Service) GetPrices(ctx context.Context, assetID uuid.UUID, full bool) ([]*model.Price, error) {
	kind := "prices"
	if full {
		kind = "prices_full"
	}
	return cached(s.cache, ctx, kind, assetID.String(), cacheTTLPrices, false, func() ([]*model.Price, error) {
		prices, err := s.repos.Price.FindByAsset(ctx, assetID)
		if err != nil {
			return nil, err
		}
		vals := make([]model.Price, 0, len(prices))
		for i := len(prices) - 1; i >= 0; i-- {
			vals = append(vals, *prices[i])
		}
		if !full {
			vals = series.Prices(vals, s.seriesMaxPoints)
		}
		out := make([]*model.Price, 0, len(vals))
		for i := len(vals) - 1; i >= 0; i-- {
			out = append(out, &vals[i])
		}
		return out, nil
	})
}

func (s *Service) canAccessPortfolio(ctx context.Context, portfolio *model.Portfolio, userID uuid.UUID) bool {
	if portfolio.UserID == userID {
		return true
	}
	return false // TODO: check portfolio_shares
}
