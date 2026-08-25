package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PortfolioSummary struct {
	PortfolioID   string          `json:"portfolio_id"`
	PortfolioName string          `json:"portfolio_name"`
	TotalValue    decimal.Decimal `json:"total_value"`
	TotalCost     decimal.Decimal `json:"total_cost"`
	GainLoss      decimal.Decimal `json:"gain_loss"`
	GainLossPct   decimal.Decimal `json:"gain_loss_pct"`
	RealizedGL     decimal.Decimal `json:"realized_gl"`
	UnrealizedGL   decimal.Decimal `json:"unrealized_gl"`
	AssetCount     int             `json:"asset_count"`
	MissingCountry int             `json:"missing_country"`
	MissingCategory int            `json:"missing_category"`
	StaleCount     int             `json:"stale_count"`
	Holdings       []AssetHolding  `json:"holdings"`
}

type AssetAllocation struct {
	AssetID  string          `json:"asset_id"`
	Ticker   string          `json:"ticker"`
	Name     string          `json:"name"`
	Value    decimal.Decimal `json:"value"`
	AllocPct decimal.Decimal `json:"alloc_pct"`
}

type PortfolioPerformance struct {
	Date  time.Time       `json:"date"`
	Value decimal.Decimal `json:"value"`
}

type AssetROI struct {
	AssetID       string          `json:"asset_id"`
	Ticker        string          `json:"ticker"`
	Name          string          `json:"name"`
	ROI           decimal.Decimal `json:"roi"`
	TotalInvested decimal.Decimal `json:"total_invested"`
	CurrentValue  decimal.Decimal `json:"current_value"`
	Realized      decimal.Decimal `json:"realized"`
	FXMissing     bool            `json:"fx_missing"`
}

// Holding is one position of an asset within a portfolio.
type Holding struct {
	PortfolioID string          `json:"-"`
	AssetID     string          `json:"asset_id"`
	Ticker      string          `json:"ticker"`
	Name        string          `json:"name"`
	Currency    string          `json:"currency"`
	Qty         decimal.Decimal `json:"qty"`
	Cost        decimal.Decimal `json:"cost"`     // cost basis in portfolio currency
	CostCCY     decimal.Decimal `json:"cost_ccy"` // cost basis in asset currency
	Realized    decimal.Decimal `json:"realized"` // realized P&L in portfolio currency
	RealizedCCY decimal.Decimal `json:"realized_ccy"`
	AvgCost         decimal.Decimal `json:"avg_cost"`
	LastClose       decimal.Decimal `json:"last_close"` // latest close in asset currency
	HasPrice        bool            `json:"has_price"`
	Country         string          `json:"country"`
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	PriceFetchedAt  *time.Time      `json:"price_fetched_at,omitempty"`
}

type AssetHolding struct {
	AssetID     string          `json:"asset_id"`
	Ticker      string          `json:"ticker"`
	Name        string          `json:"name"`
	Currency    string          `json:"currency"`
	Qty         decimal.Decimal `json:"qty"`
	Cost        decimal.Decimal `json:"cost"`
	CostCCY     decimal.Decimal `json:"cost_ccy"`
	Value       decimal.Decimal `json:"value"`
	ValuePF     decimal.Decimal `json:"value_pf"`
	Realized    decimal.Decimal `json:"realized"`
	RealizedCCY decimal.Decimal `json:"realized_ccy"`
	Unrealized  decimal.Decimal `json:"unrealized"`
	ROI         decimal.Decimal `json:"roi"`
	FXMissing   bool            `json:"fx_missing"`
	Stale       bool            `json:"stale"`
	Closed      bool            `json:"closed"`
}

type CurrencyPerformance struct {
	Currency    string          `json:"currency"`
	Invested    decimal.Decimal `json:"invested"`
	Value       decimal.Decimal `json:"value"`
	GainLoss    decimal.Decimal `json:"gain_loss"`
	GainLossPct decimal.Decimal `json:"gain_loss_pct"`
	Realized    decimal.Decimal `json:"realized"`
}

type PortfolioPerformanceSummary struct {
	PortfolioID   string          `json:"portfolio_id"`
	PortfolioName string          `json:"portfolio_name"`
	Currency      string          `json:"currency"`
	Invested      decimal.Decimal `json:"invested"`
	Value         decimal.Decimal `json:"value"`
	GainLoss      decimal.Decimal `json:"gain_loss"`
	GainLossPct   decimal.Decimal `json:"gain_loss_pct"`
	RealizedGL    decimal.Decimal `json:"realized_gl"`
	AssetCount    int             `json:"asset_count"`
	FXMissing     int             `json:"fx_missing"`
}

type AssetPerformance struct {
	AssetID    string          `json:"asset_id"`
	Ticker     string          `json:"ticker"`
	Name       string          `json:"name"`
	Currency   string          `json:"currency"`
	Qty        decimal.Decimal `json:"qty"`
	Invested   decimal.Decimal `json:"invested"` // in asset currency
	Value      decimal.Decimal `json:"value"`    // in asset currency
	GainLoss   decimal.Decimal `json:"gain_loss"`
	ROI        decimal.Decimal `json:"roi"`
	FXMissing  bool            `json:"fx_missing"`
	ValuePF    decimal.Decimal `json:"value_pf"` // consolidated in portfolio currency
	Realized   decimal.Decimal `json:"realized"` // in asset currency
	RealizedPF decimal.Decimal `json:"realized_pf"`
}

type PortfolioAssets struct {
	PortfolioID   string             `json:"portfolio_id"`
	PortfolioName string             `json:"portfolio_name"`
	Currency      string             `json:"currency"`
	Assets        []AssetPerformance `json:"assets"`
}

type PositionPoint struct {
	Date        time.Time       `json:"date"`
	Qty         decimal.Decimal `json:"qty"`
	CostBasis   decimal.Decimal `json:"cost_basis"`
	MarketValue decimal.Decimal `json:"market_value"`
	Realized    decimal.Decimal `json:"realized"`
}

type SplitInfo struct {
	Date  time.Time `json:"date"`
	Ratio string    `json:"ratio"`
}

type AssetPositionSeries struct {
	AssetID  string          `json:"asset_id"`
	Ticker   string          `json:"ticker"`
	Name     string          `json:"name"`
	Currency string          `json:"currency"`
	Series   []PositionPoint `json:"series"`
	Splits   []SplitInfo     `json:"splits"`
}

type PortfolioPositionHistory struct {
	PortfolioID   string                `json:"portfolio_id"`
	PortfolioName string                `json:"portfolio_name"`
	Currency      string                `json:"currency"`
	Series        []PositionPoint       `json:"series"`
	Assets        []AssetPositionSeries `json:"assets"`
	Splits        []SplitInfo           `json:"splits"`
}

type PortfolioHistory struct {
	PortfolioID   string                 `json:"portfolio_id"`
	PortfolioName string                 `json:"portfolio_name"`
	Currency      string                 `json:"currency"`
	Series        []PortfolioPerformance `json:"series"`
}

type Dashboard struct {
	ByCurrency []CurrencyPerformance         `json:"by_currency"`
	Portfolios []PortfolioPerformanceSummary `json:"portfolios"`
	Assets     []PortfolioAssets             `json:"assets"`
	History    []PortfolioHistory            `json:"history"`
}
