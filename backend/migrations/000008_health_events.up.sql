-- UP
CREATE TABLE health_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id uuid,
    event_type TEXT NOT NULL, -- 'fetch', 'fx_refresh', 'split', 'lookup'
    status TEXT NOT NULL, -- 'success', 'error', 'fallback'
    code TEXT, -- 'rate_limited', 'http_429', 'timeout', 'invalid_ticker'
    message TEXT,
    duration_ms INTEGER,
    error_code TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_health_events_asset_created ON health_events (asset_id, created_at DESC);
CREATE INDEX idx_health_events_created ON health_events (created_at DESC);
