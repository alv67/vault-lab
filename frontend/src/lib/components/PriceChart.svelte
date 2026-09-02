<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { LineChart } from 'echarts/charts'
  import { DataZoomComponent, GridComponent, TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatCurrency } from '$lib/format'

  use([LineChart, DataZoomComponent, GridComponent, TooltipComponent, CanvasRenderer])

  interface PricePoint {
    date: string
    close: string
  }

  interface TooltipRow {
    marker: string
    seriesName: string
    value: unknown
    axisValue?: string | number
  }

  type ZoomSpec =
    | { type: 'startValue'; startValue: string }
    | 'MAX'
    | null

  let {
    series = [] as PricePoint[],
    currency = 'USD',
    zoomSpec = null as ZoomSpec,
    onDataZoom = null as (() => void) | null,
  } = $props()

  const zoomOption = $derived.by(() => {
    if (zoomSpec === 'MAX') return { start: 0, end: 100 }
    if (zoomSpec && zoomSpec.type === 'startValue') return { startValue: zoomSpec.startValue }
    return {}
  })

  const options = $derived.by((): EChartsOption => ({
    color: ['#2563eb'],
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
            return `${p.marker}${p.seriesName}: <b>${formatCurrency(Number(v), currency)}</b>`
          })
        const dateLabel = date ? `<div>${new Date(date).toLocaleDateString()}</div>` : ''
        return `${dateLabel}${lines.join('<br/>')}`
      },
    },
    grid: { left: 48, right: 16, top: 24, bottom: 52 },
    dataZoom: [
      {
        type: 'inside',
        xAxisIndex: 0,
        ...zoomOption,
      },
      {
        type: 'slider',
        xAxisIndex: 0,
        bottom: 0,
        ...zoomOption,
      },
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
    series: [
      {
        name: 'close',
        type: 'line',
        data: series.map((p) => [p.date, parseFloat(p.close)]),
        connectNulls: true,
        symbol: 'none',
        sampling: 'lttb',
        lineStyle: { width: 2 },
      },
    ],
  }))
</script>

{#if series.length === 0}
  <div class="flex h-[340px] w-full items-center justify-center text-sm text-gray-400">
    Nessun dato prezzi disponibile
  </div>
{:else}
  <div class="h-[340px] w-full">
    <Chart {init} {options} notMerge={false} ondatazoom={() => onDataZoom?.()} />
  </div>
{/if}