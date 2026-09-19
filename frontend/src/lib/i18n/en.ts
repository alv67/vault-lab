/**
 * English dictionary (EPIC K.1b, decision D1) — the canonical shape every
 * locale must match. Keys live in a two-level nested structure (`group.key`
 * in the file); `index.svelte.ts` flattens them into dot-joined lookup keys
 * (`nav.dashboard`) exposed through the `MessageKey` union, so `t()` calls
 * are checked at compile time and `it.ts` cannot drift (it is verified with
 * `satisfies Dictionary`).
 *
 * Conventions:
 * - add new strings here first, then mirror them in `it.ts`;
 * - every new string from K.1 on ships in both languages; legacy pages keep
 *   their mixed EN/IT copy until their own migration sweep (progressive);
 * - values are plain strings with optional `{name}` placeholders; labels
 *   use sentence case without trailing punctuation.
 */
export const en = {
  common: {
    language: 'Language',
  },
  nav: {
    /** Accessible name of the sidebar/drawer `<nav>` landmark. */
    main: 'Main',
    dashboard: 'Dashboard',
    /** Phone bottom-nav destination for `/` (EPIC K.2, decision D2). */
    overview: 'Overview',
    portfolios: 'Portfolios',
    assets: 'Assets',
    /** Bottom-nav item that opens the off-canvas "More" sheet (D2). */
    more: 'More',
    /** Accessible name of the fixed phone bottom navigation landmark. */
    bottomNav: 'Primary navigation',
    /**
     * Price-sync health page (decision D7). Referenced once, from the
     * `adminItems` config in `SidebarNav.svelte`, so the whole entry —
     * label included — can be relocated to an Administration menu later.
     */
    dataSync: 'Data & Sync',
    settings: 'Settings',
    /** Sidebar section headers (visible only when the sidebar is expanded). */
    sectionAdmin: 'Admin',
    sectionSettings: 'Settings',
    skipToContent: 'Skip to content',
  },
  header: {
    expandSidebar: 'Expand sidebar',
    collapseSidebar: 'Collapse sidebar',
  },
  fab: {
    /** Accessible name of the phone-only quick-actions floating button. */
    open: 'Open quick actions',
  },
  quickActions: {
    /** Title of the sheet opened by the Fab (decision D2). */
    title: 'Quick actions',
    addTransaction: 'Add transaction',
    addTransactionHint: 'Record a buy, sell or dividend',
    addAsset: 'Add asset',
    addAssetHint: 'Search Yahoo and register it',
    refreshPrices: 'Refresh prices',
    refreshPricesHint: 'Fetch the latest quotes now',
    /** Reserved slot until EPIC J.1 ships: rendered disabled. */
    enterPrice: 'Enter price',
    comingSoon: 'Coming soon',
    /** Fab refresh feedback (toasts). */
    refreshSuccess: 'Prices updated',
    refreshError: 'Price refresh failed',
    refreshRateLimited: 'Yahoo Finance is rate-limiting requests: some prices may be stale',
    refreshIssues: '{count} price updates failed (Yahoo)',
  },
  user: {
    /** Compact icon-only trigger (rail / mobile header). */
    accountMenu: 'Account menu',
    signOut: 'Sign out',
    /** Fallback shown when the account has no name/email. */
    fallbackName: 'User',
  },
  theme: {
    /** Field label on Preferences and group name of the header popup. */
    group: 'Theme',
    light: 'Light',
    dark: 'Dark',
    system: 'System',
    /** Trigger aria-labels; `{theme}`/`{resolved}` carry a lower-cased label. */
    aria: 'Theme: {theme}',
    ariaSystem: 'Theme: {theme}, currently {resolved}',
  },
  settingsTabs: {
    /** Accessible name of the Settings section tab bar. */
    sections: 'Settings sections',
    profile: 'Profile',
    password: 'Password',
    preferences: 'Preferences',
    currencies: 'Currencies',
  },
  preferences: {
    title: 'Preferences',
    themeHint: 'Light, dark, or follow the device setting (System).',
    languageHint: 'Applied immediately and remembered on this device.',
  },
}

/** Canonical dictionary shape derived from the English source of truth. */
export type Dictionary = typeof en
