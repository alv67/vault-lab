<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { PieChart } from 'echarts/charts'
  import { TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatCurrency, formatPercent } from '$lib/format'
  import { resolvePalette } from '$lib/chartPalette'
  import { VAULTLAB_CHART_THEMES } from '$lib/chartTheme'
  import { resolved } from '$lib/stores/theme.svelte'
  import { t } from '$lib/i18n/index.svelte'
  import ChartTableToggle from '$lib/components/ui/ChartTableToggle.svelte'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'

  use([PieChart, TooltipComponent, CanvasRenderer])

  interface TooltipItem {
    marker: string
    name: string
    value: unknown
  }

  interface AllocationDatum {
    name: string
    value: number
  }

  interface AllocationRow extends AllocationDatum {
    weight: number
  }

  let {
    data = [] as AllocationDatum[],
    currency = 'USD',
    title = 'Allocation',
    showLegend = false,
    showValue = true,
    showTableToggle = true,
  }: {
    data?: AllocationDatum[]
    currency?: string
    title?: string
    showLegend?: boolean
    /** Hide the formatted amount in the tooltip (e.g. when slices mix currencies). */
    showValue?: boolean
    /** Hide the shared Chart ⇄ Table toggle (EPIC K.5b, spec §9.1);
     * shown by default like in every other chart wrapper. */
    showTableToggle?: boolean
  } = $props()

  // Non-positive rows are dropped and the weights are recomputed over the
  // remaining total, so the visible slices always sum to 100%.
  const rows = $derived.by((): AllocationRow[] => {
    const positive = data.filter((d) => Number(d.value) > 0)
    const total = positive.reduce((acc, d) => acc + Number(d.value), 0)
    return positive.map((d) => ({ name: d.name, value: Number(d.value), weight: (Number(d.value) / total) * 100 }))
  })

  // ── "View as table" (EPIC K.5b, spec §9.1) ──────────────────────────────
  // The table lists the same recomputed `rows` (name / amount / share of the
  // visible total); the amount column follows `showValue`, mirroring the
  // tooltip's mixed-currency muting. Switching unmounts the canvas (out of
  // the a11y tree) and empty data keeps showing the empty state.
  let view = $state<'chart' | 'table'>('chart')
  const showTable = $derived(showTableToggle && view === 'table')

  const palette = $derived(resolvePalette(resolved()))
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
        const row = rows.find((r) => r.name === p.name)
        if (!row) return ''
        const valueLine = showValue ? `Valore: <b>${formatCurrency(row.value, currency)}</b><br/>` : ''
        return `${p.marker}${p.name}<br/>${valueLine}Peso: <b>${formatPercent(row.weight)}</b>`
      },
    },
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
          color: labelColor,
          textBorderColor: 'transparent',
          textBorderWidth: 0,
        },
        labelLine: { length: 10, length2: 10 },
        data: rows.map((r) => ({ name: r.name, value: r.value })),
      },
    ],
  }))
</script>

{#if rows.length === 0}
  <div class="flex h-[240px] w-full items-center justify-center text-sm text-muted-foreground">
    Nessuna allocazione
  </div>
{:else}
  {#if showTableToggle}
    <div class="mb-2 flex justify-end">
      <ChartTableToggle name={title} bind:view />
    </div>
  {/if}
  {#if showTable}
    <div class="overflow-x-auto">
      <Table>
        <caption class="sr-only">{t('chartView.caption', { name: title })}</caption>
        <THead>
          <Tr>
            <Th>{t('chartView.colName')}</Th>
            {#if showValue}
              <Th align="right">{t('chartView.colValue')}</Th>
            {/if}
            <Th align="right">{t('chartView.colWeight')}</Th>
          </Tr>
        </THead>
        <TBody>
          {#each rows as r (r.name)}
            <Tr>
              <Td class="font-medium">{r.name}</Td>
              {#if showValue}
                <Td align="right">{formatCurrency(r.value, currency)}</Td>
              {/if}
              <Td align="right">{formatPercent(r.weight)}</Td>
            </Tr>
          {/each}
        </TBody>
      </Table>
    </div>
  {:else}
    <div class="h-[240px] w-full">
      {#key resolved()}
        <Chart {init} {options} theme={VAULTLAB_CHART_THEMES[resolved()]} />
      {/key}
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
            <span class="ml-auto text-muted-foreground">{formatPercent(r.weight)}</span>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
{/if}
