# VaultLab — Piano di Sviluppo

## 1. Vision & Architettura

**VaultLab** è una webapp self-hosted pensata per homelab, multi-utente, per il tracciamento di investimenti e finanze personali.

### Stack tecnologico

| Livello | Tecnologia | Motivazione |
|---------|-----------|-------------|
| **Backend** | Go 1.23 (con Chi) | Performance, binary singolo, container minimale, ideale per homelab |
| **Frontend** | SvelteKit 2 + Svelte 5 + TypeScript + Vite | SPA moderna, runes API, routing basato su file |
| **Database** | PostgreSQL 16 | Dati finanziari relazionali, CTE per statistiche |
| **Cache/Jobs** | Redis 7 | Rate-limiting Yahoo, caching prezzi |
| **Container** | Docker/Podman + Compose | Homelab standard |
| **Auth** | JWT + refresh token | Self-hosted, no dipendenze esterne |
| **Grafici** | ECharts 5 | ROI, trend, distribuzione |
| **API Design** | RESTful | |

### Perché Go?
- Binary unico, basso consumo di RAM/CPU (ideale per homelab)
- Build veloci, deploy semplice
- Ottimo supporto concorrenza per fetch prezzi multi-fonte

---

## 2. Roadmap (fasi)

### FASE 0 — Setup progetto
- [x] Struttura repository (monorepo con backend Go + frontend SvelteKit)
- [ ] Docker + docker-compose con Postgres + Redis
- [ ] CI/CD base (GitHub Actions per build e test)
- [ ] Task runner / Makefile per comandi comuni

### FASE 1 — Core: Auth & Gestione Investimenti
- [ ] Modello dati: User, Portfolio, Asset, Transaction
- [ ] Registrazione/Login multi-utente (JWT)
- [ ] CRUD portafogli e transazioni (acquisto/vendita)
- [ ] Integrazione prezzi via API esterne (Yahoo Finance, Alpha Vantage, ecc.)
- [ ] Dashboard di base: valore portafoglio, gain/loss

