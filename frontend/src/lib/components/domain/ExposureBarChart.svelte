<script lang="ts">
  import type { EChartsOption } from 'echarts'
  import { Chart } from 'svelte-echarts'
  import { init, use } from 'echarts/core'
  import { BarChart } from 'echarts/charts'
  import { GridComponent, TooltipComponent } from 'echarts/components'
  import { CanvasRenderer } from 'echarts/renderers'
  import { formatCurrency, formatPercent } from '$lib/format'
  import { resolvePalette } from '$lib/chartPalette'
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

  use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

  /** One horizontal bar. `value`/`weight` are decimal strings, like the rest
   * of the API (weight in percentage points). */
  export interface ExposureBarRow {
    name: string
    value: string
    weight: string
  }

  interface TooltipItem {
    marker: string
    name: string
    dataIndex: number
  }

  interface BarLabelItem {
    dataIndex: number
  }

  let {
    rows = [] as ExposureBarRow[],
    currency = 'USD',
    label = undefined as string | undefined,
    note = undefined as string | undefined,
    colorFor = undefined as ((name: string) => string | undefined) | undefined,
    labelFor = undefined as ((name: string) => string) | undefined,
    maxVisibleRows = undefined as number | undefined,
    showTableToggle = true,
  }: {
    rows?: ExposureBarRow[]
    currency?: string
    /** Optional heading rendered above the bars. */
    label?: string
    /** Optional muted caption (e.g. an equity-universe coverage note). */
    note?: string
    /** Per-row color override (e.g. greying out "Other"); returning
     * undefined (or omitting the prop) falls back to the chart palette. */
    colorFor?: (name: string) => string | undefined
    /** Maps a row's raw name to its axis label (e.g. an ISO country code to
     * its full name via `countryDisplayName`). Rows keep the raw name for
     * `colorFor`; when the mapped label differs from the raw name the
     * tooltip shows both (e.g. "United States (US)"). Unset = labels are
     * rendered verbatim (sector charts). */
    labelFor?: (name: string) => string
    /** When set, the canvas still renders every row at full height but is
     * wrapped in a viewport capped at `maxVisibleRows` rows with vertical
     * scrolling (dashboard country list). Unset = all rows visible. Table
     * mode deliberately ignores the cap: the page is the only scroll
     * container (spec §5.3 "collapse, don't shrink", EPIC K bug-fix). */
    maxVisibleRows?: number
    /** Hide the shared Chart ⇄ Table toggle (EPIC K.5b, spec §9.1). Only
     * needed by callers that already render the same rows as a list right
     * below the chart; every other card shows the toggle by default. */
    showTableToggle?: boolean
  } = $props()

  // Defensive shaping: drop non-positive rows and re-sort descending by value
  // even though the backend already sends sorted, non-zero data. The category
  // axis below is `inverse`d, so the first row ends up on top.
  const sorted = $derived(
    [...rows]
      .filter((r) => Number(r.value) > 0)
      .sort((a, b) => Number(b.value) - Number(a.value)),
  )

  // Bar colors come from the resolved palette (re-evaluated on theme flips;
  // the {#key} block also re-inits the chart with the new ECharts theme),
  // with the optional per-row colorFor override.
  const palette = $derived(resolvePalette(resolved()))
  const labelColor = $derived(VAULTLAB_CHART_THEMES[resolved()].textStyle.color)

  // Axis label for a row: the mapped display name when `labelFor` is set
  // (e.g. full country names), else the raw row name (sector charts).
  function displayName(name: string): string {
    return labelFor?.(name) ?? name
  }

  // One compact row per bar instead of a fixed canvas height.
  const ROW_HEIGHT = 30
  const height = $derived(Math.max(150, sorted.length * ROW_HEIGHT + 10))

  // Cap for the optional scroll viewport: the same height the canvas would
  // have with exactly `maxVisibleRows` rows, so lists at or below the cap
  // never show a scrollbar (undefined = no viewport, chart fully visible).
  const scrollCap = $derived(
    maxVisibleRows && maxVisibleRows > 0 ? maxVisibleRows * ROW_HEIGHT + 10 : undefined,
  )

  // ── "View as table" (EPIC K.5b, spec §9.1) ──────────────────────────────
  // The table renders the same shaped `sorted` rows the bars plot; switching
  // to it unmounts the canvas (out of the a11y tree), and empty data keeps
  // showing the plain empty state instead of an empty table. Table mode has
  // NO `maxVisibleRows` cap (EPIC K bug-fix): every row renders and the page
  // scrolls — no nested vertical scroll area inside the already-scrolling
  // card; below `sm` each row collapses to a stacked key–value grid so the
  // table never needs its own horizontal scroll either (spec §5.3).
  let view = $state<'chart' | 'table'>('chart')
  const showTable = $derived(showTableToggle && view === 'table')
  const caption = $derived(
    label ? t('chartView.caption', { name: label }) : t('chartView.captionGeneric'),
  )

  const options = $derived.by((): EChartsOption => ({
    color: palette,
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const raw = Array.isArray(params) ? params : [params]
        const p = raw[0] as TooltipItem
        const row = sorted[p.dataIndex]
        if (!row) return ''
        // "United States (US)" when labelFor maps the raw name; plain name
        // (sector charts) when the label is the row name itself.
        const shown = displayName(row.name)
        const title = shown === row.name ? shown : `${shown} (${row.name})`
        return `${p.marker}${title}<br/>${t('chartView.colValue')}: <b>${formatCurrency(row.value, currency)}</b><br/>${t('chartView.colWeight')}: <b>${formatPercent(row.weight)}</b>`
      },
    },
    grid: { left: 8, right: 60, top: 8, bottom: 8, containLabel: true },
    // Amounts are readable from the labels/tooltip, so the scale itself stays
    // hidden: the bars only need to be comparable to each other.
    xAxis: { type: 'value', show: false },
    yAxis: {
      type: 'category',
      inverse: true,
      data: sorted.map((r) => displayName(r.name)),
      // Mapped labels (full country names) need more room than raw codes;
      // charts without `labelFor` keep the previous compact axis.
      axisLabel: { fontSize: 11, width: labelFor ? 140 : 110, overflow: 'truncate' },
      axisTick: { show: false },
      axisLine: { show: false },
    },
    series: [
      {
        name: label ?? t('chartView.seriesExposure'),
        type: 'bar',
        barMaxWidth: 18,
        itemStyle: { borderRadius: [0, 4, 4, 0] },
        // Weight % printed at the end of each bar; the exact amount lives in
        // the tooltip.
        label: {
          show: true,
          position: 'right',
          fontSize: 11,
          color: labelColor,
          formatter: (params: unknown) => {
            const p = params as BarLabelItem
            const row = sorted[p.dataIndex]
            return row ? formatPercent(row.weight) : ''
          },
        },
        data: sorted.map((r, i) => ({
          value: Number(r.value),
          itemStyle: { color: colorFor?.(r.name) ?? palette[i % palette.length] },
        })),
      },
    ],
  }))
