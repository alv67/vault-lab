# VaultLab — Stato Progetto (28 Ago 2026)

## Infrastruttura

| Servizio | Stack | Note |
|----------|-------|------|
| Backend | Go 1.23 + Chi + pgx + golang-migrate | Containerizzato |
| Frontend | SvelteKit 5 + TypeScript + Tailwind + ECharts | Containerizzato (nginx) |
| Database | PostgreSQL 16 | Con docker volume |
| Cache | Redis 7 | Caching dashboard/series, rate-limit Yahoo |
| Worker | Go (prezzi) | Container separato |
| Python Service | FastAPI + uvicorn + requests + bs4 | ETF metadata da JustETF (esposizione paesi/regioni + settori, ticker→ISIN) — EPIC B.5 |
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

### EPIC B — Distribuzione geo/settore + FX history + Asset detail (#36) — 12 sub-issues
| Issue | Titolo | Componente | Stato |
|-------|--------|------------|-------|
| #7  | B.1 — Migration + region/sector weight model + `fx_history` + `exchange` | Backend | ✅ model + migrazioni (exchange, exposure, history) |
| #8  | B.2 — GICS/sector backfill + Yahoo v10 fetch-profile | Backend | ✅ chiusa da design change: `category_id`/`categories` rimossi (migrazione `000012`, PR #51); fetch-profile + `assets.sector` normalizzato coperti altrove; residuo (backfill `missing_sector`) accorpato a B.3 |
| #9  | B.3 — Country backfill (exposure via `assetProfile`) + sector backfill + ISO normalization + region mapping | Backend | ✅ implementata: country = domicilio emittente (fix cross-listing alla creazione), `geo.NormalizeCountry`/`RegionForCountry`, validazione ISO su create/update, `POST /assets/backfill-meta` che riempie E corregge i legacy (1AAPL.MI → US) — in PR #55 |
| #10 | B.4 — ETF weight editor (frontend): regions/sectors grid + "Try scrape" | Frontend | ✅ editor tabelle + pie chart sulla pagina asset; scrape differito a B.5 |
| #11 | B.5 — Python microservice: ETF metadata da JustETF | Python | ✅ implementata: `python-service` (FastAPI) con search ticker/ISIN + exposure paesi/regioni e settori; backend `POST /assets/{id}/fetch-etf-exposure` con auto-resolve ISIN — in PR #55 |
| #12 | B.6 — Endpoint /allocation/geography (weighted sum by region) | Backend | ✅ implementata: 8 macro-regioni + `Other`, zero-filled — in PR #60 |
| #13 | B.7 — Endpoint /allocation/sector (weighted sum by GICS) | Backend | ✅ implementata: 11 settori GICS, zero-filled — in PR #60 |
| #14 | B.8 — Frontend GeographyChart + SectorChart + dashboard/portfolio widgets | Frontend | ⏳ pianificato |
| #44 | B.9 — FX rate history + series engine per-date | Backend | ⏳ pianificato |
| #45 | B.10 — Asset detail page (`/assets/[id]`) + exchange field | Full-stack | ✅ completa |
| #49 | B.11 — Asset class: colonna `asset_class` + auto-detect Yahoo + override manuale | Full-stack | ✅ implementata (da validare) |
| #50 | B.12 — Allocazione per classi: `GET /allocation/class` + donut | Full-stack | ✅ implementata (da validare) |

**Ordine di implementazione**:
1. Data layer: B.1 → B.9 → B.3
2. Backend: B.3 (backfill) → B.5 → B.6/B.7 ✅  <small>(B.2 chiusa: superata in B.11/B.12, residuo in B.3)</small>
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

### B.3 + B.5 — Esposizione ETF da JustETF (branch `feat/B.3-B.5-meta-backfill`, PR #55 open)
- **B.3 (country backfill + ISO)** — `country` degli stock = domicilio emittente (fix cross-listing
  alla creazione, es. 1AAPL.MI → US), validazione ISO alpha-2 su create/update,
  `POST /assets/backfill-meta` che riempie e corregge i legacy. Package `geo` esteso:
  mappature paese→ISO (South Korea, Saudi Arabia, UAE, Thailand, Malaysia + EM comuni) e alias
  settori JustETF→GICS (Finance→Financials, Consumer Non-Cyclicals→Consumer Staples, etc.).
- **B.5 (python-service)** — microservizio `python-service/` (FastAPI + uvicorn + requests + bs4,
  immagine `python:3.12-slim`):
  - `GET /api/v1/etf/search?q={ticker|nome}` — ticker/ISIN via JustETF quick-search; rimuove il
    suffisso borsa (`.MI/.DE/.L/...`) come richiede il sito;
  - `GET /api/v1/etf/{isin}/exposure` — paesi/regioni e settori **completi** replicando gli AJAX
    Wicket "Show more" di JustETF (niente browser a runtime, Playwright usato solo per discovery);
  - `GET /api/v1/etf/{isin}/holdings` (stub) e `GET /healthz`; mai crash → errori `502` JSON.
- **Backend Go** — interfaccia `price.ETFFetcher` + `JustETFFetcher` (client del python-service),
  aggregazione `AggregateRegions`/`AggregateSectors` (paesi→macro-regioni + alias settori),
  nuova rotta **`POST /assets/{id}/fetch-etf-exposure`**: se l'asset non ha ISIN lo **auto-risolve
  dal ticker** (preferenza ticker esatto, poi similarità nome con `asset.Name`; il valore viene
  persistito sull'asset), poi scarica paesi/regioni + settori e li salva (`asset_region_weights` /
  `asset_sector_weights`). Config `VAULT_PYTHON_SERVICE_URL`; servizio `python-service` presente
  sia in `docker-compose.yml` sia in `docker-compose.test.yml`.
