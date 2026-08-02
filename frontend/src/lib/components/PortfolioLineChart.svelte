<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { LineChart } from 'echarts/charts'
  import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatCurrency } from '$lib/format'
  import type { PortfolioHistory } from '$lib/services/api'

  use([LineChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

  const DEFAULT_COLORS = [
    '#3b82f6', '#10b981', '#f59e0b', '#ef4444',
    '#8b5cf6', '#ec4899', '#14b8a6', '#f97316',
  ]

  interface ChartRow {
    date: string
    [portfolioId: string]: string | number | null
  }

  interface TooltipRow {
    marker: string
    seriesName: string
    seriesId: string
    value: unknown
    axisValue?: string | number
  }

  let {
    data = [] as ChartRow[],
    histories = [] as PortfolioHistory[],
    colors = DEFAULT_COLORS,
  } = $props()

  const currencyById = $derived(
    Object.fromEntries(histories.map((h) => [h.portfolio_id, h.currency])),
  )

  const options = $derived.by((): EChartsOption => ({
    color: colors,
    tooltip: {
      trigger: 'axis',
      formatter: (params: unknown) => {
        const raw = Array.isArray(params) ? params : [params]
        const rows = raw as TooltipRow[]
        const date = rows[0]?.axisValue
        const lines = rows
          .filter((p) => p.value != null)
          .map((p) => {
            const value = Number(p.value)
            const currency = currencyById[p.seriesId] ?? 'USD'
            return `${p.marker}${p.seriesName}: <b>${formatCurrency(value, currency)}</b>`
          })
        const dateLabel = date ? `<div>${new Date(date).toLocaleDateString()}</div>` : ''
        return `${dateLabel}${lines.join('<br/>')}`
      },
    },
    legend: {
      data: histories.map((h) => h.portfolio_name),
      top: 0,
    },
    grid: { left: 48, right: 16, top: 40, bottom: 28 },
    xAxis: {
      type: 'category',
      data: data.map((d) => d.date),
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
      data: data.map((d) => d[h.portfolio_id] ?? null),
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2 },
    })),
  }))
</script>

<div class="h-[320px] w-full">
  <Chart {init} {options} />
</div>
