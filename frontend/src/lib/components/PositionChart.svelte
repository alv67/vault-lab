<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { LineChart } from 'echarts/charts'
  import {
    DataZoomComponent,
    GridComponent,
    LegendComponent,
    MarkLineComponent,
    TooltipComponent,
  } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatCurrency } from '$lib/format'
  import type { PositionPoint, SplitInfo } from '$lib/services/api'

  use([
    LineChart,
    DataZoomComponent,
    GridComponent,
    LegendComponent,
    MarkLineComponent,
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
    series = [] as PositionPoint[],
    currency = 'USD',
    splits = [] as SplitInfo[],
  } = $props()

  const LINES = [
    { key: 'cost_basis', name: 'Cost basis', color: '#64748b', step: 'end' },
    { key: 'market_value', name: 'Market value', color: '#16a34a', step: undefined },
    { key: 'realized', name: 'Realized', color: '#f59e0b', step: 'end' },
  ] as const

  const options = $derived.by((): EChartsOption => ({
    color: LINES.map((l) => l.color),
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
    legend: {
      data: LINES.map((l) => l.name),
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
    series: LINES.map((l) => ({
      name: l.name,
      type: 'line',
      data: series.map((p) => [p.date, Number(p[l.key])]),
      sampling: 'lttb',
      smooth: l.step === 'end' ? false : true,
      step: l.step,
      showSymbol: true,
      symbolSize: 4,
      lineStyle: { width: 2 },
      ...(l.key === 'market_value' && splits.length > 0
        ? {
            markLine: {
              symbol: 'none',
              silent: true,
              lineStyle: { type: 'dashed', color: '#7c3aed', width: 1 },
              label: {
                show: true,
                position: 'insideEndTop',
                formatter: '{b}',
                color: '#7c3aed',
                fontSize: 10,
              },
              data: splits.map((s) => ({
                xAxis: new Date(s.date).getTime(),
                name: s.ratio,
              })),
            },
          }
        : {}),
    })),
  }))
</script>

{#if series.length === 0}
  <div class="flex h-[340px] w-full items-center justify-center text-sm text-gray-400">
    No data
  </div>
{:else}
  <div class="h-[340px] w-full">
    <Chart {init} {options} />
  </div>
{/if}
