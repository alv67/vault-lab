# VaultLab Backend — Walkthrough

> Documento di spiegazione del backend per chi conosce bene Python e un po' di Go.
> Le analogie Python sono evidenziate per aiutare il passaggio tra i due mondi.

---

## 1. Visione globale

```
┌──────────────┐        ┌─────────────────────────────────────────┐
│   frontend   │        │              backend (Go)               │
│ React (port  │  HTTP  │  cmd/server/main.go  ── porta 8080      │
│ 3000)        │ ──────►│    chi router → handler → service → repo│
└──────────────┘        │                 │                       │
                        │  cmd/worker/main.go (processo separato) │
                        │    fetch prezzi da Yahoo ogni 1h        │
                        └────────────────┬────────────────────────┘
                                         │
                    ┌────────────────────┼────────────────────┐
                    ▼                    ▼                    ▼
              PostgreSQL 16          Redis 7           Yahoo Finance (HTTP)
              dati veri             c'è ma è quasi    lookup ticker, prezzi,
                                     inutilizzato      storico, FX, split
```

Quattro servizi in `docker-compose.yml`: **postgres** (dati), **redis** (dichiarato ma di fatto usato poco — il "cache" è in una tabella Postgres), **backend** (API REST), **worker** (aggiorna prezzi in background).

Il backend è un **binario unico Go** che, a seconda dell'argomento CLI, fa due cose (`cmd/server/main.go`):
- `server migrate` → esegue le migrazioni SQL e termina
- `server` → avvia l'API HTTP

Nel container il comando è `sleep 3 && /server migrate && /server` (`docker-compose.yml:49-50`), quindi le migrazioni girano prima dell'avvio.

---

## 2. Architettura a strati (il cuore del design)

Quattro strati, ciascuno col suo ruolo e con dipendenze sempre in giù, mai in su:

```
HTTP (chi router)
   │  JSON in/out
   ▼
handler   ──  risponde alle richieste HTTP, decodifica body, estrae claims JWT
  │  chiama metodi di dominio
  ▼
service   ──  business logic: login, calcolo ROI, AVCO, FX, orchestrazione
  │  query parametrizzate
  ▼
repository ──  SQL su Postgres
```

- **handler** (`internal/handler/`) — "glue" HTTP. Conosce `http.Request`/`ResponseWriter`, non sa nulla di SQL.
- **service** (`internal/service/service.go`) — la logica di dominio. Non sa nulla di HTTP.
- **repository** (`internal/repository/`) — SQL puro, parametrizzato con `$1`, `$2`, ...

Equivalente Python: lo strato FastAPI route → use-case → repository/DAO.

### Dependency injection manuale

Tutto è cablato a mano, senza framework DI:

```go
repos := repository.New(dbPool)                                   // strato dati
fetcher := price.NewYahooFetcher(repos, cfg.PriceFetchInterval)   // client Yahoo
svc := service.New(repos, jwtAuth, fetcher, cfg.LookupCacheTTL)   // logica
h := handler.New(svc, jwtAuth)                                    // HTTP
```

Niente magic: equivalenti di `Depends`, qui si passa esplicitamente ai costruttori.

---

## 3. Entry point — `cmd/server/main.go`

Sequenza di avvio:

1. **Config** — `config.Load()` legge env vars con fallback. In Python: `os.getenv`.
2. **Pool Postgres** — `pgxpool.New(ctx, cfg.DSN())`. `pgxpool` è un **connection pool** (come SQLAlchemy engine + pool): si prende una connessione, si usa, si restituisce.
3. **Redis client** — creato, ma se non risponde il server **continua comunque** (continuing without cache).
4. **Router chi** — middleware in catena: logger, panic recover, RequestID, RealIP, timeout 30s, CORS.
5. **Routes** — `setupRoutes`: route senza auth (`/auth/*`) e gruppo protetto da `jwtAuth.Middleware`.

Il graceful shutdown (aspetta SEGNALI `SIGINT`/`SIGTERM`, poi `srv.Shutdown(ctx)` con timeout) è il pattern idiomatico Go.

---

## 4. Routing e una richiesta tipica end-to-end

Prendiamo `GET /api/v1/dashboard` — l'endpoint più ricco.

1. **Router**: `r.Get("/dashboard", h.GetDashboard)` nel gruppo protetto.
2. **Middleware JWT**: valida il token `Bearer`, mette i claims nel `context`.
3. **Handler**: estrae i claims, chiama il service, invia JSON.
4. **Service `GetDashboard`**: orchestrazione — portafogli → holdings (AVCO) → tassi FX → serie storiche.
5. **Repository**: SQL, es. `FindByUser` fa `LEFT JOIN portfolio_shares` (già predisposto per sharing futuro).

Pattern `for rows.Next()` + `rows.Scan(&x, &y)` è l'equivalente di `cursor.fetchall()`. Go richiede il **mapping manuale** colonna → struct, senza magic.

---

