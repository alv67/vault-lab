-- VaultLab migration 000003
-- FX rates cached as USD->quote pairs. Cross rates (A->B) are computed as
-- (USD->B)/(USD->A) in the application. Only the latest close rate is kept,
-- refreshed like asset prices.

CREATE TABLE fx_rates (
    base_currency TEXT NOT NULL DEFAULT 'USD',
    quote_currency TEXT NOT NULL,
    rate NUMERIC(18, 8) NOT NULL,
    source TEXT NOT NULL DEFAULT 'yahoo',
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (base_currency, quote_currency)
);
