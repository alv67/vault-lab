# Release Notes

This file collects, in order, the features and bug fixes shipped in every
release. The lines are meant to be reused directly in the product release
notes: **one line per feature/fix**, written from the end user's point of
view (what they see and use in the app — no internal/backend details).

> Convention: when you close a PR that changes the product, add the relevant
> lines at the top (or under the current release section). When you publish,
> create a new section with the version and date.

## v0.1.0 — 25 Aug 2026 (first official release)

### Features
- Multi-user registration and login (JWT access + rotating refresh tokens)
- Account settings: edit your name and email, and change your password
- Portfolios: create, edit, delete, export and import
- Assets: create, edit, delete with ticker autocomplete and automatic Yahoo price sync
- Transactions (buy / sell / dividend / split / fee)
- Dashboard: total value, gain/loss, allocation, performance and per-asset ROI
- Portfolio history chart on the dashboard showing how the portfolio value changed over time
- Performance history chart per portfolio with invested amount (cost basis), current value of the still-invested assets and historical realized gain/loss
- Portfolios list with the current value of every portfolio/asset
- Multi-currency support (EUR / USD / GBP / CHF) with a configurable currency whitelist
- Per-currency invested amounts on the dashboard
- Automatic and manual price updates with a price-sync health dashboard
- Automatic periodic refresh of asset prices and exchange rates, updating values and the history up to the latest run

## v0.2.0 — 30 Aug 2026

### Features
- Consistent values in the portfolio summary even when an exchange rate is missing
- Geography and sector distribution charts for portfolios and the dashboard
- Asset classes and allocation by investment class
- Historical exchange rates, so series and charts stay correct over time
- Asset detail page with references, exposure and full price history
- Automatic ETF exposure (countries/regions and sectors) and ticker-to-ISIN lookup

## v0.3.0 — 11 Sep 2026

### Features
- Choose how each asset gets its prices: `Yahoo`, `Manual` or `None` (avoids Yahoo errors for non-Yahoo tickers such as some bonds)
- The asset price chart now loads the full history and zooms in place using the 1M/3M/1Y/YTD/MAX selectors (no unnecessary reloads)
- New `YTD` (year-to-date) range on the asset price chart
- Stock split markers shown on the asset price chart (e.g. `Split 4:1`)
- Asset exposure (regions/sectors) is edited in a dedicated modal with validated weight tables and one-click fill from JustETF and Yahoo
- Edit the geographic distribution with a per-country list (add/remove countries and set each weight), alongside the regions and sectors
- Geographic and sector exposure can also be filled from Morningstar (official regions), alongside JustETF and Yahoo
- Every distribution shows where its data comes from and when it was last updated (e.g. `from Morningstar (2026-09-05)`)
- Provider lookups are cached, so opening the prefill again is immediate

### Fixes
- Reopening an exposure editor now always starts from the saved data: unsaved changes are discarded