### FASE 2 — Statistiche & Visualizzazioni
- [ ] ROI per asset e per portafoglio
- [ ] Distribuzione territoriale (per sede legale dell'asset)
- [ ] Distribuzione per categoria industriale (GICS)
- [ ] Grafici: andamento storico, composizione portafoglio
- [ ] Report periodici (mensile/trimestrale)
- [ ] Storico tassi di cambio (FX history, per-date nei series)
- [x] Pagina dettaglio asset (metadati, storico prezzi, distribuzioni geo/settoriali) — **EPIC B.10 (#45)**
- [ ] Microservizio Python per metadata ETF (JustETF scraping) — **EPIC B.5 (#11)**
- [ ] Endpoint allocazione geografica (weighted sum by region) — **EPIC B.6 (#12)**
- [ ] Endpoint allocazione settore (weighted sum by GICS) — **EPIC B.7 (#13)**

### FASE 3 — Multi-tenancy & Family Sharing
- [ ] Gestione permessi: utenti con ruoli (viewer, editor, admin)
- [ ] Condivisione portafogli tra familiari
- [ ] Viste aggregate famiglia

### FASE 4 — Finanza Personale (estensioni future)
- [ ] Tracciamento spese / categorie
- [ ] Budget mensile
- [ ] Risparmi e obiettivi
- [ ] Reportistica finanziaria unificata

### FASE 5 — Produzione Homelab
- [ ] reverse proxy (Traefik / Caddy) con SSL
- [ ] Backup automatico DB
- [ ] Healthcheck e monitoring
- [ ] Documentazione deploy

---

## 3. Modello Dati (bozza)

### Core
```
User         → id, email, name, password_hash, role, created_at
Portfolio    → id, user_id, name, description, currency, created_at
Asset        → id, isin, ticker, name, type, asset_class, country, exchange, currency, sector, industry
Transaction  → id, portfolio_id, asset_id, type (buy/sell), quantity, price, date, fees, notes
Price        → id, asset_id, date, open, high, low, close, volume, source
FxHistory    → base_currency, quote_currency, date, rate, source
AssetRegion  → asset_id, region, weight, source
AssetSector  → asset_id, sector, weight, source
```

### Finanza (Fase 4)
```
Expense      → id, user_id, category_id, amount, date, description, recurring
Budget       → id, user_id, category_id, amount, period (monthly/yearly)
Goal         → id, user_id, name, target_amount, current_amount, deadline
```

---

## 4. Principi di design

1. **Privacy-first**: tutto rimane in homelab, niente dato esce
2. **API-first**: ogni funzionalità backend è accessibile via API
3. **Tutto containerizzato**: `docker compose up` per far partire tutto
4. **Minimal dependencies**: poche librerie esterne, facile da mantenere
5. **Offline-resilient**: gestione gracevole quando le fonti prezzi non rispondono
6. **Mobile-friendly**: interfaccia responsive (PWA opzionale)

---

## 5. Struttura directory

```
vault-lab/
├── docker-compose.yml
├── Makefile
├── backend/
│   ├── cmd/
│   │   ├── server/main.go
│   │   └── worker/main.go
│   ├── internal/
│   │   ├── auth/        # JWT, middleware
│   │   ├── handler/     # HTTP handlers
│   │   ├── model/       # Struct/entity
│   │   ├── repository/  # DB queries
│   │   ├── service/     # Business logic
│   │   ├── price/       # Price fetcher (Yahoo, etc.)
│   │   ├── geo/         # Macro-regioni, settori GICS, mappature paese/regione
│   │   ├── position/    # AVCO engine
│   │   └── series/      # Materialized daily series
│   ├── migrations/      # SQL migrations
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── app.html
│   │   ├── app.css
│   │   ├── lib/
│   │   │   ├── components/   # Svelte 5 components
│   │   │   ├── stores/       # State management (runes)
│   │   │   ├── services/     # API client (fetch + JWT refresh)
│   │   │   └── format.ts     # Number/date formatters
│   │   └── routes/           # SvelteKit file-based routing
│   │       ├── +page.svelte  # Dashboard
│   │       ├── login/
│   │       ├── assets/
│   │       ├── portfolios/
│   │       └── settings/
│   ├── svelte.config.js
│   ├── vite.config.ts
│   ├── Dockerfile
│   └── package.json
└── docs/
    ├── BACKEND-GUIDE.en.md
    ├── BACKEND-GUIDE.it.md
    ├── DATABASE-GUIDE.en.md
    └── DATABASE-GUIDE.it.md
```

---

## Stato attuale (28 Ago 2026)

**Release v0.1.0** pubblicata su `main` (prima release ufficiale).

Fase 0 e Fase 1 completate (incluso EPIC A — data correctness & security). Lo sviluppo attivo
procede su `develop` (feature branch `feat/B.10-asset-detail` per la parte EPIC B completata).
Realizzate le parti di EPIC B afferenti alla **pagina dettaglio asset** (#45) e ai relativi
layer dati/backend (asset meta, exposure geo/settore, storico prezzi completo). Restano da
completare in EPIC B: B.5 (microservizio JustETF), B.6/B.7 (endpoint allocazione geo/settore),
B.8 (chart dashboard/portfolio), B.9 (FX history) e il seed GICS/populate `assets.sector` (B.2/B.3).
Fatte intanto B.11/B.12 (asset class: colonna `asset_class`, auto-detect Yahoo, endpoint e donut
"allocazione per classi") e rimossa la vecchia classificazione `category_id`/`categories`.
Poi EPIC C (metric di rischio), EPIC D/E (design system e pagine dominio), e i rimanenti item
di condivisione/CSV della Fase 1. Vedi STATUS.md per lo stato dettagliato.

