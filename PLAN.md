# VaultLab — Piano di Sviluppo

## 1. Vision & Architettura

**VaultLab** è una webapp self-hosted pensata per homelab, multi-utente, per il tracciamento di investimenti e finanze personali.

### Stack tecnologico proposto

| Livello | Tecnologia | Motivazione |
|---------|-----------|-------------|
| **Backend** | Go (con Chi/Gin) | Performance, binary singolo, container minimale, ideale per homelab |
| **Frontend** | React + TypeScript + Vite | SPA moderna, ecosistema ricco di librerie per grafici |
| **Database** | PostgreSQL | Dati finanziari relazionali, CTE per statistiche, estensione TimescaleDB opzionale |
| **Cache/Jobs** | Redis | Sessioni, code per aggiornamento prezzi |
| **Container** | Docker + docker-compose | Homelab standard |
| **Auth** | JWT + refresh token | Self-hosted, no dipendenze esterne |
| **Grafici** | Recharts / D3.js | ROI, trend, distribuzione |
| **API Design** | RESTful (poi GraphQL opzionale) |

### Perché Go?
- Binary unico, basso consumo di RAM/CPU (ideale per homelab)
- Build veloci, deploy semplice
- Ottimo supporto concorrenza per fetch prezzi multi-fonte

---

## 2. Roadmap (fasi)

### FASE 0 — Setup progetto
- [ ] Struttura repository (monorepo con backend Go + frontend React)
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
Asset        → id, isin, ticker, name, category_id, country, currency
Category     → id, name, sector (GICS classification)
Transaction  → id, portfolio_id, asset_id, type (buy/sell), quantity, price, date, fees, notes
Price        → id, asset_id, date, open, high, low, close, volume, source
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

## 5. Struttura directory proposta

```
vault-lab/
├── docker-compose.yml
├── Makefile
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/        # JWT, middleware
│   │   ├── handler/     # HTTP handlers
│   │   ├── model/       # Struct/entity
│   │   ├── repository/  # DB queries
│   │   ├── service/     # Business logic
│   │   └── price/       # Price fetcher (Yahoo, etc.)
│   ├── migrations/      # SQL migrations
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/    # API client
│   │   └── hooks/
│   ├── Dockerfile
│   └── package.json
└── docs/
    └── ARCHITECTURE.md
```

---

## Stato attuale (25 Ago 2026)

**Release v0.1.0** pubblicata su `main` (prima release ufficiale).

Fase 0 e Fase 1 completate (incluso EPIC A — data correctness & security). Lo sviluppo attivo procede su `develop`.
Prossime aree (vedi STATUS.md): EPIC B (esposizione geografica/settore), EPIC C (metric di rischio), EPIC D/E (design system e pagine dominio), e i rimanenti item di condivisione/CSV della Fase 1.

