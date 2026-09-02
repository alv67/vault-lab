const BASE_URL = '/api/v1'

// In-memory GET cache with TTL (~60s). Module-level so it survives component
// mounts/unmounts and page navigations within the SPA, avoiding refetching
// heavy endpoints (dashboard, summaries, history) on every mount.
const CACHE_TTL_MS = 60_000

interface CacheEntry {
  data: unknown
  expiresAt: number
}

const getCache = new Map<string, CacheEntry>()

// Deep-copy cached data so callers never share mutable references with the
// cache entry (a caller mutating a response would otherwise poison it).
// Falls back to JSON round-trip, then to the raw reference: it must never
// break the request.
function cloneCached<T>(data: unknown): T {
  try {
    if (typeof structuredClone === 'function') {
      return structuredClone(data) as T
    }
  } catch {
    // structuredClone unavailable/failed — fall through to JSON copy
  }
  try {
    return JSON.parse(JSON.stringify(data)) as T
  } catch {
    // Last resort: return the reference rather than failing the request.
    return data as T
  }
}

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
  asset_class: string
  country: string
  currency: string
  exchange?: string
  sector?: string
  industry?: string
  price_source?: string
}

export interface AssetQuote {
  currency: string
  last_close: string
  last_date: string
  change_1d: string
  change_1w: string
  change_1m: string
  change_1y: string
  change_ytd: string
  has_data: boolean
}

// Body accettato da PATCH /assets/{id}. Campi omessi = invariati; stringa
// vuota = svuota (isin, country, exchange, sector, industry).
export interface AssetPatch {
  ticker?: string
  isin?: string
  name?: string
  type?: string
  asset_class?: string
  country?: string
  currency?: string
  exchange?: string
  sector?: string
  industry?: string
  price_source?: string
}

export interface Price {
  id: string
  asset_id: string
  date: string
  open: string
  high: string
  low: string
  close: string
  volume: number
  source: string
  created_at: string
}

export interface ExposureRow {
  name: string
  weight: string
}

export interface AssetExposure {
  regions: ExposureRow[]
  sectors: ExposureRow[]
  isin?: string
}

// Body accettato da PUT /assets/{id}/exposure. Le dimensioni sono
// indipendenti: omettendo una chiave la relativa distribuzione non viene
// modificata.
export interface AssetExposurePatch {
  regions?: ExposureRow[]
  sectors?: ExposureRow[]
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

export interface AssetClassSlice {
  class: string
  value: string
  weight: string
}

export interface PortfolioClassAllocation {
  currency: string
  classes: AssetClassSlice[]
}

export interface RegionAllocation {
  region: string
  value: string
  weight: string
}

export interface SectorAllocation {
  sector: string
  value: string
  weight: string
}

export interface PortfolioGeographyAllocation {
  currency: string
  regions: RegionAllocation[]
  covered_value?: string
  excluded_value?: string
}

export interface PortfolioSectorAllocation {
  currency: string
  sectors: SectorAllocation[]
  covered_value?: string
  excluded_value?: string
}

export interface DashboardAllocation {
  currency: string
  regions: RegionAllocation[]
  sectors: SectorAllocation[]
  covered_value?: string
  excluded_value?: string
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

export interface FetchIssue {
  symbol: string
  code: string
  message: string
}

export interface RefreshReport {
  refreshed: string[]
  issues: FetchIssue[]
  rate_limited: boolean
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
  asset_class: string
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
  const method = options.method ?? 'GET'
  const url = buildUrl(path, options.params)
  const cacheKey = `${method} ${url}`

  // Any non-GET invalidates the whole cache. This is intentionally coarse:
  // dashboard/statistics endpoints aggregate across all portfolios/assets, so
  // a single mutation (transaction, portfolio edit, or /prices/refresh) can
  // change any cached GET. Clearing everything is simpler and safer than
  // tracking fine-grained dependencies, at the cost of a refetch after each
  // mutation.
  if (method !== 'GET') {
    getCache.clear()
  } else {
    const cached = getCache.get(cacheKey)
    if (cached) {
      if (cached.expiresAt > Date.now()) {
        return cloneCached<T>(cached.data)
      }
      getCache.delete(cacheKey)
    }
  }

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = localStorage.getItem('access_token')
  if (token) headers.Authorization = `Bearer ${token}`

  const init = (): RequestInit => ({
    method,
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
    // Do not cache failures (4xx/5xx): the normal throw/retry flow applies.
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

  const data: T = await res.json()
  if (method === 'GET') {
    getCache.set(cacheKey, { data, expiresAt: Date.now() + CACHE_TTL_MS })
  }
  return cloneCached<T>(data)
}

export const api = {
  get: (url: string, init: RequestInit = {}) => request(url, { ...init, method: 'GET' }),
  post: (url: string, body: unknown, init: RequestInit = {}) => request(url, { ...init, method: 'POST', body }),
  put: (url: string, body: unknown, init: RequestInit = {}) => request(url, { ...init, method: 'PUT', body }),
  patch: (url: string, body: unknown, init: RequestInit = {}) => request(url, { ...init, method: 'PATCH', body }),
  delete: (url: string, init: RequestInit = {}) => request(url, { ...init, method: 'DELETE' }),
};

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
  classAllocation: (id: string) =>
    request<PortfolioClassAllocation>(`/portfolios/${id}/allocation/class`),
  geographyAllocation: (id: string) =>
    request<PortfolioGeographyAllocation>(`/portfolios/${id}/allocation/geography`),
  sectorAllocation: (id: string) =>
    request<PortfolioSectorAllocation>(`/portfolios/${id}/allocation/sector`),
  dashboardAllocation: () => request<DashboardAllocation>('/dashboard/allocation'),
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
  update: (id: string, patch: AssetPatch) =>
    request<Asset>(`/assets/${id}`, { method: 'PATCH', body: patch }),
  quote: (id: string) => request<AssetQuote>(`/assets/${id}/quote`),
  splits: (id: string) => request<SplitInfo[]>(`/assets/${id}/splits`),
  fetchProfile: (id: string) =>
    request<Asset>(`/assets/${id}/fetch-profile`, { method: 'POST' }),
  exposure: (id: string) => request<AssetExposure>(`/assets/${id}/exposure`),
  saveExposure: (id: string, exposure: AssetExposurePatch) =>
    request<AssetExposure>(`/assets/${id}/exposure`, { method: 'PUT', body: exposure }),
  fetchExposure: (id: string) =>
    request<AssetExposure>(`/assets/${id}/fetch-exposure`, { method: 'POST' }),
  fetchETFExposure: (id: string) =>
    request<AssetExposure>(`/assets/${id}/fetch-etf-exposure`, { method: 'POST' }),
  backfillHistory: (id: string) =>
    request<{ status: string }>(`/assets/${id}/backfill-history`, { method: 'POST' }),
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
    request<RefreshReport>('/prices/refresh', {
      method: 'POST',
      params: portfolioId ? { portfolio_id: portfolioId } : {},
    }),
  byAsset: (assetId: string) => request<Price[]>(`/prices/${assetId}?full=1`),
}

export const settingsApi = {
  listCurrencies: () => request<CurrencyList>('/settings/currencies'),
  addCurrency: (code: string, name?: string) =>
    request<Currency>('/settings/currencies', { method: 'POST', body: { code, name } }),
  deleteCurrency: (code: string) =>
    request<void>(`/settings/currencies/${encodeURIComponent(code)}`, { method: 'DELETE' }),
}
