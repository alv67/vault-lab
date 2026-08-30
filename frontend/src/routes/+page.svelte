<script module lang="ts">
  let sessionRefreshed = false
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { SvelteSet } from 'svelte/reactivity'
  import { resolve } from '$app/paths'
  import { portfolioApi, pricesApi, type Dashboard, type DashboardAllocation } from '$lib/services/api'
  import { toast } from '$lib/stores/toast.svelte'
  import PortfolioLineChart from '$lib/components/PortfolioLineChart.svelte'
  import GeographyChart from '$lib/components/domain/GeographyChart.svelte'
  import SectorChart from '$lib/components/domain/SectorChart.svelte'
  import { formatCurrency, formatPercent } from '$lib/format'
  import { ChevronDown, ChevronRight } from 'lucide-svelte'

  const COLORS = [
    '#3b82f6', '#10b981', '#f59e0b', '#ef4444',
    '#8b5cf6', '#ec4899', '#14b8a6', '#f97316',
  ]

  let dash = $state<Dashboard | null>(null)
  let alloc = $state<DashboardAllocation | null>(null)
  let loading = $state(true)
  let expanded = new SvelteSet<string>()
  let initialized = false

  function glClass(value?: string | number): string {
    return Number(value ?? 0) >= 0 ? 'text-green-600' : 'text-red-600'
  }

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
          if (report.rate_limited) {
            toast.warning('Yahoo Finance ha limitato le richieste: alcuni prezzi non aggiornati')
          } else if (report.issues.length > 0) {
            toast.warning(`${report.issues.length} aggiornamenti prezzi non riusciti (Yahoo)`)
          }
          return portfolioApi.dashboard()
        })
        .then((fresh) => { dash = fresh })
        .catch(() => { /* keep current data */ })
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

  const chartData = $derived.by(() => {
    const histories = dash?.history ?? []
    if (!histories.length) return []
    const dateSet: Record<string, boolean> = {}
    histories.forEach((h) => h.series.forEach((pt) => (dateSet[pt.date.slice(0, 10)] = true)))
    const dates = Object.keys(dateSet).sort()
    return dates.map((date) => {
      const row: { date: string } & Record<string, string | number | null> = { date }
      histories.forEach((h) => {
        const pt = h.series.find((s) => s.date.slice(0, 10) === date)
        row[h.portfolio_id] = pt ? Number(pt.value) : null
      })
      return row
    })
  })

  const hasMultipleCurrencies = $derived((dash?.by_currency?.length ?? 0) > 1)
</script>

