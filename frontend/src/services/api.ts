import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true
      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken) {
        try {
          const { data } = await axios.post('/api/v1/auth/refresh', {
            refresh_token: refreshToken,
          })
          localStorage.setItem('access_token', data.access_token)
          localStorage.setItem('refresh_token', data.refresh_token)
          original.headers.Authorization = `Bearer ${data.access_token}`
          return api(original)
        } catch {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          window.location.href = '/login'
        }
      }
    }
    return Promise.reject(error)
  },
)

export default api

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

export interface Transaction {
  id: string
  portfolio_id: string
  asset_id: string
  asset_ticker: string
  asset_name: string
  type: 'buy' | 'sell' | 'dividend' | 'split' | 'fee'
  quantity: string
  price: string
  currency: string
  exchange_rate: string
  fees: string
  date: string
  notes: string
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
}

export interface PortfolioPerformanceSummary {
  portfolio_id: string
  portfolio_name: string
  currency: string
  invested: string
  value: string
  gain_loss: string
  gain_loss_pct: string
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
}

export interface PortfolioAssets {
  portfolio_id: string
  portfolio_name: string
  currency: string
  assets: AssetPerformance[]
}

export interface PortfolioHistory {
  portfolio_id: string
  portfolio_name: string
  currency: string
  series: PortfolioPerformance[]
}

export interface Dashboard {
  by_currency: CurrencyPerformance[]
  portfolios: PortfolioPerformanceSummary[]
  assets: PortfolioAssets[]
  history: PortfolioHistory[]
}

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}

export const authApi = {
  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password }),
  register: (email: string, name: string, password: string) =>
    api.post<User>('/auth/register', { email, name, password }),
  me: () => api.get<User>('/users/me'),
}

export const portfolioApi = {
  list: () => api.get<Portfolio[]>('/portfolios'),
  create: (data: Partial<Portfolio>) => api.post<Portfolio>('/portfolios', data),
  get: (id: string) => api.get<Portfolio>(`/portfolios/${id}`),
  update: (id: string, data: Partial<Portfolio>) => api.patch(`/portfolios/${id}`, data),
  delete: (id: string) => api.delete(`/portfolios/${id}`),
  summary: (id: string) => api.get<PortfolioSummary>(`/portfolios/${id}/summary`),
  allocation: (id: string) => api.get<AssetAllocation[]>(`/portfolios/${id}/allocation`),
  performance: (id: string) => api.get<PortfolioPerformance[]>(`/portfolios/${id}/performance`),
  roi: (id: string) => api.get<AssetROI[]>(`/portfolios/${id}/roi`),
  dashboard: () => api.get<Dashboard>('/dashboard'),
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

export const assetApi = {
  list: () => api.get<Asset[]>('/assets'),
  search: (q: string) => api.get<Asset[]>(`/assets/search?q=${q}`),
  lookup: (q: string) => api.get<AssetLookupResult[]>(`/assets/lookup?q=${q}`),
  meta: (ticker: string) => api.get<AssetMeta>(`/assets/meta?ticker=${ticker}`),
  get: (id: string) => api.get<Asset>(`/assets/${id}`),
  create: (data: Partial<Asset>) => api.post<Asset>('/assets', data),
  remove: (id: string) => api.delete(`/assets/${id}`),
}

export const transactionApi = {
  list: (portfolioId: string) =>
    api.get<Transaction[]>(`/portfolios/${portfolioId}/transactions`),
  create: (portfolioId: string, data: Partial<Transaction>) =>
    api.post<Transaction>(`/portfolios/${portfolioId}/transactions`, data),
  update: (id: string, data: Partial<Transaction>) =>
    api.patch<Transaction>(`/transactions/${id}`, data),
  remove: (id: string) => api.delete(`/transactions/${id}`),
}

export const pricesApi = {
  refresh: (portfolioId?: string) =>
    api.post<{ refreshed: string[] }>('/prices/refresh', null, {
      params: portfolioId ? { portfolio_id: portfolioId } : {},
    }),
}
