-- VaultLab migration 000007
-- Materialized daily position series so dashboard/history reads are fast.

CREATE TABLE portfolio_series (
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    date         DATE NOT NULL,
    qty          NUMERIC NOT NULL DEFAULT 0,
    cost_basis   NUMERIC NOT NULL DEFAULT 0,
    market_value NUMERIC NOT NULL DEFAULT 0,
    realized     NUMERIC NOT NULL DEFAULT 0,
    PRIMARY KEY (portfolio_id, date)
);

CREATE TABLE asset_series (
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    asset_id     UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    date         DATE NOT NULL,
    qty          NUMERIC NOT NULL DEFAULT 0,
    cost_basis   NUMERIC NOT NULL DEFAULT 0,
    market_value NUMERIC NOT NULL DEFAULT 0,
    realized     NUMERIC NOT NULL DEFAULT 0,
    PRIMARY KEY (portfolio_id, asset_id, date)
);

CREATE INDEX asset_series_pf_date_idx ON asset_series (portfolio_id, date);
