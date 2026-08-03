const BASE_URL = '/api/v1'

export interface User {
  id: string
  email: string
  name: string
  role: string
  created_at: string
}

export interface Portfolio {
  id: string
  user_id: string
  name: string
  description: string
  currency: string
  created_at: string
  updated_at: string
}

export interface Asset {
  id: string
  ticker: string
  isin: string
  name: string
  type: string
  category_id: string
  country: string
  currency: string
}

export interface Currency {
  code: string
  name: string
  symbol?: string
  enabled?: boolean
  sort?: number
  created_at?: string
}

export interface CurrencyList {
  currencies: Currency[]
}

export interface Transaction {
  id: string
  portfolio_id: string
  asset_id: string
  asset_ticker: string
  asset_name: string
  type: 'buy' | 'sell' | 'dividend' | 'split' | 'fee'
  quantity: string
  price: string
  fees: string
  date: string
  notes: string
}

export interface AssetHolding {
  asset_id: string
  ticker: string
  name: string
  currency: string
  qty: string
  cost: string
  cost_ccy: string
  value: string
  value_pf: string
  realized: string
  realized_ccy: string
  unrealized: string
  roi: string
  fx_missing: boolean
  closed: boolean
}

export interface PortfolioSummary {
  portfolio_id: string
  portfolio_name: string
  total_value: string
  total_cost: string
  gain_loss: string
  gain_loss_pct: string
  realized_gl: string
  unrealized_gl: string
  asset_count: number
  holdings: AssetHolding[]
}

export interface AssetAllocation {
  asset_id: string
  ticker: string
  name: string
  value: string
  alloc_pct: string
}

export interface PortfolioPerformance {
  date: string
  value: string
}

export interface AssetROI {
  asset_id: string
  ticker: string
  name: string
  roi: string
  total_invested: string
  current_value: string
}

export interface CurrencyPerformance {
  currency: string
  invested: string
  value: string
  gain_loss: string
  gain_loss_pct: string
  realized: string
}

export interface PortfolioPerformanceSummary {
  portfolio_id: string
  portfolio_name: string
  currency: string
  invested: string
  value: string
  gain_loss: string
  gain_loss_pct: string
  realized_gl: string
  asset_count: number
  fx_missing: number
}

export interface AssetPerformance {
  asset_id: string
  ticker: string
  name: string
  currency: string
  qty: string
  invested: string
  value: string
  gain_loss: string
  roi: string
  fx_missing: boolean
  value_pf: string
  realized: string
  realized_pf: string
}

export interface PortfolioAssets {
  portfolio_id: string
  portfolio_name: string
  currency: string
  assets: AssetPerformance[]
}

export interface DashboardHistory {
  portfolio_id: string
  portfolio_name: string
  currency: string
  series: PortfolioPerformance[]
}

export interface PositionPoint {
  date: string
  qty: string
  cost_basis: string
  market_value: string
  realized: string
}

export interface SplitInfo {
  date: string // ISO
  ratio: string // es. "7:1", "4:1"
}

export interface AssetPositionSeries {
  asset_id: string
  ticker: string
  name: string
  currency: string
  series: PositionPoint[]
  splits: SplitInfo[]
}

export interface PortfolioHistory {
  portfolio_id: string
  portfolio_name: string
  currency: string
  series: PositionPoint[]
  splits: SplitInfo[]
  assets: AssetPositionSeries[]
}

export interface PortfolioExportDocument {
  version: number
  exported_at: string
  portfolio: {
    name: string
    description: string
    currency: string
  }
  assets: {
    ticker: string
    name: string
    type: string
    currency: string
    isin?: string
  }[]
  transactions: {
    date: string
    type: string
    asset_ticker: string
    quantity: string
    price: string
    fees?: string
    notes?: string
  }[]
}

export interface Dashboard {
  by_currency: CurrencyPerformance[]
  portfolios: PortfolioPerformanceSummary[]
  assets: PortfolioAssets[]
  history: DashboardHistory[]
}

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}

export interface AssetLookupResult {
  ticker: string
  name: string
  type: string
  currency: string
  exchange: string
}

export interface AssetMeta {
  ticker: string
  name: string
  type: string
  currency: string
  exchange: string
  country: string
}

interface RequestOptions {
  method?: string
  body?: unknown
  params?: Record<string, string>
}