## 5. Il motore AVCO (`internal/position/position.go`)

Calcola la posizione di un asset usando la **logica AVCO (Average Cost)** — il prezzo medio di carico, come fa il tuo broker.

### Lo `State`

```go
type State struct {
    Qty         decimal.Decimal  // quantità attuale
    Avg         decimal.Decimal  // prezzo medio di carico (valuta portafoglio)
    avgCCY      decimal.Decimal  // prezzo medio (valuta strumento)
    Cost        decimal.Decimal  // totale investito
    CostCCY     decimal.Decimal  // idem in valuta strumento
    Realized    decimal.Decimal  // plus/minusvalenza realizzata
    realizedCCY decimal.Decimal  // idem in valuta strumento
}
```

### `Apply` — macchina a stati per transazione

- **Buy**: `Cost += qty×prezzo + fee`, ricalcola `Avg = Cost / nuovaQty` (nuovo prezzo medio di carico).
- **Sell**: costo del venduto = `Avg × qty`; lo sottrae dal costo; differenza con l'incasso → `Realized`.
- **Split**: quantità × rapporto, e **avg / rapporto** (il costo totale non cambia).
- **Fee**: somma al costo.
- **Dividend**: somma a `Realized`.

Equivalente Python: dataclass immutabile + funzioni pure applicate in sequenza.

### `Walk`

Raggruppa le transazioni per chiave `portfolioID|assetID`, le ordina per data e **inietta gli split** (a livello asset) nella timeline. Lo split si applica a tutti i portafogli che tengono quell'strumento.

---

## 6. Dove l'AVCO viene usato

**`HoldingsDetailed`** (`repository/portfolio.go`) è il punto nevralgico:

1. Query base: ogni strumento con transazioni + ultimo close (`LEFT JOIN LATERAL`).
2. Prende tutte le transazioni ordinate e gli split.
3. `states := position.Walk(txs, splitEvents)` — l'AVCO gira in Go.
4. Copia quantità / costo / realizzo / media dallo `State` alle struct.

I dati grezzi li prende da Postgres, ma **l'aritmetica finanziaria la fa il codice Go**, non SQL. I vecchi metodi SQL (`GetSummary`, `GetAllocation`, `GetROI`) esistono ancora su `portfolio.go` ma il service **non li usa più**.

### Serie storica giornaliera (`portfolioSeries`, service.go)

Per il grafico costruisce un punto per ogni giorno dalla prima transazione a oggi:

```go
for i, d := range dates {
    // applica le transazioni fino alla data d
    // applica gli split fino alla data d
    // prende l'ultimo prezzo disponibile ≤ d
    mv := st.Qty.Mul(priceByAsset[aid][priceDates[pricePos-1]]).Mul(rawFactor).Mul(factor)
    agg[i].MarketValue = mv   // sommato al totale del portafoglio
}
```

Nota: `rawFactor` (i prezzi storici Yahoo sono **split-adjusted**) e `factor` (FX asset → valuta portafoglio).

---

## 7. Valute e FX

1. Il fetcher/`RefreshFX` fotografia i **cross rate USD→X** da Yahoo (`USDCHF=X`) e li salva in `fx_rates` (migrazione 000003). Solo USD come base.
2. `loadRates` carica i tassi per le valute presenti nelle holding.
3. `fxFactor` converte A→B via cross: `(USD→B) / (USD→A)`. Se manca, `ok=false` → `fx_missing`.

> Non serve una tabella N×N: bastano righe per valuta (chiave per valuta + "cross via USD").

Da notare: la colonna `currency`/`exchange_rate` su `transactions` è stata **droppata** (migrazione 000005); oggi il prezzo di transazione è nella valuta dello strumento, e l'AVCO tiene separati `Cost` e `CostCCY`.

---

## 8. Il pacchetto prezzi (`internal/price/`)

### YahooFetcher

Ha: un `http.Client` con timeout 15s, due mappe di cooldown (history/split) per strumento, e un **`sync.Mutex`** che protegge quelle mappe (perché più goroutine condividono la memoria — equivalente di un lock Python per le map).

### Refresh intelligente (`RefreshStale`)

Tre criteri per evitare chiamate inutili a Yahoo:

1. nessun prezzo mai → fetch
2. fetch recente (< `PriceFetchInterval`) → skip (throttle)
3. ultimo close già al giorno feriale atteso → skip

Poi `fetchQuotes` una richiesta per ticker con `time.Sleep(500ms)` tra le chiamate (rate limiting).

### Storico (`EnsureHistory`) — backfill

Riporta i prezzi dalla prima transazione a oggi, con logica di ripresa (fetch completo se manca, altrimenti fetch da `latest` a oggi). Upsert idempotente per data.

### Split (`EnsureSplits`)

Fetcha gli eventi split da Yahoo e li **upserta** (idempotente su `asset_id,date`); il worker li ri-controlla a ogni intervallo ma con cooldown per strumento.

