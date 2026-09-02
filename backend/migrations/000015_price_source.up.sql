-- UP
ALTER TABLE assets ADD COLUMN price_source TEXT NOT NULL DEFAULT 'yahoo' CHECK (price_source IN ('yahoo','manual','none'));

