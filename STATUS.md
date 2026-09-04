# VaultLab — Stato Progetto (28 Ago 2026)

## Infrastruttura

| Servizio | Stack | Note |
|----------|-------|------|
| Backend | Go 1.23 + Chi + pgx + golang-migrate | Containerizzato |
| Frontend | SvelteKit 5 + TypeScript + Tailwind + ECharts | Containerizzato (nginx) |
| Database | PostgreSQL 16 | Con docker volume |
| Cache | Redis 7 | Caching dashboard/series, rate-limit Yahoo |
| Worker | Go (prezzi) | Container separato |
| Python Service | FastAPI + uvicorn + requests + bs4 + selenium (chromium headless) | ETF metadata da JustETF (esposizione paesi + settori) + Morningstar (esposizione paesi + regioni ufficiali + settori via resolver custom; auto-resolve ticker→ISIN per mercato) — EPIC B.5, B.14 |
| Container | podman + podman-compose su macOS | |

## Release

**v0.1.0** — prima release ufficiale su `main` (25 Ago 2026).
**v0.2.0** — seconda release su `main` (30 Ago 2026): EPIC A (data correctness & security)
e EPIC B completo (distribuzione geo/settoriale, asset class, FX history, charts).
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
| #12 | B.6 — Endpoint /allocation/geography (weighted sum by region) | Backend | ✅ implementata: 8 macro-regioni + `Other` (poi allineate a Morningstar: 10 + `Other`), zero-filled — in PR #60 |
| #13 | B.7 — Endpoint /allocation/sector (weighted sum by GICS) | Backend | ✅ implementata: 11 settori GICS, zero-filled — in PR #60 |
| #14 | B.8 — Frontend GeographyChart + SectorChart + dashboard/portfolio widgets | Frontend | ✅ implementata (charts dashboard/portfolio, universo equity-only + coverage, pulsante JustETF + ISIN) — in PR #62 |
| #44 | B.9 — FX rate history + series engine per-date | Backend | ✅ implementata — in PR #61 |
| #45 | B.10 — Asset detail page (`/assets/[id]`) + exchange field | Full-stack | ✅ completa |
| #49 | B.11 — Asset class: colonna `asset_class` + auto-detect Yahoo + override manuale | Full-stack | ✅ implementata (da validare) |
| #50 | B.12 — Allocazione per classi: `GET /allocation/class` + donut | Full-stack | ✅ implementata (da validare) |
| #58 | B.13 — Per-country exposure storage: tabella `asset_country_weights`, 3 dimensioni | Backend + Frontend | ✅ implementata |
| #59 | B.14 — Morningstar come fonte esposizione: resolver custom (bootstrap Chromium headless per WAF+JWT), rotta backend, prefill frontend | Backend + Frontend | ✅ implementata |

**Ordine di implementazione**:
1. Data layer: B.1 → B.9 → B.3
2. Backend: B.3 (backfill) → B.5 → B.6/B.7 ✅  <small>(B.2 chiusa: superata in B.11/B.12, residuo in B.3)</small>
3. Frontend: B.10 → B.4 → B.8

### Completato in questa sessione (EPIC B, parte)
- **Pagina asset detail `/assets/[id]`** (B.10) su branch `feat/B.10-asset-detail`:
  - Caratteristiche editabili: Ticker, ISIN, Nome, Tipo, Valuta, Exchange (+ metadati `exchange`, `sector`, `industry` nel modello)
  - Card metriche + grafico prezzi ECharts con selettore periodo (1M/3M/1Y/MAX)
  - Tabelle distribuzione **geo** (8 macro-regioni, poi allineate a Morningstar in B.14 → 10 + `Other`) e **settore** (11 settori GICS) **modificabili**,
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
  - **Morningstar** (`GET /api/v1/etf/{isin}/morningstar-exposure`) — stesso breakdown paesi/
    settori recuperato da **Morningstar** con un **resolver custom** (`app/morningstar.py`, senza
    mstarpy): il dominio `global.morningstar.com` usato da mstarpy è hard-blocked (403), quindi il
    resolver esegue un **bootstrap Chromium headless** (sotto Xvfb nel container, chromedriver
    esplicito per aarch64) che risolve il challenge AWS WAF su `www.morningstar.com` e ottiene il
    **Bearer JWT** da `/api/v2/stores/maas/token` (cache ~1h), poi le chiamate dati viaggiano via
    `requests` con token+cookie verso `www.us-api.morningstar.com/sal/sal-service/etf/...`
    (`portfolio/v2/sector/{sid}/data` → settori, `portfolio/regionalSectorIncludeCountries/{sid}/data`
    → paesi, `portfolio/regionalSector/{sid}/data` → regioni). ISIN→securityId via
    `www.morningstar.com/api/v2/search?q={isin}`. Parser difensivo: `fundPortfolio.countries`
    (name camelCase→nome leggibile, percent) e bucket `EQUITY`/`FIXEDINCOME` → settori GICS
    (scelta bucket col peso maggiore, chiavi non-GICS scartate), conversione 0-1→0-100,
    pesi paesi tenuti come riportati (Morningstar espone la lista paesi completa, 51 voci
    con molte a 0 e una quota residuale non esposta come paese; il residuo confluisce
    nella regione `Other / Not Classified` lato backend), ordinamento pesi desc.
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

