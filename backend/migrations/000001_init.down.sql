DROP INDEX IF EXISTS idx_prices_date;
DROP INDEX IF EXISTS idx_prices_asset;
DROP TABLE IF EXISTS prices;

DROP INDEX IF EXISTS idx_transactions_date;
DROP INDEX IF EXISTS idx_transactions_asset;
DROP INDEX IF EXISTS idx_transactions_portfolio;
DROP TABLE IF EXISTS transactions;

DROP TABLE IF EXISTS portfolio_shares;

DROP INDEX IF EXISTS idx_portfolios_user;
DROP TABLE IF EXISTS portfolios;

DROP INDEX IF EXISTS idx_assets_type;
DROP INDEX IF EXISTS idx_assets_ticker;
DROP TABLE IF EXISTS assets;

DROP TABLE IF EXISTS categories;

DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS "pgcrypto";
