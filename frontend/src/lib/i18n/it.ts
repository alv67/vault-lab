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
} satisfies Dictionary
