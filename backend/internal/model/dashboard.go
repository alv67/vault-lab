package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type PortfolioSummary struct {
	PortfolioID    string          `json:"portfolio_id"`
	PortfolioName  string          `json:"portfolio_name"`
	TotalValue     decimal.Decimal `json:"total_value"`
	TotalCost      decimal.Decimal `json:"total_cost"`
	GainLoss       decimal.Decimal `json:"gain_loss"`
	GainLossPct    decimal.Decimal `json:"gain_loss_pct"`
	RealizedGL     decimal.Decimal `json:"realized_gl"`
	UnrealizedGL   decimal.Decimal `json:"unrealized_gl"`
	AssetCount     int             `json:"asset_count"`
}

type AssetAllocation struct {
	AssetID   string          `json:"asset_id"`
	Ticker    string          `json:"ticker"`
	Name      string          `json:"name"`
	Value     decimal.Decimal `json:"value"`
	AllocPct  decimal.Decimal `json:"alloc_pct"`
}

type PortfolioPerformance struct {
	Date  time.Time       `json:"date"`
	Value decimal.Decimal `json:"value"`
}

type AssetROI struct {
	AssetID    string          `json:"asset_id"`
	Ticker     string          `json:"ticker"`
	Name       string          `json:"name"`
	ROI        decimal.Decimal `json:"roi"`
	TotalInvested decimal.Decimal `json:"total_invested"`
	CurrentValue decimal.Decimal `json:"current_value"`
}
