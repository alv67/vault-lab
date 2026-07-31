package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/amelamela/vault-lab/internal/auth"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/price"
	"github.com/amelamela/vault-lab/internal/repository"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already registered")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrAssetInUse         = errors.New("asset is used in transactions")
)

type Service struct {
	repos          *repository.Repository
	jwtAuth        *auth.JWTAuth
	fetcher        *price.YahooFetcher
	lookupCacheTTL time.Duration
}

func New(repos *repository.Repository, jwtAuth *auth.JWTAuth, fetcher *price.YahooFetcher, lookupCacheTTL time.Duration) *Service {
	return &Service{repos: repos, jwtAuth: jwtAuth, fetcher: fetcher, lookupCacheTTL: lookupCacheTTL}
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

func (s *Service) CreateAsset(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	return s.repos.Asset.Create(ctx, asset)
}

func (s *Service) GetAsset(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	return s.repos.Asset.FindByID(ctx, id)
}

func (s *Service) SearchAssets(ctx context.Context, query string) ([]*model.Asset, error) {
	return s.repos.Asset.Search(ctx, query)
}

func (s *Service) GetAssetMeta(ctx context.Context, ticker string) (*price.AssetMeta, error) {
	key := "meta:" + strings.ToUpper(strings.TrimSpace(ticker))

	if cached, err := s.repos.Lookup.Get(ctx, key, s.lookupCacheTTL); err == nil {
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
		if err := s.repos.Lookup.Set(ctx, key, data); err != nil {
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

	if cached, err := s.repos.Lookup.Get(ctx, key, s.lookupCacheTTL); err == nil {
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
		if err := s.repos.Lookup.Set(ctx, key, data); err != nil {
			log.Warn().Err(err).Str("query", key).Msg("failed to cache lookup results")
		}
	}

	return results, nil
}

// RefreshPrices refreshes stale stored closes for the given portfolio (or all
// assets when portfolioID is nil) plus the USD->X FX rates, hitting Yahoo
// Finance only when needed.
func (s *Service) RefreshPrices(ctx context.Context, portfolioID *uuid.UUID) ([]string, error) {
	var refreshed []string
	var err error
	if portfolioID != nil {
		refreshed, err = s.fetcher.RefreshStaleForPortfolio(ctx, *portfolioID)
	} else {
		var assets []*model.Asset
		assets, err = s.repos.Asset.List(ctx)
		if err != nil {
			return nil, err
		}
		refreshed, err = s.fetcher.RefreshStale(ctx, assets)
	}
	if err != nil {
		return refreshed, err
	}
	if err := s.fetcher.RefreshFX(ctx); err != nil {
		return refreshed, err
	}
	return refreshed, nil
}

func (s *Service) ListAssets(ctx context.Context) ([]*model.Asset, error) {
	return s.repos.Asset.List(ctx)
}

func (s *Service) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, currency string) (*model.Portfolio, error) {
	p := &model.Portfolio{
		UserID:      userID,
		Name:        name,
		Description: description,
		Currency:    currency,
	}
	return s.repos.Portfolio.Create(ctx, p)
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
	return s.repos.Portfolio.Update(ctx, p)
}

func (s *Service) DeletePortfolio(ctx context.Context, id uuid.UUID) error {
	return s.repos.Portfolio.Delete(ctx, id)
}

func (s *Service) AddTransaction(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	return s.repos.Transaction.Create(ctx, tx)
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
	if tx.ExchangeRate.IsZero() {
		tx.ExchangeRate = decimal.NewFromInt(1)
	}
	return s.repos.Transaction.Update(ctx, tx)
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
	return s.repos.Transaction.Delete(ctx, id)
}

func (s *Service) GetPortfolioSummary(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioSummary, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
	if err != nil {
		return nil, err
	}
	rates, err := s.loadRates(ctx, holdings)
	if err != nil {
		return nil, err
	}

	summary := &model.PortfolioSummary{
		PortfolioID:   portfolioID.String(),
		PortfolioName: p.Name,
		AssetCount:    len(holdings),
	}
	var totalValue, totalCost decimal.Decimal
	for _, h := range holdings {
		totalCost = totalCost.Add(h.Cost)
		if !h.HasPrice {
			continue
		}
		value := h.Qty.Mul(h.LastClose)
		if factor, ok := fxFactor(rates, h.Currency, p.Currency); ok {
			totalValue = totalValue.Add(value.Mul(factor))
		}
	}
	summary.TotalCost = totalCost
	summary.TotalValue = totalValue
	summary.GainLoss = totalValue.Sub(totalCost)
	if totalCost.IsPositive() {
		summary.GainLossPct = summary.GainLoss.Div(totalCost).Mul(decimal.NewFromInt(100))
	}
	return summary, nil
}

func (s *Service) GetPortfolioAllocation(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetAllocation, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
	if err != nil {
		return nil, err
	}
	rates, err := s.loadRates(ctx, holdings)
	if err != nil {
		return nil, err
	}

	var allocs []*model.AssetAllocation
	var total decimal.Decimal
	for _, h := range holdings {
		if !h.HasPrice {
			continue
		}
		value := h.Qty.Mul(h.LastClose)
		factor, ok := fxFactor(rates, h.Currency, p.Currency)
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
}

func (s *Service) GetPortfolioPerformance(ctx context.Context, portfolioID uuid.UUID) ([]*model.PortfolioPerformance, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	return s.repos.Portfolio.PerformanceSeries(ctx, portfolioID, p.Currency)
}

func (s *Service) GetPortfolioROI(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetROI, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
	if err != nil {
		return nil, err
	}
	rates, err := s.loadRates(ctx, holdings)
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
		}
		if h.HasPrice {
			value := h.Qty.Mul(h.LastClose)
			if factor, ok := fxFactor(rates, h.Currency, p.Currency); ok {
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
}

// GetDashboard returns the consolidated dashboard for a user: performance
// grouped by currency, per-portfolio summaries, assets grouped per portfolio
// and per-portfolio historical series.
func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (*model.Dashboard, error) {
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
	rates, err := s.loadRates(ctx, holdings)
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
			AssetID:  h.AssetID,
			Ticker:   h.Ticker,
			Name:     h.Name,
			Currency: h.Currency,
			Qty:      h.Qty,
			Invested: h.CostCCY,
		}
		if h.HasPrice {
			value := h.Qty.Mul(h.LastClose)
			ap.Value = value
			ap.GainLoss = value.Sub(h.CostCCY)
			if h.CostCCY.IsPositive() {
				ap.ROI = ap.GainLoss.Div(h.CostCCY).Mul(decimal.NewFromInt(100))
			}
			if p != nil {
				if factor, ok := fxFactor(rates, h.Currency, p.Currency); ok {
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

		series, err := s.repos.Portfolio.PerformanceSeries(ctx, p.ID, p.Currency)
		if err != nil {
			return nil, err
		}
		seriesVals := make([]model.PortfolioPerformance, 0, len(series))
		for _, sp := range series {
			seriesVals = append(seriesVals, *sp)
		}
		dash.History = append(dash.History, model.PortfolioHistory{
			PortfolioID:   p.ID.String(),
			PortfolioName: p.Name,
			Currency:      p.Currency,
			Series:        seriesVals,
		})
	}

	return dash, nil
}

// loadRates loads USD->X rates for every currency appearing in the holdings
// plus the portfolios' reference currencies.
func (s *Service) loadRates(ctx context.Context, holdings []*model.Holding) (map[string]decimal.Decimal, error) {
	quotes := map[string]bool{}
	for _, h := range holdings {
		quotes[h.Currency] = true
	}
	// Portfolio currencies are not directly in holdings; fetch all needed from
	// the portfolios involved via the caller. For simplicity include USD.
	quotes["USD"] = true
	list := make([]string, 0, len(quotes))
	for c := range quotes {
		if c != "" {
			list = append(list, c)
		}
	}
	return s.repos.FX.LatestByQuotes(ctx, list)
}

// fxFactor returns the factor to convert an amount from currency `from` to
// currency `to`, computed via USD cross rates. Returns ok=false when a needed
// rate is missing.
func fxFactor(rates map[string]decimal.Decimal, from, to string) (decimal.Decimal, bool) {
	if from == to {
		return decimal.NewFromInt(1), true
	}
	usdToFrom, okFrom := usdRate(rates, from)
	if !okFrom {
		return decimal.Zero, false
	}
	usdToTo, okTo := usdRate(rates, to)
	if !okTo {
		return decimal.Zero, false
	}
	if usdToFrom.IsZero() {
		return decimal.Zero, false
	}
	return usdToTo.Div(usdToFrom), true
}

func usdRate(rates map[string]decimal.Decimal, currency string) (decimal.Decimal, bool) {
	if currency == "USD" {
		return decimal.NewFromInt(1), true
	}
	rate, ok := rates[currency]
	return rate, ok
}

func mustUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

func (s *Service) GetPrices(ctx context.Context, assetID uuid.UUID) ([]*model.Price, error) {
	return s.repos.Price.FindByAsset(ctx, assetID)
}

func (s *Service) canAccessPortfolio(ctx context.Context, portfolio *model.Portfolio, userID uuid.UUID) bool {
	if portfolio.UserID == userID {
		return true
	}
	return false // TODO: check portfolio_shares
}
