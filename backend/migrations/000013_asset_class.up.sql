-- UP
ALTER TABLE assets ADD COLUMN asset_class TEXT NOT NULL DEFAULT 'other' CHECK (asset_class IN ('equity','bond','commodity','currency','crypto','real_estate','mixed','other'));
