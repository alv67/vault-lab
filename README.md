# VaultLab

**Self-hosted, multi-user personal finance and investment suite for homelabs.**

Track your investments, monitor asset performance, and gain insights into your financial portfolio — all from your own infrastructure.

## Features

- **Multi-user** — Family-friendly with role-based access (owner, admin, editor, viewer)
- **Portfolio management** — Multiple portfolios per user, shared access
- **Asset tracking** — Stocks, ETFs, bonds, crypto, commodities with auto-complete via Yahoo Finance lookup
- **Transaction history** — Buy, sell, dividends with multi-currency support
- **Dashboard** — Portfolio value, gain/loss, asset allocation, performance charts, ROI by asset
- **Self-contained** — Everything runs via `podman-compose up`

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23, Chi router, pgx (PostgreSQL), golang-migrate |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS, Recharts |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Container | Docker / Podman + Compose |
| Auth | JWT (access + refresh tokens) |

## Quick Start

```bash
podman-compose up -d --build
```

Then open http://localhost:3000.

> Requires Podman (or Docker) with Compose support.

## Project Status

Active development — Phase 1 (core investment tracking) in progress.  
See the [`develop`](https://github.com/alv67/vault-lab/tree/develop) branch for source code.

## License

MIT
