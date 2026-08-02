-- VaultLab migration 000004
-- Stock split events per asset (source of truth). Ex-date is the effective day.

CREATE TABLE splits (
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    numerator NUMERIC(18, 8) NOT NULL,
    denominator NUMERIC(18, 8) NOT NULL,
    source TEXT NOT NULL DEFAULT 'yahoo',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (asset_id, date)
);
