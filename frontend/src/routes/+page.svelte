<script module lang="ts">
  let sessionRefreshed = false
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { resolve } from '$app/paths'
  import {
    portfolioApi,
    pricesApi,
    type Dashboard,
    type DashboardAllocation,
    type DashboardPerformance,
    type PortfolioPerformanceSummary,
  } from '$lib/services/api'
  import { toast } from '$lib/stores/toast.svelte'
  import AllocationDonut from '$lib/components/domain/AllocationDonut.svelte'
  import CapitalChart from '$lib/components/domain/CapitalChart.svelte'
  import InvestmentsTable from '$lib/components/domain/InvestmentsTable.svelte'
  import PerformanceChart from '$lib/components/domain/PerformanceChart.svelte'
  import ClassDonut from '$lib/components/domain/ClassDonut.svelte'
  import ExposureBarChart, { type ExposureBarRow } from '$lib/components/domain/ExposureBarChart.svelte'
  import { countryDisplayName } from '$lib/countryNames'
  import { chartSemanticColors } from '$lib/chartPalette'
  import { resolved } from '$lib/stores/theme.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import EmptyState from '$lib/components/ui/EmptyState.svelte'
  import Spinner from '$lib/components/ui/Spinner.svelte'
  import { formatCurrency, formatPercent } from '$lib/format'
  import { pnlColorClass } from '$lib/ui-colors'

  let dash = $state<Dashboard | null>(null)
  let alloc = $state<DashboardAllocation | null>(null)
  let loading = $state(true)
  let lastUpdate = $state('')

  // Performance + Capital invested cards (EPIC I.3): vault-wide percentage
  // return and invested-capital buckets in the base currency, monthly by
  // default, switchable to annual. One fetch drives both charts. Isolated
  // like the allocation card: a failed fetch just renders the charts' empty
  // states.
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
  // issued request may write the state.
  let perfReq = 0
  async function loadPerformance(g: 'month' | 'year'): Promise<void> {
    const req = ++perfReq
    perfLoading = true
    try {
      const res = await portfolioApi.dashboardPerformance(g)
      if (req === perfReq) perf = res
    } catch {
      if (req === perfReq) perf = null
    } finally {
      if (req === perfReq) perfLoading = false
    }
  }

  // Runs once on mount with the default granularity and refetches whenever
  // the Monthly/Annual toggle changes.
  $effect(() => {
    void loadPerformance(granularity)
  })

  onMount(async () => {
    try {
      dash = await portfolioApi.dashboard()
    } catch {
      dash = null
    } finally {
      loading = false
    }

    // L'allocazione complessiva (classi, regioni, settori, paesi) è isolata:
    // se l'endpoint non è disponibile la card viene omessa senza bloccare il
    // resto della dashboard.
    try {
      alloc = await portfolioApi.dashboardAllocation()
    } catch {
      alloc = null
    }

    if (!sessionRefreshed) {
      sessionRefreshed = true
      pricesApi.refresh()
        .then((report) => {
          lastUpdate = report.finished_at
          if (report.rate_limited) {
            toast.warning('Yahoo Finance ha limitato le richieste: alcuni prezzi non aggiornati')
          } else if (report.issues.length > 0) {
            toast.warning(`${report.issues.length} aggiornamenti prezzi non riusciti (Yahoo)`)
          }
          return portfolioApi.dashboard()
        })
        .then((fresh) => {
          dash = fresh
          // The POST above cleared the GET cache and new prices can move the
          // performance buckets: refresh both the Performance and Capital
          // invested cards (they share this one fetch).
          void loadPerformance(granularity)
        })
        .catch(() => { /* keep current data, omit the "Prices updated" line */ })
    }
  })

  const hasMultipleCurrencies = $derived((dash?.by_currency?.length ?? 0) > 1)
  const pricesUpdatedLabel = $derived(lastUpdate ? new Date(lastUpdate).toLocaleString() : '')
  const portfolioSlices = $derived(
    (dash?.portfolios ?? []).map((p) => ({ name: p.portfolio_name, value: Number(p.active.value) })),
  )

  // EPIC I.5 consolidated "Invested assets" table: open positions aggregated
  // across portfolios in the base currency, already sorted by descending
  // value server-side (kept as-is, no client re-sort). `?? []` also covers
  // older backends that still omit the field.
  const investedAssets = $derived(dash?.invested_assets ?? [])

  // EPIC I.4 "Allocazione complessiva" card: region, sector and country
  // payloads are mapped onto the generic ExposureBarRow shape consumed by
  // ExposureBarChart (countries keep their raw ISO codes as row identity; the
  // chart maps them to full names via `labelFor`, see the country panel below).
  const regionBarRows = $derived<ExposureBarRow[]>(
    (alloc?.regions ?? []).map((r) => ({ name: r.region, value: r.value, weight: r.weight })),
  )
  const sectorBarRows = $derived<ExposureBarRow[]>(
    (alloc?.sectors ?? []).map((s) => ({ name: s.sector, value: s.value, weight: s.weight })),
  )
  const countryBarRows = $derived<ExposureBarRow[]>(
    (alloc?.countries ?? []).map((c) => ({ name: c.country, value: c.value, weight: c.weight })),
  )

  // Equity-universe coverage note for the bar charts, mirroring the long
  // covered/excluded note the donut charts render themselves; shown only when
  // non-equity holdings were actually excluded.
  const equityUniverseNote = $derived.by(() => {
    const covered = Number(alloc?.covered_value || 0)
    const excluded = Number(alloc?.excluded_value || 0)
    const total = covered + excluded
    if (total <= 0 || excluded <= 0) return undefined
    return `Universo azionario: ${((covered / total) * 100).toFixed(1)}% del portafoglio`
  })

  // Aggregated "Other" buckets are muted grey, like the slice treatment in the
  // donut charts. Read via chartSemanticColors so it re-evaluates on theme
  // flips (the {#key} blocks inside the charts re-init them anyway).
  function otherGrey(name: string): string | undefined {
    return name === 'Other' || name === 'Other / Not Classified'
      ? chartSemanticColors(resolved()).other
      : undefined
  }

  // The compact per-card closed line is dropped when the portfolio never sold
  // a lot. Closed dividends are now folded into `proceeds` by the backend, so
  // they no longer drive the visibility on their own.
  function hasClosedActivity(p: PortfolioPerformanceSummary): boolean {
    return Number(p.closed.invested) !== 0
  }
