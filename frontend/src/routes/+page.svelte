<script module lang="ts">
  let sessionRefreshed = false
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { SvelteSet } from 'svelte/reactivity'
  import { resolve } from '$app/paths'
  import {
    portfolioApi,
    pricesApi,
    type Dashboard,
    type DashboardAllocation,
    type PortfolioAssets,
    type PortfolioPerformanceSummary,
  } from '$lib/services/api'
  import { toast } from '$lib/stores/toast.svelte'
  import PortfolioLineChart from '$lib/components/PortfolioLineChart.svelte'
  import AllocationDonut from '$lib/components/domain/AllocationDonut.svelte'
  import PositionTable, { type PositionRow } from '$lib/components/domain/PositionTable.svelte'
  import GeographyChart from '$lib/components/domain/GeographyChart.svelte'
  import SectorChart from '$lib/components/domain/SectorChart.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import StatCard from '$lib/components/ui/StatCard.svelte'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'
  import EmptyState from '$lib/components/ui/EmptyState.svelte'
  import Spinner from '$lib/components/ui/Spinner.svelte'
  import { formatCurrency, formatPercent } from '$lib/format'
  import { pnlColorClass } from '$lib/ui-colors'
  import { ChevronDown, ChevronRight } from 'lucide-svelte'

  let dash = $state<Dashboard | null>(null)
  let alloc = $state<DashboardAllocation | null>(null)
  let loading = $state(true)
  let lastUpdate = $state('')
  let expanded = new SvelteSet<string>()
  let initialized = false

  onMount(async () => {
    try {
      dash = await portfolioApi.dashboard()
    } catch {
      dash = null
    } finally {
      loading = false
    }

    // L'allocazione complessiva geo/settore è isolata: se l'endpoint non è
    // disponibile la card viene omessa senza bloccare il resto della dashboard.
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
        .then((fresh) => { dash = fresh })
        .catch(() => { /* keep current data, omit the "Prices updated" line */ })
    }

    const firstPortfolioId = dash?.assets?.[0]?.portfolio_id
    if (firstPortfolioId && !initialized) {
      initialized = true
      expanded.add(firstPortfolioId)
    }
  })

  function toggle(id: string): void {
    if (expanded.has(id)) {
      expanded.delete(id)
    } else {
      expanded.add(id)
    }
  }

  const hasMultipleCurrencies = $derived((dash?.by_currency?.length ?? 0) > 1)
  const pricesUpdatedLabel = $derived(lastUpdate ? new Date(lastUpdate).toLocaleString() : '')
  const portfolioSlices = $derived(
    (dash?.portfolios ?? []).map((p) => ({ name: p.portfolio_name, value: Number(p.active.value) })),
  )

  // The compact per-card closed line is dropped when the portfolio never sold
  // a lot. Closed dividends are now folded into `proceeds` by the backend, so
  // they no longer drive the visibility on their own.
  function hasClosedActivity(p: PortfolioPerformanceSummary): boolean {
    return Number(p.closed.invested) !== 0
  }

  // value/realized come from the *_pf fields, consolidated in the portfolio
  // currency, so the table stays currency-consistent across FX assets.
  function positionRows(pa: PortfolioAssets): PositionRow[] {
    return pa.assets.map((a) => ({
      assetId: a.asset_id,
      ticker: a.ticker,
      name: a.name,
      qty: Number(a.qty),
      value: Number(a.value_pf),
      realized: Number(a.realized_pf ?? a.realized),
      roi: Number(a.roi),
    }))
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
      <div class="space-y-4">
        {#if dash.summary}
          <!-- Primary KPIs: consolidated totals in the user's base currency
               (EPIC I.1), collected into a single "Investments" card with one
               row per breakdown group (EPIC I.2): Active = lots still held
               (dividends of open positions kept in their own column), Closed =
               sold lots (proceeds already fold in the dividends of fully-closed
               positions, so the Dividends cell is empty). The per-currency loop
               below stays a secondary breakdown, shown only when portfolios
               actually differ. -->
          <Card class="p-4">
            <h2 class="mb-3 font-semibold">Investments</h2>
            <div class="overflow-x-auto">
              <Table aria-label="Investments">
                <THead>
                  <Tr>
                    <Th class="sr-only">Group</Th>
                    <Th align="right">Invested</Th>
                    <Th align="right">Value / Proceeds</Th>
                    <Th align="right">Gain/Loss</Th>
                    <Th align="right">%</Th>
                    <Th align="right">Dividends</Th>
                  </Tr>
                </THead>
                <TBody>
                  <Tr>
                    <Td class="font-medium">Active</Td>
                    <Td align="right">{formatCurrency(dash.summary.active.invested, dash.base_currency)}</Td>
                    <Td align="right">{formatCurrency(dash.summary.active.value, dash.base_currency)}</Td>
                    <Td align="right" class="font-medium {pnlColorClass(dash.summary.active.gain_loss)}">
                      {formatCurrency(dash.summary.active.gain_loss, dash.base_currency)}
                    </Td>
                    <Td align="right" class={pnlColorClass(dash.summary.active.gain_loss_pct)}>
                      {formatPercent(dash.summary.active.gain_loss_pct)}
                    </Td>
                    <Td align="right">{formatCurrency(dash.summary.active.dividends, dash.base_currency)}</Td>
                  </Tr>
                  <Tr>
                    <Td class="font-medium">Closed</Td>
                    <Td align="right">{formatCurrency(dash.summary.closed.invested, dash.base_currency)}</Td>
                    <Td align="right">{formatCurrency(dash.summary.closed.proceeds, dash.base_currency)}</Td>
                    <Td align="right" class="font-medium {pnlColorClass(dash.summary.closed.realized)}">
                      {formatCurrency(dash.summary.closed.realized, dash.base_currency)}
                    </Td>
                    <Td align="right" class={pnlColorClass(dash.summary.closed.realized_pct)}>
                      {formatPercent(dash.summary.closed.realized_pct)}
                    </Td>
                    <Td align="right" class="text-muted-foreground">—</Td>
                  </Tr>
                </TBody>
              </Table>
            </div>
          </Card>
        {/if}
        {#if !dash.summary || hasMultipleCurrencies}
          {#each dash.by_currency as c (c.currency)}
            <div class="space-y-2">
              {#if hasMultipleCurrencies}
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{c.currency}</p>
              {/if}
              <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
                <StatCard label="Invested" value={formatCurrency(c.invested, c.currency)} />
                <StatCard
                  label="Current Value"
                  value={formatCurrency(c.value, c.currency)}
                  valueClass="text-2xl sm:text-3xl"
                />
                <StatCard
                  label="Gain/Loss"
                  value={formatCurrency(c.gain_loss, c.currency)}
                  delta={formatPercent(c.gain_loss_pct)}
                  deltaValue={Number(c.gain_loss_pct)}
                />
                <StatCard label="ROI" value={formatPercent(c.gain_loss_pct)} />
              </div>
            </div>
          {/each}
        {/if}
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <Card class="p-4">
          <h2 class="mb-4 font-semibold">Portfolio History</h2>
          {#if dash.history?.some((h) => h.series?.length)}
            <PortfolioLineChart histories={dash.history} />
          {:else}
            <p class="text-sm text-muted-foreground">No price history yet</p>
          {/if}
        </Card>

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
      </div>

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
          <div class="grid gap-4 md:grid-cols-2">
            <GeographyChart data={alloc.regions} currency={alloc.currency} covered={alloc.covered_value} excluded={alloc.excluded_value} />
            <SectorChart data={alloc.sectors} currency={alloc.currency} covered={alloc.covered_value} excluded={alloc.excluded_value} />
          </div>
        {/if}
      </div>

      {#each dash.assets as pa (pa.portfolio_id)}
        <div class="rounded-card border-border bg-surface p-4 shadow-card">
          <button onclick={() => toggle(pa.portfolio_id)} class="flex w-full items-center gap-2 text-left">
            {#if expanded.has(pa.portfolio_id)}
              <ChevronDown class="h-4 w-4 text-muted-foreground" />
            {:else}
              <ChevronRight class="h-4 w-4 text-muted-foreground" />
            {/if}
            <span class="font-semibold">{pa.portfolio_name}</span>
            <span class="ml-1 text-xs text-muted-foreground">({pa.currency})</span>
            <span class="ml-auto text-sm text-muted-foreground tabular-nums">
              {formatCurrency(
                dash.portfolios.find((p) => p.portfolio_id === pa.portfolio_id)?.active.value ?? 0,
                pa.currency,
              )}
            </span>
          </button>

          {#if expanded.has(pa.portfolio_id)}
            <div class="mt-3">
              {#if pa.assets.length === 0}
                <p class="text-sm text-muted-foreground">No assets in this portfolio.</p>
              {:else}
                <PositionTable rows={positionRows(pa)} currency={pa.currency} linkAssets showRealized />
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
