# Release Notes

## Unreleased

### Nuove funzionalità
- La sezione allocazione del dettaglio portafoglio ora replica la card "Allocazione complessiva" della dashboard: una ciambella per classe di attività e barre decrescenti per regioni, settori e paesi, tutto nella valuta del portafoglio (solo azioni, ordinate per valore)
- Anche nel dettaglio del portafoglio compare lo stesso riepilogo degli investimenti attivi e chiusi della dashboard, nella valuta del portafoglio: le posizioni aperte con investito, valore attuale, guadagno/perdita e dividendi, e la parte chiusa con il costo dei lotti venduti, l'incasso e il guadagno/perdita realizzata
- La dashboard mostra ora un'unica tabella degli asset investiti: una riga per ciascun asset, aggregata su tutti i tuoi portafogli ed espressa nella tua valuta base, con importo investito, valore attuale e P/L %, ordinata per valore
- Scegli la tua valuta base in Impostazioni → Profilo: la dashboard ora consolida totali e allocazioni in quella valuta
- L'allocazione della dashboard ora mostra anche la ripartizione per classe di investimento e per singolo paese, accanto a quelle già disponibili per settore e macro-regione: i valori sono aggregati su tutti i tuoi portafogli ed espressi nella tua valuta base
- La dashboard ora separa investimenti attivi e chiusi, a livello di vault e per singolo portafoglio: per la parte chiusa vedi il costo dei lotti venduti, l'incasso e il guadagno/perdita realizzata sul capitale; le righe dei dividendi compaiono insieme agli investimenti attivi e, una volta chiusa completamente una posizione, confluiscono nell'incasso
- Nuovo grafico delle performance aggregato nella dashboard: un unico grafico per tutti i tuoi portafogli, nella tua valuta base, che mostra il rendimento percentuale time-weighted di ogni mese o anno (con una linea del TWR cumulativo) accanto a un grafico del capitale che confronta il denaro investito con il valore attuale, con selettore mensile/annuale. Il rendimento è misurato giorno per giorno, quindi versamenti, vendite, dividendi e commissioni non lo distorcono più e chiudere o riaprire completamente una posizione non produce più picchi assurdi; le posizioni senza prezzo di mercato sono mantenute al costo e non compaiono mai come falsa perdita

### Correzioni
- L'importazione di un portafoglio esportato da una versione precedente dell'app non fallisce più: le informazioni mancanti vengono completate con valori predefiniti sensati
- I portafogli con sole posizioni chiuse non mostrano più un guadagno/perdita fuorviante del -100%: le posizioni chiuse non entrano nei valori attivi e gli importi non hanno più residui di arrotondamento

## v0.4.0 — 13 Set 2026

### Nuove funzionalità
- Nuovo tema scuro, attivo di default, con le opzioni Chiaro / Scuro / Sistema
- Navigazione ridisegnata: sidebar collassabile, header in alto con selettore del tema e menu utente, e drawer a scomparsa su mobile
- Colori coerenti con il tema su tutte le pagine e i grafici, così l'interfaccia è leggibile sia in chiaro sia in scuro
- Le azioni distruttive ora usano una finestra di conferma dell'app invece del prompt nativo del browser
- Le notifiche (toast) sono state ridisegnate in linea col tema e rese accessibili agli screen reader
- La creazione di asset/portafogli e l'import di un portafoglio ora avvengono in finestre modali coerenti con il design dell'app
- Ridisegnata la schermata di accesso/registrazione con il logo VaultLab, lo switch Sign in / Register, la validazione inline dei campi e la conferma password in registrazione
- Dashboard ridisegnata: card KPI per valuta, donut dell'allocazione e card dei portafogli cliccabili
- L'header della dashboard mostra quando i prezzi sono stati aggiornati l'ultima volta
- Dettaglio portafoglio: aggiungere/modificare una transazione ora avviene in una finestra con validazione inline e totale live, e l'eliminazione è confermata nell'app
- Le posizioni e le transazioni del portafoglio usano le tabelle condivise del design system (le posizioni ora mostrano anche l'ultimo prezzo), e le azioni della pagina stanno in un header sticky
- Impostazioni riorganizzate in tab: Profile, Password, Currencies e Health
- Le valute gestite ora si scelgono da una lista, con il nome compilato automaticamente
- Il cambio password ora valida inline ed evidenzia il campo in errore (es. password corrente sbagliata)
- La lista degli eventi di health della sincronizzazione prezzi ora è paginata, così si può consultare tutto lo storico
- La pagina di health della sincronizzazione prezzi è stata spostata in una sezione Admin dedicata nella sidebar

### Correzioni
- Le allocazioni del portafoglio si aggiornano subito dopo aver aggiunto, modificato o eliminato una transazione, senza ricaricare la pagina
- Il grafico storico dei portafogli in dashboard disegna ogni portafoglio come una linea continua su una timeline reale, e si può zoomare e spostare come i grafici del portafoglio
- Health della sincronizzazione prezzi: le card Success Rate e Rate Limited mostrano ora i valori reali (il rate poteva restare bloccato su `N/A`, o mostrare `NaN%` senza dati)
- Health della sincronizzazione prezzi: i messaggi di errore ora indicano il tipo di richiesta (chart / spark / search / fx) e il relativo ticker o valuta
- Gli asset non gestiti da Yahoo (manual / none) non generano più errori di sincronizzazione nella dashboard di health
- Health della sincronizzazione prezzi: i totali non si azzerano più al riavvio e si possono limitare a Today / Last 24h / Last 100 eventi

## v0.3.0 — 11 Set 2026

### Nuove funzionalità
- Scegli come ogni asset riceve i suoi prezzi: `Yahoo`, `Manual` o `None` (evita errori Yahoo per ticker non-Yahoo come alcuni bond)
- Il grafico storico dell'asset ora carica tutto lo storico e fa lo zoom in-place usando i selettori 1M/3M/1Y/YTD/MAX (nessun ricaricamento inutile)
- Nuovo range `YTD` (da inizio anno) sul grafico storico dell'asset
- Marcatori degli split mostrati sul grafico storico dell'asset (es. `Split 4:1`)
- L'esposizione (regioni/settori) dell'asset si modifica in una modale dedicata, con tabelle dei pesi validate e compilazione automatica da JustETF e Yahoo
- Modifica la distribuzione geografica con un elenco per-paese (aggiungi/rimuovi paesi e imposta ogni peso), oltre a regioni e settori
- L'esposizione geografica e settoriale può essere compilata anche da Morningstar (regioni ufficiali), oltre a JustETF e Yahoo
- Ogni distribuzione mostra da dove arrivano i dati e quando sono stati aggiornati l'ultima volta (es. `da Morningstar (2026-09-05)`)
- Le letture dai provider sono in cache, quindi riaprire i prefill è immediato

### Correzioni
- Riaprire un editor di esposizione ora riparte sempre dai dati salvati: le modifiche non salvate vengono scartate

## v0.2.0 — 30 Ago 2026

### Nuove funzionalità
- Valori coerenti nel riepilogo del portafoglio anche quando manca un tasso di cambio
- Chart di distribuzione geografica e settoriale per portafogli e dashboard
- Classi di asset e allocazione per classe di investimento
- Storico dei tassi di cambio, così serie e chart restano corretti nel tempo
- Pagina dettaglio asset con riferimenti, esposizione e storico prezzi completo
- Esposizione ETF automatica (paesi/regioni e settori) con ricerca del codice ISIN partendo dal ticker

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
