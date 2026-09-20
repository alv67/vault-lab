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
  preferences: {
    title: 'Preferenze',
    themeHint: 'Chiaro, scuro, oppure in base alle impostazioni del dispositivo (Sistema).',
    languageHint: 'Applicata subito e ricordata su questo dispositivo.',
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
} satisfies Dictionary
