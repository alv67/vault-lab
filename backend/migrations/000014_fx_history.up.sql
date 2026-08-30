-- VaultLab migration 000014
-- Per-date USD->quote FX rate history, mirroring the prices table. The series
-- engine resolves rates as of each portfolio day from this history, falling
-- back to the latest fx_rates snapshot when a day has no stored rate.

CREATE TABLE fx_history (
    base_currency TEXT NOT NULL DEFAULT 'USD',
    quote_currency TEXT NOT NULL,
    date DATE NOT NULL,
    rate NUMERIC(18, 8) NOT NULL,
    source TEXT NOT NULL DEFAULT 'yahoo',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (base_currency, quote_currency, date)
);