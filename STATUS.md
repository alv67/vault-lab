# VaultLab — Stato Progetto (17 Set 2026)

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
**v0.3.0** — terza release su `main` (11 Set 2026): asset editing overhaul (price_source
Yahoo/Manual/None, chart in-place con YTD e marcatori split, modale esposizione) ed editing
dell'esposizione per-paese con fonte Morningstar/JustETF, cache e provenienza.
**v0.4.0** — quarta release su `main` (13 Set 2026): design system & dark mode (EPIC D),
rebuild delle pagine e dei componenti di dominio (EPIC E: asset/portafogli/modali/login/
impostazioni a tab), dashboard e dettaglio portafoglio rinnovati, Health più chiaro
(periodo Today/24h/100, paginazione, fix N/A) e CI GitHub Actions.
**v0.5.0** — quinta release su `main` (17 Set 2026): **EPIC I completa** (dashboard & portfolio
v2) — valuta base utente con aggregazione FX, riepilogo attivo/chiuso, grafico performance
time-weighted (barre mensili/annuali + linea cumulata) e grafico del capitale, allocazioni per
classe/settore/paese/macro-regione, tabella asset investiti consolidata, KPI e allocazioni del
dettaglio portafoglio allineati alla dashboard, performance a barre e transazioni paginate;
inclusi i fix import di export vecchi (#99) e P/L fittizio -100% sulle posizioni chiuse (#100).
Flusso: branch → PR su `develop` → merge → tag `v0.1.x`/`v0.2.0`/`v0.3.0`/`v0.4.0`/`v0.5.0` su `main`.

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
  persistito sull'asset), poi scarica paesi/regioni + settori e li restituisce come
  **anteprima non persistente** (vedi «Redesign modale distribuzione geografica»: la
  persistenza avviene solo via `PUT /assets/{id}/exposure`). Config `VAULT_PYTHON_SERVICE_URL`;
  servizio `python-service` presente
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
  (`POST /assets/{id}/fetch-etf-exposure`) che scarica regioni+settori come
  **anteprima** (non persiste; salvataggio solo col PUT) e sincronizza l'**ISIN**
  risolto nel form; campo ISIN anche nel form di
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
    tenere solo codici ISO canonici; la somma dei paesi **non deve** essere
    esattamente 100 (i pesi paesi sono informativi): quando i countries sono
    forniti il backend ricalcola e persiste le regions dalla mappatura
    paese→regione, imputando il residuo (100 − somma) nella regione
    `Other / Not Classified` (coerenza automatica a somma 100 per le regioni).
    _(Auto-derivation al save poi **rimossa** dal redesign modale: vedi
    «Redesign modale distribuzione geografica (paesi-first)」.)_
  - JustETF ora espone i countries raw (normalizzati a codici ISO) invece di
    scartarli; il backend li include nella preview e li archivia in
    `asset_country_weights` solo al salvataggio (`PUT`).
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
- **Resilienza WAF Morningstar (fix post-#59)**: la WAF può revocare la
  sessione cacheata prima della scadenza del JWT e rispondere con una pagina
  HTML di challenge (JSON non decodificabile → 502 opaco "upstream fetch
  failed"). Ora l'intero flow (ISIN→securityId + le tre chiamate SAL + parsing)
  viene **ritentato una volta** con credenziali fresche:
  `_invalidate_bootstrap_cache`/`_retry_on_waf_challenge` in
  `python-service/app/morningstar.py`, con nuovo errore `MorningstarWafError`.
  Se la challenge persiste, `morningstar-exposure` risponde 502 con messaggio
  esplicito ("Morningstar WAF challenge could not be passed for ISIN …; try
  again in a few seconds") e l'auto-resolve ticker→ISIN degrada a JustETF.
  Gli altri errori di trasporto restano mappati come prima (nessun retry).
  pytest aggiornato a 57 test, tutti verdi.
- **Resilienza bootstrap Chrome (fix post-#59, oltre al retry WAF)**: bootstrap
  falliti lasciano processi Chromium/chromedriver/crashpad orfani che bloccano
  ogni sessione successiva ("session not created: Chrome instance exited",
  zombie `<defunct>` persistenti nel container). Il ramo browser di
  `_session_credentials` ora ritenta **una sola volta** dopo pulizia best-effort
  dei processi orfani (`_kill_stale_browsers`, solo `pkill` — mai glob su /tmp);
  al secondo crash risponde 502 con messaggio chiaro ("Chrome headless could not
  start in the sandbox; try again in a few seconds"). pytest aggiornato a 62
  test, tutti verdi.
- **Fallback multi-listing Morningstar (fix #76)**: il resolver usava solo il
  **primo** securityID restituito da `/api/v2/search`; se quel listing non ha
  dati SAL (risposta `206`/`Can't get SecurityInfo`, es. la quotazione XAMS di
  `IE00B5L8K969`/CSEMAS) il JSON non era decodificabile e il caso veniva
  scambiato per challenge WAF → re-bootstrap del browser (con processi Chromium
  orfani lasciati indietro) e 502 fuorviante "WAF challenge could not be
  passed", nonostante gli altri listing dello stesso ISIN avessero dati validi.
  Ora `_search_security_ids` restituisce **tutti** i securityID dei fondi che
  matchano l'ISIN (in ordine di ricerca, deduplicati) e `_fetch_exposure_once`
  li prova in sequenza fermandosi al primo con paesi validi; il nuovo errore
  non ritentabile `MorningstarSecurityUnavailable` fa passare al candidato
  successivo **senza** re-bootstrap, e se nessun listing ha dati l'errore è un
  chiaro `MorningstarDataError` ("No SAL data found for ISIN … on any of N
  Morningstar listings"), mai `MorningstarWafError`. Solo una pagina HTML di
  challenge (JSON non decodificabile senza marcatore missing-data) resta
  soggetta al retry con sessione fresca. pytest aggiornato a 70 test, tutti
  verdi.
- **Cache esposizione provider (post-B.14)**: `FetchETFExposure` e
  `FetchMorningstarExposure` cachano il payload grezzo del provider in Redis
  (chiave `vl:lookup:exposure:<source>:<ISIN>`, TTL `VAULT_EXPOSURE_CACHE_TTL`
  default 7 giorni): la prima richiesta su un ISIN esegue il fetch pesante, le
  successive rispondono dalla cache; `?refresh=1` forza il refetch e riscrive
  la cache (risultati senza paesi mai cachati; Yahoo `fetch-exposure`
  invariato). 4 nuovi test service (`TestFetch*Cache*`/`EmptyResultNotCached`).

### Redesign modale distribuzione geografica (paesi-first)
- **Frontend** (`ExposureGeoModal`, `ExposurePie`, `ProvenanceBadge`): la modale
  passa a **due colonne** — **Paesi a sinistra**, **Regioni a destra** (impilati
  paesi-primo su mobile).
  - **Paesi**: lista a **partenza vuota** (non più 89 righe zero-filled); ogni
    riga è `codice · nome · barra orizzontale · input peso · cestino`, ordinata
    per peso desc (riordino su add/remove/blur, mai mentre digiti — la barra si
    anima live); select + "Aggiungi" per inserire codici canonici; footer con
    "Totale X%" + meter; **save disabilitato se somma > 100** (sotto 100 ok);
    **nessun donut**.
  - **Regioni**: **tabella fissa delle 10 regioni canoniche** (niente
    add/remove; riga "Other / Not Classified" **rimossa** — filtrata alla
    pagina, non entra mai in `regionsEdit`); **donut aperto** con la nuova prop
    `complete={false}` su `ExposurePie` (una fetta residua trasparente mantiene
    veritieri gli angoli quando la somma < 100, invece della fetta grigia
    Other); **save disabilitato se somma > 100** (sostituisce la vecchia regola
    100 ± 0,5).
  - **Badge di provenienza** (`ProvenanceBadge`): pillola con puntino colorato
    per fonte — `manuale`/`da JustETF`/`da Morningstar`(` regioni ufficiali`)/
    `calcolato dai paesi`/`da JustETF via paesi`; prefill/derive impostano il
    badge, ogni modifica manuale lo riporta a "manuale". Stato session-scoped
    (niente persistenza), badge nascosto a reload.
  - Coerenza: anche la **card geografica** della pagina usa `complete={false}`
    e filtra Other da donut/legenda regioni.
  - **Salvataggio paesi**: il backend **non ricalcola più le regioni** al save
    dei `countries` (auto-derivation rimossa); le regioni memorizzate restano
    invariate (badge di provenienza intatto) e si aggiornano solo manualmente
    con «Calcola da paesi».
- **Backend**: `SaveAssetExposure` accetta ora **regioni con somma ≤ 100** e,
  quando il client non invia Other, **inieietta il residuo in
  `Other / Not Classified`** prima di persistere (invariante «regioni=100»
  preservata per l'aggregazione portafoglio); **countries rifiutano somma
  > 100.5** (nessun minimo, sotto 100 ok); settori invariati (100 ± 0,5).
  Auto-derivation al save dei paesi **rimossa**: ogni dimensione del body `PUT`
  è salvata indipendentemente (logica estratta in `saveExposureDimensions`) e
  le regioni si ricalcolano solo via `POST /assets/{id}/exposure/derive` o
  salvando regioni esplicite.
  `derive` invariato (restituisce Other; filtra il frontend).
- **Prefill/fetch = anteprime non persistenti**: i tre endpoint di fetch
  (`POST /assets/{id}/fetch-exposure`, `/fetch-etf-exposure`,
  `/fetch-morningstar-exposure`) **non salvano più** le dimensioni di
  esposizione (countries/regions/sectors) né i campi profilo dell'asset:
  restituiscono una **preview canonica** costruita dai dati del provider
  (per `fetch-exposure` le dimensioni geo memorizzate restano lette dallo
  stato attuale). L'unica eccezione concordata resta l'**ISIN auto-risolto**,
  che continua a essere persistito sulla tabella `assets` (con `bumpRev`).
  La persistenza avviene **solo** con `PUT /assets/{id}/exposure` (pulsanti
  Salva); `POST /assets/{id}/exposure/derive` era e resta non persistente.
  Smoke e2e aggiornato: `tests/test-epic-b.sh` FASE 4 fa ora un PUT esplicito
  dei valori in preview dopo il fetch. Test:
  `TestFetchETFExposure_PreviewDoesNotPersist` / `...PersistsOnlyResolvedISIN`,
  `TestFetchMorningstarExposure_PreviewDoesNotPersist` /
  `...PersistsOnlyResolvedISIN`, `TestFetchAssetExposure_PreviewDoesNotPersist`.
- **Verifica**: Go build/vet/test green (nuovi `TestPrepareRegions`/
  `TestPrepareCountries`/`TestValidateExposureWeights`/
  `TestSaveExposureDimensions_*`); pytest 51;
  `svelte-check`/eslint/build clean; `make test-e2e` 18 PASS 0 FAIL. Test
  manuale PUT: regions somma 95 → 200 con Other=5 persistito (somma 100);
  regions 103 → 400; countries 120 → 400; countries 95 → 200 con regioni
  memorizzate invariate (nessuna riscrittura).

### Provenienza persistita dell'esposizione (follow-up B.13/B.14)
- **Backend**: nuova tabella `asset_exposure_provenance (asset_id, dimension, source, updated_at)`
  (migrazione `000017`). `PUT /assets/{id}/exposure` accetta `countries_source` /
  `regions_source` / `sectors_source` (default `manual`) e registra la provenienza di
  ogni dimensione salvata; `GET`/`PUT /assets/{id}/exposure` rispondono la mappa
  `provenance` (`omitempty`) così i badge "da Morningstar (2026-09-05)" dell'UI
  sopravvivono al reload. Gli fetch/prefill restano anteprime senza provenienza.
  3 nuovi test service + fake aggiornato; Go build/vet/test green.

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

### EPIC D — Design system & dark mode (#37) — ✅ Completata
Branch unico `feat/D-design-system`, 5 commit:
- `feat(tokens)` — token semantici (CSS custom properties HSL in `app.css` mappate in `tailwind.config.js` con `<alpha-value>`), store tema a 3 modalità (light/dark/system, **default dark**), script anti-FOUC in `app.html`, `lib/chartTheme.ts` (temi ECharts `vaultlab-light`/`vaultlab-dark`) e `lib/chartPalette.ts` (palette risolta a runtime).
- `feat(ui)` — sweep dei colori hardcoded (~333 classi palette + hex) verso i token su tutte le pagine/componenti; grafici dark-aware.
- `feat(ui)` — primitive in `src/lib/components/ui/` (Button, Input, Field, Select, Card, Badge, Modal, ConfirmDialog, Spinner, Skeleton, EmptyState, Table, SegmentedControl, StatCard) + refactor Toaster/ProvenanceBadge; i 4 `confirm()` nativi sostituiti da `ConfirmDialog`.
- `feat(shell)` — `AppShell` responsive: sidebar collassabile (stato persistito in localStorage), header sticky, UserMenu, ThemeToggle a 3 modalità, MobileDrawer accessibile; `Layout.svelte` rimosso.
- `feat(theme)` — dark mode di default + rimozione della pagina dev `/settings/theme-tokens`.
- Fix: label dei donut leggibili in dark (le label ECharts non ereditavano il `textStyle` del tema → fill scuro + bordo bianco).
- Documentazione: `docs/FRONTEND-GUIDE.en/it.md` (styling/tema, chart, layout) e `docs/RELEASE-NOTES.en/it.md`.

### Ondata 0 — Fix & infrastruttura (13 Set 2026)

- H.1 (#32) — GitHub Actions: `.github/workflows/ci.yml` con due job (backend: `go build` + `go vet` + `go test`; frontend: `npm ci` + `npm run check` + `npm run lint`) su push e pull request verso `develop`/`main`. Fase 0 completa.
- H.4 (#46) — Price sync health: campo `has_data` nel summary; la card Success Rate mostra **N/A** (niente più `NaN%`) quando non ci sono eventi nel periodo; `formatRate` robusto a null/undefined/NaN.
- H.5 (#47) — Logging API: `FetchIssue` con `request_type` (`chart`/`spark`/`search`/`fx`) e `asset_id`; `HealthEvent.AssetID` popolato sugli eventi per-asset; messaggi di successo con l'elenco dei ticker; registrazione degli eventi `search` (lookup) prima assenti.
- E.9 (#71) — Le allocazioni del portafoglio (classi, geo, settori) vengono rifetchate dopo create/update/delete di una transazione, senza reload.
- H.2 (#33, parziale) — Nuovi unit test per le allocazioni backend: `GetPortfolioAllocation` (multi-valuta, FX mancante, skip qty/prezzo) e `GetPortfolioClassAllocation` (raggruppamento/ordinamento, skip FX), più allocazione settoriale ETF (`backend/internal/service/allocation_test.go`).
- H.9 (#90) — Asset non-Yahoo (`manual`/`none`) non più interrogati per history/split: filtro `price_source` in `GetPortfolioHistory`, `SyncAssetData`/`syncAssets` e `BackfillAssetHistory` (no-op). Niente più eventi health `history_fetch`/`split_fetch` per questi asset.
- H.11 (#93) — Health summary calcolato dal DB (`health_events`) su finestra selezionabile **Today / Last 24h / Last 100 events** (rimossi i contatori Redis orari); `period` e `has_data` nella risposta, selettore nella pagina Health.
- H.10 (#91, parziale) — Paginazione della lista eventi Health (`limit`/`offset`, `events_total`; UI 50/pagina con Previous/Next). Restano: copertura di tutte le chiamate esterne (meta/profilo, JustETF/Morningstar, successi) e retention di `health_events`.

### Ondata 1 — EPIC E ✅ completata

- E.8 (#54) — Chart storico portafogli in dashboard: asse `time` con i punti di ogni portafoglio (niente più unione di date con `null` che spezzava le linee), `connectNulls` + `sampling: lttb`, `dataZoom` inside/slider come `PositionChart`. `PortfolioLineChart` non prende più la prop `data` (usato solo dalla dashboard).
- E.3 (#20) — Componenti di dominio riusabili: `AssetSearchAutocomplete`, `CurrencySelect`, `CreateAssetModal`, `CreatePortfolioModal`, `ImportPortfolioModal`. Pagine assets/portfolios ripulite dai form inline (Badge per il tipo, `ui/Button`/`ui/Card`/`ui/EmptyState`), login ridisegnato (logo, `SegmentedControl` Sign in/Register, `Field`/`Input`, validazione inline, conferma password in registrazione).
- E.1 (#18) — Dashboard ridisegnata: KPI `ui/StatCard` per valuta, `AllocationDonut` per portafoglio, portafogli come card cliccabili, `PositionTable` condiviso nell'accordion, `EmptyState`/`Spinner`. Backend: `finished_at` in `RefreshReport`, mostrato nell'header come "Prices updated".
- E.2 (#19) — Dettaglio portafoglio: `AssetCombobox`, `TransactionTable`, `AddTransactionModal` (form + validazione inline + totale live + delete con ConfirmDialog). KPI con `StatCard`, posizioni con `PositionTable` condiviso (con colonna Price), azioni in header sticky; refetch post-mutation (E.9) preservato.
- E.4 (#21) — Settings divisa in tab via subroute (`/settings` Profile, `/settings/password`, `/settings/currencies`, `/settings/health`) con `SettingsTabs`; cambio password con validazione inline e mappatura errori `401`/`400` sui campi. Sweep a11y/numerico: `ui/Th` con `scope="col"`, tabelle assets/health/valute con primitive `ui/Table`, `aria-label` sui bottoni icona, `tabular-nums`/allineamento a destra sui numeri.

### I.1 — Valuta base utente + aggregazione FX (branch `feat/I.1-base-currency`)
Enabler di EPIC I (sblocca I.2–I.5). La dashboard e le sue viste aggregate esprimono i
totali nella **valuta base dell'utente** (default `EUR`).
- **Backend**:
  - Migrazione `000018_users_base_currency`: colonna `users.base_currency` (default `EUR`).
  - `model.User`/`repository/user.go`: `base_currency` in SELECT/RETURNING/Update.
  - `UpdateProfile` accetta `base_currency` (opzionale, validata contro la whitelist
    `supported_currencies` enabled via `EnabledByCodes`; vuota = mantiene il valore).
  - `PATCH /users/me` accetta il campo `base_currency` (400 se non è una valuta abilitata).
  - `GET /dashboard` risponde ora `base_currency` + `summary` aggregato in valuta base
    (`invested`/`value`/`gain_loss`/`gain_loss_pct`/`realized`/`fx_missing_count`/
    `fx_missing_value`); le serie `history` sono convertite per-data in valuta base
    (`series.LoadDateRates` + `dateRates.Factor`, punti senza FX scartati). `by_currency`,
    `portfolios`, `assets` restano invariati (serviranno a I.2/I.5).
  - `GET /dashboard/allocation` ora esprime geo/settori nella valuta base (prima USD).
  - `series.LoadRates` esteso (variadic `extra ...string`) per includere le valute dei
    portafogli oltre a quelle degli asset + base + USD.
  - FX mancante gestito esplicitamente (escluso dai totali + conteggiato), coerente con
    `GetPortfolioSummary` (EPIC A).
- **Frontend**:
  - `api.ts`: `User.base_currency`, `Dashboard.base_currency`/`summary` (`DashboardSummary`),
    `authApi.updateProfile` con `base_currency?`.
  - `auth.svelte.ts`: `updateProfile(name, email, baseCurrency?)`.
  - Settings → Profile: `CurrencySelect` "Base currency" (popolata da
    `settingsApi.listCurrencies()`), salvata con `PATCH /users/me`.
  - Dashboard: KPI primari in valuta base (`summary`), ripartizione per-valuta mostrata
    solo quando ci sono più valute; donut "Allocation by portfolio" in valuta base.
- **Verifica**: Go build/vet/test green (7 nuovi test service: `UpdateProfile` base currency,
  `GetDashboard` summary multi-valuta + FX mancante + conversione history, `GetDashboardAllocation`
  in valuta base); `svelte-check`/eslint clean.

### I.2 — Dashboard attivo vs chiuso (branch `feat/I.1-base-currency`)
Il riepilogo dashboard (vault e per-portafoglio) separa ora le quote di investimento
**attive** da quelle **chiuse**, in valuta base a livello vault.
- **Backend**:
  - `position.State`: aggiunti i cumulati dei lotti chiusi `ClosedCost`/`ClosedCostCCY`
    (costo AVCO dei venduto), `Proceeds`/`ProceedsCCY` (incasso netto) e
    `Dividends`/`DividendsCCY`; `TxSell` e `TxDividend` li accumulano senza toccare la
    logica `Realized` esistente.
  - `model.Holding`: propagati i sei campi da `HoldingsDetailed`.
  - `model`: nuovi `ActiveBreakdown` (`invested`/`value`/`gain_loss`/`gain_loss_pct`) e
    `ClosedBreakdown` (`invested`/`proceeds`/`realized` = proceeds − invested, `dividends`
    separate dal capitale). `DashboardSummary` e `PortfolioPerformanceSummary` sostituiscono
    i campi flat I.1 con gli oggetti annidati `active`/`closed` (breaking per la UI, frontend
    da adeguare); `by_currency` e `assets` invariati.
  - `GetDashboard`: granularità per porzione di lotto (quantità vendute → chiuso, quantità
    residue → attivo); conversione per-importo con gli stessi criteri FX-missing di I.1
    (importo non convertibile escluso dai totali e contato in `fx_missing_count`/
    `fx_missing_value`, solo importi nonnulli).
- **Verifica**: Go build/vet/test green; nuovo `position_test.go` (venduto totale/parziale,
  dividendi separati, `Walk`) + test service `TestGetDashboard_ActiveClosedBreakdown` e
  `TestGetDashboard_SummaryInBaseCurrency` aggiornati alla forma annidata.
- **Documentazione**: `docs/BACKEND-GUIDE.en/it.md` (cap. 7/8, paragrafo valuta base) e
  `docs/RELEASE-NOTES.en/it.md`.

### I.3 — Dashboard: grafico performance + capitale (branch `feat/I.1-base-currency`)
- **Backend**: nuovo `GET /dashboard/performance?granularity=month|year` → bucket
  `{period, return, twr, invested, value}`. Rendimento **time-weighted (TWR)** puro: `V(d)` =
  valore di mercato (asset prezzati + bond non quotati portati al costo), flussi esterni
  `buy/sell/dividend/fee`, rendimento giornaliero composto; `invested` = capitale netto,
  `value` = market value. Rimosso il vecchio `history` dal dashboard.
- **Frontend**: card **Performance** (barre `return` % + linea TWR cumulata) e card **Capital
  invested** (`invested` vs `value`), toggle Mensile/Annuale (`PerformanceChart`,
  `CapitalChart`); rimosso `PortfolioLineChart`. Rimosse anche le righe per-valuta
  (`by_currency`) dalla dashboard.
- **Verifica**: `service_test.go` (TWR, liquidazione/riapertura, bond al costo, dividendi,
  multi-valuta) + `svelte-check`/lint.

### I.4 — Allocazione dashboard estesa (branch `feat/I.1-base-currency`)
- **Backend**: `GET /dashboard/allocation` esteso con `classes` (per asset class) e `countries`
  (per paese, equity-only, non-zero, descending), in valuta base.
- **Frontend**: card "Allocazione complessiva" → classi (donut, `ClassDonut`) + regioni/settori/
  paesi a **barre orizzontali** (`ExposureBarChart`); paesi con nome completo e ~10 righe
  visibili + scroll.

### I.5 — Dashboard: tabella asset investiti consolidata (branch `feat/I.1-base-currency`)
- **Backend**: `GET /dashboard` espone `invested_assets` (per asset, aggregato su tutti i
  portafogli, valuta base, sole posizioni aperte, ordinato per valore; asset senza prezzo al
  costo con `has_price=false`).
- **Frontend**: card **Invested assets** che sostituisce gli accordion per-portafoglio.

### I.6 — Dettaglio portafoglio: KPI allineati alla dashboard (branch `feat/I.1-base-currency`)
- **Backend**: `GET /portfolios/{id}/summary` espone `active`/`closed` (stessa forma della
  dashboard, in valuta portafoglio).
- **Frontend**: card condivisa **`InvestmentsTable`** (estratto) usata da dashboard e dettaglio;
  la vecchia riga KPI Value/Realized/Open G/L/Assets è sostituita dalla card Active/Closed +
  riga asset.

### I.7 — Dettaglio portafoglio: allocazioni come la dashboard (branch `feat/I.1-base-currency`)
- **Backend**: `GET /portfolios/{id}/allocation/geography` esteso con `countries` (equity-only,
  descending, valuta portafoglio).
- **Frontend**: sezione allocazione = classi (donut) + regioni/settori/paesi a barre
  (`ClassDonut`/`ExposureBarChart`); rimossi i componenti `GeographyChart`/`SectorChart`.

### I.8 — Dettaglio portafoglio: performance a barre (branch `feat/I.1-base-currency`)
- **Backend**: `GET /portfolios/{id}/performance/buckets?granularity=month|year` (TWR, valuta
  portafoglio, ownership + cache).
- **Frontend**: card **Performance** (barre % + linea TWR, toggle Mensile/Annuale); il
  `PositionChart` "Performance history" resta come vista secondaria.

### I.9 — Transazioni paginate (branch `feat/I.1-base-currency`)
- **Backend**: `GET /portfolios/{id}/transactions?limit=&offset=` → `{transactions, total,
  limit, offset}` (default 20, max 100, ordine `date DESC, created_at DESC, id DESC`,
  ownership check); script e2e aggiornati.
- **Frontend**: paginatore sotto la tabella (range + prev/next), refetch della pagina corrente
  dopo le mutazioni.

### Fix nella stessa PR
- **#99 — import export vecchi**: l'import non fallisce più se il documento non ha `price_source`
  (default `yahoo`), export esteso con `price_source`/`asset_class`, versione documento gestita.
- **#100 — posizioni chiuse e residui**: la riga Active non conta più il costo residuo (AVCO) di
  posizioni chiuse; arrotondamento degli importi. Niente più P/L fittizio -100%.

### EPIC I — stato
Tutte le sub-issue **I.1–I.9 completate** e rilasciate in **v0.5.0** (PR #98 mergiata su
`develop`/`main`). Nota: la gestione del **capitale disponibile / versamenti-prelievi** (conto
titoli) è tracciata a parte nell'issue **#101** e sarà una PR separata.

## EPIC J — Nuove asset class: bond, certificati, fondi pensione, conti deposito (#113) — pianificata

Estendere VaultLab agli **investimenti a reddito fisso e non quotati**: obbligazioni (tipo, cedola,
scadenza, esposizione geo/settoriale), certificati d'investimento (prodotti strutturati), piani
pensionistici complementari (fondi pensione/PIP) e conti deposito. L'analisi finanziaria
(30 Ago 2026) ha verificato che le quattro classi sono già tracciabili con il modello attuale
(`price_source` `manual`/`none`, TWR al costo), a patto di colmare due gap trasversali:
**inserimento prezzo manuale in UI** (assente) e tipo **`cash`** non selezionabile nel form di
creazione.

| Issue | Titolo | Componente | Priorità |
|-------|--------|------------|----------|
| #105 | J.1 — Inserimento prezzo manuale (endpoint + UI) | Backend + Frontend | MVP (sblocca tutte le classi) |
| #106 | J.2 — Metadati asset fixed income (scadenza, emittente, `attributes` JSONB) | Backend + Frontend | MVP |
| #107 | J.3 — Tipo `cash` in UI + nuovo tipo `certificate` | Backend + Frontend | MVP |
| #108 | J.4 — Esposizione geo/settoriale per fixed income (opt-in `exposure_kind`) | Backend | Post-MVP |
| #109 | J.5 — Maturazione interessi conti deposito | Backend + Frontend | Post-MVP |
| #110 | J.6 — Metriche bond: duration, YTM, current yield | Backend | Post-MVP |
| #111 | J.7 — Allocazione per merito di credito (`asset_credit_weights`) | Backend + Frontend | Post-MVP |
| #112 | J.8 — Wrapper piani pensionistici (comparti/sub-fondi) | Backend + Frontend | Post-MVP (bassa) |

**Prima PR consigliata**: J.1 + J.2 + J.3 insieme — set minimale e non regressivo che abilita
un'esperienza first-class per tutte e quattro le classi.

**Decisioni aperte** (da risolvere in implementazione): cedola vs dividendo come tipo a sé
(MVP: `coupon` = `TxDividend` + note); `asset_class` dei certificati (`other` vs nuovo
`structured`); policy di esposizione opt-in (`exposure_kind`) per non impattare i portafogli
esistenti; gestione prezzo clean/dirty per i bond (MVP: prezzo inserito usato as-is).

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
