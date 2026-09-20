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
    /** Generic confirm-dialog labels (used e.g. by the portfolio delete). */
    delete: 'Delete',
    cancel: 'Cancel',
    /** Close affordance on sheets/dialogs (e.g. the transaction sheet, K.4c). */
    close: 'Close',
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
  /**
   * Vault Overview hero zone (EPIC K.3a, redesign spec §6.1 zone A). The
   * `Investments`/`Performance` card copy that predates the dictionary stays
   * hardcoded until the dashboard sweep (progressive migration, D1).
   */
  hero: {
    netValue: 'Net value',
    allTime: 'all-time',
    /** Headline of the value-vs-invested chart next to the hero number. */
    valueVsInvested: 'Value vs invested',
    /** Progressive disclosure wrapping the Active/Closed table. */
    breakdown: 'Breakdown',
    realized: 'Realized',
    dividends: 'Dividends',
    invested: 'Invested',
  },
  /** Period chips on the hero chart (decision D10: bucket-driven ranges). */
  period: {
    oneYear: '1Y',
    threeYears: '3Y',
    all: 'ALL',
    group: 'Chart period',
  },
  /**
   * Data-quality strip chips (spec §6.1/§8.5). Each chip is a link to the
   * fixing surface, so labels name the problem, not the destination.
   */
  quality: {
    fxMissing: '{amount} excluded — missing FX ({count} holdings)',
    rateLimited: 'Some prices not updated (Yahoo rate limit)',
    refreshIssues: '{count} price updates failed',
    refreshFailed: 'Price refresh failed — values may be stale',
  },
  /** Freshness stamp near the hero (spec §8.5), from the session refresh. */
  freshness: {
    asOf: 'Prices as of {time}',
    refreshing: 'Refreshing prices…',
    /** Plain-text tooltip on the completed stamp. */
    hint: 'Consolidated values use these prices',
    partialHint: 'Prices as of {time} — some updates failed or were rate-limited',
  },
  /** First-run checklist replacing the empty-vault EmptyState (D8). */
  checklist: {
    title: 'Set up your vault',
    intro: 'Three steps to start tracking your investments.',
    stepPortfolio: 'Create a portfolio',
    stepPortfolioHint: 'Group investments by goal or account.',
    stepAsset: 'Add an asset',
    stepAssetHint: 'Search Yahoo and register what you own.',
    stepTransaction: 'Record a transaction',
    stepTransactionHint: 'Open a portfolio and log a buy.',
    done: 'Done',
    current: 'Current step',
    pending: 'Not started',
  },
  /** Vault ⇄ portfolio scope switcher in the Overview header (D3). */
  scope: {
    label: 'Scope',
    all: 'All portfolios (Vault)',
  },
  /**
   * Portfolio-card sparklines on the vault Overview (EPIC K.3b, spec §6.1
   * zone C). The shape is supplementary (the card already carries value and
   * P/L), so these are accessible names, not visible copy.
   */
  sparkline: {
    /** Generic fallback when the caller has no better label. */
    trend: 'Value trend',
    /** Per-card aria-label with the portfolio name interpolated. */
    valueTrend: '{name} value trend',
  },
  /**
   * Portfolio-detail shell (EPIC K.4a, spec §6.2): sticky-header chrome and
   * the four tier-2 tabs. Card/section copy that predates the dictionary
   * stays hardcoded until the page's migration sweep (progressive, D1).
   */
  portfolio: {
    /** Accessible name of the portfolio tab bar. */
    tabsLabel: 'Portfolio sections',
    tabOverview: 'Overview',
    tabPositions: 'Positions',
    tabActivity: 'Activity',
    tabAllocation: 'Allocation',
    /** Back link in the sticky header, to the portfolios list. */
    back: 'All portfolios',
    addTransaction: 'Add transaction',
    /** Accessible name of the `⋯` actions menu in the sticky header. */
    actionsMenu: 'Portfolio actions',
    export: 'Export',
    import: 'Import',
    delete: 'Delete portfolio',
    deleteConfirm: 'Delete this portfolio? All its transactions will be lost.',
    deleted: 'Portfolio deleted',
    /** Overview allocation digest → link to the Allocation tab. */
    viewAllocation: 'View full allocation',
  },
  /**
   * Portfolio Activity tab filter row (EPIC K.4c, spec §6.2/§8.2): type
   * chips, asset picker and date range, persisted in the tab's URL query.
   * The "Transactions" card heading predates the dictionary and stays
   * hardcoded until the page's migration sweep (progressive, D1).
   */
  activity: {
    /** Accessible name of the transaction-type chip radiogroup. */
    typeGroup: 'Filter by transaction type',
    typeAll: 'All',
    typeBuy: 'Buy',
    typeSell: 'Sell',
    typeDividend: 'Dividend',
    typeSplit: 'Split',
    typeFee: 'Fee',
    /** Visible inline label of the asset picker. */
    asset: 'Asset',
    assetAll: 'All assets',
    /** Placeholder row for a deep-linked asset outside this portfolio. */
    assetUnknown: 'Asset (not in this portfolio)',
    /** Visible inline labels of the inclusive date-range inputs. */
    from: 'From',
    to: 'To',
    clearFilters: 'Clear filters',
    /** Filtered-empty state (the row exists but matches no filter). */
    emptyFiltered: 'No transactions match these filters',
    emptyFilteredHint: 'Try widening the date range or clearing a filter.',
  },
  /**
   * Transaction form + delete feedback (EPIC K.4c): titles are shared by
   * the `ui/Modal` (≥ `sm`) and the `ui/Sheet` (phone) containers; the
   * delete strings drive the undo-based flow (decision D11). The form's
   * inner field/button copy predates the dictionary (progressive sweep).
   */
  tx: {
    titleNew: 'New Transaction',
    titleEdit: 'Edit Transaction',
    deleted: 'Transaction deleted',
    /** Undo action inside the delete toast (5 s window). */
    undo: 'Undo',
    undoFailed: 'The transaction could not be restored',
  },
  /**
   * Asset-detail shell (EPIC K.4b, spec §6.3): sticky-header chrome, the
   * three tier-2 tabs, the "Where held" Overview block, the danger-zone
   * and reserved-section labels. Card copy that predates the dictionary
   * stays hardcoded until the page's migration sweep (progressive, D1).
   */
  asset: {
    /** Accessible name of the asset tab bar. */
    tabsLabel: 'Asset sections',
    tabOverview: 'Overview',
    tabExposure: 'Exposure',
    tabData: 'Data',
    /** Back link in the sticky header, to the assets library list. */
    back: 'Assets',
    /** Accessible name of the `⋯` actions menu in the sticky header. */
    actionsMenu: 'Asset actions',
    refreshMeta: 'Update from Yahoo',
    backfillHistory: 'Backfill full history',
    delete: 'Delete asset',
    deleteConfirm: 'Delete {ticker}? This cannot be undone.',
    deleted: 'Asset deleted',
    /** Muted date under the quote chips: `{date}` is locale-formatted. */
    priceUpdated: 'Updated {date}',
    /** Compact quote-delta chips (1D/1W/1M/1Y/YTD). */
    chip1d: '1D',
    chip1w: '1W',
    chip1m: '1M',
    chip1y: '1Y',
    chipYtd: 'YTD',
    /** Overview "Where held" block (spec §4.2 decision 5). */
    whereHeld: 'Where held',
    whereHeldEmpty: 'Not held in any portfolio',
    whereHeldUnavailable: 'Holding portfolios are unavailable right now',
    heldPortfolio: 'Portfolio',
    heldQty: 'Qty',
    heldCost: 'Cost',
    heldValue: 'Value',
    heldGl: 'Gain/Loss',
    /** Overview read-only identity grid. */
    quickFacts: 'Quick facts',
    factIsin: 'ISIN',
    factType: 'Type',
    factClass: 'Class',
    factCurrency: 'Currency',
    factExchange: 'Exchange',
    factPriceSource: 'Price source',
    /** Data tab danger zone (same actions as the header `⋯` menu). */
    dangerZone: 'Danger zone',
    /** Reserved EPIC J placeholders (no behaviour yet). */
    manualPrice: 'Manual price entry',
    manualPriceHint: 'Record dated prices by hand for assets without an automatic feed.',
    fixedIncome: 'Fixed-income attributes',
    fixedIncomeHint: 'Issuer, maturity and coupon details for bonds.',
  },
}

/** Canonical dictionary shape derived from the English source of truth. */
export type Dictionary = typeof en
