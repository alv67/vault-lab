-- UP
CREATE TABLE asset_exposure_provenance (
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    dimension TEXT NOT NULL,
    source TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (asset_id, dimension)
);
