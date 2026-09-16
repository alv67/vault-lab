<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { LineChart } from 'echarts/charts'
  import {
    DataZoomComponent,
    GridComponent,
    LegendComponent,
    TooltipComponent,
  } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatCurrency } from '$lib/format'
  import type { PerformanceBucket } from '$lib/services/api'
  import { chartSemanticColors } from '$lib/chartPalette'
  import { VAULTLAB_CHART_THEMES } from '$lib/chartTheme'
  import { resolved } from '$lib/stores/theme.svelte'

  use([
    LineChart,
    DataZoomComponent,
    GridComponent,
    LegendComponent,
    TooltipComponent,
    CanvasRenderer,
  ])

  interface TooltipRow {
    marker: string
    seriesName: string
    value: unknown
    axisValue?: string | number
  }

  let {
    buckets = [] as PerformanceBucket[],
    currency = 'USD',
    granularity = 'month' as 'month' | 'year',
  } = $props()

  // Line colors come from the semantic chart tokens (grey cost basis for the
  // invested capital, green market value), re-evaluated on theme flips (the
  // {#key} block below also re-inits the chart with the new ECharts theme),
  // same convention as PositionChart / PerformanceChart.
  const semantic = $derived(chartSemanticColors(resolved()))

  /** "2025-06" → "Jun 2025" (month name follows the browser locale);
   * "2025" stays "2025" on the annual granularity. */
  function formatPeriod(period: string): string {
    if (granularity !== 'month') return period
    const [year, month] = period.split('-').map(Number)
    if (!year || !month) return period
    return new Date(year, month - 1, 1).toLocaleDateString(undefined, {
      month: 'short',
      year: 'numeric',
    })
  }

  const options = $derived.by((): EChartsOption => {
    const rows = buckets ?? []
    return {
      tooltip: {
        trigger: 'axis',
        formatter: (params: unknown) => {
          const raw = Array.isArray(params) ? params : [params]
          const series = raw as TooltipRow[]
          const period = series[0]?.axisValue ?? ''
          const lines = series
            .filter((p) => p.value != null)
            .map((p) => {
              const v = Array.isArray(p.value) ? p.value[1] : p.value
              return `${p.marker}${p.seriesName}: <b>${formatCurrency(Number(v), currency)}</b>`
            })
          return `<div>${period}</div>${lines.join('<br/>')}`
        },
      },
      legend: {
        data: ['Invested', 'Value'],
        top: 0,
      },
      grid: { left: 48, right: 16, top: 40, bottom: 52 },
      // Long monthly ranges stay usable: wheel/drag zoom plus the slider.
      dataZoom: [
        { type: 'inside', xAxisIndex: 0 },
        { type: 'slider', xAxisIndex: 0, bottom: 0 },
      ],
      xAxis: {
        type: 'category',
        data: rows.map((b) => formatPeriod(b.period)),
        axisLabel: { fontSize: 11, hideOverlap: true },
      },
      yAxis: {
        type: 'value',
        axisLabel: { fontSize: 11 },
      },
      series: [
        {
          name: 'Invested',
          type: 'line',
          // Net invested capital at the end of each bucket: it only moves on
          // cash flows, so a stepped line mirrors the PositionChart cost basis.
          data: rows.map((b) => Number(b.invested)),
          step: 'end',
          smooth: false,
          symbolSize: 5,
          lineStyle: { width: 2, color: semantic.costBasis },
          itemStyle: { color: semantic.costBasis },
        },
        {
          name: 'Value',
          type: 'line',
          // Market value at the end of each bucket.
          data: rows.map((b) => Number(b.value)),
          smooth: true,
          symbolSize: 5,
          lineStyle: { width: 2, color: semantic.marketValue },
          itemStyle: { color: semantic.marketValue },
        },
      ],
    }
  })
</script>

{#if (buckets ?? []).length === 0}
  <div class="flex h-[340px] w-full items-center justify-center text-sm text-muted-foreground">
    No data
  </div>
{:else}
  <div class="h-[340px] w-full">
    <!-- {#key} re-inits the chart when the theme flips so the ECharts theme
         object passed below is picked up (svelte-echarts only reads `theme`
         at init time). -->
    {#key resolved()}
      <Chart {init} {options} theme={VAULTLAB_CHART_THEMES[resolved()]} />
    {/key}
  </div>
{/if}
