const currencySymbols: Record<string, string> = {
  USD: '$',
  EUR: '€',
  GBP: '£',
  CHF: 'CHF',
  JPY: '¥',
  CAD: 'C$',
  AUD: 'A$',
  CNY: '¥',
  SEK: 'kr',
  NOK: 'kr',
  DKK: 'kr',
  PLN: 'zł',
  KRW: '₩',
  INR: '₹',
  BRL: 'R$',
  MXN: 'MX$',
  SGD: 'S$',
  NZD: 'NZ$',
  HKD: 'HK$',
  TRY: '₺',
  RUB: '₽',
  ZAR: 'R',
}

export function currencySymbol(code: string): string {
  return currencySymbols[code] || code
}

export function formatCurrency(amount: number | string, currency: string = 'USD'): string {
  const sym = currencySymbol(currency)
  const val = typeof amount === 'string' ? Number(amount) : amount
  return `${sym}${val.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

export function formatPercent(value: number | string): string {
  const val = typeof value === 'string' ? Number(value) : value
  return `${val.toFixed(2)}%`
}

/** Like `formatPercent` but always prefixes an explicit `+` on positive
 * values (e.g. `+3.42%`), used by the dashboard return chart tooltips where
 * the sign carries the meaning. */
export function formatSignedPercent(value: number | string): string {
  const val = typeof value === 'string' ? Number(value) : value
  return `${val > 0 ? '+' : ''}${val.toFixed(2)}%`
}

// Etichette italiane per le classi di asset (valore backend → label UI).
export const ASSET_CLASS_LABELS: Record<string, string> = {
  equity: 'Azioni',
  bond: 'Obbligazioni',
  commodity: 'Materie prime',
  currency: 'Valute',
  crypto: 'Crypto',
  real_estate: 'Immobiliare',
  mixed: 'Misto',
  other: 'Altro',
}

// Etichette dei tipi asset (stesse coppie valore/label del form "Type" della
// pagina asset pre-K.4b): centralizzate perché header, quick facts e select
// del tab Data le leggano tutte dalla stessa mappa.
export const ASSET_TYPE_LABELS: Record<string, string> = {
  stock: 'Stock',
  etf: 'ETF',
  bond: 'Bond',
  mutual_fund: 'Mutual fund',
  crypto: 'Crypto',
  commodity: 'Commodity',
}

export const PRICE_SOURCE_LABELS: Record<string, string> = {
  yahoo: 'Yahoo Finance',
  manual: 'Prezzo manuale',
  none: 'Nessun prezzo',
}
