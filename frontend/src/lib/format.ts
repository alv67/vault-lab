const currencySymbols: Record<string, string> = {
  USD: '$',
  EUR: '€',
  GBP: '£',
  CHF: 'CHF',
  JPY: '¥',
  CAD: 'C$',
  AUD: 'A$',
  CNY: '¥',
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