- **Asset duplicato** — `POST /assets` con ticker già esistente ora risponde **409 Conflict** con
  messaggio chiaro e l'id dell'asset esistente (`asset_id` + `id`).
- **Verifica** — pytest (`python-service/tests`, 17 test), Go `build/vet/test green`; e2e su stack
  isolato `vaultlab-test`: XMME (14 paesi/13 settori, regioni sommano 100), VWCE e `SMEA.MI` senza
  ISIN auto-risolti correttamente (SMEA.MI → `IE00B4K48X80` iShares Core MSCI Europe).
  Test manuali via **`tests/api-test.http`** (estensione REST Client in VS Code) sullo stack test
  (porta 8081).

### Decisione ISIN
Verificato: **Yahoo non espone l'ISIN** (nessun campo in `assetProfile`/`fundProfile`/`price`).
`investing.com` è bloccato da Cloudflare (403) e Morningstar richiede API a pagamento
o scraping fragile token-gated. Decisione B.5: il campo `isin` resta **editabile a mano**
nella pagina asset, ma per gli ETF è ora **automatizzato** via JustETF: `POST /assets/{id}/fetch-etf-exposure`
(anche alla creazione, se l'ISIN è vuoto) risolve ticker→ISIN e lo persiste sull'asset.

### Asset class + allocazione per classi (B.11, B.12)
- **Rimossa la classificazione single-category** (`category_id` → tabella `categories`): inadatta agli ETF
  multi-settore; la distribuzione settoriale è già coperta da `asset_sector_weights`/`assets.sector`.
  Migrazione `000012_remove_category` (drop colonna + tabella).
- **Nuova `assets.asset_class`** (migrazione `000013`, check enum): `equity`, `bond`, `commodity`,
  `currency`, `crypto`, `real_estate`, `mixed`, `other`. Etichetta primaria esclusiva; per i
  multi-classe si usa `mixed`.
- **Auto-detect da Yahoo**: l'endpoint non espone più `assetClass` (modulo `quote`
  inesistente in quoteSummary v10), quindi la classe viene derivata da
  `geo.ClassifyAssetClass` (categoria fondo Morningstar `defaultKeyStatistics.category`/
  `fundProfile.categoryName` + euristica sul nome; BND→bond, GLD→commodity, VWCE.DE→equity),
  con default dal tipo (`stock`→equity, `bond`→bond, `commodity`→commodity, `cash`→currency,
  `crypto`→crypto). Il recupero della classe è **accorpato al recupero info asset**
  (`GET /assets/meta`, usato alla creazione e da "Aggiorna da Yahoo") e **non** alla lettura
  dei settori (`fetch-exposure` non la tocca più); **override manuale** nell'editor asset
  che vince sempre (aggiornato solo se vuota o `other`).
- Metrica data-quality: `missing_category` → **`missing_sector`** (asset con `sector` non valorizzato).
- **Endpoint** `GET /portfolios/{id}/allocation/class` (somma pesata sul valore in valuta portafoglio)
  e widget donut "Allocazione per classi" nella pagina portfolio.

### B.6 + B.7 — Allocazione geo/settoriale a livello portafoglio (branch `feat/B.6-B.7-allocation`, PR #60)
- **Endpoint** `GET /portfolios/{id}/allocation/geography` (somma pesata per **macro-regione**) e
  `GET /portfolios/{id}/allocation/sector` (somma pesata per **settore GICS**): stessi principi di
  `/allocation/class` (pesi per-asset da `asset_region_weights`/`asset_sector_weights` × valore in
  valuta portafoglio, conversioni FX incluse).
- Risultato con **8 macro-regioni (+ `Other`)** e **11 settori GICS (+ `Other`)**, zero-filled per
  evitare buchi: bucket sempre in ordine fisso, percentuali che sommano a 100. `Other` raccoglie
  pesi non mappabili (asset country/region o settore non riconosciuti); per i settori un peso
  complessivo 0 finisce tutto in `Other`.
- Porto vuoto o totale zero → righe a zero senza bucket `Other` (niente denominatori artificiali).
- **Verifica**: e2e-unit in `backend/internal/service/service_test.go` (ETF completo, fallback stock
  su domicilio, conversione FX, portafoglio vuoto, bucket `Other`); test manuale di smoke su stack
  isolato `vaultlab-test` con **`tests/test-epic-b.sh`** (20 check PASS) usando prezzi seminati da
  **`tests/seed-prices.sql`** (Yahoo è disabilitato sullo stack test, quindi i prezzi si scrivono
  solo via SQL) e la raccolta **`tests/api-test.http`** estesa.

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
# Test manuali API (estensione REST Client in VS Code) sullo stack test:
#   tests/api-test.http — richieste in ordine contro http://localhost:8081/api/v1
# Smoke test EPIC B sulle allocazioni (stack test, porta 8081):
#   tests/test-epic-b.sh [--step | --no-seed]
# Per ricreare container dopo modifiche:
podman-compose stop <service>
podman rm <container>
podman-compose up -d --build <service>
```
