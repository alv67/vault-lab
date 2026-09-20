<script lang="ts">
  import ExposureBarChart, { type ExposureBarRow } from '$lib/components/domain/ExposureBarChart.svelte'
  import ClassDonut from '$lib/components/domain/ClassDonut.svelte'
  import { countryDisplayName } from '$lib/countryNames'
  import { chartSemanticColors } from '$lib/chartPalette'
  import { resolved } from '$lib/stores/theme.svelte'
  import { getPortfolioPage } from '../context'

  /**
   * Allocation tab (EPIC K.4a): the EPIC I.7 (#86) section promoted to its
   * own deep-linkable route — the asset-class donut plus the equity-only
   * region, sector and country bars, mirroring the dashboard's
   * "Allocazione complessiva" card, all in the portfolio currency. The
   * three payloads (and their per-endpoint isolated error flags) come from
   * the shell context; a failed allocation call still only shows its own
   * "non disponibile" panel without blocking the rest of the tab.
   */
  const ctx = getPortfolioPage()
  const currency = $derived(ctx.currency)

  // EPIC I.7 (#86): the region/sector/country payloads are mapped onto the
  // generic ExposureBarRow shape consumed by ExposureBarChart (countries
  // keep their raw ISO codes as row identity; the chart maps them to full
  // names via `labelFor`, see the country panel below).
  const regionBarRows = $derived<ExposureBarRow[]>(
    (ctx.geoAlloc?.regions ?? []).map((r) => ({ name: r.region, value: r.value, weight: r.weight })),
  )
  const sectorBarRows = $derived<ExposureBarRow[]>(
    (ctx.sectorAlloc?.sectors ?? []).map((s) => ({ name: s.sector, value: s.value, weight: s.weight })),
  )
  const countryBarRows = $derived<ExposureBarRow[]>(
    (ctx.geoAlloc?.countries ?? []).map((c) => ({ name: c.country, value: c.value, weight: c.weight })),
  )

  // Equity-universe coverage note for the bar panels, like the dashboard:
  // shown only when non-equity holdings were actually excluded. Geography
  // (regions + countries) and sectors come from two isolated endpoints, so
  // each payload carries its own covered/excluded note.
  function equityUniverseNote(covered?: string, excluded?: string): string | undefined {
    const coveredNum = Number(covered || 0)
    const excludedNum = Number(excluded || 0)
    const total = coveredNum + excludedNum
    if (total <= 0 || excludedNum <= 0) return undefined
    return `Universo azionario: ${((coveredNum / total) * 100).toFixed(1)}% del portafoglio`
  }
  const geoUniverseNote = $derived(
    equityUniverseNote(ctx.geoAlloc?.covered_value, ctx.geoAlloc?.excluded_value),
  )
  const sectorUniverseNote = $derived(
    equityUniverseNote(ctx.sectorAlloc?.covered_value, ctx.sectorAlloc?.excluded_value),
  )

  // Aggregated "Other" buckets are muted grey, like the slice treatment in
  // the donut charts (same helper as the dashboard card). Read via
  // chartSemanticColors so it re-evaluates on theme flips (the {#key} blocks
  // inside the charts re-init them anyway).
  function otherGrey(name: string): string | undefined {
    return name === 'Other' || name === 'Other / Not Classified'
      ? chartSemanticColors(resolved()).other
      : undefined
  }
</script>

<div class="grid gap-4 lg:grid-cols-2">
  <div class="rounded-card border-border bg-surface p-4 shadow-card">
    {#if ctx.classAllocError}
      <h3 class="mb-3 font-semibold">Classi di attività</h3>
      <p class="text-sm text-muted-foreground">Allocazione per classi non disponibile</p>
    {:else}
      <ClassDonut
        data={ctx.classAlloc?.classes ?? []}
        currency={ctx.classAlloc?.currency || currency}
        label="Classi di attività"
      />
    {/if}
  </div>
  <div class="rounded-card border-border bg-surface p-4 shadow-card">
    {#if ctx.sectorAllocError}
      <h3 class="mb-1 font-semibold">Settori (solo equity)</h3>
      <p class="text-sm text-muted-foreground">Allocazione settoriale non disponibile</p>
    {:else}
      <ExposureBarChart
        rows={sectorBarRows}
        currency={ctx.sectorAlloc?.currency || currency}
        label="Settori (solo equity)"
        note={sectorUniverseNote}
        colorFor={otherGrey}
      />
    {/if}
  </div>
  <div class="rounded-card border-border bg-surface p-4 shadow-card">
    {#if ctx.geoAllocError}
      <h3 class="mb-1 font-semibold">Regioni (solo equity)</h3>
      <p class="text-sm text-muted-foreground">Allocazione geografica non disponibile</p>
    {:else}
      <ExposureBarChart
        rows={regionBarRows}
        currency={ctx.geoAlloc?.currency || currency}
        label="Regioni (solo equity)"
        note={geoUniverseNote}
        colorFor={otherGrey}
      />
    {/if}
  </div>
  <div class="rounded-card border-border bg-surface p-4 shadow-card">
    {#if ctx.geoAllocError}
      <h3 class="mb-1 font-semibold">Paesi (solo equity)</h3>
      <p class="text-sm text-muted-foreground">Allocazione geografica non disponibile</p>
    {:else}
      <ExposureBarChart
        rows={countryBarRows}
        currency={ctx.geoAlloc?.currency || currency}
        label="Paesi (solo equity)"
        note={geoUniverseNote}
        colorFor={otherGrey}
        labelFor={countryDisplayName}
        maxVisibleRows={10}
      />
    {/if}
  </div>
</div>
