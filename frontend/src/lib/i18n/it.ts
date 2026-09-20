import type { Dictionary } from './en'

/**
 * Italian dictionary — the product default language (decision D1). Kept
 * structurally identical to `en.ts` (canonical): `satisfies Dictionary`
 * fails the build on any missing, extra or misspelled key.
 *
 * Values are shown when the active locale is `it`; `{placeholders}` stay
 * ASCII so interpolated labels compose inside aria-labels.
 */
export const it = {
  common: {
    language: 'Lingua',
    /** Etichette generiche dei dialoghi di conferma (es. eliminazione portafoglio). */
    delete: 'Elimina',
    cancel: 'Annulla',
    /** Chiusura di sheet/dialoghi (es. sheet transazione, K.4c). */
    close: 'Chiudi',
  },
  nav: {
    main: 'Principale',
    dashboard: 'Dashboard',
    overview: 'Panoramica',
    portfolios: 'Portafogli',
    assets: 'Asset',
    more: 'Altro',
    bottomNav: 'Navigazione principale',
    dataSync: 'Dati e sincronizzazione',
    settings: 'Impostazioni',
    sectionAdmin: 'Admin',
    sectionSettings: 'Impostazioni',
    skipToContent: 'Vai al contenuto',
  },
  header: {
    expandSidebar: 'Espandi la barra laterale',
    collapseSidebar: 'Comprimi la barra laterale',
  },
  fab: {
    open: 'Apri le azioni rapide',
  },
  quickActions: {
    title: 'Azioni rapide',
    addTransaction: 'Aggiungi transazione',
    addTransactionHint: 'Registra un acquisto, una vendita o un dividendo',
    addAsset: 'Aggiungi asset',
    addAssetHint: 'Cerca su Yahoo e registralo',
    refreshPrices: 'Aggiorna prezzi',
    refreshPricesHint: 'Recupera subito le ultime quotazioni',
    enterPrice: 'Inserisci prezzo',
    comingSoon: 'In arrivo',
    refreshSuccess: 'Prezzi aggiornati',
    refreshError: 'Aggiornamento dei prezzi non riuscito',
    refreshRateLimited: 'Yahoo Finance sta limitando le richieste: alcuni prezzi potrebbero non essere aggiornati',
    refreshIssues: '{count} aggiornamenti prezzi non riusciti (Yahoo)',
  },
  user: {
    accountMenu: 'Menu account',
    signOut: 'Esci',
    fallbackName: 'Utente',
  },
  theme: {
    group: 'Tema',
    light: 'Chiaro',
    dark: 'Scuro',
    system: 'Sistema',
    aria: 'Tema: {theme}',
    ariaSystem: 'Tema: {theme}, attualmente {resolved}',
  },
  settingsTabs: {
    sections: 'Sezioni delle impostazioni',
    profile: 'Profilo',
    password: 'Password',
    preferences: 'Preferenze',
    currencies: 'Valute',
  },
  /**
   * Pannello comandi globale (EPIC K.5a, spec §8.1). Le etichette di
   * destinazioni/azioni sono riusate dai gruppi `nav.*`, `settingsTabs.*`,
   * `quickActions.*`, `theme.*` e `preferences.palette*`; qui c'è solo la
   * copy propria del pannello. `{query}` porta il testo digitato.
   */
  commandPalette: {
    /** Nome accessibile del dialog. */
    title: 'Pannello comandi',
    /** aria-label del pulsante di ricerca nell'header. */
    trigger: 'Apri il pannello comandi',
    /** Etichetta visibile (lg+) e aria-label dell'input di ricerca. */
    inputLabel: 'Cerca',
    placeholder: 'Cerca pagine, portafogli, asset…',
    sectionGoTo: 'Vai a',
    sectionAssets: 'Asset',
    sectionActions: 'Azioni',
    noResults: 'Nessun risultato',
    yahooRow: 'Cerca su Yahoo “{query}”',
    searching: 'Ricerca su Yahoo in corso…',
    /** Azioni (gli hint mostrano lo stato di destinazione dei toggle). */
    toggleTheme: 'Cambia tema',
    toggleCvd: 'Attiva/disattiva palette per daltonici',
    toggleSidebar: 'Mostra/nascondi barra laterale',
    /** Suggestioni da tastiera, dopo un glifo <kbd>. */
    hintNavigate: 'per navigare',
    hintSelect: 'per selezionare',
    hintClose: 'per chiudere',
  },
  preferences: {
    title: 'Preferenze',
    themeHint: 'Chiaro, scuro, oppure in base alle impostazioni del dispositivo (Sistema).',
    languageHint: 'Applicata subito e ricordata su questo dispositivo.',
    /** Controllo della palette utile/perdita (decisione D6, EPIC K.5c): vale
     *  anche come nome accessibile della tablist SegmentedControl. Le
     *  etichette restano corte ("Verde/Rosso" / "Blu/Arancione") per non far
     *  traboccare il controllo nella card a larghezza telefono (bug-fix
     *  EPIC K); il pannello comandi le riusa come hint dello stato di
     *  destinazione del toggle, e `paletteHint` porta la spiegazione estesa. */
    colorGroup: 'Colori utile/perdita',
    paletteClassic: 'Verde/Rosso',
    paletteCvd: 'Blu/Arancione',
    paletteHint: 'Sostituisce verde/rosso con blu/arancione in testi e grafici. Segni e frecce ▲▼ restano sempre.',
  },
  /**
   * Superfici di allocazione (bug-fix EPIC K, sweep progressivo D1): la card
   * "Allocazione complessiva" del vault in dashboard, il tab Allocazione del
   * portafoglio e il suo digest in Panoramica — titoli dei pannelli, stati di
   * errore isolati e nota di copertura dell'universo azionario (`{pct}` porta
   * una cifra decimale).
   */
  allocation: {
    title: 'Allocazione complessiva',
    unavailable: 'Allocazione non disponibile',
    classUnavailable: 'Allocazione per classi non disponibile',
    sectorUnavailable: 'Allocazione settoriale non disponibile',
    geoUnavailable: 'Allocazione geografica non disponibile',
    assetClasses: 'Classi di attività',
    sectorsEquity: 'Settori (solo equity)',
    regionsEquity: 'Regioni (solo equity)',
    countriesEquity: 'Paesi (solo equity)',
    /** Nota di copertura dell'universo azionario mostrata quando c'era esclusioni. */
    equityUniverse: 'Universo azionario: {pct}% del portafoglio',
    /** Torta dashboard: valore del vault ripartito per portafoglio. */
    byPortfolio: 'Allocazione per portafoglio',
    mixedCurrencies:
      'I portafogli usano valute diverse: i valori non sono confrontabili, le quote sono indicative.',
  },
  /**
   * Tab Esposizione degli asset e relative modali di modifica (bug-fix EPIC
   * K): titoli delle card, intestazioni dei pannelli, pulsanti di modifica e
   * banner riservato all'equity.
   */
  exposure: {
    geoTitle: 'Distribuzione geografica',
    sectorTitle: 'Distribuzione settoriale',
    countries: 'Paesi',
    regions: 'Regioni',
    sectors: 'Settori',
    noCountries: 'Nessun paese inserito',
    noCountriesHint: 'Aggiungi un paese qui sotto, oppure usa un prefill JustETF / Morningstar',
    modify: 'Modifica',
    editGeo: 'Modifica distribuzione geografica',
    editSectors: 'Modifica distribuzione settoriale',
    /** Banner che sostituisce le card per asset fuori dall'universo azionario. */
    universeTitle: 'Distribuzione geografica e settoriale',
    universeHint:
      'Questa distribuzione si applica solo agli asset azionari (azioni ed ETF/fondi di classe equity).',
    universeClassHint: "Imposta la classe 'Azioni' o 'Immobiliare' nelle Caratteristiche per attivarla.",
  },
  /**
   * Badge di provenienza (bug-fix EPIC K): etichetta della pill, spiegazione
   * nel tooltip/aria connessa e connettore di data (`{date}` è il stamp
   * YYYY-MM-DD grezzo). Una chiave piatta per id di provenienza
   * (`manualLabel`/`manualDesc`, …) perché il runtime supporta esattamente
   * due livelli di annidamento.
   */
  provenance: {
    updatedTo: 'aggiornato al {date}',
    manualLabel: 'manuale',
    manualDesc: 'Dati modificati manualmente',
    justetfLabel: 'da JustETF',
    justetfDesc: 'Lista paesi importata da JustETF, non modificata manualmente',
    morningstarLabel: 'da Morningstar',
    morningstarDesc: 'Dati importati da Morningstar, non modificati manualmente',
    morningstarRegionsLabel: 'da Morningstar (regioni ufficiali)',
    morningstarRegionsDesc: 'Regioni ufficiali importate da Morningstar, non modificate manualmente',
    yahooLabel: 'da Yahoo',
    yahooDesc: 'Settori importati da Yahoo, non modificati manualmente',
    derivedLabel: 'calcolato dai paesi',
    derivedDesc: 'Regioni calcolate a partire dai pesi dei paesi',
    derivedEtfLabel: 'da JustETF via paesi',
    derivedEtfDesc: 'Regioni calcolate dai paesi importati da JustETF',
  },
  hero: {
    netValue: 'Valore netto',
    allTime: 'complessivo',
    valueVsInvested: 'Valore vs investito',
    breakdown: 'Dettaglio',
    realized: 'Realizzato',
    dividends: 'Dividendi',
    invested: 'Investito',
  },
  period: {
    oneYear: '1Y',
    threeYears: '3Y',
    all: 'TUTTO',
    group: 'Periodo del grafico',
  },
  quality: {
    fxMissing: '{amount} esclusi — cambio mancante ({count} posizioni)',
    rateLimited: 'Alcuni prezzi non aggiornati (limitazione Yahoo)',
    refreshIssues: '{count} aggiornamenti prezzi non riusciti',
    refreshFailed: 'Aggiornamento prezzi non riuscito: i valori potrebbero essere obsoleti',
  },
  freshness: {
    asOf: 'Prezzi alle {time}',
    refreshing: 'Aggiornamento dei prezzi…',
    hint: 'I valori consolidati usano questi prezzi',
    partialHint: 'Prezzi alle {time}: alcuni aggiornamenti non riusciti o limitati',
  },
  checklist: {
    title: 'Configura il tuo vault',
    intro: 'Tre passi per iniziare a tracciare i tuoi investimenti.',
    stepPortfolio: 'Crea un portafoglio',
    stepPortfolioHint: 'Raggruppa gli investimenti per obiettivo o conto.',
    stepAsset: 'Aggiungi un asset',
    stepAssetHint: 'Cerca su Yahoo e registra ciò che possiedi.',
    stepTransaction: 'Registra una transazione',
    stepTransactionHint: 'Apri un portafoglio e inserisci un acquisto.',
    done: 'Completato',
    current: 'Passo corrente',
    pending: 'Non iniziato',
  },
  scope: {
    label: 'Ambito',
    all: 'Tutti i portafogli (Vault)',
  },
  sparkline: {
    trend: 'Andamento del valore',
    valueTrend: 'Andamento del valore di {name}',
  },
  /**
   * Toggle "vedi come tabella" dei grafici (EPIC K.5b, spec §9.1):
   * etichette e nomi accessibili del `ui/ChartTableToggle` condiviso, il
   * template del `<caption>` sr-only delle viste tabellari e le intestazioni
   * di colonna. `{name}` porta il titolo proprio del grafico, se ce l'ha.
   */
  chartView: {
    chart: 'Grafico',
    table: 'Tabella',
    aria: 'Vista dati del grafico',
    ariaNamed: 'Vista dati del grafico — {name}',
    caption: '{name} — dati del grafico',
    captionGeneric: 'Dati del grafico',
    /** Intestazioni di colonna condivise dalle viste tabellari (usate anche
     *  come etichette `Valore:`/`Peso:` nei tooltip ECharts, bug-fix EPIC K). */
    colName: 'Nome',
    colValue: 'Valore',
    colWeight: 'Peso',
    colPeriod: 'Periodo',
    colReturn: 'Rendimento',
    colCumulative: 'TWR cumulativo',
    colInvested: 'Investito',
    namePerformance: 'Performance',
    nameCapital: 'Investito contro valore',
    /** Stati vuoti dei wrapper di grafici di allocazione/esposizione. */
    noData: 'Nessun dato',
    noDistribution: 'Nessuna distribuzione',
    noClassAllocation: 'Nessuna allocazione per classi',
    noAllocation: 'Nessuna allocazione',
    /** Nomi di serie di riserva (identità in legenda/tooltip) quando il
     *  wrapper non ha un titolo proprio da riusare. */
    seriesExposure: 'Esposizione',
    seriesClassAllocation: 'Allocazione per classi',
  },
  portfolio: {
    tabsLabel: 'Sezioni del portafoglio',
    tabOverview: 'Panoramica',
    tabPositions: 'Posizioni',
    tabActivity: 'Attività',
    tabAllocation: 'Allocazione',
    back: 'Tutti i portafogli',
    addTransaction: 'Aggiungi transazione',
    actionsMenu: 'Azioni sul portafoglio',
    export: 'Esporta',
    import: 'Importa',
    delete: 'Elimina portafoglio',
    deleteConfirm: 'Eliminare questo portafoglio? Tutte le sue transazioni andranno perse.',
    deleted: 'Portafoglio eliminato',
    viewAllocation: 'Vedi allocazione completa',
  },
  activity: {
    typeGroup: 'Filtra per tipo di transazione',
    typeAll: 'Tutte',
    typeBuy: 'Acquisto',
    typeSell: 'Vendita',
    typeDividend: 'Dividendo',
    typeSplit: 'Split',
    typeFee: 'Commissione',
    asset: 'Asset',
    assetAll: 'Tutti gli asset',
    assetUnknown: 'Asset (non presente in questo portafoglio)',
    from: 'Dal',
    to: 'Al',
    clearFilters: 'Cancella filtri',
    emptyFiltered: 'Nessuna transazione corrisponde ai filtri',
    emptyFilteredHint: 'Prova a estendere le date o a rimuovere un filtro.',
  },
  tx: {
    titleNew: 'Nuova transazione',
    titleEdit: 'Modifica transazione',
    deleted: 'Transazione eliminata',
    undo: 'Annulla',
    undoFailed: 'Impossibile ripristinare la transazione',
  },
  asset: {
    tabsLabel: 'Sezioni dettaglio asset',
    tabOverview: 'Panoramica',
    tabExposure: 'Esposizione',
    tabData: 'Dati',
    back: 'Asset',
    actionsMenu: 'Azioni asset',
    refreshMeta: 'Aggiorna da Yahoo',
    backfillHistory: 'Backfill storico completo',
    delete: 'Elimina asset',
    deleteConfirm: 'Eliminare {ticker}? Questa azione non può essere annullata.',
    deleted: 'Asset eliminato',
    priceUpdated: 'Aggiornato il {date}',
    chip1d: '1G',
    chip1w: '1S',
    chip1m: '1M',
    chip1y: '1Y',
    chipYtd: 'YTD',
    whereHeld: 'Dove è detenuto',
    whereHeldEmpty: 'Non è detenuto in nessun portafoglio',
    whereHeldUnavailable: 'Le detenzioni nei portafogli non sono disponibili al momento',
    heldPortfolio: 'Portafoglio',
    heldQty: 'Quantità',
    heldCost: 'Costo',
    heldValue: 'Valore',
    heldGl: 'Guadagno/Perdita',
    quickFacts: 'Dati principali',
    factIsin: 'ISIN',
    factType: 'Tipo',
    factClass: 'Classe',
    factCurrency: 'Valuta',
    factExchange: 'Mercato',
    factPriceSource: 'Fonte prezzo',
    dangerZone: 'Zona pericolosa',
    manualPrice: 'Inserimento prezzo manuale',
    manualPriceHint: 'Inserisci prezzi datati a mano per gli asset senza fonte prezzi automatica.',
    fixedIncome: 'Attributi obbligazionari',
    fixedIncomeHint: 'Dettagli di emittente, scadenza e cedola per le obbligazioni.',
  },
} satisfies Dictionary
