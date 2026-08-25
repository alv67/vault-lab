# VaultLab — Stato Progetto (25 Ago 2026)

## Infrastruttura

| Servizio | Stack | Note |
|----------|-------|------|
| Backend | Go 1.23 + Chi + pgx + golang-migrate | Containerizzato |
| Frontend | SvelteKit 5 + TypeScript + Tailwind + ECharts | Containerizzato (nginx) |
| Database | PostgreSQL 16 | Con docker volume |
| Cache | Redis 7 | Caching dashboard/series, rate-limit Yahoo |
| Worker | Go (prezzi) | Container separato |
| Container | podman + podman-compose su macOS | |

## Release

**v0.1.0** — prima release ufficiale su `main` (25 Ago 2026).
Flusso: branch → PR su `develop` → merge → tag `v0.1.x`/`v0.2.0` su `main`.

## Fase 0 — ✅ Completata

- [x] Struttura monorepo (backend + frontend + docker)
- [x] Docker Compose (Postgres, Redis, backend, frontend, worker)
- [x] Go module con tutte le dipendenze
- [x] Frontend migrato a SvelteKit (da React/Vite)
- [x] Auth JWT (access + refresh token, rotazione automatica)
- [x] API REST (auth, assets, portfolios, transactions, prices, settings, dashboard, health)
- [x] DB schema (users, assets, categories, portfolios, portfolio_shares, transactions, prices, fx_rates, splits, series, health_events)
- [x] Makefile aggiornato per podman-compose

## Fase 1 — ✅ Completata

- [x] Registrazione/Login multi-utente
- [x] CRUD portafogli + export/import
- [x] CRUD asset con autocomplete e sync Yahoo Finance
- [x] Transazioni (buy/sell/dividend/split/fee) con JOIN asset
- [x] Dashboard: valore totale, gain/loss, allocazione, performance, ROI per asset
- [x] Valuta dinamica (€/$/£/CHF) e whitelist valute configurabile
- [x] Prezzi Yahoo con cache Redis, rate-limit/backoff, serie materializzate, health dashboard
- [x] `POST /api/v1/prices/refresh`
- [x] EPIC A — Data correctness & security (#3 #4 #5 #6):
  - Aggregazione coerente con FX mancante (bucket esplicito + forward-fill serie)
  - Enforce ownership su tutti gli endpoint analitici
  - Metric data-quality nel summary (`missing_country`, `missing_category`, `stale_count`, `fx_missing_*`)
  - Rimosso codice SQL morto in `repository/portfolio.go`

### Da fare (Fase 1/2)
- [ ] Sharing portafogli tra utenti (portfolio_shares)
- [ ] Inserimento prezzo manuale nella UI
- [ ] Ricerca asset con autocomplete nel form transazioni
- [ ] Import CSV transazioni (l'import JSON esiste già)

## Problemi Aperti

### Yahoo Finance rate-limited (429)
Mitigato con rate-limit/backoff e throttling, ma Yahoo può comunque bloccare il container podman.
- Da macOS host → funziona
- Se il worker resta bloccato, valutare: script esterno su macOS (curl/cron) che POSTa i prezzi al backend,
  oppure API alternativa (Finnhub, Alpha Vantage con API key)

### Ticker europei
Asset europei su Yahoo usano suffisso exchange (es. VWCE.DE, VWCE.AS). L'utente deve sapere il
ticker corretto. Da documentare o aggiungere selezione exchange nell'autocomplete.

## Fase 2 — Pianificata

- EPIC B — Distribuzione territoriale e per settore GICS (#36)
- EPIC C — Metric di rischio: Sharpe, max drawdown, volatilità, regressione, Monte Carlo (#39)
- EPIC E — Pagine e componenti dominio (rebuilt dashboard, tabelle, modali)
- EPIC D — Design system & dark mode

## Fase 3 — Pianificata

- Multi-tenancy familiare (portfolio_shares)
- Gestione permessi e condivisione

## Fase 4 — Futura

- Tracciamento spese e categorie
- Budget mensile
- Obiettivi di risparmio

## Comandi Utili

```bash
make up              # Avvia tutto con podman-compose
make down            # Ferma tutti i servizi
make reset           # Ferma i servizi e cancella i volumi dati (fresh start)
make logs            # Log in tempo reale
make migrate         # Esegui migration DB
make test            # Test Go (in container)
make test-e2e        # Test end-to-end su stack isolato
make frontend-dev    # Sviluppo frontend con hot-reload
# Per ricreare container dopo modifiche:
podman-compose stop <service>
podman rm <container>
podman-compose up -d --build <service>
```
