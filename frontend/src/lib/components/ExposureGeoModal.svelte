<script lang="ts">
  import { tick, untrack } from 'svelte'
  import { Calculator, Globe2, Info, Loader2, X, Plus, Trash2 } from 'lucide-svelte'
  import type { ExposureRow } from '$lib/services/api'
  import { CANONICAL_COUNTRIES, countryDisplayName } from '$lib/countryNames'
  import ExposurePie from './ExposurePie.svelte'
  import ProvenanceBadge from './ProvenanceBadge.svelte'
  import { CHART_PALETTE, colorForRow } from '$lib/chartPalette'

  let {
    open = $bindable(false),
    onClose,
    regionsEdit = $bindable([] as ExposureRow[]),
    countriesEdit = $bindable([] as ExposureRow[]),
    sumRegions = 0,
    sumCountries = 0,
    savingRegions = false,
    savingCountries = false,
    saveRegions,
    saveCountries,
    fetchingETF = false,
    fetchingMorningstar = false,
    derivingRegions = false,
    prefillCountriesFromETF,
    deriveRegionsFromCountries,
    prefillRegionsFromMorningstar,
    prefillCountriesFromMorningstar,
    countriesSource = null as string | null,
    regionsSource = null as string | null,
    onCountriesDirty,
    onRegionsDirty,
    assetType = 'stock',
  }: {
    open: boolean
    onClose: () => void
    regionsEdit: ExposureRow[]
    countriesEdit: ExposureRow[]
    sumRegions: number
    sumCountries: number
    savingRegions: boolean
    savingCountries: boolean
    saveRegions: () => void
    saveCountries: () => void
    fetchingETF: boolean
    fetchingMorningstar: boolean
    derivingRegions: boolean
    prefillCountriesFromETF: () => void
    deriveRegionsFromCountries: () => void
    prefillRegionsFromMorningstar: () => void
    prefillCountriesFromMorningstar: () => void
    countriesSource: string | null
    regionsSource: string | null
    onCountriesDirty: () => void
    onRegionsDirty: () => void
    assetType: string
  } = $props()

  /** Country codes currently present in the countries edit list. */
  const presentCodes = $derived(new Set(countriesEdit.map((r) => r.name)))

  /** Canonical codes not yet in the edit list — available for the add
   *  dropdown, ordered by friendly country name (not ISO code). */
  const availableCodes = $derived(
    CANONICAL_COUNTRIES.filter((c) => !presentCodes.has(c)).sort((a, b) =>
      countryDisplayName(a).localeCompare(countryDisplayName(b), 'it'),
    ),
  )

  let addCountryCode = $state('')

  // Keep the dropdown selection valid: if the current choice is already present
  // (e.g. after a Morningstar prefill), reset to the first available code.
  $effect(() => {
    const codes = availableCodes
    if (codes.length > 0 && !codes.includes(addCountryCode)) {
      addCountryCode = codes[0]
    }
  })

  // ---------------------------------------------------------------------------
  // Countries sorting
  //
  // The rows are the page-owned edit list (already filtered to weight > 0 at
  // assignment time; the only zero-weight rows are ones the user just added).
  // Display order is a snapshot of codes, re-sorted ONLY on list replacement
  // (load/prefill/save), add, remove and on input commit (change/blur) — never
  // while typing, so rows don't jump under the cursor; the animated bar gives
  // the live feedback instead.
  // ---------------------------------------------------------------------------
  let orderedCodes = $state<string[]>([])

  function resort(): void {
    orderedCodes = [...countriesEdit]
      .sort((a, b) => (Number(b.weight) || 0) - (Number(a.weight) || 0))
      .map((r) => r.name)
  }

  $effect(() => {
    // Reading each row's name (never its weight) tracks list replacements
    // without reacting to in-place weight edits.
    const names = countriesEdit.map((r) => r.name)
    untrack(() => {
      if (names.length === 0) orderedCodes = []
      else resort()
    })
  })

  const visibleCountries = $derived.by(() =>
    orderedCodes
      .map((code) => countriesEdit.find((r) => r.name === code))
      .filter((r): r is ExposureRow => r !== undefined),
  )

  // Largest visible weight: bars are scaled proportionally against it, same
  // formula as the page card bar list.
  const maxCountryWeight = $derived(
    visibleCountries.reduce((max, r) => Math.max(max, Number(r.weight) || 0), 0),
  )

  // Save/validation rule for both dimensions: sums up to 100 are valid (the
  // backend absorbs the residual), sums above 100 (± float epsilon) block the
  // save with an inline alert.
  const countriesOver = $derived(sumCountries > 100 + 1e-9)
  const regionsOver = $derived(sumRegions > 100 + 1e-9)

  function totalColorClass(sum: number, over: boolean): string {
    if (over) return 'text-red-600'
    if (sum >= 99.5) return 'text-green-600'
    return 'text-gray-900'
  }

  function removeCountry(code: string): void {
    countriesEdit = countriesEdit.filter((r) => r.name !== code)
    onCountriesDirty()
    resort()
  }

  async function addCountry(): Promise<void> {
    if (!addCountryCode || presentCodes.has(addCountryCode)) return
    const code = addCountryCode
    countriesEdit = [...countriesEdit, { name: code, weight: '0' }]
    onCountriesDirty()
    resort()
    // Reset selection to the first country by name (availableCodes is
    // name-ordered, so the dropdown always proposes the alphabetically
    // first not-yet-added country).
    const next = availableCodes[0] ?? ''
    addCountryCode = next
    await tick()
    // Land the user straight on the new row's weight input.
    const input = document.querySelector<HTMLInputElement>(
      `[data-code="${code}"] input[type='number']`,
    )
    input?.focus()
    input?.select()
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (e.key === 'Escape' && open) {
      onClose()
    }
  }

  function handleBackdropClick(e: MouseEvent): void {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    onclick={handleBackdropClick}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
    role="dialog"
    aria-modal="true"
    aria-label="Modifica distribuzione geografica"
    tabindex="-1"
  >
    <!-- max-w-6xl (1152px): with the lg:grid-cols-2 split each box gets ~536px
         of width, so the regions table (~264px after the 224px donut) has room
         for long region names like "Africa / Middle East" on one line. -->
    <div
      class="relative mx-4 max-h-[90vh] w-full max-w-6xl overflow-y-auto rounded-xl bg-white p-6 shadow-xl"
    >
      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-lg font-semibold">Modifica distribuzione geografica</h2>
        <button
          onclick={onClose}
          class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
          aria-label="Chiudi"
        >
          <X class="h-5 w-5" />
        </button>
      </div>

      <div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
        <!-- Countries (regions update only manually, via "Calcola da paesi") -->
        <div class="flex flex-col rounded-xl border border-gray-200 bg-gray-50 p-4">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <h3 class="font-medium">Paesi</h3>
              <ProvenanceBadge source={countriesSource} />
            </div>
            <div class="flex items-center gap-1.5">
              <button
                onclick={prefillCountriesFromETF}
                disabled={assetType !== 'etf' || fetchingETF || fetchingMorningstar}
                title="Prefill da JustETF"
                aria-label="Prefill paesi da JustETF"
                class="rounded-lg border border-gray-300 bg-white p-1.5 shadow-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {#if fetchingETF}
                  <Loader2 class="h-5 w-5 animate-spin text-gray-500" />
                {:else}
                  <img
                    class="h-5 w-5 rounded"
                    src="https://www.google.com/s2/favicons?domain=justetf.com&sz=32"
                    alt="JustETF"
                  />
                {/if}
              </button>
              <button
                onclick={prefillCountriesFromMorningstar}
                disabled={assetType !== 'etf' || fetchingETF || fetchingMorningstar}
                title="Prefill da Morningstar"
                aria-label="Prefill paesi da Morningstar"
                class="rounded-lg border border-gray-300 bg-white p-1.5 shadow-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {#if fetchingMorningstar}
                  <Loader2 class="h-5 w-5 animate-spin text-gray-500" />
                {:else}
                  <img
                    class="h-5 w-5 rounded"
                    src="https://www.google.com/s2/favicons?domain=www.morningstar.com&sz=32"
                    alt="Morningstar"
                  />
                {/if}
              </button>
            </div>
          </div>

          <!-- Fixed-height middle area: EXACTLY the same quota as the regions
               box (h-[38rem] fits the header row + 10 region rows + side
               donut), so both boxes and their footers (hr / total / messages
               / Save) stay aligned. The taller quota keeps the regions table
               fully visible without a scrollbar; the countries list scrolls
               INSIDE this quota only as a last resort (min-h-0 flex-1 on the
               ul): few countries leave empty space below, many never widen or
               lengthen the box. -->
          <div class="flex h-[38rem] flex-col">
            <!-- Table-style header row, columns aligned with the rows below. -->
            <div
              class="flex items-center gap-3 border-b border-gray-300 pb-2 pl-2 pr-3 text-sm text-gray-500"
            >
              <span class="min-w-0 flex-1 truncate">Paese</span>
              <span class="w-20 shrink-0 text-right">Peso %</span>
              <!-- Spacer matching the remove-button column of the rows. -->
              <span class="w-[22px] shrink-0" aria-hidden="true"></span>
            </div>

            {#if visibleCountries.length === 0}
              <div class="flex min-h-0 flex-1 flex-col items-center justify-center py-10 text-center">
                <Globe2 class="h-8 w-8 text-gray-300" />
                <p class="text-sm text-gray-500">Nessun paese inserito</p>
                <p class="text-xs text-gray-500">
                  Aggiungi un paese qui sotto, oppure usa un prefill JustETF / Morningstar
                </p>
              </div>
            {:else}
              <ul
                class="min-h-0 flex-1 divide-y divide-gray-200 overflow-y-auto py-1 pr-1"
              >
                {#each visibleCountries as row, i (row.name)}
                  {@const weight = Number(row.weight) || 0}
                  {@const displayName = countryDisplayName(row.name)}
                  {@const barPct =
                    maxCountryWeight > 0 ? ((weight / maxCountryWeight) * 100).toFixed(1) : '0'}
                  <li
                    data-code={row.name}
                    class="group flex items-center gap-3 rounded-lg px-2 py-2 transition-colors hover:bg-gray-100 motion-reduce:transition-none"
                  >
                    <span class="w-7 shrink-0 text-xs font-medium text-gray-500">{row.name}</span>
                    <span
                      class="w-32 shrink-0 truncate text-sm text-gray-700 sm:w-40"
                      title="{row.name} — {displayName}"
                    >{displayName}</span>
                    <div
                      class="h-2.5 min-w-0 flex-1 overflow-hidden rounded-full bg-gray-200"
                      aria-hidden="true"
                    >
                      <div
                        class="h-full rounded-full transition-[width] duration-200 motion-reduce:transition-none"
                        style="width: {barPct}%; background-color: {CHART_PALETTE[i % CHART_PALETTE.length]};"
                      ></div>
                    </div>
                    <input
                      type="number"
                      min="0"
                      max="100"
                      step="0.1"
                      inputmode="decimal"
                      value={row.weight}
                      aria-label="Peso di {displayName}"
                      oninput={(e) => {
                        row.weight = e.currentTarget.value
                        onCountriesDirty()
                      }}
                      onchange={resort}
                      class="w-20 shrink-0 rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-right text-sm tabular-nums focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
                    />
                    <button
                      onclick={() => removeCountry(row.name)}
                      class="shrink-0 rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-500"
                      title="Rimuovi {displayName}"
                      aria-label="Rimuovi {displayName}"
                    >
                      <Trash2 class="h-3.5 w-3.5" />
                    </button>
                  </li>
                {/each}
              </ul>
            {/if}

            {#if availableCodes.length > 0}
              <div class="flex items-center gap-2 pt-3">
                <select
                  bind:value={addCountryCode}
                  aria-label="Paese da aggiungere"
                  class="min-w-0 flex-1 rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
                >
                  {#each availableCodes as code (code)}
                    <option value={code}>{code} — {countryDisplayName(code)}</option>
                  {/each}
                </select>
                <button
                  onclick={addCountry}
                  disabled={!addCountryCode}
                  class="flex shrink-0 items-center gap-1 rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
                >
                  <Plus class="h-4 w-4" />
                  Aggiungi
                </button>
              </div>
            {/if}
          </div>

          <!-- Footer rows (identical in both boxes so they line up):
               separator, fixed-height total, reserved 3-line message area,
               fixed-height Save area. -->
          <div class="mt-3 border-t border-gray-200" aria-hidden="true"></div>

          <div class="mt-3 h-8">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-700">Totale</span>
              <span
                class="text-sm font-semibold tabular-nums {totalColorClass(sumCountries, countriesOver)}"
              >{sumCountries.toFixed(2)}%</span>
            </div>
            <div class="mt-1.5 h-1.5 overflow-hidden rounded-full bg-gray-200" aria-hidden="true">
              <div
                class="h-full rounded-full transition-[width] duration-200 motion-reduce:transition-none"
                style="width: {Math.min(sumCountries, 100).toFixed(2)}%; background-color: {countriesOver
                  ? '#ef4444'
                  : '#2563eb'};"
              ></div>
            </div>
          </div>

          <div class="h-[3.75rem] overflow-hidden pt-1 text-xs leading-5">
            {#if countriesOver}
              <p role="alert" class="text-red-600">
                La somma supera il 100% — attuale {sumCountries.toFixed(2)}%.
                Riduci i pesi per salvare.
              </p>
            {:else if sumCountries < 99.5}
              <p class="text-gray-500">
                Residuo non attribuito: {(100 - sumCountries).toFixed(2)}%.
              </p>
            {/if}
          </div>

          <div class="flex h-12 items-center justify-end">
            <button
              onclick={saveCountries}
              disabled={savingCountries || countriesOver}
              title={countriesOver
                ? 'La somma dei pesi supera il 100%: riduci i pesi per poter salvare'
                : undefined}
              class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {#if savingCountries}
                <Loader2 class="h-4 w-4 animate-spin" />
              {/if}
              {savingCountries ? 'Salvataggio...' : 'Salva'}
            </button>
          </div>
        </div>

        <!-- Regions -->
        <div class="flex flex-col rounded-xl border border-gray-200 bg-gray-50 p-4">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <h3 class="font-medium">Regioni</h3>
              <ProvenanceBadge source={regionsSource} />
            </div>
            <div class="flex items-center gap-1.5">
              <button
                onclick={deriveRegionsFromCountries}
                disabled={derivingRegions || fetchingMorningstar}
                title="Calcola da paesi"
                aria-label="Calcola regioni dai paesi"
                class="rounded-lg border border-gray-300 bg-white p-1.5 shadow-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {#if derivingRegions}
                  <Loader2 class="h-5 w-5 animate-spin text-gray-500" />
                {:else}
                  <Calculator class="h-5 w-5 text-gray-500" />
                {/if}
              </button>
              <button
                onclick={prefillRegionsFromMorningstar}
                disabled={assetType !== 'etf' || derivingRegions || fetchingMorningstar}
                title="Prefill da Morningstar"
                aria-label="Prefill regioni da Morningstar"
                class="rounded-lg border border-gray-300 bg-white p-1.5 shadow-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                {#if fetchingMorningstar}
                  <Loader2 class="h-5 w-5 animate-spin text-gray-500" />
                {:else}
                  <img
                    class="h-5 w-5 rounded"
                    src="https://www.google.com/s2/favicons?domain=www.morningstar.com&sz=32"
                    alt="Morningstar"
                  />
                {/if}
              </button>
            </div>
          </div>

          <!-- Fixed-height middle area: same quota as the countries box
               (h-[38rem] fits the header row + 10 region rows and the side
               donut completely, no scrollbar). Stacked on small screens it
               scrolls within this quota instead of growing the box. -->
          <div class="flex h-[38rem] flex-col gap-4 overflow-y-auto md:flex-row">
            <div class="min-w-0 flex-1">
              <table class="w-full text-left text-sm">
                <thead>
                  <tr class="border-b text-gray-500">
                    <th class="pb-2">Area geografica</th>
                    <th class="pb-2 text-right">Peso %</th>
                  </tr>
                </thead>
                <tbody>
                  {#each regionsEdit as r (r.name)}
                    <tr class="border-b last:border-0 hover:bg-gray-100/70">
                      <td class="py-2">
                        <!-- nowrap only once the panel truly reaches max-w-6xl
                             (viewport ≥ 1184px = 1152 + 2×mx-4); below that the
                             two-column table is too narrow for nowrap without
                             overflowing into the donut, so wrapping is allowed. -->
                        <span class="flex items-center gap-2 min-[1184px]:whitespace-nowrap">
                          <span
                            class="inline-block h-3 w-3 shrink-0 rounded"
                            style="background-color: {colorForRow(r, regionsEdit)};"
                          ></span>
                          {r.name}
                        </span>
                      </td>
                      <td class="py-2 text-right">
                        <input
                          type="number"
                          min="0"
                          max="100"
                          step="0.1"
                          inputmode="decimal"
                          value={r.weight}
                          aria-label="Peso di {r.name}"
                          oninput={(e) => {
                            r.weight = e.currentTarget.value
                            onRegionsDirty()
                          }}
                          class="w-24 shrink-0 rounded-lg border border-gray-300 bg-white px-2.5 py-1.5 text-right text-sm tabular-nums focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
                        />
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
            <div class="w-48 shrink-0 md:w-56">
              <ExposurePie data={regionsEdit} title="Distribuzione geografica" mute complete={false} />
            </div>
          </div>

          <!-- Footer rows (identical in both boxes so they line up):
               separator, fixed-height total, reserved 3-line message area,
               fixed-height Save area. -->
          <div class="mt-3 border-t border-gray-200" aria-hidden="true"></div>

          <div class="mt-3 h-8">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-700">Totale</span>
              <span
                class="text-sm font-semibold tabular-nums {totalColorClass(sumRegions, regionsOver)}"
              >{sumRegions.toFixed(2)}%</span>
            </div>
            <div class="mt-1.5 h-1.5 overflow-hidden rounded-full bg-gray-200" aria-hidden="true">
              <div
                class="h-full rounded-full transition-[width] duration-200 motion-reduce:transition-none"
                style="width: {Math.min(sumRegions, 100).toFixed(2)}%; background-color: {regionsOver
                  ? '#ef4444'
                  : '#2563eb'};"
              ></div>
            </div>
          </div>

          <div class="h-[3.75rem] overflow-hidden pt-1 text-xs leading-5">
            {#if regionsOver}
              <p role="alert" class="text-red-600">
                La somma supera il 100% — attuale {sumRegions.toFixed(2)}%.
                Riduci i pesi per salvare.
              </p>
            {:else if sumRegions < 99.5}
              <p class="flex items-center gap-1.5 text-gray-500">
                <Info class="h-3.5 w-3.5 shrink-0" />
                Residuo non classificato: {(100 - sumRegions).toFixed(2)}% — escluso dal grafico.
              </p>
            {/if}
          </div>

          <div class="flex h-12 items-center justify-end">
            <button
              onclick={saveRegions}
              disabled={savingRegions || regionsOver}
              title={regionsOver
                ? 'La somma dei pesi supera il 100%: riduci i pesi per poter salvare'
                : undefined}
              class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {#if savingRegions}
                <Loader2 class="h-4 w-4 animate-spin" />
              {/if}
              {savingRegions ? 'Salvataggio...' : 'Salva'}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
