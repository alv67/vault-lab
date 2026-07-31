-- VaultLab migration 000002
-- Reduce Yahoo Finance API calls.

-- Cache for Yahoo Finance search results (ticker autocomplete).
CREATE TABLE lookup_cache (
    query TEXT PRIMARY KEY,
    results JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tracks the last successful price fetch per asset. Used to decide when a
-- stored close value is stale and must be refreshed (e.g. on page open).
ALTER TABLE assets ADD COLUMN price_fetched_at TIMESTAMPTZ;
