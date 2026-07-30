import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { portfolioApi } from '@/services/api'
import { Link } from 'react-router-dom'
import toast from 'react-hot-toast'
import { Plus, ExternalLink } from 'lucide-react'

export default function PortfoliosPage() {
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [currency, setCurrency] = useState('USD')
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['portfolios'],
    queryFn: () => portfolioApi.list(),
  })

  const createMutation = useMutation({
    mutationFn: () =>
      portfolioApi.create({ name, description, currency }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['portfolios'] })
      setShowCreate(false)
      setName('')
      setDescription('')
      toast.success('Portfolio created')
    },
    onError: () => toast.error('Failed to create portfolio'),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => portfolioApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['portfolios'] })
      toast.success('Portfolio deleted')
    },
  })

  return (
    <div className="p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Portfolios</h1>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          New Portfolio
        </button>
      </div>

      {showCreate && (
        <div className="mb-6 rounded-xl border bg-white p-4">
          <h2 className="mb-4 font-semibold">Create Portfolio</h2>
          <div className="space-y-3">
            <input
              type="text"
              placeholder="Portfolio name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-lg border px-3 py-2 text-sm"
            />
            <textarea
              placeholder="Description (optional)"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-lg border px-3 py-2 text-sm"
            />
            <select
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              className="w-full rounded-lg border px-3 py-2 text-sm"
            >
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
              <option value="GBP">GBP</option>
              <option value="CHF">CHF</option>
            </select>
            <button
              onClick={() => createMutation.mutate()}
              disabled={!name || createMutation.isPending}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <p className="text-gray-500">Loading...</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data?.data?.map((p) => (
            <div
              key={p.id}
              className="rounded-xl border bg-white p-4 shadow-sm"
            >
              <div className="mb-2 flex items-start justify-between">
                <div>
                  <h3 className="font-semibold">{p.name}</h3>
                  <p className="text-xs text-gray-500">{p.currency}</p>
                </div>
                <Link
                  to={`/portfolios/${p.id}`}
                  className="text-blue-600 hover:text-blue-800"
                >
                  <ExternalLink className="h-4 w-4" />
                </Link>
              </div>
              {p.description && (
                <p className="mb-3 text-sm text-gray-600">{p.description}</p>
              )}
              <button
                onClick={() => {
                  if (confirm('Delete this portfolio?')) {
                    deleteMutation.mutate(p.id)
                  }
                }}
                className="text-xs text-red-500 hover:underline"
              >
                Delete
              </button>
            </div>
          ))}
          {data?.data?.length === 0 && (
            <p className="col-span-full text-center text-gray-400">
              No portfolios yet. Create one!
            </p>
          )}
        </div>
      )}
    </div>
  )
}
