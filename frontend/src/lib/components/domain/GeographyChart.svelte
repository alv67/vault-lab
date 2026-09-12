<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { PieChart } from 'echarts/charts'
  import { LegendComponent, TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatCurrency, formatPercent } from '$lib/format'
  import type { RegionAllocation } from '$lib/services/api'
  import { chartSemanticColors, resolvePalette } from '$lib/chartPalette'
  import { VAULTLAB_CHART_THEMES } from '$lib/chartTheme'
  import { resolved } from '$lib/stores/theme.svelte'

  use([PieChart, LegendComponent, TooltipComponent, CanvasRenderer])

  interface TooltipItem {
    marker: string
    name: string
    value: unknown
  }

  let {
    data = [] as RegionAllocation[],
    currency = 'USD',
    covered = undefined,
    excluded = undefined,
  }: { data?: RegionAllocation[]; currency?: string; covered?: string; excluded?: string } = $props()

  // Le righe con peso zero restano visibili nella tabella ma non nel donut.
  const rows = $derived(data.filter((r) => Number(r.weight) > 0))

  const coveragePct = $derived.by(() => {
    const c = Number(covered || 0)
    const e = Number(excluded || 0)
    const tot = c + e
    return tot > 0 ? (c / tot) * 100 : 100
  })

  // Palette and "Other" grey come from the chart tokens and are re-evaluated
  // on theme flips; the {#key} block also re-inits the chart with the new
  // ECharts theme.
  const palette = $derived(resolvePalette(resolved()))
  const otherColor = $derived(chartSemanticColors(resolved()).other)

  const options = $derived.by((): EChartsOption => ({
    color: palette,
    tooltip: {
      trigger: 'item',
      formatter: (params: unknown) => {
        const p = params as TooltipItem
        const row = rows.find((r) => r.region === p.name)
        if (!row) return ''
        return `${p.marker}${p.name}<br/>Valore: <b>${formatCurrency(row.value, currency)}</b><br/>Peso: <b>${formatPercent(row.weight)}</b>`
      },
    },
    legend: rows.length > 0 && rows.length <= 6
      ? { bottom: 0, type: 'scroll', textStyle: { fontSize: 11 } }
      : undefined,
    series: [
      {
        name: 'Allocazione geografica',
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: true,
        label: {
          formatter: '{b}: {d}%',
          fontSize: 11,
        },
        labelLine: { length: 10, length2: 10 },
        data: rows.map((r) => ({
          name: r.region,
          value: Number(r.weight),
          // La fetta "Other" (macro-regioni fuori dalle 8 canoniche) è
          // evidenziata in grigio spento per distinguerla dalle altre.
          itemStyle: r.region === 'Other' || r.region === 'Other / Not Classified'
            ? { color: otherColor }
            : undefined,
        })),
      },
    ],
  }))
</script>

<div class="rounded-card border-border bg-surface p-4 shadow-card">
  <h2 class="mb-4 font-semibold">Allocazione geografica</h2>
  {#if rows.length === 0}
    <p class="text-sm text-muted-foreground">Nessuna allocazione per geografia</p>
  {:else}
    {#if covered !== undefined && excluded !== undefined && Number(excluded || 0) > 0}
      <p class="mb-3 text-xs text-muted-foreground">
        Copre il {coveragePct.toFixed(1)}% del portafoglio: il resto è in bond, crypto e strumenti non azionari, esclusi per natura.
      </p>
    {/if}
    <div class="flex flex-col gap-4 md:flex-row">
      <div class="w-full md:w-1/2 lg:w-1/3">
        <div class="h-[280px] w-full">
          {#key resolved()}
            <Chart {init} {options} theme={VAULTLAB_CHART_THEMES[resolved()]} />
          {/key}
        </div>
      </div>
      <div class="flex-1 overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="border-b border-border text-muted-foreground">
              <th class="pb-2">Territorio</th>
              <th class="pb-2 text-right">Valore</th>
              <th class="pb-2 text-right">Peso %</th>
            </tr>
          </thead>
          <tbody>
            {#each data as c (c.region)}
              <tr class="border-b border-border last:border-0">
                <td class="py-2 font-medium">{c.region}</td>
                <td class="py-2 text-right">{formatCurrency(c.value, currency)}</td>
                <td class="py-2 text-right font-medium">{formatPercent(c.weight)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>