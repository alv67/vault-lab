# VaultLab — Stato Progetto (28 Ago 2026)

## Infrastruttura

| Servizio | Stack | Note |
|----------|-------|------|
| Backend | Go 1.23 + Chi + pgx + golang-migrate | Containerizzato |
| Frontend | SvelteKit 5 + TypeScript + Tailwind + ECharts | Containerizzato (nginx) |
| Database | PostgreSQL 16 | Con docker volume |
| Cache | Redis 7 | Caching dashboard/series, rate-limit Yahoo |
| Worker | Go (prezzi) | Container separato |
| Python Service | FastAPI + uvirror + requests + bs4 | ETF metadata (JustETF) — in arrivo con EPIC B (B.5) |
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
- [x] DB schema (users, assets, portfolios, portfolio_shares, transactions, prices, fx_rates, splits, series, health_events)
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
  - Metric data-quality nel summary (`missing_country`, `missing_sector`, `stale_count`, `fx_missing_*`)
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

## Fase 2 — In corso

### EPIC B — Distribuzione geo/settore + FX history + Asset detail (#36) — 10 sub-issues
| Issue | Titolo | Componente | Stato |
|-------|--------|------------|-------|
| #7  | B.1 — Migration + region/sector weight model + `fx_history` + `exchange` | Backend | ✅ model + migrazioni (exchange, exposure, history) |
| #8  | B.2 — GICS/sector backfill + Yahoo v10 fetch-profile (`category_id` rimosso) | Backend | ⏳ fetch-profile e sector weightings fatti; seed GICS + populate `assets.sector` normalizzato differito |
| #9  | B.3 — Country backfill + ISO normalization + country→macro-region mapping | Backend | ⏳ `geo.RegionForCountry` + stock default, backfill completo differito |
| #10 | B.4 — ETF weight editor (frontend): regions/sectors grid + "Try scrape" | Frontend | ✅ editor tabelle + pie chart sulla pagina asset; scrape differito a B.5 |
| #11 | B.5 — Python microservice: ETF metadata da JustETF | Python | ⏳ pianificato (prefill JustETF differito) |
| #12 | B.6 — Endpoint /allocation/geography (weighted sum by region) | Backend | ⏳ pianificato (endpoint exposure asset già fatti) |
| #13 | B.7 — Endpoint /allocation/sector (weighted sum by GICS) | Backend | ⏳ pianificato (endpoint exposure asset già fatti) |
| #14 | B.8 — Frontend GeographyChart + SectorChart + dashboard/portfolio widgets | Frontend | ⏳ pianificato |
| #44 | B.9 — FX rate history + series engine per-date | Backend | ⏳ pianificato |
| #45 | B.10 — Asset detail page (`/assets/[id]`) + exchange field | Full-stack | ✅ completa |
| #49 | B.11 — Asset class: colonna `asset_class` + auto-detect Yahoo + override manuale | Full-stack | ✅ implementata (da validare) |
| #50 | B.12 — Allocazione per classi: `GET /allocation/class` + donut | Full-stack | ✅ implementata (da validare) |

**Ordine di implementazione**:
1. Data layer: B.1 → B.9 → B.3
2. Backend: B.2 → B.5 → B.6/B.7
3. Frontend: B.10 → B.4 → B.8

### Completato in questa sessione (EPIC B, parte)
- **Pagina asset detail `/assets/[id]`** (B.10) su branch `feat/B.10-asset-detail`:
  - Caratteristiche editabili: Ticker, ISIN, Nome, Tipo, Valuta, Exchange (+ metadati `exchange`, `sector`, `industry` nel modello)
  - Card metriche + grafico prezzi ECharts con selettore periodo (1M/3M/1Y/MAX)
  - Tabelle distribuzione **geo** (8 macro-regioni) e **settore** (11 settori GICS) **modificabili**,
    con validazione somma=100% e grafici a ciambella affiancati
  - Pulsanti da menu hamburger: **"Aggiorna da Yahoo"** (meta: profile + sector weightings)
    e **"Backfill storico completo"** (storico prezzi completo da Yahoo)
  - Backend: migrazioni `000009` (asset meta), `000010` (exposure weights), `000011` (history_backfilled);
    package `geo` (macro-regioni + settori GICS + mappatura paese→regione);
    repository exposure; service `UpdateAsset`/`GetAssetQuote`/`FetchAssetProfile`/
    `GetAssetExposure`/`SaveAssetExposure`/`FetchAssetExposure`/`BackfillAssetHistory`;
    fix invalidazione cache (`bumpRev`) su sync dati
  - Endpoint: PATCH `/assets/{id}`, GET `/assets/{id}/quote`, POST `/assets/{id}/fetch-profile`,
    GET/PUT `/assets/{id}/exposure`, POST `/assets/{id}/fetch-exposure`,
    POST `/assets/{id}/backfill-history`

### Decisione ISIN
Verificato: **Yahoo non espone l'ISIN** (nessun campo in `assetProfile`/`fundProfile`/`price`).
`investing.com` è bloccato da Cloudflare (403) e Morningstar richiede API a pagamento
o scraping fragile token-gated. Decisione: il campo `isin` resta **editabile a mano**
nella pagina asset; l'automazione ticker→ISIN è rimandata a **B.5** (microservizio JustETF).

### Asset class + allocazione per classi (B.11, B.12)
- **Rimossa la classificazione single-category** (`category_id` → tabella `categories`): inadatta agli ETF
  multi-settore; la distribuzione settoriale è già coperta da `asset_sector_weights`/`assets.sector`.
  Migrazione `000012_remove_category` (drop colonna + tabella).
- **Nuova `assets.asset_class`** (migrazione `000013`, check enum): `equity`, `bond`, `commodity`,
  `currency`, `crypto`, `real_estate`, `mixed`, `other`. Etichetta primaria esclusiva; per i
  multi-classe si usa `mixed`.
- **Auto-detect da Yahoo** (`assetClass` della quote, mapper `MapYahooAssetClass`) con default dal tipo
  (`stock`→equity, `bond`→bond, `commodity`→commodity, `cash`→currency, `crypto`→crypto);
  **override manuale** nell'editor asset detail che vince sull'auto-detect.
- Metrica data-quality: `missing_category` → **`missing_sector`** (asset con `sector` non valorizzato).
- **Endpoint** `GET /portfolios/{id}/allocation/class` (somma pesata sul valore in valuta portafoglio)
  e widget donut "Allocazione per classi" nella pagina portfolio.

### Altri EPIC Fase 2
- EPIC C (#39) — Metric di rischio: Sharpe, max drawdown, volatilità, regressione, Monte Carlo
- EPIC E (#38) — Pagine e componenti dominio (rebuilt dashboard, tabelle, modali)
- EPIC D (#37) — Design system & dark mode

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
