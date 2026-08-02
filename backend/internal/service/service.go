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
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/position"
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
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidInput       = errors.New("invalid input")
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
		if factor, ok := fxFactor(rates, h.Currency, p.Currency); ok {
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
			if factor, ok := fxFactor(rates, h.Currency, p.Currency); ok {
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
		if !h.Qty.IsPositive() {
			continue
		}
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
	agg, _, err := s.portfolioSeries(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	perf := make([]*model.PortfolioPerformance, 0, len(agg))
	for _, pt := range agg {
		perf = append(perf, &model.PortfolioPerformance{Date: pt.Date, Value: pt.MarketValue})
	}
	return perf, nil
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
			Realized:      h.Realized,
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

// GetPortfolioHistory returns the running AVCO cost basis, market value and
// realized P&L per date for a portfolio and for each asset in it.
func (s *Service) GetPortfolioHistory(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioPositionHistory, error) {
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
			From:   dayOf(assetTxs[0].Date),
		})
		assetPtrs = append(assetPtrs, &model.Asset{ID: aid, Ticker: assetTxs[0].AssetTicker})
	}
	if err := s.fetcher.EnsureHistory(ctx, historyAssets); err != nil {
		log.Warn().Err(err).Msg("history ensure failed")
	}
	if err := s.fetcher.EnsureSplits(ctx, assetPtrs); err != nil {
		log.Warn().Err(err).Msg("splits ensure failed")
	}

	agg, assets, err := s.portfolioSeries(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	splitSeen := map[time.Time]bool{}
	aggSplits := []model.SplitInfo{}
	for _, a := range assets {
		for _, sp := range a.Splits {
			if !splitSeen[sp.Date] {
				splitSeen[sp.Date] = true
				aggSplits = append(aggSplits, sp)
			}
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
}

// portfolioSeries builds the daily position series for a portfolio using the
// running AVCO engine, split adjustments and split-adjusted prices. It does
// not trigger any Yahoo fetches.
func (s *Service) portfolioSeries(ctx context.Context, portfolioID uuid.UUID) ([]model.PositionPoint, []model.AssetPositionSeries, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, portfolioID)
	if err != nil {
		return nil, nil, err
	}
	txs, err := s.repos.Transaction.FindByPortfoliosAsc(ctx, []uuid.UUID{portfolioID})
	if err != nil {
		return nil, nil, err
	}
	holdings, err := s.repos.Portfolio.HoldingsDetailed(ctx, []uuid.UUID{portfolioID})
	if err != nil {
		return nil, nil, err
	}
	rates, err := s.loadRates(ctx, holdings)
	if err != nil {
		return nil, nil, err
	}
	prices, err := s.repos.Price.FindForPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, nil, err
	}

	txByAsset := map[uuid.UUID][]model.TransactionWithAsset{}
	for _, tx := range txs {
		txByAsset[tx.AssetID] = append(txByAsset[tx.AssetID], tx)
	}

	assetIDs := make([]uuid.UUID, 0, len(txByAsset))
	for id := range txByAsset {
		assetIDs = append(assetIDs, id)
	}

	splitRows, err := s.repos.Split.FindByAssets(ctx, assetIDs)
	if err != nil {
		return nil, nil, err
	}
	splitsByAsset := map[uuid.UUID][]model.Split{}
	for _, sp := range splitRows {
		splitsByAsset[sp.AssetID] = append(splitsByAsset[sp.AssetID], *sp)
	}
	for _, list := range splitsByAsset {
		sort.Slice(list, func(i, j int) bool { return list[i].Date.Before(list[j].Date) })
	}

	assetCurrencies := map[uuid.UUID]string{}
	for _, h := range holdings {
		if h.PortfolioID == portfolioID.String() {
			assetCurrencies[mustUUID(h.AssetID)] = h.Currency
		}
	}

	priceByAsset := map[uuid.UUID]map[time.Time]decimal.Decimal{}
	priceDatesByAsset := map[uuid.UUID][]time.Time{}
	for _, pr := range prices {
		d := dayOf(pr.Date)
		if priceByAsset[pr.AssetID] == nil {
			priceByAsset[pr.AssetID] = map[time.Time]decimal.Decimal{}
		}
		priceByAsset[pr.AssetID][d] = pr.Close
		priceDatesByAsset[pr.AssetID] = append(priceDatesByAsset[pr.AssetID], d)
	}
	for _, ds := range priceDatesByAsset {
		sort.Slice(ds, func(i, j int) bool { return ds[i].Before(ds[j]) })
	}

	var dates []time.Time
	if len(txs) > 0 {
		firstDate := dayOf(txs[0].Date)
		lastDate := dayOf(time.Now())
		dates = make([]time.Time, 0, int(lastDate.Sub(firstDate)/(24*time.Hour))+1)
		for d := firstDate; !d.After(lastDate); d = d.Add(24 * time.Hour) {
			dates = append(dates, d)
		}
	}

	agg := make([]model.PositionPoint, len(dates))
	assets := []model.AssetPositionSeries{}

	sort.Slice(assetIDs, func(i, j int) bool {
		return txByAsset[assetIDs[i]][0].AssetTicker < txByAsset[assetIDs[j]][0].AssetTicker
	})

	for _, aid := range assetIDs {
		assetTxs := txByAsset[aid]
		currency := assetCurrencies[aid]
		factor, hasFX := fxFactor(rates, currency, p.Currency)
		st := &position.State{}
		pos := 0
		series := make([]model.PositionPoint, 0, len(dates))

		priceDates := priceDatesByAsset[aid]
		pricePos := 0

		splits := splitsByAsset[aid]
		firstTx := dayOf(assetTxs[0].Date)
		splitPos := 0
		for splitPos < len(splits) && !dayOf(splits[splitPos].Date).After(firstTx) {
			splitPos++
		}
		rawFactor := decimal.NewFromInt(1)
		for _, sp := range splits[splitPos:] {
			rawFactor = rawFactor.Mul(sp.Numerator.Div(sp.Denominator))
		}
		splitInfo := make([]model.SplitInfo, 0, len(splits)-splitPos)
		for _, sp := range splits[splitPos:] {
			splitInfo = append(splitInfo, model.SplitInfo{
				Date:  dayOf(sp.Date),
				Ratio: fmt.Sprintf("%s:%s", sp.Numerator.String(), sp.Denominator.String()),
			})
		}
		for i, d := range dates {
			for pos < len(assetTxs) && !dayOf(assetTxs[pos].Date).After(d) {
				position.Apply(st, assetTxs[pos])
				pos++
			}
			for splitPos < len(splits) && !dayOf(splits[splitPos].Date).After(d) {
				ratio := splits[splitPos].Numerator.Div(splits[splitPos].Denominator)
				position.Apply(st, model.TransactionWithAsset{
					PortfolioID: portfolioID,
					AssetID:     aid,
					Type:        model.TxSplit,
					Quantity:    ratio,
					Date:        splits[splitPos].Date,
					CreatedAt:   splits[splitPos].Date,
				})
				rawFactor = rawFactor.Div(ratio)
				splitPos++
			}
			for pricePos < len(priceDates) && !priceDates[pricePos].After(d) {
				pricePos++
			}
			mv := decimal.Zero
			if hasFX && st.Qty.IsPositive() && pricePos > 0 {
				mv = st.Qty.Mul(priceByAsset[aid][priceDates[pricePos-1]]).Mul(rawFactor).Mul(factor)
			}
			pt := model.PositionPoint{
				Date:        d,
				Qty:         st.Qty,
				CostBasis:   st.Cost,
				MarketValue: mv,
				Realized:    st.Realized,
			}
			if !d.Before(firstTx) {
				series = append(series, pt)
			}
			agg[i].Date = d
			agg[i].CostBasis = agg[i].CostBasis.Add(pt.CostBasis)
			agg[i].MarketValue = agg[i].MarketValue.Add(pt.MarketValue)
			agg[i].Realized = agg[i].Realized.Add(pt.Realized)
		}
		assets = append(assets, model.AssetPositionSeries{
			AssetID:  aid.String(),
			Ticker:   assetTxs[0].AssetTicker,
			Name:     assetTxs[0].AssetName,
			Currency: currency,
			Series:   series,
			Splits:   splitInfo,
		})
	}

	return agg, assets, nil
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

		agg, _, err := s.portfolioSeries(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		seriesVals := make([]model.PortfolioPerformance, 0, len(agg))
		for _, pt := range agg {
			seriesVals = append(seriesVals, model.PortfolioPerformance{Date: pt.Date, Value: pt.MarketValue})
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

func dayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
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
