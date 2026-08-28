# VaultLab

**Self-hosted, multi-user personal finance and investment suite for homelabs.**

Track your investments, monitor asset performance, and gain insights into your financial portfolio — all from your own infrastructure.

## Features

- **Multi-user** — Family-friendly with role-based access (owner, admin, editor, viewer)
- **Portfolio management** — Multiple portfolios per user, export/import, ownership enforced on all endpoints
- **Asset tracking** — Stocks, ETFs, bonds, crypto, commodities with auto-complete and sync via Yahoo Finance
- **Asset detail page** — Editable metadata (exchange, ISIN), full price history with backfill, and editable geographic/sector exposure with charts
- **Transaction history** — Buy, sell, dividends, splits, fees with multi-currency support
- **Dashboard** — Portfolio value, gain/loss, allocation, performance charts, ROI by asset
- **Market prices** — Yahoo Finance with Redis caching, rate-limit/backoff, series materialization, price health dashboard
- **Data quality** — Summary exposes staleness / missing-country / missing-sector / missing-FX metrics
- **Self-contained** — Everything runs via `podman-compose up`

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23, Chi router, pgx (PostgreSQL), golang-migrate |
| Frontend | SvelteKit 5, TypeScript, Tailwind CSS, ECharts |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Container | Docker / Podman + Compose |
| Auth | JWT (access + refresh tokens, rotation) |

## Quick Start

```bash
podman-compose up -d --build
```

Then open http://localhost:3000.

> Requires Podman (or Docker) with Compose support.

## Testing

```bash
make test-e2e        # End-to-end API tests on an isolated stack (no data pollution)
make test            # Go unit tests (inside the backend container)
```

## Project Status

First official release: **v0.1.0** on [`main`](https://github.com/alv67/vault-lab/tree/main).

Active development on the [`develop`](https://github.com/alv67/vault-lab/tree/develop) branch — see [STATUS.md](STATUS.md) and [PLAN.md](PLAN.md) for the roadmap.

## License

MIT
