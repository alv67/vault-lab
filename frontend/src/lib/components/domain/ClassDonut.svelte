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
  import { t } from '$lib/i18n/index.svelte'
  import { cx } from '$lib/components/ui/utils'
  import ChartTableToggle from '$lib/components/ui/ChartTableToggle.svelte'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'

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
    showTableToggle = true,
  }: {
    data?: AssetClassSlice[]
    currency?: string
    /** Optional heading rendered above the donut. */
    label?: string
    /** Hide the shared Chart ⇄ Table toggle (EPIC K.5b, spec §9.1);
     * shown by default like in every other chart wrapper. */
    showTableToggle?: boolean
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

  // ── "View as table" (EPIC K.5b, spec §9.1) ──────────────────────────────
  // The table lists the same shaped `rows` the slices are drawn from; the
  // switch unmounts the canvas (out of the a11y tree) and empty data keeps
  // showing the empty state instead of an empty table.
  let view = $state<'chart' | 'table'>('chart')
  const showTable = $derived(showTableToggle && view === 'table')
  const caption = $derived(
    label ? t('chartView.caption', { name: label }) : t('chartView.captionGeneric'),
  )

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

{#if label || (showTableToggle && rows.length > 0)}
  <div
    class={cx(
      'mb-3 flex flex-wrap items-center gap-2',
      label ? 'justify-between' : 'justify-end',
    )}
  >
    {#if label}
      <h3 class="font-semibold">{label}</h3>
    {/if}
    {#if showTableToggle && rows.length > 0}
      <ChartTableToggle name={label} bind:view />
    {/if}
  </div>
{/if}
{#if rows.length === 0}
  <div class="flex h-[280px] w-full items-center justify-center text-sm text-muted-foreground">
    Nessuna allocazione per classi
  </div>
{:else if showTable}
  <div class="overflow-x-auto">
    <Table>
      <caption class="sr-only">{caption}</caption>
      <THead>
        <Tr>
          <Th>{t('chartView.colName')}</Th>
          <Th align="right">{t('chartView.colValue')}</Th>
          <Th align="right">{t('chartView.colWeight')}</Th>
        </Tr>
      </THead>
      <TBody>
        {#each rows as r (r.class)}
          <Tr>
            <!-- Friendly class label, same mapping the slice names use. -->
            <Td class="font-medium">{labelFor(r.class)}</Td>
            <Td align="right">{formatCurrency(r.value, currency)}</Td>
            <Td align="right">{formatPercent(r.weight)}</Td>
          </Tr>
        {/each}
      </TBody>
    </Table>
  </div>
{:else}
  <div class="h-[280px] w-full">
    {#key resolved()}
      <Chart {init} {options} theme={VAULTLAB_CHART_THEMES[resolved()]} />
    {/key}
  </div>
{/if}