Da B.14 l'auto-resolve è **market-aware**: quando il ticker porta un suffisso di
mercato riconosciuto (`.MI`, `.DE`, `.L`, `.SW`, ...), l'ISIN viene risolto via
**Morningstar** cercando la quotazione su quell'exchange (mappa suffisso→codice
Morningstar in `python-service/app/morningstar.py`); senza suffisso si usa
JustETF. È necessario perché lo stesso ticker può avere **ISIN diversi a seconda
del listino** (es. `EQQQ` = `IE0032077012` sul listino tedesco vs altre quotazioni),
e Morningstar permette di cercare sul mercato esatto.

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
- Risultato con **10 macro-regioni (+ `Other`, allineate a Morningstar)** e **11 settori GICS (+ `Other`)**, zero-filled per
  evitare buchi: bucket sempre in ordine fisso, percentuali che sommano a 100. `Other` raccoglie
  pesi non mappabili (asset country/region o settore non riconosciuti); per i settori un peso
  complessivo 0 finisce tutto in `Other`.
- Porto vuoto o totale zero → righe a zero senza bucket `Other` (niente denominatori artificiali).
- **Verifica**: e2e-unit in `backend/internal/service/service_test.go` (ETF completo, fallback stock
  su domicilio, conversione FX, portafoglio vuoto, bucket `Other`); test manuale di smoke su stack
  isolato `vaultlab-test` con **`tests/test-epic-b.sh`** (20 check PASS) usando prezzi seminati da
  **`tests/seed-prices.sql`** (Yahoo è disabilitato sullo stack test, quindi i prezzi si scrivono
  solo via SQL) e la raccolta **`tests/api-test.http`** estesa.

### B.8 + B.9 — Chart dashboard/portfolio e FX history (branch `feat/B.8-allocation-charts` PR #62, `feat/B.9-fx-history` PR #61)
- **B.8 (issue #14, PR #62)** — widget di allocazione geo/settoriale:
  `GeographyChart` + `SectorChart` (donut a 12 colori + tabella righe complete)
  sulla pagina portafoglio e nella card "Allocazione complessiva" della
  dashboard, alimentati da `GET /portfolios/{id}/allocation/geography`,
  `/allocation/sector` e `GET /dashboard/allocation` (aggregato USD).
- **Universo equity-only (follow-up B.8)** — le allocazioni geo/settoriali
  coprono solo l'equity (`exposureEligible`: stock sempre; etf/mutual_fund solo
  con `asset_class` `equity`/`real_estate`). Bond, crypto, commodity, valute e
  fondi non classificati sono **esclusi** (mai in `Other`); geography/sector/
  dashboard espongono `covered_value`/`excluded_value` e i grafici mostrano la
  nota di copertura.
- **Editor asset (follow-up B.8)** — pulsante **"Carica da JustETF"**
  (`POST /assets/{id}/fetch-etf-exposure`) che scarica regioni+settori e
  sincronizza l'**ISIN** risolto nel form; campo ISIN anche nel form di
  creazione asset; banner "solo asset azionari" quando l'asset non è
  azionabile.
