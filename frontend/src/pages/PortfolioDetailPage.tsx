import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { portfolioApi, transactionApi, assetApi, pricesApi, type Transaction } from '@/services/api'
import toast from 'react-hot-toast'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { formatCurrency, formatPercent } from '@/lib/format'

const defaultForm = () => ({
  asset_id: '',
  type: 'buy' as string,
  quantity: '',
  price: '',
  date: new Date().toISOString().split('T')[0],
  fees: '0',
  notes: '',
})

export default function PortfolioDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const refreshed = useRef<Set<string>>(new Set())
  const [showTx, setShowTx] = useState(false)
  const [editingTx, setEditingTx] = useState<Transaction | null>(null)
  const [txForm, setTxForm] = useState(defaultForm())

  useEffect(() => {
    if (!id || refreshed.current.has(id)) return
    refreshed.current.add(id)
    pricesApi
      .refresh(id)
      .then(() => {
        queryClient.invalidateQueries({ queryKey: ['portfolio-summary', id] })
      })
      .catch(() => {})
  }, [id, queryClient])

  const { data: portfolio } = useQuery({
    queryKey: ['portfolio', id],
    queryFn: () => portfolioApi.get(id!),
    enabled: !!id,
  })

  const { data: summary } = useQuery({
    queryKey: ['portfolio-summary', id],
    queryFn: () => portfolioApi.summary(id!),
    enabled: !!id,
  })

  const { data: transactions } = useQuery({
    queryKey: ['transactions', id],
    queryFn: () => transactionApi.list(id!),
    enabled: !!id,
  })

  const s = summary?.data
  const currency = portfolio?.data?.currency || 'USD'
  const gainLossClass = s && Number(s.gain_loss) >= 0 ? 'text-green-600' : 'text-red-600'

  const { data: assets } = useQuery({
    queryKey: ['assets'],
    queryFn: () => assetApi.list(),
  })

  const closeTx = () => {
    setShowTx(false)
    setEditingTx(null)
    setTxForm(defaultForm())
  }

  const openNewTx = () => {
    setEditingTx(null)
    setTxForm(defaultForm())
    setShowTx(true)
  }

  const startEdit = (tx: Transaction) => {
    setEditingTx(tx)
    setTxForm({
      asset_id: tx.asset_id,
      type: tx.type,
      quantity: tx.quantity,
      price: tx.price,
      date: new Date(tx.date).toISOString().split('T')[0],
      fees: tx.fees || '0',
      notes: tx.notes || '',
    })
    setShowTx(true)
  }

  const txMutation = useMutation({
    mutationFn: () => {
      const payload = {
        asset_id: txForm.asset_id,
        type: txForm.type as any,
        quantity: txForm.quantity,
        price: txForm.price,
        date: new Date(txForm.date).toISOString(),
        fees: txForm.fees,
        notes: txForm.notes,
      }
      if (editingTx) {
        return transactionApi.update(editingTx.id, {
          ...payload,
          currency: editingTx.currency,
          exchange_rate: editingTx.exchange_rate,
        })
      }
      return transactionApi.create(id!, payload)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions', id] })
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary', id] })
      queryClient.invalidateQueries({ queryKey: ['dashboard'] })
      closeTx()
      toast.success(editingTx ? 'Transaction updated' : 'Transaction added')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (txId: string) => transactionApi.remove(txId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions', id] })
      queryClient.invalidateQueries({ queryKey: ['portfolio-summary', id] })
      queryClient.invalidateQueries({ queryKey: ['dashboard'] })
      closeTx()
      toast.success('Transaction deleted')
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Delete failed')
    },
  })

  return (
    <div className="p-6">
      <h1 className="mb-2 text-2xl font-bold">{portfolio?.data?.name}</h1>
      <p className="mb-6 text-sm text-gray-500">{portfolio?.data?.description}</p>

      {/* Summary */}
      <div className="mb-6 grid grid-cols-3 gap-4">
        <div className="rounded-xl bg-white p-4 shadow">
          <p className="text-sm text-gray-500">Value</p>
          <p className="text-xl font-bold">{formatCurrency(s?.total_value || 0, currency)}</p>
        </div>
        <div className="rounded-xl bg-white p-4 shadow">
          <p className="text-sm text-gray-500">Gain/Loss</p>
          <p className={`text-xl font-bold ${gainLossClass}`}>
            {formatCurrency(s?.gain_loss || 0, currency)} ({formatPercent(s?.gain_loss_pct || 0)})
          </p>
        </div>
        <div className="rounded-xl bg-white p-4 shadow">
          <p className="text-sm text-gray-500">Assets</p>
          <p className="text-xl font-bold">{s?.asset_count || 0}</p>
        </div>
      </div>

      {/* Add transaction button */}
      <button
        onClick={() => (showTx ? closeTx() : openNewTx())}
        className="mb-4 flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
      >
        <Plus className="h-4 w-4" />
        Add Transaction
      </button>

      {showTx && (
        <div className="mb-6 rounded-xl border bg-white p-4">
          <h3 className="mb-3 font-semibold">{editingTx ? 'Edit Transaction' : 'New Transaction'}</h3>
          <div className="grid grid-cols-2 gap-3">
            <select
              value={txForm.asset_id}
              onChange={(e) => setTxForm({ ...txForm, asset_id: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            >
              <option value="">Select asset</option>
              {assets?.data?.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.ticker} - {a.name}
                </option>
              ))}
            </select>
            <select
              value={txForm.type}
              onChange={(e) => setTxForm({ ...txForm, type: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            >
              <option value="buy">Buy</option>
              <option value="sell">Sell</option>
              <option value="dividend">Dividend</option>
            </select>
            <input
              type="number"
              placeholder="Quantity"
              value={txForm.quantity}
              onChange={(e) => setTxForm({ ...txForm, quantity: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            />
            <input
              type="number"
              step="0.01"
              placeholder="Price"
              value={txForm.price}
              onChange={(e) => setTxForm({ ...txForm, price: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            />
            <input
              type="date"
              value={txForm.date}
              onChange={(e) => setTxForm({ ...txForm, date: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            />
            <input
              type="number"
              step="0.01"
              placeholder="Fees"
              value={txForm.fees}
              onChange={(e) => setTxForm({ ...txForm, fees: e.target.value })}
              className="rounded-lg border px-3 py-2 text-sm"
            />
          </div>
          <div className="mt-3 flex items-center gap-3">
            <button
              onClick={() => txMutation.mutate()}
              disabled={!txForm.asset_id || !txForm.quantity || !txForm.price || txMutation.isPending}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {txMutation.isPending ? 'Saving...' : editingTx ? 'Save Changes' : 'Save'}
            </button>
            {editingTx && (
              <button
                onClick={() => {
                  if (window.confirm('Delete this transaction?')) {
                    deleteMutation.mutate(editingTx.id)
                  }
                }}
                disabled={deleteMutation.isPending}
                className="flex items-center gap-1.5 rounded-lg bg-red-600 px-4 py-2 text-sm text-white hover:bg-red-700 disabled:opacity-50"
              >
                <Trash2 className="h-4 w-4" />
                Delete
              </button>
            )}
            <button
              onClick={closeTx}
              className="rounded-lg px-4 py-2 text-sm text-gray-600 hover:bg-gray-100"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Transactions */}
      <div className="rounded-xl bg-white p-4 shadow">
        <h2 className="mb-4 font-semibold">Transactions</h2>
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b text-gray-500">
              <th className="pb-2">Date</th>
              <th className="pb-2">Asset</th>
              <th className="pb-2">Type</th>
              <th className="pb-2">Quantity</th>
              <th className="pb-2">Price</th>
              <th className="pb-2">Total</th>
              <th className="pb-2 text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {transactions?.data?.map((tx) => (
              <tr key={tx.id} className="border-b last:border-0">
                <td className="py-2">{new Date(tx.date).toLocaleDateString()}</td>
                <td className="py-2">
                  <span className="font-medium">{tx.asset_ticker}</span>
                  <span className="ml-1 text-xs text-gray-500">{tx.asset_name}</span>
                </td>
                <td className="py-2">
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                      tx.type === 'buy'
                        ? 'bg-green-100 text-green-700'
                        : tx.type === 'sell'
                          ? 'bg-red-100 text-red-700'
                          : 'bg-blue-100 text-blue-700'
                    }`}
                  >
                    {tx.type}
                  </span>
                </td>
                <td className="py-2">{tx.quantity}</td>
                <td className="py-2">{formatCurrency(tx.price, currency)}</td>
                <td className="py-2">
                  {formatCurrency(Number(tx.quantity) * Number(tx.price), currency)}
                </td>
                <td className="py-2 text-right">
                  <button
                    onClick={() => startEdit(tx)}
                    className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                    title="Edit transaction"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