function buildUrl(path: string, params?: Record<string, string>): string {
  const url = new URL(BASE_URL + path, window.location.origin)
  if (params) {
    Object.entries(params).forEach(([key, value]) => url.searchParams.set(key, value))
  }
  return url.toString()
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const url = buildUrl(path, options.params)
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = localStorage.getItem('access_token')
  if (token) headers.Authorization = `Bearer ${token}`

  const init = (): RequestInit => ({
    method: options.method ?? 'GET',
    headers,
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  })

  let res = await fetch(url, init())

  if (res.status === 401 && !path.includes('/auth/')) {
    const refreshToken = localStorage.getItem('refresh_token')
    if (refreshToken) {
      try {
        const refreshRes = await fetch(`${BASE_URL}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        })
        if (refreshRes.ok) {
          const data = await refreshRes.json()
          localStorage.setItem('access_token', data.access_token)
          localStorage.setItem('refresh_token', data.refresh_token)
          headers.Authorization = `Bearer ${data.access_token}`
          res = await fetch(url, init())
        } else {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          window.location.replace('/login')
        }
      } catch {
        window.location.replace('/login')
      }
    }
  }

  if (!res.ok) {
    let message = 'Something went wrong'
    try {
      const data = await res.json()
      if (data?.error) message = data.error
    } catch {
      // keep default message
    }
    throw Object.assign(new Error(message), { status: res.status })
  }

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export const authApi = {
  login: (email: string, password: string) =>
    request<AuthResponse>('/auth/login', { method: 'POST', body: { email, password } }),
  register: (email: string, name: string, password: string) =>
    request<User>('/auth/register', { method: 'POST', body: { email, name, password } }),
  me: () => request<User>('/users/me'),
  updateProfile: (data: { name: string; email: string }) =>
    request<User>('/users/me', { method: 'PATCH', body: data }),
  changePassword: (data: { current_password: string; new_password: string }) =>
    request<void>('/users/me/password', { method: 'POST', body: data }),
}

export const portfolioApi = {
  list: () => request<Portfolio[]>('/portfolios'),
  create: (data: Partial<Portfolio>) =>
    request<Portfolio>('/portfolios', { method: 'POST', body: data }),
  get: (id: string) => request<Portfolio>(`/portfolios/${id}`),
  update: (id: string, data: Partial<Portfolio>) =>
    request<Portfolio>(`/portfolios/${id}`, { method: 'PATCH', body: data }),
  delete: (id: string) => request<void>(`/portfolios/${id}`, { method: 'DELETE' }),
  summary: (id: string) => request<PortfolioSummary>(`/portfolios/${id}/summary`),
  allocation: (id: string) => request<AssetAllocation[]>(`/portfolios/${id}/allocation`),
  performance: (id: string) => request<PortfolioPerformance[]>(`/portfolios/${id}/performance`),
  roi: (id: string) => request<AssetROI[]>(`/portfolios/${id}/roi`),
  history: (id: string) => request<PortfolioHistory>(`/portfolios/${id}/history`),
  dashboard: () => request<Dashboard>('/dashboard'),
  exportDoc: (id: string) => request<PortfolioExportDocument>(`/portfolios/${id}/export`),
  importDoc: (data: {
    document: PortfolioExportDocument
    mode: 'new' | 'overwrite'
    name?: string
    target_portfolio_id?: string
  }) => request<Portfolio>('/portfolios/import', { method: 'POST', body: data }),
}

export const assetApi = {
  list: () => request<Asset[]>('/assets'),
  search: (q: string) => request<Asset[]>(`/assets/search?q=${q}`),
  lookup: (q: string) => request<AssetLookupResult[]>(`/assets/lookup?q=${q}`),
  meta: (ticker: string) => request<AssetMeta>(`/assets/meta?ticker=${ticker}`),
  get: (id: string) => request<Asset>(`/assets/${id}`),
  create: (data: Partial<Asset>) => request<Asset>('/assets', { method: 'POST', body: data }),
  remove: (id: string) => request<void>(`/assets/${id}`, { method: 'DELETE' }),
  sync: () => request<{ status: string }>('/assets/sync', { method: 'POST' }),
}

export const transactionApi = {
  list: (portfolioId: string) => request<Transaction[]>(`/portfolios/${portfolioId}/transactions`),
  create: (portfolioId: string, data: Partial<Transaction>) =>
    request<Transaction>(`/portfolios/${portfolioId}/transactions`, { method: 'POST', body: data }),
  update: (id: string, data: Partial<Transaction>) =>
    request<Transaction>(`/transactions/${id}`, { method: 'PATCH', body: data }),
  remove: (id: string) => request<void>(`/transactions/${id}`, { method: 'DELETE' }),
}

export const pricesApi = {
  refresh: (portfolioId?: string) =>
    request<{ refreshed: string[] }>('/prices/refresh', {
      method: 'POST',
      params: portfolioId ? { portfolio_id: portfolioId } : {},
    }),
}

export const settingsApi = {
  listCurrencies: () => request<CurrencyList>('/settings/currencies'),
  addCurrency: (code: string, name?: string) =>
    request<Currency>('/settings/currencies', { method: 'POST', body: { code, name } }),
  deleteCurrency: (code: string) =>
    request<void>(`/settings/currencies/${encodeURIComponent(code)}`, { method: 'DELETE' }),
}