</script>

<div class="p-6">
  <div class="mb-6">
    <h1 class="text-2xl font-bold">Dashboard</h1>
    {#if pricesUpdatedLabel}
      <p class="mt-1 text-sm text-muted-foreground">Prices updated: {pricesUpdatedLabel}</p>
    {/if}
  </div>

  {#if loading}
    <div class="flex justify-center py-24 text-muted-foreground">
      <Spinner size="lg" />
    </div>
  {:else if !dash?.portfolios?.length}
    <EmptyState
      dashed
      title="No portfolios yet"
      description="Create a portfolio to start tracking your investments."
    >
      {#snippet action()}
        <Button href={resolve('/portfolios')}>Create your first portfolio</Button>
      {/snippet}
    </EmptyState>
  {:else}
    <div class="space-y-6">
      {#if dash.summary}
        <!-- Primary KPIs: consolidated totals in the user's base currency
             (EPIC I.1), rendered by the shared "Investments" card (EPIC I.2
             active/closed split) also used by the portfolio detail (EPIC I.6,
             #85): Active = lots still held (dividends of open positions kept
             in their own column), Closed = sold lots (proceeds already fold
             in the dividends of fully-closed positions, so the Dividends
             cell is empty). -->
        <InvestmentsTable
          active={dash.summary.active}
          closed={dash.summary.closed}
          currency={dash.base_currency}
        />
      {/if}

      <div class="grid gap-4 lg:grid-cols-2">
        <!-- Percentage return card: bars = per-bucket time-weighted return,
             line = cumulative TWR. The Monthly/Annual control lives here and
             drives `granularity`, which also feeds the Capital invested card
             below (both share the same `dashboardPerformance` fetch). -->
        <Card class="p-4">
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

        <!-- Capital invested card: invested (stepped) vs market value lines in
             the base currency, over the SAME buckets/granularity. -->
        <Card class="p-4">
          <h2 class="mb-4 font-semibold">Capital invested</h2>
          {#if perfLoading}
            <div class="flex h-[340px] items-center justify-center text-muted-foreground">
              <Spinner />
            </div>
          {:else}
            <CapitalChart
              buckets={perf?.buckets ?? []}
              currency={perf?.currency || dash.base_currency || 'USD'}
              granularity={granularity}
            />
          {/if}
        </Card>
      </div>

      <Card class="p-4">
        <h2 class="mb-4 font-semibold">Allocation by portfolio</h2>
        <AllocationDonut
          data={portfolioSlices}
          title="Allocation by portfolio"
          currency={dash.base_currency || dash.by_currency[0]?.currency || 'USD'}
          showValue={!hasMultipleCurrencies}
        />
        {#if hasMultipleCurrencies}
          <p class="mt-2 text-xs text-muted-foreground">
            Portfolios use different currencies: values are not comparable, shares are indicative.
          </p>
        {/if}
      </Card>

      <div>
        <h2 class="mb-4 font-semibold">Portfolios</h2>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {#each dash.portfolios as p (p.portfolio_id)}
            <Card class="transition-colors hover:border-accent">
              <a href={resolve(`/portfolios/${p.portfolio_id}`)} class="block p-4">
                <div class="flex items-baseline justify-between gap-2">
                  <span class="truncate font-semibold">{p.portfolio_name}</span>
                  <span class="shrink-0 text-xs text-muted-foreground">{p.currency}</span>
                </div>
                <p class="mt-2 text-lg font-bold tabular-nums">{formatCurrency(p.active.value, p.currency)}</p>
                <div class="mt-1 flex items-center justify-between text-sm">
                  <span class="font-medium tabular-nums {pnlColorClass(p.active.gain_loss)}">
                    {formatCurrency(p.active.gain_loss, p.currency)}
                  </span>
                  <span class="font-medium tabular-nums {pnlColorClass(p.active.gain_loss_pct)}">
                    {formatPercent(p.active.gain_loss_pct)}
                  </span>
                </div>
                {#if hasClosedActivity(p)}
                  <p class="mt-2 text-xs tabular-nums text-muted-foreground">
                    Closed: {formatCurrency(p.closed.invested, p.currency)} ·
                    {formatCurrency(p.closed.proceeds, p.currency)} ·
                    <span class="font-medium {pnlColorClass(p.closed.realized)}">
                      {formatCurrency(p.closed.realized, p.currency)}
                    </span>
                  </p>
                {/if}
                <p class="mt-2 text-xs text-muted-foreground">{p.asset_count} assets</p>
              </a>
            </Card>
          {/each}
        </div>
      </div>

      <div class="rounded-card border-border bg-surface p-4 shadow-card">
        <h2 class="mb-4 font-semibold">Allocazione complessiva</h2>
        {#if alloc == null}
          <p class="text-sm text-muted-foreground">Allocazione non disponibile</p>
        {:else}
          <!-- EPIC I.4 layout: asset-class donut over the whole vault plus the
               equity-only breakdown (sector, region and country bars), in a
               responsive 2-column grid; regions sit next to countries in the
               bottom row. The inner panels reuse the shared card surface
               styling; the portfolio detail "Allocazione" section mirrors this
               exact layout in EPIC I.7 (#86). -->
          <div class="grid gap-4 lg:grid-cols-2">
            <div class="rounded-card border-border bg-surface p-4 shadow-card">
              <ClassDonut data={alloc.classes ?? []} currency={alloc.currency} label="Classi di attività" />
            </div>
            <div class="rounded-card border-border bg-surface p-4 shadow-card">
              <ExposureBarChart
                rows={sectorBarRows}
                currency={alloc.currency}
                label="Settori (solo equity)"
                note={equityUniverseNote}
                colorFor={otherGrey}
              />
            </div>
            <div class="rounded-card border-border bg-surface p-4 shadow-card">
              <ExposureBarChart
                rows={regionBarRows}
                currency={alloc.currency}
                label="Regioni (solo equity)"
                note={equityUniverseNote}
                colorFor={otherGrey}
              />
            </div>
            <div class="rounded-card border-border bg-surface p-4 shadow-card">
              <ExposureBarChart
                rows={countryBarRows}
                currency={alloc.currency}
                label="Paesi (solo equity)"
                note={equityUniverseNote}
                colorFor={otherGrey}
                labelFor={countryDisplayName}
                maxVisibleRows={10}
              />
            </div>
          </div>
        {/if}
      </div>

      <!-- EPIC I.5: consolidated invested-assets table replacing the old
           per-portfolio accordions. One row per open asset merged across all
           portfolios, in the base currency, ordered by value descending as
           returned by the backend. -->
      <Card class="p-4">
        <h2 class="mb-3 font-semibold">Invested assets</h2>
        {#if investedAssets.length === 0}
          <EmptyState
            dashed
            title="No invested assets yet"
            description="Open positions will appear here once you record transactions in your portfolios."
          />
        {:else}
          <div class="overflow-x-auto">
            <Table aria-label="Invested assets">
              <THead>
                <Tr>
                  <Th>Asset</Th>
                  <Th align="right">Invested</Th>
                  <Th align="right">Value</Th>
                  <Th align="right">Gain/Loss</Th>
                  <Th align="right">P/L %</Th>
                </Tr>
              </THead>
              <TBody>
                {#each investedAssets as a (a.asset_id)}
                  <Tr>
                    <Td>
                      <a
                        href={resolve(`/assets/${a.asset_id}`)}
                        class="font-medium text-accent-text hover:underline"
                      >
                        {a.ticker}
                      </a>
                      {#if !a.has_price}
                        <Badge
                          variant="neutral"
                          class="ml-1.5 align-middle"
                          title="No price data: value is carried at cost, so its P/L is 0"
                        >
                          no price
                        </Badge>
                      {/if}
                      <span class="block text-xs text-muted-foreground">{a.name}</span>
                    </Td>
                    <Td align="right">{formatCurrency(a.invested, dash.base_currency)}</Td>
                    <Td align="right">{formatCurrency(a.value, dash.base_currency)}</Td>
                    <Td align="right" class="font-medium {pnlColorClass(a.gain_loss)}">
                      {formatCurrency(a.gain_loss, dash.base_currency)}
                    </Td>
                    <Td align="right" class={pnlColorClass(a.gain_loss_pct)}>
                      {formatPercent(a.gain_loss_pct)}
                    </Td>
                  </Tr>
                {/each}
              </TBody>
            </Table>
          </div>
        {/if}
      </Card>
    </div>
  {/if}
</div>
