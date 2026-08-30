-- UP
CREATE TABLE asset_region_weights (
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    region TEXT NOT NULL,
    weight NUMERIC(10,4) NOT NULL DEFAULT 0,
    PRIMARY KEY (asset_id, region)
);
CREATE TABLE asset_sector_weights (
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    sector TEXT NOT NULL,
    weight NUMERIC(10,4) NOT NULL DEFAULT 0,
    PRIMARY KEY (asset_id, sector)
);
CREATE INDEX idx_asset_region_weights_asset ON asset_region_weights (asset_id);
CREATE INDEX idx_asset_sector_weights_asset ON asset_sector_weights (asset_id);
