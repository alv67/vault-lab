import { useQuery } from '@tanstack/react-query'
import { portfolioApi } from '@/services/api'
import { Link } from 'react-router-dom'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import { formatCurrency, formatPercent } from '@/lib/format'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

export default function DashboardPage() {
  const { data: portfolios } = useQuery({
    queryKey: ['portfolios'],
    queryFn: () => portfolioApi.list(),
  })

  const firstPortfolio = portfolios?.data?.[0]
  const firstPortfolioId = firstPortfolio?.id
  const currency = firstPortfolio?.currency || 'USD'

  const { data: summary } = useQuery({
    queryKey: ['portfolio-summary', firstPortfolioId],
    queryFn: () => portfolioApi.summary(firstPortfolioId!),
    enabled: !!firstPortfolioId,
  })

  const { data: allocation } = useQuery({
    queryKey: ['portfolio-allocation', firstPortfolioId],
    queryFn: () => portfolioApi.allocation(firstPortfolioId!),
    enabled: !!firstPortfolioId,
  })

  const { data: performance } = useQuery({
    queryKey: ['portfolio-performance', firstPortfolioId],
    queryFn: () => portfolioApi.performance(firstPortfolioId!),
    enabled: !!firstPortfolioId,
  })

  const { data: roi } = useQuery({
    queryKey: ['portfolio-roi', firstPortfolioId],
    queryFn: () => portfolioApi.roi(firstPortfolioId!),
    enabled: !!firstPortfolioId,
  })

  const s = summary?.data
  const gainLossClass = s && Number(s.gain_loss) >= 0 ? 'text-green-600' : 'text-red-600'

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">Dashboard</h1>

      {!firstPortfolioId ? (
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
        <>
          {/* Summary cards */}
          <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div className="rounded-xl bg-white p-4 shadow">
              <p className="text-sm text-gray-500">Total Value</p>
                    <p className="text-2xl font-bold">
                        {formatCurrency(s?.total_value || 0, currency)}
                      </p>
            </div>
            <div className="rounded-xl bg-white p-4 shadow">
              <p className="text-sm text-gray-500">Total Cost</p>
              <p className="text-2xl font-bold">
                {formatCurrency(s?.total_cost || 0, currency)}
              </p>
            </div>
            <div className="rounded-xl bg-white p-4 shadow">
              <p className="text-sm text-gray-500">Gain/Loss</p>
              <p className={`text-2xl font-bold ${gainLossClass}`}>
                {formatCurrency(s?.gain_loss || 0, currency)}
                <span className="ml-1 text-sm">
                  ({formatPercent(s?.gain_loss_pct || 0)})
                </span>
              </p>
            </div>
            <div className="rounded-xl bg-white p-4 shadow">
              <p className="text-sm text-gray-500">Assets</p>
              <p className="text-2xl font-bold">{s?.asset_count || 0}</p>
            </div>
          </div>

          <div className="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Performance chart */}
            <div className="rounded-xl bg-white p-4 shadow">
              <h2 className="mb-4 font-semibold">Portfolio Performance</h2>
              {performance?.data?.length ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={performance.data}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="date"
                      tick={{ fontSize: 12 }}
                      tickFormatter={(v) => new Date(v).toLocaleDateString()}
                    />
                    <YAxis tick={{ fontSize: 12 }} />
                    <Tooltip />
                    <Line
                      type="monotone"
                      dataKey="value"
                      stroke="#3b82f6"
                      dot={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <p className="text-sm text-gray-400">No data yet</p>
              )}
            </div>

            {/* Allocation pie chart */}
            <div className="rounded-xl bg-white p-4 shadow">
              <h2 className="mb-4 font-semibold">Asset Allocation</h2>
              {allocation?.data?.length ? (
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie
                      data={allocation.data}
                      dataKey="value"
                      nameKey="ticker"
                      cx="50%"
                      cy="50%"
                      outerRadius={100}
                      label={({ ticker, alloc_pct }) =>
                        `${ticker} (${Number(alloc_pct).toFixed(1)}%)`
                      }
                    >
                      {allocation.data.map((_, i) => (
                        <Cell
                          key={i}
                          fill={COLORS[i % COLORS.length]}
                        />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              ) : (
                <p className="text-sm text-gray-400">No data yet</p>
              )}
            </div>
          </div>

          {/* ROI table */}
          <div className="rounded-xl bg-white p-4 shadow">
            <h2 className="mb-4 font-semibold">ROI by Asset</h2>
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b text-gray-500">
                  <th className="pb-2">Ticker</th>
                  <th className="pb-2">Name</th>
                  <th className="pb-2">Invested</th>
                  <th className="pb-2">Value</th>
                  <th className="pb-2">ROI</th>
                </tr>
              </thead>
              <tbody>
                {roi?.data?.map((r) => (
                  <tr key={r.asset_id} className="border-b last:border-0">
                    <td className="py-2 font-medium">{r.ticker}</td>
                    <td className="py-2 text-gray-600">{r.name}</td>
                    <td className="py-2">{formatCurrency(r.total_invested, currency)}</td>
                    <td className="py-2">{formatCurrency(r.current_value, currency)}</td>
                    <td
                      className={`py-2 font-medium ${Number(r.roi) >= 0 ? 'text-green-600' : 'text-red-600'}`}
                    >
                      {formatPercent(r.roi)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
