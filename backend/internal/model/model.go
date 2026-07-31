package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Portfolio struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AssetType string

const (
	AssetTypeStock      AssetType = "stock"
	AssetTypeETF        AssetType = "etf"
	AssetTypeBond       AssetType = "bond"
	AssetTypeMutualFund AssetType = "mutual_fund"
	AssetTypeCrypto     AssetType = "crypto"
	AssetTypeCommodity  AssetType = "commodity"
	AssetTypeCash       AssetType = "cash"
)

type Asset struct {
	ID             uuid.UUID  `json:"id"`
	Ticker         string     `json:"ticker"`
	ISIN           string     `json:"isin,omitempty"`
	Name           string     `json:"name"`
	Type           AssetType  `json:"type"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	Country        string     `json:"country,omitempty"`
	Currency       string     `json:"currency"`
	CreatedAt      time.Time  `json:"created_at"`
	PriceFetchedAt *time.Time `json:"price_fetched_at,omitempty"`
}

type Category struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Sector string    `json:"sector,omitempty"`
}

type TransactionType string

const (
	TxBuy      TransactionType = "buy"
	TxSell     TransactionType = "sell"
	TxDividend TransactionType = "dividend"
	TxSplit    TransactionType = "split"
	TxFee      TransactionType = "fee"
)

type Transaction struct {
	ID           uuid.UUID       `json:"id"`
	PortfolioID  uuid.UUID       `json:"portfolio_id"`
	AssetID      uuid.UUID       `json:"asset_id"`
	Type         TransactionType `json:"type"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Currency     string          `json:"currency"`
	ExchangeRate decimal.Decimal `json:"exchange_rate,omitempty"`
	Fees         decimal.Decimal `json:"fees,omitempty"`
	Date         time.Time       `json:"date"`
	Notes        string          `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type TransactionWithAsset struct {
	ID           uuid.UUID       `json:"id"`
	PortfolioID  uuid.UUID       `json:"portfolio_id"`
	AssetID      uuid.UUID       `json:"asset_id"`
	AssetTicker  string          `json:"asset_ticker"`
	AssetName    string          `json:"asset_name"`
	Type         TransactionType `json:"type"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Currency     string          `json:"currency"`
	ExchangeRate decimal.Decimal `json:"exchange_rate,omitempty"`
	Fees         decimal.Decimal `json:"fees,omitempty"`
	Date         time.Time       `json:"date"`
	Notes        string          `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type Price struct {
	ID        uuid.UUID       `json:"id"`
	AssetID   uuid.UUID       `json:"asset_id"`
	Date      time.Time       `json:"date"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	Volume    int64           `json:"volume"`
	Source    string          `json:"source"`
	CreatedAt time.Time       `json:"created_at"`
}

type PortfolioShare struct {
	PortfolioID uuid.UUID `json:"portfolio_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        Role      `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}