- **B.9 (issue #44, PR #61)** — storico tassi di cambio per-data (`fx_history`)
  integrato nel series engine per conversioni storiche per-date.
- **Verifica** — Go build/vet/test green; smoke su stack isolato
  `vaultlab-test` con `tests/test-epic-b.sh` (20 PASS; gli ETF sono creati con
  `asset_class: equity` per rispettare l'universo strict) + esclusione bond
  verificata end-to-end (covered=2600, excluded=10000, pesi somma 100).

### B.13 + B.14 — Exposure 3 dimensioni + Morningstar (issues #58, #59)
- **B.13 — Per-country exposure storage** (#58): nuova tabella
  `asset_country_weights (asset_id, country, weight)` (migrazione
  `000016_country_weights`). L'esposizione ora ha 3 dimensioni:
`countries` (pesi per paese ISO-3166 alpha-2), `regions` (macro-regioni)
    e `sectors` (settori GICS). Il package `geo` espone `var Countries` (~89
    codici ISO canonici). Il repository `ExposureRepository` aggiunge
    `FindCountries`/`ReplaceCountries`/`FindCountriesByAssets`.
    - **Tassonomia regioni allineata a Morningstar**: 10 macro-regioni +
      `Other / Not Classified`. Rispetto alla vecchia lista a 8 regioni:
      **United Kingdom** (GB) separata da Europe Developed; **Japan** (JP)
      separata da Asia Developed; **Australasia** (AU, NZ) separata; **TW/KR**
      spostate da Asia Emerging a **Asia Developed**; **SI** spostata da
      Europe Emerging a Europe Developed (coerente con Morningstar).
      Le chiavi Morningstar (`northAmerica`, `unitedKingdom`, `japan`,
      `australasia`, ...) mappano 1:1 sui nomi canonici.
  - `GET /assets/{id}/exposure` restituisce countries zero-filled sulla lista
    canonica completa (più regions e sectors).
  - `PUT /assets/{id}/exposure` accetta `countries?` (dimensioni indipendenti),
    tiene solo codici ISO canonici; la somma dei paesi **non deve** essere
    esattamente 100 (i pesi paesi sono informativi): quando i countries sono
    forniti il backend ricalcola e persiste le regions dalla mappatura
    paese→regione, imputando il residuo (100 − somma) nella regione
    `Other / Not Classified` (coerenza automatica a somma 100 per le regioni).
  - JustETF ora salva i countries raw (normalizzati a codici ISO) invece di
    scartarli, e il backend li archivia in `asset_country_weights`.
- **B.14 — Morningstar come fonte esposizione** (#59):
  - **python-service**: nuovo endpoint `GET /api/v1/etf/{isin}/morningstar-exposure`
    (modulo `python-service/app/morningstar.py`, **resolver custom senza mstarpy**).
    Il dominio `global.morningstar.com` delle SAL API informali usate da mstarpy è
    hard-blocked (403 "Request blocked") dall'IP di questo ambiente; il resolver
    esegue un **bootstrap Chromium headless** (installato nell'immagine via apt:
    `chromium`, `chromium-driver`, `xvfb`, `fonts-liberation`; chromedriver passato
    esplicitamente perché Selenium Manager non supporta `linux/aarch64`). Il browser
    carica `www.morningstar.com` finché il challenge AWS WAF non si risolve, legge i
    cookie di sessione e il **Bearer JWT** da `/api/v2/stores/maas/token` (cache in
    memoria, scadenza ~1h da `exp` del JWT). Le chiamate dati avvengono poi via
    `requests` con Bearer+cookie verso
    `www.us-api.morningstar.com/sal/sal-service/etf/...`:
    `portfolio/v2/sector/{sid}/data` (settori), `portfolio/regionalSectorIncludeCountries/{sid}/data`
    (paesi), `portfolio/regionalSector/{sid}/data` (regioni ufficiali),
    ISIN→securityId via `www.morningstar.com/api/v2/search?q={isin}`.
    Parser: paesi camelCase→nomi leggibili, settori dal bucket col peso maggiore
    (EQUITY per fondi azionari, FIXEDINCOME per obbligazionari; chiavi non-GICS come
    `cashAndEquivalents`/`government` scartate), regioni mappate sui nomi canonici
    (`_MORNINGSTAR_REGION_NAMES`). I pesi paesi sono tenuti come
    riportati (Morningstar espone la lista paesi completa, 51 voci, ma con una
    quota residuale "Other" non esposta come paese → la somma è ~95%): il residuo
    confluisce nella regione `Other / Not Classified` lato backend. L'auto-resolve
    ticker→ISIN è **market-aware** (mappa suffisso→exchange, es. `.MI`→`XMIL`):
    `search_etf` usa Morningstar per i ticker con suffisso riconosciuto e JustETF
    per i ticker senza suffisso. Variabili env opzionali
    `MORNINGSTAR_BEARER`/`MORNINGSTAR_COOKIES` per bypassare il browser nei test.
  - **Backend Go**: interfaccia `ETFFetcher` estesa con `FetchMorningstarExposure`;
    nuovo metodo `Service.FetchMorningstarExposure`; nuovo handler + rotta
    `POST /assets/{id}/fetch-morningstar-exposure` (solo ETF, auto-risolve ISIN).
    Quando il python-service restituisce le regioni ufficiali Morningstar, queste
    vengono usate direttamente (ordine canonico); altrimenti si ricade sulla
    derivazione da paesi. Nuovo endpoint **`POST /assets/{id}/exposure/derive`**
    (`{countries}` → `{regions}`, senza persistenza): usa la stessa
    `AggregateRegions` con il residuo in `Other / Not Classified`.
  - **Frontend**: la pagina asset detail è riorganizzata in **due card** —
    "Distribuzione geografica" (top 15 paesi a barre orizzontali + donut
    regioni) e "Distribuzione settoriale" (donut settori) — e l'editing è
    separato in **due modali**: `ExposureGeoModal` (regioni + paesi) e
    `ExposureSectorModal` (settori, prefill JustETF/Yahoo). I pulsanti di prefill
    seguono la fonte dei dati:
    - box **paesi**: **"Prefill da JustETF"** (JustETF scarica i paesi) e
      **"Prefill da Morningstar"** (paesi [+ settori]);
    - box **regioni**: **"Calcola da paesi"** (`POST /assets/{id}/exposure/derive`,
      ricalcola le regioni dai paesi correnti senza salvare) e
      **"Prefill da Morningstar"** (regioni ufficiali Morningstar);
    - box **settori**: "Prefill da JustETF" e "Prefill da Yahoo".
    Modifica add/remove
    dei paesi dalla lista canonica supportata; display usa nomi paese amichevoli
    da `frontend/src/lib/countryNames.ts`.
- **Verifica**: Go build/vet/test green; python-service pytest (51 test) green;
  `svelte-check`/eslint clean; e2e su stack isolato `vaultlab-test` con
  `make test-e2e` (18 PASS, 0 FAIL). Morningstar verificato end-to-end sullo stack
  test: `POST /assets/{id}/fetch-morningstar-exposure` per SMEA restituisce paesi
  canonici (zero-fill, somma raw ~95%) + settori GICS (somma 100) + **regioni
  ufficiali** (es. United Kingdom 21.27, Europe Developed ex-UK 75.85 per un ETF
  europeo); `POST /assets/{id}/exposure/derive` verifica la nuova tassonomia
  (GB→United Kingdom, JP→Japan, TW+KR→Asia Developed, residuo→Other).

### Altri EPIC Fase 2
- EPIC G.7 (#53) — Asset con ticker non-Yahoo: no richiesta prezzo e no errori
  - Campo `price_source` su `assets` (`yahoo`/`manual`/`none`, default `yahoo`):
    migrazione `000015`, modello, repo, validazione service, handler PATCH.
  - Il worker e `RefreshStale` filtrano solo asset `yahoo`; gli asset `manual`/
    `none` sono saltati del tutto (nessuna chiamata Yahoo, nessun errore di health).
  - Implementata: backend (model, repo, service, handler, price fetcher, docs).
- F.9 (#52) — Chart storico asset: zoom in-place + selettore YTD
  - `PriceChart` carica sempre tutto lo storico; i selettori 1M/3M/1Y/YTD/MAX
    fanno zoom in-place (coppia `start`/`end` percentuali) senza ricaricare dati.
  - Uno zoom/spostamento manuale deseleziona il pulsante attivo e preserva la vista.
- F.10 (#64) — Pagina asset: grafici + modale di modifica esposizione
  - **Ristrutturata dopo B.13/B.14**: l'unica card "Distribuzione" con una sola
    `ExposureModal` è diventata **due card** — "Distribuzione geografica" (paesi
    a barre + donut regioni) e "Distribuzione settoriale" (donut settori) — con
    **due modali separate**: `ExposureGeoModal.svelte` (regioni + paesi) e
    `ExposureSectorModal.svelte` (settori). I prefill vivono solo nelle modali e
    popolano una dimensione alla volta: JustETF → regioni; Morningstar → paesi
    (+ settori); Yahoo → settori.
- Splits sul chart asset — nuovi `GET /assets/{id}/splits` (service `AssetSplits`,
  handler) e `markLine` viola etichettati con il rapporto sul `PriceChart`,
  come nel `PositionChart` del portafoglio.
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
