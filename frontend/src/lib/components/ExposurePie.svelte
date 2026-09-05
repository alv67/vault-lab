<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { PieChart } from 'echarts/charts'
  import { TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatPercent } from '$lib/format'
  import type { ExposureRow } from '$lib/services/api'
  import { CHART_PALETTE } from '$lib/chartPalette'

  use([PieChart, TooltipComponent, CanvasRenderer])

  interface TooltipItem {
    marker: string
    name: string
    value: unknown
  }

  // Shape of a pie data item: real rows plus the optional invisible residual
  // slice used in "open" (complete = false) mode. The extra keys mirror
  // ECharts' PieDataItemOption so the array stays assignable to `data`.
  interface PieSlice {
    name: string
    value: number
    itemStyle?: { opacity: number }
    label?: { show: boolean }
    labelLine?: { show: boolean }
    emphasis?: { disabled: boolean }
    silent?: boolean
  }

  let {
    data = [] as ExposureRow[],
    title = 'Distribuzione',
    showLegend = false,
    mute = false,
    // When false the donut is rendered "open": if the passed rows sum below 100
    // a transparent residual slice is appended so visible arcs are proportional
    // to the real percentages (default true = legacy behavior, no residual).
    complete = true,
  } = $props()

  const rows = $derived(data.filter((r) => Number(r.weight) > 0))

  // Rows are normalized weights (> 0), so their sum is always > 0 here.
  const pieData = $derived.by((): PieSlice[] => {
    const slices: PieSlice[] = rows.map((r) => ({ name: r.name, value: Number(r.weight) }))
    if (!complete) {
      const sum = slices.reduce((acc, s) => acc + s.value, 0)
      if (sum < 100 - 0.5) {
        slices.push({
          name: '',
          value: 100 - sum,
          itemStyle: { opacity: 0 },
          label: { show: false },
          labelLine: { show: false },
          emphasis: { disabled: true },
          silent: true,
        })
      }
    }
    return slices
  })

  const options = $derived.by((): EChartsOption => ({
    color: CHART_PALETTE,
    tooltip: {
      trigger: 'item',
      formatter: (params: unknown) => {
        const p = params as TooltipItem
        // Guard: the synthetic residual slice has an empty name and must never
        // surface in a tooltip.
        if (!p.name) return ''
        return `${p.marker}${p.name}: <b>${formatPercent(Number(p.value))}</b>`
      },
    },
    series: [
      {
        name: title,
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: true,
        label: mute ? { show: false } : { formatter: '{b}: {d}%', fontSize: 11 },
        labelLine: mute ? { show: false } : { length: 10, length2: 10 },
        data: pieData,
      },
    ],
  }))

  const palette = CHART_PALETTE
</script>

{#if rows.length === 0}
  <div class="flex h-[280px] w-full items-center justify-center text-sm text-gray-400">
    Nessuna distribuzione
  </div>
{:else}
  <div class="h-[240px] w-full">
    <Chart {init} {options} />
  </div>
  {#if showLegend}
    <div class="mt-2 grid grid-cols-1 gap-x-3 gap-y-1 sm:grid-cols-2">
      {#each rows as r, i (r.name)}
        <div class="flex items-center gap-2 text-xs">
          <span
            class="inline-block h-2.5 w-2.5 shrink-0 rounded-sm"
            style="background-color: {palette[i % palette.length]};"
          ></span>
          <span class="truncate">{r.name}</span>
          <span class="ml-auto text-gray-500">{formatPercent(Number(r.weight))}</span>
        </div>
      {/each}
    </div>
  {/if}
{/if}