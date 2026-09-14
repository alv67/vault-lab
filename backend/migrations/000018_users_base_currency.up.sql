-- VaultLab migration 000018
-- Per-user base currency. All dashboard aggregations (summary, history,
-- allocation) are converted into it. EUR is the default, matching the
-- historical behavior of the supported-currencies seed.
ALTER TABLE users ADD COLUMN base_currency TEXT NOT NULL DEFAULT 'EUR';
