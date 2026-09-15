# Release Notes

## Unreleased

### Features
- Choose your base currency in Settings → Profile: the dashboard now consolidates your totals, portfolio history and allocations into that currency
- The dashboard now separates active and closed investments at vault and per-portfolio level: for the closed part you see the cost of the sold lots, the proceeds and the realized capital gain/loss; dividend rows are shown with the active investments and, once a position is fully closed, they are included in the proceeds

### Fixes
- Importing a portfolio exported by an older version of the app no longer fails: missing information is filled with sensible defaults
- Portfolios with only closed positions no longer show a misleading -100% gain/loss: closed positions are kept out of the active figures and the amounts no longer carry rounding residues

## v0.4.0 — 13 Sep 2026

### Features
- New dark theme, enabled by default, with Light / Dark / System options
- Redesigned navigation: collapsible sidebar, top header with the theme selector and the user menu, and a slide-in drawer on mobile
- Consistent theme-aware colors across every page and chart, so the app is readable in both light and dark mode
- Destructive actions now use an in-app confirmation dialog instead of the browser's native prompt
- Notifications (toasts) restyled to match the theme and made accessible to screen readers
- Creating assets/portfolios and importing a portfolio now happen in modal dialogs consistent with the app design
- Redesigned sign-in / registration screen with the VaultLab logo, a Sign in / Register switch, inline field validation and a password confirmation on registration
- Dashboard rebuilt: KPI cards per currency, an allocation donut and clickable portfolio cards
- The dashboard header shows when prices were last updated
- Portfolio detail: adding/editing a transaction now happens in a modal with inline validation and a live total, and deletion is confirmed in-app
- Portfolio positions and transactions use the shared design-system tables (positions now also show the latest price), and the page actions sit in a sticky header
- Settings reorganized into tabs: Profile, Password, Currencies and Health
- Managed currencies are now picked from a list, with the name filled in automatically
- Changing your password now validates inline and highlights the field at fault (e.g. wrong current password)
- The price-sync health events list is now paginated, so the full history can be browsed
- The price-sync health page moved to a dedicated Admin section in the sidebar

### Fixes
- Portfolio allocations update immediately after adding, editing or deleting a transaction, with no page reload
- Dashboard portfolio history chart draws every portfolio as a continuous line over a real timeline, and can be zoomed and panned like the portfolio charts
- Price-sync health: the Success Rate and Rate Limited cards now show the real values (the rate could stay stuck on `N/A`, or show `NaN%` when there was no data)
- Price-sync health: failed sync messages now state the request type (chart / spark / search / fx) and the related ticker or currency
- Assets not priced by Yahoo (manual / none) no longer generate sync errors in the price-sync health dashboard
- Price-sync health: the summary totals no longer reset on restart and can be scoped to Today / Last 24h / Last 100 events

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

## v0.2.0 — 30 Aug 2026

### Features
- Consistent values in the portfolio summary even when an exchange rate is missing
- Geography and sector distribution charts for portfolios and the dashboard
- Asset classes and allocation by investment class
- Historical exchange rates, so series and charts stay correct over time
- Asset detail page with references, exposure and full price history
- Automatic ETF exposure (countries/regions and sectors) and ticker-to-ISIN lookup

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
