-- VaultLab migration 000006
-- Supported currencies as a persisted, configurable whitelist. Seeded with the
-- two always-available currencies; the rest is populated via `/settings/currencies`.

CREATE TABLE supported_currencies (
    code       TEXT PRIMARY KEY,
    name       TEXT NOT NULL DEFAULT '',
    symbol     TEXT NOT NULL DEFAULT '',
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    sort       INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO supported_currencies (code, name, symbol, sort) VALUES
    ('USD', 'US Dollar', '$', 1),
    ('EUR', 'Euro', '€', 2)
ON CONFLICT (code) DO NOTHING;