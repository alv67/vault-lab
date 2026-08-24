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
	"golang.org/x/crypto/bcrypt"

	"github.com/amelamela/vault-lab/internal/auth"
	"github.com/amelamela/vault-lab/internal/cache"
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
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidInput       = errors.New("invalid input")
	ErrCurrencyExists     = errors.New("currency already in whitelist")
	ErrCurrencyInUse      = errors.New("currency is used by assets or portfolios")
	ErrCurrencyProtected  = errors.New("currency cannot be removed")
	ErrCurrencyNotManaged = errors.New("currency conversion not available")
)

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
	Health          *HealthService
}

func New(repos *repository.Repository, jwtAuth *auth.JWTAuth, fetcher *price.YahooFetcher, lookupCacheTTL time.Duration, c *cache.Cache, seriesMaxPoints int, health *HealthService) *Service {
	if seriesMaxPoints <= 0 {
		seriesMaxPoints = 500
	}
	return &Service{repos: repos, jwtAuth: jwtAuth, fetcher: fetcher, lookupCacheTTL: lookupCacheTTL, cache: c, seriesMaxPoints: seriesMaxPoints, Health: health}
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
		rates, err := series.LoadRates(ctx, s.repos, holdings)
		if err != nil {
			return nil, err
		}

		summary := &model.PortfolioSummary{
			PortfolioID:   portfolioID.String(),
			PortfolioName: p.Name,
			AssetCount:    len(holdings),
		}
		var totalValue, totalCost, totalRealized decimal.Decimal
		for _, h := range holdings {
			totalRealized = totalRealized.Add(h.Realized)
			if !h.Qty.IsPositive() {
				continue
			}
			if !h.HasPrice {
				continue
			}
			totalCost = totalCost.Add(h.Cost)
			value := h.Qty.Mul(h.LastClose)
			if factor, ok := series.FxFactor(rates, h.Currency, p.Currency); ok {
				totalValue = totalValue.Add(value.Mul(factor))
			}
		}
		summary.TotalCost = totalCost
		summary.TotalValue = totalValue
		summary.GainLoss = totalValue.Sub(totalCost)
		summary.RealizedGL = totalRealized
		summary.UnrealizedGL = summary.GainLoss
		if totalCost.IsPositive() {
			summary.GainLossPct = summary.GainLoss.Div(totalCost).Mul(decimal.NewFromInt(100))
		}
		summary.Holdings = make([]model.AssetHolding, 0, len(holdings))
		for _, h := range holdings {
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
				Closed:      !h.Qty.IsPositive(),
			}
			if h.HasPrice && h.Qty.IsPositive() {
				value := h.Qty.Mul(h.LastClose)
				ah.Value = value
				if factor, ok := series.FxFactor(rates, h.Currency, p.Currency); ok {
					ah.ValuePF = value.Mul(factor)
				} else {
					ah.FXMissing = true
				}
				ah.Unrealized = ah.ValuePF.Sub(h.Cost)
				if h.Cost.IsPositive() {
					ah.ROI = ah.Unrealized.Div(h.Cost).Mul(decimal.NewFromInt(100))
				}
			}
			summary.Holdings = append(summary.Holdings, ah)
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
		rates, err := series.LoadRates(ctx, s.repos, holdings)
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
			if !ok {
				continue
			}
			value = value.Mul(factor)
			total = total.Add(value)
			allocs = append(allocs, &model.AssetAllocation{
				AssetID: h.AssetID,
				Ticker:  h.Ticker,
				Name:    h.Name,
				Value:   value,
			})
		}
		for _, a := range allocs {
			if total.IsPositive() {
				a.AllocPct = a.Value.Div(total).Mul(decimal.NewFromInt(100))
			}
		}
		return allocs, nil
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
		rates, err := series.LoadRates(ctx, s.repos, holdings)
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
				a = &model.Asset{Ticker: ticker, Name: ticker, Type: model.AssetTypeStock, Currency: "USD"}
				for i := range doc.Assets {
					if strings.EqualFold(doc.Assets[i].Ticker, ticker) {
						if doc.Assets[i].Name != "" {
							a.Name = doc.Assets[i].Name
						}
						a.ISIN = doc.Assets[i].ISIN
						if doc.Assets[i].Type != "" {
							a.Type = doc.Assets[i].Type
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
	return series.RecomputeAll(ctx, s.repos)
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
		historyAssets = append(historyAssets, price.HistoryAsset{ID: a.ID, Ticker: a.Ticker, From: from})
	}
	if err := s.fetcher.EnsureSplits(ctx, assets); err != nil {
		return err
	}
	return s.fetcher.EnsureHistory(ctx, historyAssets)
}

func (s *Service) syncAssetBackground(assetID uuid.UUID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := s.syncAssets(ctx, []uuid.UUID{assetID}); err != nil {
			log.Warn().Err(err).Str("asset_id", assetID.String()).Msg("asset background sync failed")
		}
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
		rates, err := series.LoadRates(ctx, s.repos, holdings)
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

func (s *Service) GetPrices(ctx context.Context, assetID uuid.UUID) ([]*model.Price, error) {
	return cached(s.cache, ctx, "prices", assetID.String(), cacheTTLPrices, false, func() ([]*model.Price, error) {
		prices, err := s.repos.Price.FindByAsset(ctx, assetID)
		if err != nil {
			return nil, err
		}
		vals := make([]model.Price, 0, len(prices))
		for i := len(prices) - 1; i >= 0; i-- {
			vals = append(vals, *prices[i])
		}
		vals = series.Prices(vals, s.seriesMaxPoints)
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
