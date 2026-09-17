<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { BarChart, LineChart } from 'echarts/charts'
  import {
    DataZoomComponent,
    GridComponent,
    LegendComponent,
    TooltipComponent,
  } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatSignedPercent } from '$lib/format'
  import type { PerformanceBucket } from '$lib/services/api'
  import { chartSemanticColors } from '$lib/chartPalette'
  import { VAULTLAB_CHART_THEMES } from '$lib/chartTheme'
  import { resolved } from '$lib/stores/theme.svelte'

  use([
    BarChart,
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
    granularity = 'month' as 'month' | 'year',
  } = $props()

  // Bar/line colors come from the semantic chart tokens, re-evaluated on
  // theme flips (the {#key} block below also re-inits the chart with the new
  // ECharts theme), same convention as PositionChart.
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
          // Both series are percentages (return and cumulative TWR): no
          // currency formatting here, just a signed `+3.42%`.
          const lines = series
            .filter((p) => p.value != null)
            .map((p) => {
              const v = Array.isArray(p.value) ? p.value[1] : p.value
              return `${p.marker}${p.seriesName}: <b>${formatSignedPercent(Number(v))}</b>`
            })
          return `<div>${period}</div>${lines.join('<br/>')}`
        },
      },
      legend: {
        data: ['Gain/Loss', 'Cumulative'],
        top: 0,
      },
      grid: { left: 56, right: 16, top: 40, bottom: 52 },
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
        // Percentage axis: the values are already % numbers (e.g. 3.42).
        axisLabel: { fontSize: 11, formatter: '{value}%' },
      },
      series: [
        {
          name: 'Gain/Loss',
          type: 'bar',
          // Time-weighted return generated inside each bucket: green bar
          // when positive, red when negative (per-bar itemStyle, mirrors
          // `pnlColorClass`).
          data: rows.map((b) => {
            const ret = Number(b.return)
            return { value: ret, itemStyle: { color: ret >= 0 ? semantic.positive : semantic.negative } }
          }),
          itemStyle: { color: semantic.positive },
          barMaxWidth: 28,
        },
        {
          name: 'Cumulative',
          type: 'line',
          // Cumulative time-weighted return (TWR) up to the end of each
          // bucket; reuses the amber cumulative-line semantic token.
          data: rows.map((b) => Number(b.twr)),
          smooth: false,
          symbolSize: 5,
          lineStyle: { width: 2, color: semantic.realized },
          itemStyle: { color: semantic.realized },
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
