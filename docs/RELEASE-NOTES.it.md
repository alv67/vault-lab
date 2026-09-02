# Release Notes

Qui raccogliamo, in ordine, le feature e i bug fix che entrano in ogni release.
Le righe sono pensate per essere riutilizzate direttamente nelle release note
del prodotto: **una riga per feature/fix**, scritte dal punto di vista
dell'utente finale (ciò che vede e usa nell'app — niente dettagli
interni/backend).

> Convenzione: quando chiudi una PR con modifiche di prodotto, aggiungi le
> righe relative in cima (oppure sotto la sezione della release corrente).
> Alla pubblicazione, crea una nuova sezione con versione e data.

## v0.1.0 — 25 Ago 2026 (prima release ufficiale)

### Nuove funzionalità
- Registrazione e login multi-utente (JWT con access e refresh token in rotazione)
- Impostazioni account: modifica di nome ed email e cambio password
- Portafogli: creazione, modifica, eliminazione, esportazione e importazione
- Asset: creazione, modifica, eliminazione con autocomplete del ticker e sincronizzazione automatica dei prezzi Yahoo
- Transazioni (acquisto / vendita / dividendo / split / commissione)
- Dashboard: valore totale, gain/loss, allocazione, performance e ROI per asset
- Grafico storico del valore del portafoglio nella dashboard, per vedere come cambia nel tempo
- Grafico storico di performance del singolo portafoglio con carico investito (cost basis), valore degli asset ancora investiti e andamento storico del P/L realizzato
- Lista portafogli con il valore corrente di ogni portafoglio/asset
- Supporto multi-valuta (EUR / USD / GBP / CHF) con whitelist delle valute configurabile
- Importo investito per valuta nella dashboard
- Aggiornamento prezzi automatico e manuale con dashboard di health della sincronizzazione
- Aggiornamento automatico periodico di prezzi asset e tassi di cambio, che aggiorna valori e storico fino all'ultima esecuzione

## v0.2.0 — 30 Ago 2026

### Nuove funzionalità
- Valori coerenti nel riepilogo del portafoglio anche quando manca un tasso di cambio
- Chart di distribuzione geografica e settoriale per portafogli e dashboard
- Classi di asset e allocazione per classe di investimento
- Storico dei tassi di cambio, così serie e chart restano corretti nel tempo
- Pagina dettaglio asset con riferimenti, esposizione e storico prezzi completo
- Esposizione ETF automatica (paesi/regioni e settori) con ricerca del codice ISIN partendo dal ticker

## In corso — Asset editing overhaul (PR #65)

### Nuove funzionalità
- Scegli come ogni asset riceve i suoi prezzi: `Yahoo`, `Manual` o `None` (evita errori Yahoo per ticker non-Yahoo come alcuni bond)
- Il grafico storico dell'asset ora carica tutto lo storico e fa lo zoom in-place usando i selettori 1M/3M/1Y/YTD/MAX (nessun ricaricamento inutile)
- Nuovo range `YTD` (da inizio anno) sul grafico storico dell'asset
- Marcatori degli split mostrati sul grafico storico dell'asset (es. `Split 4:1`)
- L'esposizione (regioni/settori) dell'asset si modifica in una modale dedicata, con tabelle dei pesi validate e compilazione automatica da JustETF e Yahoo