</script>

{#if label || (showTableToggle && sorted.length > 0)}
  <div
    class={cx(
      'mb-1 flex flex-wrap items-center gap-2',
      label ? 'justify-between' : 'justify-end',
    )}
  >
    {#if label}
      <h3 class="font-semibold">{label}</h3>
    {/if}
    {#if showTableToggle && sorted.length > 0}
      <ChartTableToggle name={label} bind:view />
    {/if}
  </div>
{/if}
{#if note}
  <p class="mb-3 text-xs text-muted-foreground">{note}</p>
{/if}
{#snippet canvas()}
  <div class="w-full" style="height: {height}px">
    <!-- {#key} re-inits the chart when the theme flips so the ECharts theme
         object passed below is picked up (svelte-echarts only reads `theme`
         at init time). -->
    {#key resolved()}
      <Chart {init} {options} theme={VAULTLAB_CHART_THEMES[resolved()]} />
    {/key}
  </div>
{/snippet}
{#if sorted.length === 0}
  <div class="flex items-center justify-center text-sm text-muted-foreground" style="height: {height}px">
    {t('chartView.noData')}
  </div>
{:else if showTable}
  <!-- No wrapper scroll container (EPIC K bug-fix): the page is the only
       scroll container, so no nested scrollbar appears. The table is
       `w-full` with wrapping names, and below `sm` rows collapse into a
       stacked key–value grid (name spanning the full width, value and weight
       sharing the second line) — spec §5.3 "collapse, don't shrink". -->
  <Table class="max-sm:block table-fixed">
    <caption class="sr-only">{caption}</caption>
    <THead class="max-sm:block">
      <Tr class="max-sm:grid max-sm:grid-cols-2 max-sm:gap-x-4 max-sm:py-2">
        <Th class="max-sm:col-span-2 max-sm:py-0.5 break-words">{t('chartView.colName')}</Th>
        <Th align="right" class="max-sm:py-0.5 max-sm:text-left whitespace-nowrap">{t('chartView.colValue')}</Th>
        <Th align="right" class="max-sm:py-0.5 whitespace-nowrap">{t('chartView.colWeight')}</Th>
      </Tr>
    </THead>
    <TBody class="max-sm:block">
      {#each sorted as r (r.name)}
        {@const shown = displayName(r.name)}
        <Tr class="max-sm:grid max-sm:grid-cols-2 max-sm:gap-x-4 max-sm:py-2">
          <!-- Same label as the tooltip: friendly name plus the raw name
               in parentheses when `labelFor` maps it (e.g. "US"). -->
          <Td class="max-sm:col-span-2 max-sm:py-0.5 font-medium break-words">
            {shown === r.name ? shown : `${shown} (${r.name})`}
          </Td>
          <Td align="right" class="max-sm:min-w-0 max-sm:py-0.5 max-sm:text-left whitespace-nowrap">
            {formatCurrency(r.value, currency)}
          </Td>
          <Td align="right" class="max-sm:py-0.5 whitespace-nowrap">{formatPercent(r.weight)}</Td>
        </Tr>
      {/each}
    </TBody>
  </Table>
{:else if scrollCap != null}
  <!-- Capped view (e.g. the country bars): full-height canvas with all rows,
       scrolled vertically inside a maxVisibleRows-tall viewport. -->
  <div class="w-full overflow-y-auto" style="max-height: {scrollCap}px">
    {@render canvas()}
  </div>
{:else}
  {@render canvas()}
{/if}
