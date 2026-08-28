<script module lang="ts">
  let sessionRefreshed = false
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import { resolve } from '$app/paths'
  import { toast } from '$lib/stores/toast.svelte'
  import {
    portfolioApi,
    transactionApi,
    assetApi,
    pricesApi,
    type Portfolio,
    type PortfolioSummary,
    type PortfolioHistory,
    type Transaction,
    type Asset,
  } from '$lib/services/api'
  import { formatCurrency, formatPercent, ASSET_CLASS_LABELS } from '$lib/format'
  import PositionChart from '$lib/components/PositionChart.svelte'
  import ExposurePie from '$lib/components/ExposurePie.svelte'
  import {
    type ExposureRow,
    type PortfolioClassAllocation,
  } from '$lib/services/api'
  import { Plus, Pencil, Trash2, Download } from 'lucide-svelte'

  const id = $derived(page.params.id as string | undefined)

  type TxType = Transaction['type']
  interface TxForm {
    asset_id: string
    type: TxType
    quantity: string
    price: string
    date: string
    fees: string
    notes: string
  }

  const defaultForm = (): TxForm => ({
    asset_id: '',
    type: 'buy',
    quantity: '',
    price: '',
    date: new Date().toISOString().split('T')[0],
    fees: '0',
    notes: '',
  })

  let portfolio = $state<Portfolio | null>(null)
  let summary = $state<PortfolioSummary | null>(null)
  let transactions = $state<Transaction[] | null>(null)
  let assets = $state<Asset[] | null>(null)
  let history = $state<PortfolioHistory | null>(null)
  let classAlloc = $state<PortfolioClassAllocation | null>(null)
  let classAllocError = $state(false)
  let selectedAsset = $state('')
  let showTx = $state(false)
  let editingTx = $state<Transaction | null>(null)
  let txForm = $state<TxForm>(defaultForm())
  let txDividendAmount = $state('')
  let txSaving = $state(false)
  let deleting = $state(false)

  const currency = $derived(portfolio?.currency || 'USD')
  const classAllocRows = $derived<ExposureRow[]>(
    (classAlloc?.classes ?? []).map((c) => ({
      name: ASSET_CLASS_LABELS[c.class] ?? c.class,
      weight: c.weight,
    })),
  )
  const gainLossClass = $derived(
    summary && Number(summary.gain_loss) >= 0 ? 'text-green-600' : 'text-red-600',
  )
  const realizedClass = $derived(
    summary && Number(summary.realized_gl) >= 0 ? 'text-green-600' : 'text-red-600',
  )

  function pnlClass(value?: string | number): string {
    return Number(value ?? 0) >= 0 ? 'text-green-600' : 'text-red-600'
  }

  onMount(load)

  async function load(): Promise<void> {
    if (!id) return
    try {
      const [p, s, t, a] = await Promise.all([
        portfolioApi.get(id),
        portfolioApi.summary(id),
        transactionApi.list(id),
        assetApi.list(),
      ])
      portfolio = p
      summary = s
      transactions = t
      assets = a
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load portfolio'
      toast.error(message)
    }

    // L'allocazione per classi è isolata: se il backend non la espone ancora
    // (es. asset_class non popolati) non blocca il resto della pagina.
    try {
      classAlloc = await portfolioApi.classAllocation(id)
    } catch {
      classAllocError = true
    }

    try {
      history = await portfolioApi.history(id)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load history'
      toast.error(message)
    }

    if (!sessionRefreshed) {
      sessionRefreshed = true
      pricesApi.refresh(id)
        .then((report) => {
          if (report.rate_limited) {
            toast.warning('Yahoo Finance ha limitato le richieste: alcuni prezzi non aggiornati')
          } else if (report.issues.length > 0) {
            toast.warning(`${report.issues.length} aggiornamenti prezzi non riusciti (Yahoo)`)
          }
          return portfolioApi.summary(id)
        })
        .then((fresh) => { summary = fresh })
        .catch(() => { /* keep current data */ })
    }
  }

  function closeTx(): void {
    showTx = false
    editingTx = null
    txForm = defaultForm()
    txDividendAmount = ''
  }

  function openNewTx(): void {
    editingTx = null
    txForm = defaultForm()
    txDividendAmount = ''
    showTx = true
  }

  function startEdit(tx: Transaction): void {
    editingTx = tx
    txForm = {
      asset_id: tx.asset_id,
      type: tx.type,
      quantity: tx.quantity,
      price: tx.price,
      date: new Date(tx.date).toISOString().split('T')[0],
      fees: tx.fees || '0',
      notes: tx.notes || '',
    }
    txDividendAmount =
      tx.type === 'dividend' ? String(Number(tx.price) * Number(tx.quantity)) : ''
    showTx = true
  }

  async function saveTransaction(): Promise<void> {
    if (!id) return
    txSaving = true
    const payload = {
      asset_id: txForm.asset_id,
      type: txForm.type,
      quantity: txForm.type === 'dividend' ? '1' : txForm.quantity,
      price: txForm.type === 'dividend' ? txDividendAmount : txForm.price,
      date: new Date(txForm.date).toISOString(),
      fees: txForm.fees,
      notes: txForm.notes,
    }
    try {
      if (editingTx) {
        await transactionApi.update(editingTx.id, {
          ...payload,
        })
      } else {
        await transactionApi.create(id, payload)
      }
      transactions = await transactionApi.list(id)
      summary = await portfolioApi.summary(id)
      history = await portfolioApi.history(id)
      closeTx()
      toast.success(editingTx ? 'Transaction updated' : 'Transaction added')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed'
      toast.error(message)
    } finally {
      txSaving = false
    }
  }

  async function deleteTransaction(txId: string): Promise<void> {
    if (!id) return
    deleting = true
    try {
      await transactionApi.remove(txId)
      transactions = await transactionApi.list(id)
      summary = await portfolioApi.summary(id)
      history = await portfolioApi.history(id)
      closeTx()
      toast.success('Transaction deleted')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Delete failed'
      toast.error(message)
    } finally {
      deleting = false
    }
  }

  function handleDeleteEditing(): void {
    if (!editingTx) return
    if (confirm('Delete this transaction?')) {
      deleteTransaction(editingTx.id)
    }
  }

  function canSave(): boolean {
    if (txForm.type === 'dividend') {
      return Boolean(txForm.asset_id && txDividendAmount && Number(txDividendAmount) > 0)
    }
    return Boolean(txForm.asset_id && txForm.quantity && txForm.price)
  }

  async function exportPortfolio(): Promise<void> {
    if (!id) return
    try {
      const doc = await portfolioApi.exportDoc(id)
      const blob = new Blob([JSON.stringify(doc, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `vault-lab-${doc.portfolio.name.toLowerCase().replace(/\s+/g, '-') || 'portfolio'}.json`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Export failed'
      toast.error(message)
    }
  }
</script>

<div class="p-6">
  <h1 class="mb-2 text-2xl font-bold">{portfolio?.name ?? 'Portfolio'}</h1>
  <p class="mb-6 text-sm text-gray-500">{portfolio?.description}</p>

  <div class="mb-6 grid grid-cols-2 gap-4 md:grid-cols-4">
    <div class="rounded-xl bg-white p-4 shadow">
      <p class="text-sm text-gray-500">Value</p>
      <p class="text-xl font-bold">{formatCurrency(summary?.total_value || 0, currency)}</p>
    </div>
    <div class="rounded-xl bg-white p-4 shadow">
      <p class="text-sm text-gray-500">Realized</p>
      <p class="text-xl font-bold {realizedClass}">
        {formatCurrency(summary?.realized_gl || 0, currency)}
      </p>
    </div>
    <div class="rounded-xl bg-white p-4 shadow">
      <p class="text-sm text-gray-500">Gain/Loss (open)</p>
      <p class="text-xl font-bold {gainLossClass}">
        {formatCurrency(summary?.gain_loss || 0, currency)} ({formatPercent(summary?.gain_loss_pct || 0)})
      </p>
    </div>
    <div class="rounded-xl bg-white p-4 shadow">
      <p class="text-sm text-gray-500">Assets</p>
      <p class="text-xl font-bold">{summary?.asset_count || 0}</p>
    </div>
  </div>

  {#if summary?.holdings && summary.holdings.length > 0}
    <div class="mb-6 rounded-xl bg-white p-4 shadow">
      <h2 class="mb-4 font-semibold">Positions</h2>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="border-b text-gray-500">
              <th class="pb-2">Ticker</th>
              <th class="pb-2 text-right">Qty</th>
              <th class="pb-2 text-right">Cost</th>
              <th class="pb-2 text-right">Value</th>
              <th class="pb-2 text-right">Realized</th>
              <th class="pb-2 text-right">Unrealized</th>
              <th class="pb-2 text-right">ROI</th>
              <th class="pb-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {#each summary.holdings as h (h.asset_id)}
              <tr class="border-b last:border-0">
                <td class="py-2">
                  <a href={resolve(`/assets/${h.asset_id}`)} class="font-medium text-blue-600 hover:underline">
                    {h.ticker}
                  </a>
                  <span class="block text-xs text-gray-500">{h.name}</span>
                </td>
                <td class="py-2 text-right">{h.closed ? '-' : h.qty}</td>
                <td class="py-2 text-right">{h.closed ? '-' : formatCurrency(h.cost, currency)}</td>
                <td class="py-2 text-right">{h.closed ? '-' : formatCurrency(h.value_pf, currency)}</td>
                <td class="py-2 text-right font-medium {pnlClass(h.realized)}">
                  {formatCurrency(h.realized, currency)}
                </td>
                <td class="py-2 text-right font-medium {pnlClass(h.unrealized)}">
                  {h.closed ? '-' : formatCurrency(h.unrealized, currency)}
                </td>
                <td class="py-2 text-right font-medium {pnlClass(h.roi)}">
                  {h.closed ? '-' : formatPercent(h.roi)}
                </td>
                <td class="py-2">
                  {#if h.closed}
                    <span
                      class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600"
                    >
                      Closed
                    </span>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {:else}
    <p class="mb-6 text-sm text-gray-500">No positions</p>
  {/if}

  <div class="mb-6 rounded-xl bg-white p-4 shadow">
    <h2 class="mb-4 font-semibold">Performance history</h2>
    {#if history && history.series.length > 0}
      <div class="mb-4">
        <select bind:value={selectedAsset} class="rounded-lg border px-3 py-2 text-sm">
          <option value="">Portfolio</option>
          {#each history.assets as a (a.asset_id)}
            <option value={a.asset_id}>{a.ticker} - {a.name}</option>
          {/each}
        </select>
      </div>
      <PositionChart
        series={selectedAsset
          ? history.assets.find((a) => a.asset_id === selectedAsset)?.series ?? []
          : history.series}
        splits={selectedAsset
          ? history.assets.find((a) => a.asset_id === selectedAsset)?.splits ?? []
          : history.splits}
        {currency}
      />
    {:else}
      <p class="text-sm text-gray-400">No data</p>
    {/if}
  </div>

  <div class="mb-6 rounded-xl bg-white p-4 shadow">
    <h2 class="mb-4 font-semibold">Allocazione per classi</h2>
    {#if classAllocError}
      <p class="text-sm text-gray-400">Allocazione per classi non disponibile</p>
    {:else if classAlloc && classAllocRows.length > 0}
      <div class="flex flex-col gap-4 md:flex-row">
        <div class="w-full md:w-1/2 lg:w-1/3">
          <ExposurePie data={classAllocRows} title="Allocazione per classi" />
        </div>
        <div class="flex-1 overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="border-b text-gray-500">
                <th class="pb-2">Classe</th>
                <th class="pb-2 text-right">Valore</th>
                <th class="pb-2 text-right">Peso %</th>
              </tr>
            </thead>
            <tbody>
              {#each classAlloc.classes as c (c.class)}
                <tr class="border-b last:border-0">
                  <td class="py-2 font-medium">{ASSET_CLASS_LABELS[c.class] ?? c.class}</td>
                  <td class="py-2 text-right">
                    {formatCurrency(c.value, classAlloc.currency)}
                  </td>
                  <td class="py-2 text-right font-medium">
                    {formatPercent(c.weight)}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {:else}
      <p class="text-sm text-gray-400">Nessuna allocazione per classi</p>
    {/if}
  </div>

  <div class="mb-4 flex items-center gap-2">
    <button
      onclick={() => (showTx ? closeTx() : openNewTx())}
      class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
    >
      <Plus class="h-4 w-4" />
      Add Transaction
    </button>
    <button
      onclick={exportPortfolio}
      class="flex items-center gap-2 rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
    >
      <Download class="h-4 w-4" />
      Export
    </button>
  </div>

  {#if showTx}
    <div class="mb-6 rounded-xl border bg-white p-4">
      <h3 class="mb-3 font-semibold">{editingTx ? 'Edit Transaction' : 'New Transaction'}</h3>
      <div class="grid grid-cols-2 gap-3">
        <select
          bind:value={txForm.asset_id}
          class="rounded-lg border px-3 py-2 text-sm"
        >
          <option value="">Select asset</option>
          {#each assets ?? [] as a (a.id)}
            <option value={a.id}>{a.ticker} - {a.name}</option>
          {/each}
        </select>
        <select
          bind:value={txForm.type}
          class="rounded-lg border px-3 py-2 text-sm"
        >
          <option value="buy">Buy</option>
          <option value="sell">Sell</option>
          <option value="dividend">Dividend</option>
        </select>
        {#if txForm.type === 'dividend'}
          <input
            type="number"
            step="0.01"
            min="0"
            placeholder="Amount"
            bind:value={txDividendAmount}
            class="rounded-lg border px-3 py-2 text-sm"
          />
        {:else}
          <input
            type="number"
            placeholder="Quantity"
            bind:value={txForm.quantity}
            class="rounded-lg border px-3 py-2 text-sm"
          />
          <input
            type="number"
            step="0.01"
            placeholder="Price"
            bind:value={txForm.price}
            class="rounded-lg border px-3 py-2 text-sm"
          />
        {/if}
        <input
          type="date"
          bind:value={txForm.date}
          class="rounded-lg border px-3 py-2 text-sm"
        />
        <input
          type="number"
          step="0.01"
          placeholder="Fees"
          bind:value={txForm.fees}
          class="rounded-lg border px-3 py-2 text-sm"
        />
      </div>
      <div class="mt-3 flex items-center gap-3">
        <button
          onclick={saveTransaction}
          disabled={!canSave() || txSaving}
          class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {txSaving ? 'Saving...' : editingTx ? 'Save Changes' : 'Save'}
        </button>
        {#if editingTx}
          <button
            onclick={handleDeleteEditing}
            disabled={deleting}
            class="flex items-center gap-1.5 rounded-lg bg-red-600 px-4 py-2 text-sm text-white hover:bg-red-700 disabled:opacity-50"
          >
            <Trash2 class="h-4 w-4" />
            Delete
          </button>
        {/if}
        <button
          onclick={closeTx}
          class="rounded-lg px-4 py-2 text-sm text-gray-600 hover:bg-gray-100"
        >
          Cancel
        </button>
      </div>
    </div>
  {/if}

  <div class="rounded-xl bg-white p-4 shadow">
    <h2 class="mb-4 font-semibold">Transactions</h2>
    <table class="w-full text-left text-sm">
      <thead>
        <tr class="border-b text-gray-500">
          <th class="pb-2">Date</th>
          <th class="pb-2">Asset</th>
          <th class="pb-2">Type</th>
          <th class="pb-2">Quantity</th>
          <th class="pb-2">Price</th>
          <th class="pb-2">Total</th>
          <th class="pb-2 text-right">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each transactions ?? [] as tx (tx.id)}
          <tr class="border-b last:border-0">
            <td class="py-2">{new Date(tx.date).toLocaleDateString()}</td>
            <td class="py-2">
              <span class="font-medium">{tx.asset_ticker}</span>
              <span class="ml-1 text-xs text-gray-500">{tx.asset_name}</span>
            </td>
            <td class="py-2">
              <span
                class="rounded-full px-2 py-0.5 text-xs font-medium {tx.type === 'buy'
                  ? 'bg-green-100 text-green-700'
                  : tx.type === 'sell'
                    ? 'bg-red-100 text-red-700'
                    : 'bg-blue-100 text-blue-700'}"
              >
                {tx.type}
              </span>
            </td>
            <td class="py-2">{tx.quantity}</td>
            <td class="py-2">{formatCurrency(tx.price, currency)}</td>
            <td class="py-2">
              {formatCurrency(Number(tx.quantity) * Number(tx.price), currency)}
            </td>
            <td class="py-2 text-right">
              <button
                onclick={() => startEdit(tx)}
                class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                title="Edit transaction"
              >
                <Pencil class="h-4 w-4" />
              </button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
