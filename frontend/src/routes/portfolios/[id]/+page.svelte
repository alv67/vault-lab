<script module lang="ts">
  let sessionRefreshed = false
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
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
  import { pnlColorClass } from '$lib/ui-colors'
  import PositionChart from '$lib/components/PositionChart.svelte'
  import ExposurePie from '$lib/components/ExposurePie.svelte'
  import GeographyChart from '$lib/components/domain/GeographyChart.svelte'
  import SectorChart from '$lib/components/domain/SectorChart.svelte'
  import PositionTable, { type PositionRow } from '$lib/components/domain/PositionTable.svelte'
  import TransactionTable from '$lib/components/domain/TransactionTable.svelte'
  import AddTransactionModal from '$lib/components/domain/AddTransactionModal.svelte'
  import {
    type ExposureRow,
    type PortfolioClassAllocation,
    type PortfolioGeographyAllocation,
    type PortfolioSectorAllocation,
  } from '$lib/services/api'
  import { Plus, Download } from 'lucide-svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import StatCard from '$lib/components/ui/StatCard.svelte'

  const id = $derived(page.params.id as string | undefined)

  let portfolio = $state<Portfolio | null>(null)
  let summary = $state<PortfolioSummary | null>(null)
  let transactions = $state<Transaction[] | null>(null)
  let assets = $state<Asset[] | null>(null)
  let history = $state<PortfolioHistory | null>(null)
  let classAlloc = $state<PortfolioClassAllocation | null>(null)
  let classAllocError = $state(false)
  let geoAlloc = $state<PortfolioGeographyAllocation | null>(null)
  let sectorAlloc = $state<PortfolioSectorAllocation | null>(null)
  let geoAllocError = $state(false)
  let sectorAllocError = $state(false)
  let selectedAsset = $state('')
  let showTx = $state(false)
  let editingTx = $state<Transaction | null>(null)

  const currency = $derived(portfolio?.currency || 'USD')
  const classAllocRows = $derived<ExposureRow[]>(
    (classAlloc?.classes ?? []).map((c) => ({
      name: ASSET_CLASS_LABELS[c.class] ?? c.class,
      weight: c.weight,
    })),
  )
  const positionRows = $derived<PositionRow[]>(
    (summary?.holdings ?? []).map((h) => ({
      assetId: h.asset_id,
      ticker: h.ticker,
      name: h.name,
      qty: Number(h.qty),
      cost: Number(h.cost),
      value: Number(h.value_pf),
      realized: Number(h.realized),
      unrealized: Number(h.unrealized),
      roi: Number(h.roi),
      closed: h.closed,
      price: Number(h.last_close) > 0 ? Number(h.last_close) : undefined,
      priceCurrency: h.currency,
    })),
  )

  $effect(() => {
    if (!showTx) editingTx = null
  })

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

    await loadAllocations()

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

  async function loadAllocations(): Promise<void> {
    if (!id) return
    classAllocError = false
    geoAllocError = false
    sectorAllocError = false

    // L'allocazione per classi è isolata: se il backend non la espone ancora
    // (es. asset_class non popolati) non blocca il resto della pagina.
    try {
      classAlloc = await portfolioApi.classAllocation(id)
    } catch {
      classAllocError = true
    }

    // Anche geografia e settore sono isolati: un errore qui (endpoint non
    // disponibile o dati mancanti) non deve bloccare il resto della pagina.
    try {
      geoAlloc = await portfolioApi.geographyAllocation(id)
    } catch {
      geoAllocError = true
    }

    try {
      sectorAlloc = await portfolioApi.sectorAllocation(id)
    } catch {
      sectorAllocError = true
    }
  }

  async function reloadAfterMutation(): Promise<void> {
    if (!id) return
    try {
      const [t, s, h] = await Promise.all([
        transactionApi.list(id),
        portfolioApi.summary(id),
        portfolioApi.history(id),
      ])
      transactions = t
      summary = s
      history = h
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to refresh portfolio'
      toast.error(message)
    }
    await loadAllocations()
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
  <header class="sticky top-14 z-10 -mx-6 -mt-6 mb-6 flex flex-col gap-3 bg-background px-6 pb-3 pt-6 sm:flex-row sm:items-center sm:justify-between">
    <div class="min-w-0">
      <h1 class="text-2xl font-bold">{portfolio?.name ?? 'Portfolio'}</h1>
      {#if portfolio?.description}
        <p class="text-sm text-muted-foreground">{portfolio.description}</p>
      {/if}
    </div>
    <div class="flex shrink-0 items-center gap-2">
      <Button
        onclick={() => {
          editingTx = null
          showTx = true
        }}
      >
        <Plus class="h-4 w-4" />
        Add Transaction
      </Button>
      <Button variant="outline" onclick={exportPortfolio}>
        <Download class="h-4 w-4" />
        Export
      </Button>
    </div>
  </header>

  <div class="mb-6 grid grid-cols-2 gap-4 md:grid-cols-4">
    <StatCard label="Value" value={formatCurrency(summary?.total_value ?? 0, currency)} />
    <StatCard
      label="Realized"
      value={formatCurrency(summary?.realized_gl ?? 0, currency)}
      valueClass={pnlColorClass(summary?.realized_gl)}
    />
    <StatCard
      label="Open G/L"
      value={formatCurrency(summary?.gain_loss ?? 0, currency)}
      delta={formatPercent(summary?.gain_loss_pct ?? 0)}
      deltaValue={Number(summary?.gain_loss_pct ?? 0)}
      valueClass={pnlColorClass(summary?.gain_loss)}
    />
    <StatCard label="Assets" value={String(summary?.asset_count ?? 0)} />
  </div>

  {#if positionRows.length > 0}
    <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
      <h2 class="mb-4 font-semibold">Positions</h2>
      <PositionTable
        rows={positionRows}
        {currency}
        linkAssets
        showCost
        showRealized
        showUnrealized
        showPrice
      />
    </div>
  {:else}
    <p class="mb-6 text-sm text-muted-foreground">No positions</p>
  {/if}

  <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
    <h2 class="mb-4 font-semibold">Performance history</h2>
    {#if history && history.series.length > 0}
      <div class="mb-4">
        <select bind:value={selectedAsset} class="rounded-control border border-input px-3 py-2 text-sm">
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
      <p class="text-sm text-muted-foreground">No data</p>
    {/if}
  </div>

  <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
    <h2 class="mb-4 font-semibold">Allocazione per classi</h2>
    {#if classAllocError}
      <p class="text-sm text-muted-foreground">Allocazione per classi non disponibile</p>
    {:else if classAlloc && classAllocRows.length > 0}
      <div class="flex flex-col gap-4 md:flex-row">
        <div class="w-full md:w-1/2 lg:w-1/3">
          <ExposurePie data={classAllocRows} title="Allocazione per classi" />
        </div>
        <div class="flex-1 overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="border-b border-border text-muted-foreground">
                <th class="pb-2">Classe</th>
                <th class="pb-2 text-right">Valore</th>
                <th class="pb-2 text-right">Peso %</th>
              </tr>
            </thead>
            <tbody>
              {#each classAlloc.classes as c (c.class)}
                <tr class="border-b border-border last:border-0">
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
      <p class="text-sm text-muted-foreground">Nessuna allocazione per classi</p>
    {/if}
  </div>

  <div class="mb-6 flex flex-col gap-4 md:flex-row">
    <div class="w-full md:w-1/2">
      {#if geoAllocError}
        <div class="rounded-card border-border bg-surface p-4 shadow-card">
          <h2 class="mb-4 font-semibold">Allocazione geografica</h2>
          <p class="text-sm text-muted-foreground">Allocazione geografica non disponibile</p>
        </div>
      {:else}
        <GeographyChart data={geoAlloc?.regions ?? []} {currency} covered={geoAlloc?.covered_value} excluded={geoAlloc?.excluded_value} />
      {/if}
    </div>
    <div class="w-full md:w-1/2">
      {#if sectorAllocError}
        <div class="rounded-card border-border bg-surface p-4 shadow-card">
          <h2 class="mb-4 font-semibold">Allocazione settoriale</h2>
          <p class="text-sm text-muted-foreground">Allocazione settoriale non disponibile</p>
        </div>
      {:else}
        <SectorChart data={sectorAlloc?.sectors ?? []} {currency} covered={sectorAlloc?.covered_value} excluded={sectorAlloc?.excluded_value} />
      {/if}
    </div>
  </div>

  <div class="rounded-card border-border bg-surface p-4 shadow-card">
    <h2 class="mb-4 font-semibold">Transactions</h2>
    <TransactionTable
      transactions={transactions ?? []}
      {currency}
      onedit={(tx) => {
        editingTx = tx
        showTx = true
      }}
    />
  </div>
</div>

<AddTransactionModal
  bind:open={showTx}
  portfolioId={id ?? ''}
  assets={assets ?? []}
  {currency}
  editing={editingTx}
  onsuccess={reloadAfterMutation}
/>
