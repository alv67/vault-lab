<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { PieChart } from 'echarts/charts'
  import { LegendComponent, TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { ASSET_CLASS_LABELS, formatCurrency, formatPercent } from '$lib/format'
  import type { AssetClassSlice } from '$lib/services/api'
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
    data = [] as AssetClassSlice[],
    currency = 'USD',
    label = undefined as string | undefined,
  }: {
    data?: AssetClassSlice[]
    currency?: string
    /** Optional heading rendered above the donut. */
    label?: string
  } = $props()

  // Backend class keys are mapped to friendly labels through the shared
  // ASSET_CLASS_LABELS table; zero/negative weights are dropped.
  const rows = $derived(data.filter((r) => Number(r.weight) > 0))
  function labelFor(cls: string): string {
    return ASSET_CLASS_LABELS[cls] ?? cls
  }
  function isOther(cls: string): boolean {
    return cls.toLowerCase() === 'other'
  }

  // Palette and "Other" grey come from the chart tokens and are re-evaluated
  // on theme flips; the {#key} block also re-inits the chart with the new
  // ECharts theme.
  const palette = $derived(resolvePalette(resolved()))
  const otherColor = $derived(chartSemanticColors(resolved()).other)
  // Pie labels do not inherit the ECharts theme textStyle: without an explicit
  // color they keep the default dark fill + white text border, which is
  // unreadable on a dark card ("outlined in white").
  const labelColor = $derived(VAULTLAB_CHART_THEMES[resolved()].textStyle.color)

  const options = $derived.by((): EChartsOption => ({
    color: palette,
    tooltip: {
      trigger: 'item',
      formatter: (params: unknown) => {
        const p = params as TooltipItem
        const row = rows.find((r) => labelFor(r.class) === p.name)
        if (!row) return ''
        return `${p.marker}${p.name}<br/>Valore: <b>${formatCurrency(row.value, currency)}</b><br/>Peso: <b>${formatPercent(row.weight)}</b>`
      },
    },
    legend: rows.length > 0 && rows.length <= 6
      ? { bottom: 0, type: 'scroll', textStyle: { fontSize: 11 } }
      : undefined,
    series: [
      {
        name: 'Allocazione per classi',
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: true,
        label: {
          formatter: '{b}: {d}%',
          fontSize: 11,
          color: labelColor,
          textBorderColor: 'transparent',
          textBorderWidth: 0,
        },
        labelLine: { length: 10, length2: 10 },
        // Slice angles come from the amounts, so they stay truthful even when
        // the caller sends weights computed over a different base.
        data: rows.map((r) => ({
          name: labelFor(r.class),
          value: Number(r.value),
          // The aggregated "other" bucket is muted in grey like in the
          // sibling donut charts.
          itemStyle: isOther(r.class) ? { color: otherColor } : undefined,
        })),
      },
    ],
  }))
</script>

{#if label}
  <h3 class="mb-3 font-semibold">{label}</h3>
{/if}
{#if rows.length === 0}
  <div class="flex h-[280px] w-full items-center justify-center text-sm text-muted-foreground">
    Nessuna allocazione per classi
  </div>
{:else}
  <div class="h-[280px] w-full">
    {#key resolved()}
      <Chart {init} {options} theme={VAULTLAB_CHART_THEMES[resolved()]} />
    {/key}
  </div>
{/if}
