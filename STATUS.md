# VaultLab — Stato Progetto (30 Lug 2026)

## Infrastruttura

| Servizio | Stack | Note |
|----------|-------|------|
| Backend | Go 1.23 + Chi + pgx + golang-migrate | Containerizzato |
| Frontend | React 19 + Vite + TypeScript + Tailwind + Recharts | Containerizzato (nginx) |
| Database | PostgreSQL 16 | Con docker volume |
| Cache | Redis 7 | Per sessioni future |
| Worker | Go (prezzi) | Container separato |
| Container | podman + podman-compose su macOS | |

## Fase 0 — ✅ Completata

- [x] Struttura monorepo (backend + frontend + docker)
- [x] Docker Compose (Postgres, Redis, backend, frontend, worker)
- [x] Go module con tutte le dipendenze
- [x] Frontend (Vite + React + TypeScript + Tailwind)
- [x] Auth JWT (access + refresh token, rotazione automatica)
- [x] API REST (18 endpoint: auth, assets, portfolios, transactions, prices)
- [x] DB schema (users, assets, categories, portfolios, portfolio_shares, transactions, prices)
- [x] AuthContext per stato utente globale
- [x] Login/logout con `window.location.replace()` (bypassa React Router)
- [x] Makefile aggiornato per podman-compose

## Fase 1 — 🟡 In corso

### ✅ Completato
- [x] Registrazione/Login multi-utente
- [x] CRUD portafogli
- [x] CRUD asset con autocomplete Yahoo Finance (`/api/v1/assets/lookup?q=`)
- [x] Transazioni (buy/sell/dividend) con JOIN asset per mostrare ticker + nome
- [x] Dashboard: valore totale, gain/loss, allocazione (grafico a torta), performance (line chart), ROI per asset
- [x] Valuta dinamica (€/$/£/CHF) basata sul portafoglio
- [x] Formattazione importi e percentuali centralizzata (`src/lib/format.ts`)

### ❌ Da fare
- [ ] `POST /api/v1/prices/refresh` — endpoint per aggiornamento prezzi
- [ ] Inserimento prezzo manuale nella UI
- [ ] Sharing portafogli tra utenti (portfolio_shares)
- [ ] Import CSV transazioni
- [ ] Ricerca asset con autocomplete nel form transazioni

## Problemi Aperti

### Yahoo Finance rate-limited (429)
Il worker non riesce a fetchare i prezzi perché Yahoo blocca le richieste dal container podman.
- `v7/finance/quote` → 429 Too Many Requests
- `v8/finance/chart` → 429 Too Many Requests
- `v1/finance/search` → funziona (ma non dà prezzi)
- Da macOS host → funziona
- Soluzione proposta: script esterno su macOS (curl/cron) che fetcha prezzi e li POSTa al backend,
  oppure passare a API alternativa (Finnhub, Alpha Vantage con API key)

### Worker non logga errori
Il worker non ha prodotto log dopo "price worker started" — possibile hang su Yahoo rate-limit.
Da investigare: aggiungere timeout più esplicito e log prima/dopo ogni batch.

### Ticker europei
Asset europei su Yahoo usano suffisso exchange (es. VWCE.DE, VWCE.AS). L'utente deve sapere il
ticker corretto. Da documentare o aggiungere selezione exchange nell'autocomplete.

## Fase 2 — Pianificata

- Backtest engine (gonum)
- Simulazione Monte Carlo
- Statistiche avanzate (Sharpe ratio, drawdown, beta, alpha)
- Distribuzione territoriale e per settore GICS
- Report periodici

## Fase 3 — Pianificata

- Multi-tenancy familiare
- Gestione permessi e condivisione

## Fase 4 — Futura

- Tracciamento spese e categorie
- Budget mensile
- Obiettivi di risparmio

## Comandi Utili

```bash
make up              # Avvia tutto con podman-compose
make logs            # Log in tempo reale
make migrate         # Esegui migration DB
make frontend-dev    # Sviluppo frontend con hot-reload
# Per ricreare container dopo modifiche:
podman-compose stop <service>
podman rm <container>
podman-compose up -d --build <service>
```
