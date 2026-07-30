-- VaultLab initial schema
-- Migration 000001

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'owner' CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Asset categories (GICS-based)
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    sector TEXT NOT NULL,
    industry TEXT
);

-- Assets
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticker TEXT NOT NULL,
    isin TEXT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('stock', 'etf', 'bond', 'mutual_fund', 'crypto', 'commodity', 'cash')),
    category_id UUID REFERENCES categories(id),
    country TEXT,
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(ticker)
);

CREATE INDEX idx_assets_ticker ON assets(ticker);
CREATE INDEX idx_assets_type ON assets(type);

-- Portfolios
CREATE TABLE portfolios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_portfolios_user ON portfolios(user_id);

-- Portfolio sharing
CREATE TABLE portfolio_shares (
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'editor', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (portfolio_id, user_id)
);

-- Transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id),
    type TEXT NOT NULL CHECK (type IN ('buy', 'sell', 'dividend', 'split', 'fee')),
    quantity NUMERIC(18, 8) NOT NULL DEFAULT 0,
    price NUMERIC(18, 6) NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    exchange_rate NUMERIC(18, 8) NOT NULL DEFAULT 1,
    fees NUMERIC(18, 6) NOT NULL DEFAULT 0,
    date TIMESTAMPTZ NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_portfolio ON transactions(portfolio_id);
CREATE INDEX idx_transactions_asset ON transactions(asset_id);
CREATE INDEX idx_transactions_date ON transactions(date);

-- Prices
CREATE TABLE prices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    open NUMERIC(18, 6) NOT NULL DEFAULT 0,
    high NUMERIC(18, 6) NOT NULL DEFAULT 0,
    low NUMERIC(18, 6) NOT NULL DEFAULT 0,
    close NUMERIC(18, 6) NOT NULL DEFAULT 0,
    volume BIGINT NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'yahoo',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(asset_id, date)
);

CREATE INDEX idx_prices_asset ON prices(asset_id);
CREATE INDEX idx_prices_date ON prices(date);