<div class="p-6">
  <h1 class="mb-6 text-2xl font-bold">Dashboard</h1>

  {#if loading}
    <p class="text-gray-500">Loading...</p>
  {:else if !dash?.portfolios?.length}
    <div class="rounded-lg border-2 border-dashed p-12 text-center">
      <p class="mb-4 text-gray-500">No portfolios yet</p>
      <a
        href={resolve('/portfolios')}
        class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white"
      >
        Create your first portfolio
      </a>
    </div>
  {:else}
    <div class="space-y-6">
      <div class="rounded-xl bg-white p-4 shadow">
        <h2 class="mb-4 font-semibold">Portfolio History</h2>
        {#if chartData.length > 0}
          <PortfolioLineChart data={chartData} histories={dash.history} colors={COLORS} />
        {:else}
          <p class="text-sm text-gray-400">No price history yet</p>
        {/if}
      </div>

      <div class="rounded-xl bg-white p-4 shadow">
        <h2 class="mb-4 font-semibold">Allocazione complessiva</h2>
        {#if alloc == null}
          <p class="text-sm text-gray-400">Allocazione non disponibile</p>
        {:else}
          <div class="grid gap-4 md:grid-cols-2">
            <GeographyChart data={alloc.regions} currency={alloc.currency} covered={alloc.covered_value} excluded={alloc.excluded_value} />
            <SectorChart data={alloc.sectors} currency={alloc.currency} covered={alloc.covered_value} excluded={alloc.excluded_value} />
          </div>
        {/if}
      </div>

      {#if hasMultipleCurrencies}
        <div class="rounded-xl bg-white p-4 shadow">
          <h2 class="mb-4 font-semibold">Performance by Currency</h2>
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="border-b text-gray-500">
                <th class="pb-2">Currency</th>
                <th class="pb-2">Invested</th>
                <th class="pb-2">Value</th>
                <th class="pb-2">Gain/Loss</th>
                <th class="pb-2">Return</th>
              </tr>
            </thead>
            <tbody>
              {#each dash.by_currency as c (c.currency)}
                <tr class="border-b last:border-0">
                  <td class="py-2 font-medium">{c.currency}</td>
                  <td class="py-2">{formatCurrency(c.invested, c.currency)}</td>
                  <td class="py-2">{formatCurrency(c.value, c.currency)}</td>
                  <td class="py-2 font-medium {glClass(c.gain_loss)}">
                    {formatCurrency(c.gain_loss, c.currency)}
                  </td>
                  <td class="py-2 font-medium {glClass(c.gain_loss_pct)}">
                    {formatPercent(c.gain_loss_pct)}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

      <div class="rounded-xl bg-white p-4 shadow">
        <h2 class="mb-4 font-semibold">Portfolios</h2>
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="border-b text-gray-500">
              <th class="pb-2">Portfolio</th>
              <th class="pb-2">Currency</th>
              <th class="pb-2">Assets</th>
              <th class="pb-2">Invested</th>
              <th class="pb-2">Value</th>
              <th class="pb-2">Realized</th>
              <th class="pb-2">Gain/Loss</th>
              <th class="pb-2">Return</th>
            </tr>
          </thead>
          <tbody>
            {#each dash.portfolios as p (p.portfolio_id)}
              <tr class="border-b last:border-0">
                <td class="py-2 font-medium">{p.portfolio_name}</td>
                <td class="py-2">{p.currency}</td>
                <td class="py-2">{p.asset_count}</td>
                <td class="py-2">{formatCurrency(p.invested, p.currency)}</td>
                <td class="py-2">{formatCurrency(p.value, p.currency)}</td>
                <td class="py-2 font-medium {glClass(p.realized_gl)}">
                  {formatCurrency(p.realized_gl, p.currency)}
                </td>
                <td class="py-2 font-medium {glClass(p.gain_loss)}">
                  {formatCurrency(p.gain_loss, p.currency)}
                </td>
                <td class="py-2 font-medium {glClass(p.gain_loss_pct)}">
                  {formatPercent(p.gain_loss_pct)}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#each dash.assets as pa (pa.portfolio_id)}
        <div class="rounded-xl bg-white p-4 shadow">
          <button onclick={() => toggle(pa.portfolio_id)} class="flex w-full items-center gap-2 text-left">
            {#if expanded.has(pa.portfolio_id)}
              <ChevronDown class="h-4 w-4 text-gray-400" />
            {:else}
              <ChevronRight class="h-4 w-4 text-gray-400" />
            {/if}
            <span class="font-semibold">{pa.portfolio_name}</span>
            <span class="ml-1 text-xs text-gray-500">({pa.currency})</span>
            <span class="ml-auto text-sm text-gray-600">
              {formatCurrency(
                dash.portfolios.find((p) => p.portfolio_id === pa.portfolio_id)?.value ?? 0,
                pa.currency,
              )}
            </span>
          </button>

          {#if expanded.has(pa.portfolio_id)}
            <div class="mt-3 overflow-x-auto">
              {#if pa.assets.length === 0}
                <p class="text-sm text-gray-400">No assets in this portfolio.</p>
              {:else}
                <table class="w-full text-left text-sm">
                  <thead>
                    <tr class="border-b text-gray-500">
                      <th class="pb-2">Ticker</th>
                      <th class="pb-2">Name</th>
                      <th class="pb-2">Currency</th>
                      <th class="pb-2 text-right">Qty</th>
                      <th class="pb-2 text-right">Invested</th>
                      <th class="pb-2 text-right">Value</th>
                      <th class="pb-2 text-right">Gain/Loss</th>
                      <th class="pb-2 text-right">Realized</th>
                      <th class="pb-2 text-right">ROI</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each pa.assets as a (a.asset_id)}
                      <tr class="border-b last:border-0">
                        <td class="py-2 font-medium">
                          <a href={resolve(`/assets/${a.asset_id}`)} class="text-blue-600 hover:underline">
                            {a.ticker}
                          </a>
                        </td>
                        <td class="py-2 text-gray-600">{a.name}</td>
                        <td class="py-2">
                          {a.currency}
                          {#if a.fx_missing}
                            <span
                              class="ml-2 rounded bg-amber-100 px-1.5 py-0.5 text-xs text-amber-700"
                              title="Exchange rate not available: excluded from portfolio total"
                            >
                              cambio mancante
                            </span>
                          {/if}
                        </td>
                        <td class="py-2 text-right">{a.qty}</td>
                        <td class="py-2 text-right">{formatCurrency(a.invested, a.currency)}</td>
                        <td class="py-2 text-right">{formatCurrency(a.value, a.currency)}</td>
                        <td class="py-2 text-right font-medium {glClass(a.gain_loss)}">
                          {formatCurrency(a.gain_loss, a.currency)}
                        </td>
                        <td class="py-2 text-right font-medium {glClass(a.realized_pf ?? a.realized)}">
                          {formatCurrency(a.realized_pf ?? a.realized, pa.currency)}
                        </td>
                        <td class="py-2 text-right font-medium {glClass(a.roi)}">
                          {formatPercent(a.roi)}
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
