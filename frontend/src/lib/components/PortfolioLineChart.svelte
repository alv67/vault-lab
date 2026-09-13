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
  import type { DashboardHistory } from '$lib/services/api'
  import { resolvePalette } from '$lib/chartPalette'
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
    seriesId: string
    value: unknown
    axisValue?: string | number
  }

  let {
    histories = [] as DashboardHistory[],
    colors = null as string[] | null,
  } = $props()

  const currencyById = $derived(
    Object.fromEntries(histories.map((h) => [h.portfolio_id, h.currency])),
  )

  // Theme-aware series palette: falls back to the resolved chart tokens unless
  // the caller explicitly overrides the colors (mirrors `resolved()` so the
  // options re-evaluate — and the {#key} below re-inits the chart — on theme
  // flips).
  const seriesColors = $derived(colors ?? resolvePalette(resolved()))

  const options = $derived.by((): EChartsOption => ({
    color: seriesColors,
    tooltip: {
      trigger: 'axis',
      formatter: (params: unknown) => {
        const raw = Array.isArray(params) ? params : [params]
        const rows = raw as TooltipRow[]
        const date = rows[0]?.axisValue
        const lines = rows
          .filter((p) => p.value != null)
          .map((p) => {
            const v = Array.isArray(p.value) ? p.value[1] : p.value
            const currency = currencyById[p.seriesId] ?? 'USD'
            return `${p.marker}${p.seriesName}: <b>${formatCurrency(Number(v), currency)}</b>`
          })
        const dateLabel = date ? `<div>${new Date(date).toLocaleDateString()}</div>` : ''
        return `${dateLabel}${lines.join('<br/>')}`
      },
    },
    legend: {
      data: histories.map((h) => h.portfolio_name),
      top: 0,
    },
    grid: { left: 48, right: 16, top: 40, bottom: 52 },
    dataZoom: [
      { type: 'inside', xAxisIndex: 0 },
      { type: 'slider', xAxisIndex: 0, bottom: 0 },
    ],
    xAxis: {
      type: 'time',
      axisLabel: {
        fontSize: 11,
        formatter: (value: string | number) => new Date(value).toLocaleDateString(),
      },
    },
    yAxis: {
      type: 'value',
      axisLabel: { fontSize: 11 },
    },
    series: histories.map((h) => ({
      id: h.portfolio_id,
      name: h.portfolio_name,
      type: 'line',
      data: h.series.map((pt) => [pt.date, Number(pt.value)]),
      connectNulls: true,
      sampling: 'lttb',
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2 },
    })),
  }))
</script>

<div class="h-[320px] w-full">
  <!-- {#key} re-inits the chart when the theme flips so the ECharts theme
       object passed below is picked up (svelte-echarts only reads `theme`
       at init time). -->
  {#key resolved()}
    <Chart {init} {options} theme={VAULTLAB_CHART_THEMES[resolved()]} />
  {/key}
</div>