---

## 9. Il worker (`cmd/worker/main.go`)

Processo separato che ogni `VAULT_PRICE_FETCH_INTERVAL` chiama `fetcher.FetchAll`:

```go
ticker := time.NewTicker(interval)
// Run once immediately → poi loop con select su: ticker / ctx.Done / signal
for { select { case <-ticker.C: FetchAll(); case <-ctx.Done(): return } }
```

Processo separato e non goroutine nel server = **isolamento**: se Yahoo va giù, si riavvia il worker senza toccare l'API.

---

## 10. Auth JWT (`internal/auth/jwt.go`)

JWT stateless, HMAC-SHA256 col segreto di config. Due token:

- **access** (15 min): `sub` (user id), `email`, `role`, `token_type=access`
- **refresh** (72h): solo `sub` e `token_type=refresh`

Il middleware valida, **verifica che `token_type == "access"`**, e mette i token nel `context`. `RefreshToken` ricarica l'utente e genera una nuova coppia (rotazione automatica).

Password: `bcrypt` (`x/crypto/bcrypt`), come `passlib` in Python.

---

## 11. Transizioni atomiche — `WithTx`

```go
func (r *Repository) WithTx(ctx, fn func(*Repository) error) error {
    tx, _ := r.DB.Begin(ctx)
    defer func() { _ = tx.Rollback(ctx) }()   // se fn fallisce → rollback
    rr := &Repository{ User: &userRepo{db: tx}, ... }
    if err := fn(rr); err != nil { return err }
    return tx.Commit(ctx)
}
```

Due pattern Go notevoli:

1. **`DBTX` interface**: sia `*pgxpool.Pool` che `*pgx.Tx` implementano `Exec/Query/QueryRow`, quindi i repository funzionano identici sul pool o su una transazione.
2. La closure riceve un `Repository` clonato con i repository legati alla `tx`. `defer Rollback` dopo il commit è innocua no-op.

Usato in `ImportPortfolio`: in modalità "overwrite" cancella e ricrea il portafoglio **atomicamente** — se qualcosa altre fallisce, niente cambia.

---

## 12. Il ciclo dei dati completo

```
User crea asset ──► POST /assets ──► CreateAsset
                     syncAssetBackground (goroutine!) → EnsureSplits + EnsureHistory
Worker ogni 1h: RefreshStale (quote) + RefreshFX
User apre dashboard ──► GET /dashboard:
    HoldingsDetailed (AVCO) + FX rates + portfolioSeries → JSON
```

---

## 13. Mappa dei package

```
backend/
├── cmd/
│   ├── server/main.go      # binary API (avvio, routing, migrazioni)
│   └── worker/main.go      # binary prezzi (ticker loop)
├── internal/
│   ├── auth/jwt.go         # JWT: generazione, validazione, middleware
│   ├── config/config.go    # env vars + DSN
│   ├── handler/            # HTTP: auth.go, portfolio.go, portfolio_io.go
│   ├── model/              # struct + tag JSON (DTO e entità)
│   ├── position/           # motore AVCO (State, Apply, Walk)
│   ├── price/              # client Yahoo (quote, storico, split, FX, meta, lookup)
│   ├── repository/         # SQL (repository.go = "hub" + WithTx + DBTX)
│   └── service/            # business logic (service.go)
├── migrations/             # SQL versionato
└── go.mod
```

---

## 14. Concetti Go vs Python

| Go | Python |
|---|---|
| `package` + export (maiuscola) | modulo + `__` |
| `interface{...}` | protocol/ABC (duck typing statico) |
| `struct + tag json` | dataclass / Pydantic |
| `*pgxpool.Pool` | `SQLAlchemy engine` |
| `context.Context` | (niente di uguale) |
| `goroutine` | `threading.Thread` / `asyncio.create_task` |
| `sync.Mutex` | `threading.Lock` |
| `rows.Scan()` | `cursor.fetchall()` |
| `defer` | context manager / `try-finally` |
| errori come valori `(ok, err)` | eccezioni |
| `select { case <-ch: }` | event loop con più await |

La differenza più rilevante: **in Go gli errori sono valori**, ogni chiamata ritorna `(risultato, err)` e va controllata.

---

## 15. Note e punti aperti

- **Redis è di fatto inutilizzato**: la cache (autocomplete ticker) è in `lookup_cache` (Postgres, migrazione 0002), non in Redis. Il client Redis c'è ma nessun service lo usa.
- **Bug noto**: il worker subisce 429 da Yahoo Finance, difese già presenti: refresh-only-stale, throttle `price_fetched_at`, sleep 500ms, User-Agent browser.
- **`canAccessPortfolio`** (service.go:1188) verifica solo l'owner, ignora le `portfolio_shares` (TODO condivisione).
- I vecchi metodi SQL (`GetSummary`, `GetAllocation`, `GetROI`) sono **codice morto** per il service: rimovibili quando il motore AVCO è stabile.