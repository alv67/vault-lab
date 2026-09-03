-- UP
CREATE TABLE asset_country_weights (
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    country TEXT NOT NULL,
    weight NUMERIC(10,4) NOT NULL DEFAULT 0,
    PRIMARY KEY (asset_id, country)
);
CREATE INDEX idx_asset_country_weights_asset ON asset_country_weights (asset_id);
