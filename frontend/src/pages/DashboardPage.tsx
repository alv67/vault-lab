import { useEffect, useMemo, useRef, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { portfolioApi, pricesApi } from '@/services/api'
import { Link } from 'react-router-dom'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { formatCurrency, formatPercent } from '@/lib/format'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316']

function glClass(value?: string | number) {
  return Number(value ?? 0) >= 0 ? 'text-green-600' : 'text-red-600'
}

export default function DashboardPage() {
  const queryClient = useQueryClient()
  const refreshed = useRef(false)
  const initialized = useRef(false)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  const { data: dash } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => portfolioApi.dashboard(),
  })

  useEffect(() => {
    if (refreshed.current) return
    refreshed.current = true
    pricesApi
      .refresh()
      .then(() => {
        queryClient.invalidateQueries({ queryKey: ['dashboard'] })
      })
      .catch(() => {})
  }, [queryClient])

  const firstPortfolioId = dash?.data?.assets?.[0]?.portfolio_id
  useEffect(() => {
    if (firstPortfolioId && !initialized.current) {
      initialized.current = true
      setExpanded(new Set([firstPortfolioId]))
    }
  }, [firstPortfolioId])

  const toggle = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const chartData = useMemo(() => {
    const histories = dash?.data?.history ?? []
    if (!histories.length) return []
    const dateSet = new Set<string>()
    histories.forEach((h) => h.series.forEach((pt) => dateSet.add(pt.date.slice(0, 10))))
    const dates = [...dateSet].sort()
    return dates.map((date) => {
      const row: Record<string, string | number | null> = { date }
      histories.forEach((h) => {
        const pt = h.series.find((s) => s.date.slice(0, 10) === date)
        row[h.portfolio_id] = pt ? Number(pt.value) : null
      })
      return row
    })
  }, [dash])

  const hasMultipleCurrencies = (dash?.data?.by_currency?.length ?? 0) > 1

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">Dashboard</h1>

      {!dash?.data?.portfolios?.length ? (
        <div className="rounded-lg border-2 border-dashed p-12 text-center">
          <p className="mb-4 text-gray-500">No portfolios yet</p>
          <Link
            to="/portfolios"
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white"
          >
            Create your first portfolio
          </Link>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Historical chart, one line per portfolio */}
          <div className="rounded-xl bg-white p-4 shadow">
            <h2 className="mb-4 font-semibold">Portfolio History</h2>
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={320}>
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis
                    dataKey="date"
                    tick={{ fontSize: 12 }}
                    tickFormatter={(v: string) => new Date(v).toLocaleDateString()}
                  />
                  <YAxis tick={{ fontSize: 12 }} />
                  <Tooltip
                    formatter={(value, name) => {
                      const hist = dash.data.history.find((h) => h.portfolio_id === name)
                      return [formatCurrency(Number(value), hist?.currency), hist?.portfolio_name ?? String(name)]
                    }}
                  />
                  <Legend />
                  {dash.data.history.map((h, i) => (
                    <Line
                      key={h.portfolio_id}
                      dataKey={h.portfolio_id}
                      name={h.portfolio_name}
                      type="monotone"
                      stroke={COLORS[i % COLORS.length]}
                      dot={false}
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <p className="text-sm text-gray-400">No price history yet</p>
            )}
          </div>

          {/* Performance by currency */}
          {hasMultipleCurrencies && (
            <div className="rounded-xl bg-white p-4 shadow">
              <h2 className="mb-4 font-semibold">Performance by Currency</h2>
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b text-gray-500">
                    <th className="pb-2">Currency</th>
                    <th className="pb-2">Invested</th>
                    <th className="pb-2">Value</th>
                    <th className="pb-2">Gain/Loss</th>
                    <th className="pb-2">Return</th>
                  </tr>
                </thead>
                <tbody>
                  {dash.data.by_currency.map((c) => (
                    <tr key={c.currency} className="border-b last:border-0">
                      <td className="py-2 font-medium">{c.currency}</td>
                      <td className="py-2">{formatCurrency(c.invested, c.currency)}</td>
                      <td className="py-2">{formatCurrency(c.value, c.currency)}</td>
                      <td className={`py-2 font-medium ${glClass(c.gain_loss)}`}>
                        {formatCurrency(c.gain_loss, c.currency)}
                      </td>
                      <td className={`py-2 font-medium ${glClass(c.gain_loss_pct)}`}>
                        {formatPercent(c.gain_loss_pct)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* Per-portfolio summary table */}
          <div className="rounded-xl bg-white p-4 shadow">
            <h2 className="mb-4 font-semibold">Portfolios</h2>
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b text-gray-500">
                  <th className="pb-2">Portfolio</th>
                  <th className="pb-2">Currency</th>
                  <th className="pb-2">Assets</th>
                  <th className="pb-2">Invested</th>
                  <th className="pb-2">Value</th>
                  <th className="pb-2">Gain/Loss</th>
                  <th className="pb-2">Return</th>
                </tr>
              </thead>
              <tbody>
                {dash.data.portfolios.map((p) => (
                  <tr key={p.portfolio_id} className="border-b last:border-0">
                    <td className="py-2 font-medium">{p.portfolio_name}</td>
                    <td className="py-2">{p.currency}</td>
                    <td className="py-2">{p.asset_count}</td>
                    <td className="py-2">{formatCurrency(p.invested, p.currency)}</td>
                    <td className="py-2">{formatCurrency(p.value, p.currency)}</td>
                    <td className={`py-2 font-medium ${glClass(p.gain_loss)}`}>
                      {formatCurrency(p.gain_loss, p.currency)}
                    </td>
                    <td className={`py-2 font-medium ${glClass(p.gain_loss_pct)}`}>
                      {formatPercent(p.gain_loss_pct)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Collapsible per-portfolio asset tables */}
          {dash.data.assets.map((pa) => (
            <div key={pa.portfolio_id} className="rounded-xl bg-white p-4 shadow">
              <button
                onClick={() => toggle(pa.portfolio_id)}
                className="flex w-full items-center gap-2 text-left"
              >
                {expanded.has(pa.portfolio_id) ? (
                  <ChevronDown className="h-4 w-4 text-gray-400" />
                ) : (
                  <ChevronRight className="h-4 w-4 text-gray-400" />
                )}
                <span className="font-semibold">{pa.portfolio_name}</span>
                <span className="ml-1 text-xs text-gray-500">({pa.currency})</span>
                <span className="ml-auto text-sm text-gray-600">
                  {formatCurrency(
                    dash.data.portfolios.find((p) => p.portfolio_id === pa.portfolio_id)?.value ?? 0,
                    pa.currency,
                  )}
                </span>
              </button>

              {expanded.has(pa.portfolio_id) && (
                <div className="mt-3 overflow-x-auto">
                  {pa.assets.length === 0 ? (
                    <p className="text-sm text-gray-400">No assets in this portfolio.</p>
                  ) : (
                    <table className="w-full text-left text-sm">
                      <thead>
                        <tr className="border-b text-gray-500">
                          <th className="pb-2">Ticker</th>
                          <th className="pb-2">Name</th>
                          <th className="pb-2">Currency</th>
                          <th className="pb-2 text-right">Qty</th>
                          <th className="pb-2 text-right">Invested</th>
                          <th className="pb-2 text-right">Value</th>
                          <th className="pb-2 text-right">Gain/Loss</th>
                          <th className="pb-2 text-right">ROI</th>
                        </tr>
                      </thead>
                      <tbody>
                        {pa.assets.map((a) => (
                          <tr key={a.asset_id} className="border-b last:border-0">
                            <td className="py-2 font-medium">{a.ticker}</td>
                            <td className="py-2 text-gray-600">{a.name}</td>
                            <td className="py-2">
                              {a.currency}
                              {a.fx_missing && (
                                <span
                                  className="ml-2 rounded bg-amber-100 px-1.5 py-0.5 text-xs text-amber-700"
                                  title="Exchange rate not available: excluded from portfolio total"
                                >
                                  cambio mancante
                                </span>
                              )}
                            </td>
                            <td className="py-2 text-right">{a.qty}</td>
                            <td className="py-2 text-right">{formatCurrency(a.invested, a.currency)}</td>
                            <td className="py-2 text-right">{formatCurrency(a.value, a.currency)}</td>
                            <td className={`py-2 text-right font-medium ${glClass(a.gain_loss)}`}>
                              {formatCurrency(a.gain_loss, a.currency)}
                            </td>
                            <td className={`py-2 text-right font-medium ${glClass(a.roi)}`}>
                              {formatPercent(a.roi)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
