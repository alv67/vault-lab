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
    type DashboardPerformance,
    type Transaction,
    type Asset,
  } from '$lib/services/api'
  import PositionChart from '$lib/components/PositionChart.svelte'
  import ClassDonut from '$lib/components/domain/ClassDonut.svelte'
  import ExposureBarChart, { type ExposureBarRow } from '$lib/components/domain/ExposureBarChart.svelte'
  import InvestmentsTable from '$lib/components/domain/InvestmentsTable.svelte'
  import PerformanceChart from '$lib/components/domain/PerformanceChart.svelte'
  import PositionTable, { type PositionRow } from '$lib/components/domain/PositionTable.svelte'
  import TransactionTable from '$lib/components/domain/TransactionTable.svelte'
  import AddTransactionModal from '$lib/components/domain/AddTransactionModal.svelte'
  import {
    type PortfolioClassAllocation,
    type PortfolioGeographyAllocation,
    type PortfolioSectorAllocation,
  } from '$lib/services/api'
  import { countryDisplayName } from '$lib/countryNames'
  import { chartSemanticColors } from '$lib/chartPalette'
  import { resolved } from '$lib/stores/theme.svelte'
  import { Plus, Download } from 'lucide-svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte'
  import Spinner from '$lib/components/ui/Spinner.svelte'

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

  // EPIC I.9 (#88): the Transactions table is paginated through the
  // `TransactionPage` envelope of `GET /portfolios/{id}/transactions`
  // (order date desc). `txPage` is the 1-based page index, `txOffset` the
  // row window start; the page size mirrors the backend default (20).
  const TX_PAGE_SIZE = 20
  let txPage = $state(1)
  let txLimit = $state(TX_PAGE_SIZE)
  let txOffset = $state(0)
  let txTotal = $state(0)
  let txLoading = $state(false)

  // Same "1–20 of 137" range label and footer layout as the health page.
  const txRangeLabel = $derived(
    (transactions?.length ?? 0) === 0
      ? `0 of ${txTotal}`
      : `${txOffset + 1}–${txOffset + (transactions?.length ?? 0)} of ${txTotal}`,
  )

  // Monotonic request id (same guard as the performance card): rapid page
  // flips must never let a stale response overwrite the current window.
  let txReq = 0

  /** Fetch the current transaction page into `transactions`/`txTotal`,
   * touching nothing else on the page. If the window comes back empty while
   * rows still exist (the last row of the last page was just deleted), step
   * back to the previous page — clamped against the fresh total — and
   * refetch it within the same call. */
  async function loadTransactions(): Promise<void> {
    if (!id) return
    const req = ++txReq
    txLoading = true
    try {
      let res = await transactionApi.list(id, { limit: txLimit, offset: txOffset })
      if (req === txReq && res.transactions.length === 0 && res.total > 0 && txOffset > 0) {
        const maxPage = Math.max(1, Math.ceil(res.total / txLimit))
        txPage = Math.min(Math.max(1, txPage - 1), maxPage)
        txOffset = (txPage - 1) * txLimit
        res = await transactionApi.list(id, { limit: txLimit, offset: txOffset })
      }
      if (req === txReq) {
        transactions = res.transactions
        txTotal = res.total
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load transactions'
      if (req === txReq) toast.error(message)
    } finally {
      if (req === txReq) txLoading = false
    }
  }

  /** Jump to a 1-based page (clamped to the last page of `txTotal`) and
   * refetch only the transactions window — never the whole portfolio page. */
  function gotoTxPage(target: number): void {
    const maxPage = Math.max(1, Math.ceil(txTotal / txLimit))
    txPage = Math.min(Math.max(1, target), maxPage)
    txOffset = (txPage - 1) * txLimit
    void loadTransactions()
  }

  const currency = $derived(portfolio?.currency || 'USD')

  // EPIC I.8 (#87): the portfolio's own percentage performance card — same
  // bucket model as the dashboard "Performance" card (EPIC I.3) but in the
  // PORTFOLIO currency, fed by `performanceBuckets(id, granularity)`
  // (`GET /portfolios/{id}/performance/buckets`). Isolated like the
  // allocations: a failed fetch just shows the chart's "No data" empty
  // state, and a `Spinner` covers every load.
  let perf = $state<DashboardPerformance | null>(null)
  let perfLoading = $state(true)
  let granularity = $state<'month' | 'year'>('month')
  const perfItems = [
    { value: 'month', label: 'Monthly' },
    { value: 'year', label: 'Annual' },
  ]

  // SegmentedControl binds a plain string; the accessors keep the union type.
  function getGranularity(): string {
    return granularity
  }
  function setGranularity(value: string): void {
    if (value === 'month' || value === 'year') granularity = value
  }

  // Monotonic request id: when the toggle is flipped quickly, only the last
  // issued request may write the state (same guard as the dashboard).
  let perfReq = 0
  async function loadPerformance(g: 'month' | 'year'): Promise<void> {
    if (!id) return
    const req = ++perfReq
    perfLoading = true
    try {
      const res = await portfolioApi.performanceBuckets(id, g)
      if (req === perfReq) perf = res
    } catch {
      if (req === perfReq) perf = null
    } finally {
      if (req === perfReq) perfLoading = false
    }
  }

  // Runs once on mount with the default granularity, refetches on every
  // Monthly/Annual toggle change (and on portfolio navigation, since `id`
  // is read reactively inside `loadPerformance`).
  $effect(() => {
    void loadPerformance(granularity)
  })

  // EPIC I.7 (#86): the allocation section mirrors the dashboard's
  // "Allocazione complessiva" card. The region/sector/country payloads are
  // mapped onto the generic ExposureBarRow shape consumed by
  // ExposureBarChart (countries keep their raw ISO codes as row identity;
  // the chart maps them to full names via `labelFor`, see the panel below).
  const regionBarRows = $derived<ExposureBarRow[]>(
    (geoAlloc?.regions ?? []).map((r) => ({ name: r.region, value: r.value, weight: r.weight })),
  )
  const sectorBarRows = $derived<ExposureBarRow[]>(
    (sectorAlloc?.sectors ?? []).map((s) => ({ name: s.sector, value: s.value, weight: s.weight })),
  )
  const countryBarRows = $derived<ExposureBarRow[]>(
    (geoAlloc?.countries ?? []).map((c) => ({ name: c.country, value: c.value, weight: c.weight })),
  )

  // Equity-universe coverage note for the bar panels, like the dashboard:
  // shown only when non-equity holdings were actually excluded. Geography
  // (regions + countries) and sectors come from two isolated endpoints, so
  // each payload carries its own covered/excluded note.
  function equityUniverseNote(covered?: string, excluded?: string): string | undefined {
    const coveredNum = Number(covered || 0)
    const excludedNum = Number(excluded || 0)
    const total = coveredNum + excludedNum
    if (total <= 0 || excludedNum <= 0) return undefined
    return `Universo azionario: ${((coveredNum / total) * 100).toFixed(1)}% del portafoglio`
  }
  const geoUniverseNote = $derived(
    equityUniverseNote(geoAlloc?.covered_value, geoAlloc?.excluded_value),
  )
  const sectorUniverseNote = $derived(
    equityUniverseNote(sectorAlloc?.covered_value, sectorAlloc?.excluded_value),
  )

  // Aggregated "Other" buckets are muted grey, like the slice treatment in
  // the donut charts (same helper as the dashboard card). Read via
  // chartSemanticColors so it re-evaluates on theme flips (the {#key} blocks
  // inside the charts re-init them anyway).
  function otherGrey(name: string): string | undefined {
    return name === 'Other' || name === 'Other / Not Classified'
      ? chartSemanticColors(resolved()).other
      : undefined
  }
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
      const [p, s, txs, a] = await Promise.all([
        portfolioApi.get(id),
        portfolioApi.summary(id),
        transactionApi.list(id, { limit: txLimit, offset: txOffset }),
        assetApi.list(),
      ])
      portfolio = p
      summary = s
      transactions = txs.transactions
      txTotal = txs.total
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
        .then((fresh) => {
          summary = fresh
          // The POST above cleared the GET cache and new prices can move the
          // buckets: refresh the performance card too (same as the dashboard).
          void loadPerformance(granularity)
        })
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
      // Refetch the CURRENT transactions page (plus total, with the
      // empty-page step-back) in parallel with the summary/history.
      // `loadTransactions` never rejects (it toasts its own errors), so a
      // transactions failure cannot block the other refreshes. The
      // mutation already cleared the GET cache.
      const [s, h] = await Promise.all([
        portfolioApi.summary(id),
        portfolioApi.history(id),
        loadTransactions(),
      ])
      summary = s
      history = h
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to refresh portfolio'
      toast.error(message)
    }
    // New/edited transactions change the flows behind the TWR buckets too
    // (the mutation already cleared the GET cache) — refresh the card.
    void loadPerformance(granularity)
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

  {#if summary}
    <!-- Same "Investments" card the dashboard shows (EPIC I.6, #85), fed by
         the portfolio-currency active/closed roll-ups of
         `GET /portfolios/{id}/summary`; the old flat KPI row (Value /
         Realized / Open G/L / Assets) is gone, only the asset count stays as
         a muted secondary line under the card. -->
    <div class="mb-6">
      <InvestmentsTable active={summary.active} closed={summary.closed} {currency} />
      <p class="mt-2 text-xs text-muted-foreground">{summary.asset_count} assets</p>
    </div>
  {/if}

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

  <!-- EPIC I.8 (#87): the portfolio's own percentage performance, same shared
       `PerformanceChart` as the dashboard card (bars = per-bucket `return` %,
       line = cumulative `twr`) with its own Monthly/Annual toggle; amounts in
       the Performance bucket payloads are already in the portfolio currency,
       and both series are pure percentages so no `currency` prop is needed. -->
  <Card class="mb-6 p-4">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
      <h2 class="font-semibold">Performance</h2>
      <SegmentedControl
        items={perfItems}
        bind:value={getGranularity, setGranularity}
        ariaLabel="Performance granularity"
      />
    </div>
    {#if perfLoading}
      <div class="flex h-[340px] items-center justify-center text-muted-foreground">
        <Spinner />
      </div>
    {:else}
      <PerformanceChart buckets={perf?.buckets ?? []} granularity={granularity} />
    {/if}
  </Card>

  <!-- Secondary view (kept below the percentage chart, EPIC I.8 #87): the
       raw invested/value/realized capital lines with the per-asset selector. -->
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
    <h2 class="mb-4 font-semibold">Allocazione</h2>
    <!-- EPIC I.7 (#86): mirrors the dashboard "Allocazione complessiva" card —
         the asset-class donut plus the equity-only region, sector and country
         bars, all in the portfolio currency. Each panel keeps the endpoint's
         isolated error state: a failed allocation call shows its own
         "non disponibile" panel without blocking the section or the page. -->
    <div class="grid gap-4 lg:grid-cols-2">
      <div class="rounded-card border-border bg-surface p-4 shadow-card">
        {#if classAllocError}
          <h3 class="mb-3 font-semibold">Classi di attività</h3>
          <p class="text-sm text-muted-foreground">Allocazione per classi non disponibile</p>
        {:else}
          <ClassDonut
            data={classAlloc?.classes ?? []}
            currency={classAlloc?.currency || currency}
            label="Classi di attività"
          />
        {/if}
      </div>
      <div class="rounded-card border-border bg-surface p-4 shadow-card">
        {#if sectorAllocError}
          <h3 class="mb-1 font-semibold">Settori (solo equity)</h3>
          <p class="text-sm text-muted-foreground">Allocazione settoriale non disponibile</p>
        {:else}
          <ExposureBarChart
            rows={sectorBarRows}
            currency={sectorAlloc?.currency || currency}
            label="Settori (solo equity)"
            note={sectorUniverseNote}
            colorFor={otherGrey}
          />
        {/if}
      </div>
      <div class="rounded-card border-border bg-surface p-4 shadow-card">
        {#if geoAllocError}
          <h3 class="mb-1 font-semibold">Regioni (solo equity)</h3>
          <p class="text-sm text-muted-foreground">Allocazione geografica non disponibile</p>
        {:else}
          <ExposureBarChart
            rows={regionBarRows}
            currency={geoAlloc?.currency || currency}
            label="Regioni (solo equity)"
            note={geoUniverseNote}
            colorFor={otherGrey}
          />
        {/if}
      </div>
      <div class="rounded-card border-border bg-surface p-4 shadow-card">
        {#if geoAllocError}
          <h3 class="mb-1 font-semibold">Paesi (solo equity)</h3>
          <p class="text-sm text-muted-foreground">Allocazione geografica non disponibile</p>
        {:else}
          <ExposureBarChart
            rows={countryBarRows}
            currency={geoAlloc?.currency || currency}
            label="Paesi (solo equity)"
            note={geoUniverseNote}
            colorFor={otherGrey}
            labelFor={countryDisplayName}
            maxVisibleRows={10}
          />
        {/if}
      </div>
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
    <!-- EPIC I.9 (#88): pagination footer with the same "1–20 of 137" range
         label and Previous/Next layout used by the admin health page. -->
    <div class="mt-3 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-3">
      <span class="text-sm tabular-nums text-muted-foreground">{txRangeLabel}</span>
      <div class="flex items-center gap-2">
        <Button
          variant="secondary"
          size="sm"
          disabled={txOffset === 0 || txLoading}
          onclick={() => gotoTxPage(txPage - 1)}
        >
          Previous
        </Button>
        <Button
          variant="secondary"
          size="sm"
          disabled={txOffset + txLimit >= txTotal || txLoading}
          onclick={() => gotoTxPage(txPage + 1)}
        >
          Next
        </Button>
      </div>
    </div>
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
