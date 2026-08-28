<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { PieChart } from 'echarts/charts'
  import { LegendComponent, TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatPercent } from '$lib/format'
  import type { ExposureRow } from '$lib/services/api'

  use([PieChart, LegendComponent, TooltipComponent, CanvasRenderer])

  interface TooltipItem {
    marker: string
    name: string
    value: unknown
  }

  let { data = [] as ExposureRow[], title = 'Distribuzione' } = $props()

  const rows = $derived(data.filter((r) => Number(r.weight) > 0))

  const options = $derived.by((): EChartsOption => ({
    color: [
      '#2563eb', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
      '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16',
      '#06b6d4', '#a855f7',
    ],
    tooltip: {
      trigger: 'item',
      formatter: (params: unknown) => {
        const p = params as TooltipItem
        return `${p.marker}${p.name}: <b>${formatPercent(Number(p.value))}</b>`
      },
    },
    legend: rows.length > 0 && rows.length <= 6
      ? { bottom: 0, type: 'scroll', textStyle: { fontSize: 11 } }
      : undefined,
    series: [
      {
        name: title,
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: true,
        label: {
          formatter: '{b}: {d}%',
          fontSize: 11,
        },
        labelLine: { length: 10, length2: 10 },
        data: rows.map((r) => ({ name: r.name, value: Number(r.weight) })),
      },
    ],
  }))
</script>

{#if rows.length === 0}
  <div class="flex h-[280px] w-full items-center justify-center text-sm text-gray-400">
    Nessuna distribuzione
  </div>
{:else}
  <div class="h-[280px] w-full">
    <Chart {init} {options} />
  </div>
{/if}