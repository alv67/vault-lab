import { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { assetApi, type AssetLookupResult } from '@/services/api'
import toast from 'react-hot-toast'
import { Plus, Loader2, Search } from 'lucide-react'

export default function AssetsPage() {
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({
    ticker: '',
    name: '',
    type: 'stock',
    currency: 'USD',
    country: '',
  })
  const [lookupQuery, setLookupQuery] = useState('')
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [selectedTicker, setSelectedTicker] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const queryClient = useQueryClient()

  const { data: lookupResults, isFetching: lookupLoading } = useQuery({
    queryKey: ['asset-lookup', lookupQuery],
    queryFn: () => assetApi.lookup(lookupQuery),
    enabled: lookupQuery.length >= 1 && !selectedTicker,
  })

  useEffect(() => {
    if (!showCreate) {
      setLookupQuery('')
      setShowSuggestions(false)
      setSelectedTicker(false)
    }
  }, [showCreate])

  const handleTickerChange = (value: string) => {
    setForm({ ...form, ticker: value.toUpperCase() })
    setSelectedTicker(false)
    setLookupQuery(value)
    setShowSuggestions(value.length >= 1)
  }

  const selectSuggestion = (result: AssetLookupResult) => {
    setForm({
      ticker: result.ticker,
      name: result.name || '',
      type: result.type || 'stock',
      currency: result.currency || 'USD',
      country: '',
    })
    setSelectedTicker(true)
    setShowSuggestions(false)
    setLookupQuery(result.ticker)
  }

  const { data, isLoading } = useQuery({
    queryKey: ['assets'],
    queryFn: () => assetApi.list(),
  })

  const createMutation = useMutation({
    mutationFn: () => assetApi.create(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assets'] })
      setShowCreate(false)
      setForm({ ticker: '', name: '', type: 'stock', currency: 'USD', country: '' })
      toast.success('Asset created')
    },
  })

  return (
    <div className="p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Assets</h1>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Add Asset
        </button>
      </div>

      {showCreate && (
        <div className="mb-6 rounded-xl border bg-white p-4">
          <h2 className="mb-3 font-semibold">New Asset</h2>
          <div className="grid grid-cols-2 gap-3">
            <div className="relative">
              <input
                ref={inputRef}
                type="text"
                placeholder="Ticker (e.g. AAPL)"
                value={form.ticker}
                onChange={(e) => handleTickerChange(e.target.value)}
                onFocus={() => form.ticker.length >= 1 && setShowSuggestions(true)}
                onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
                className="w-full rounded-lg border px-3 py-2 pr-8 text-sm"
              />
              {lookupLoading && (
                <Loader2 className="absolute right-2 top-2.5 h-4 w-4 animate-spin text-gray-400" />
              )}
              {!lookupLoading && form.ticker && !selectedTicker && (
                <Search className="absolute right-2 top-2.5 h-4 w-4 text-gray-400" />
              )}

              {showSuggestions && lookupResults?.data && lookupResults.data.length > 0 && !selectedTicker && (
                <div className="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-lg border bg-white shadow-lg">
                  {lookupResults.data.map((r) => (
                    <button
                      key={r.ticker}
                      onMouseDown={() => selectSuggestion(r)}
                      className="flex w-full items-center gap-3 px-3 py-2 text-left text-sm hover:bg-gray-50"
                    >
                      <span className="font-medium">{r.ticker}</span>
                      <span className="flex-1 truncate text-gray-500">{r.name}</span>
                      <span className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">{r.exchange}</span>
                      <span className="text-xs text-gray-400">{r.type}</span>
                    </button>
                  ))}
                </div>
              )}

              {showSuggestions && lookupResults?.data && lookupResults.data.length === 0 && !selectedTicker && (
                <div className="absolute z-10 mt-1 w-full rounded-lg border bg-white p-3 text-center text-sm text-gray-400 shadow-lg">
                  No results found
                </div>
              )}
            </div>

            <input
              type="text"
              placeholder="Name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            />
            <select
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            >
              <option value="stock">Stock</option>
              <option value="etf">ETF</option>
              <option value="bond">Bond</option>
              <option value="crypto">Crypto</option>
              <option value="commodity">Commodity</option>
            </select>
            <input
              type="text"
              placeholder="Country"
              value={form.country}
              onChange={(e) => setForm({ ...form, country: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            />
            <select
              value={form.currency}
              onChange={(e) => setForm({ ...form, currency: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            >
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
              <option value="GBP">GBP</option>
              <option value="CHF">CHF</option>
            </select>
          </div>
          <button
            onClick={() => createMutation.mutate()}
            disabled={!form.ticker || !form.name || createMutation.isPending}
            className="mt-3 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {createMutation.isPending ? 'Saving...' : 'Save'}
          </button>
        </div>
      )}

      {isLoading ? (
        <p className="text-gray-500">Loading...</p>
      ) : (
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b text-gray-500">
              <th className="pb-2">Ticker</th>
              <th className="pb-2">Name</th>
              <th className="pb-2">Type</th>
              <th className="pb-2">Currency</th>
              <th className="pb-2">Country</th>
            </tr>
          </thead>
          <tbody>
            {data?.data?.map((a) => (
              <tr key={a.id} className="border-b last:border-0">
                <td className="py-2 font-medium">{a.ticker}</td>
                <td className="py-2 text-gray-600">{a.name}</td>
                <td className="py-2">
                  <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs">
                    {a.type}
                  </span>
                </td>
                <td className="py-2">{a.currency}</td>
                <td className="py-2">{a.country || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
