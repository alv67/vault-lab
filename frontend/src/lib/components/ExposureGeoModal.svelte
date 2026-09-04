<script lang="ts">
  import { Calculator, Loader2, X, Plus, Trash2 } from 'lucide-svelte'
  import type { ExposureRow } from '$lib/services/api'
  import { CANONICAL_COUNTRIES, countryDisplayName } from '$lib/countryNames'
  import ExposurePie from './ExposurePie.svelte'
  import { colorForRow } from '$lib/chartPalette'

  let {
    open = $bindable(false),
    onClose,
    regionsEdit = $bindable([] as ExposureRow[]),
    countriesEdit = $bindable([] as ExposureRow[]),
    sumRegions = 0,
    sumCountries = 0,
    regionsValid = false,
    countriesValid = false,
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
    assetType = 'stock',
  }: {
    open: boolean
    onClose: () => void
    regionsEdit: ExposureRow[]
    countriesEdit: ExposureRow[]
    sumRegions: number
    sumCountries: number
    regionsValid: boolean
    countriesValid: boolean
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
    assetType: string
  } = $props()

  /** Country codes currently present in the countries edit list. */
  const presentCodes = $derived(new Set(countriesEdit.map((r) => r.name)))

  /** Canonical codes not yet in the edit list — available for the add dropdown. */
  const availableCodes = $derived(CANONICAL_COUNTRIES.filter((c) => !presentCodes.has(c)))

  let addCountryCode = $state('')

  // Keep the dropdown selection valid: if the current choice is already present
  // (e.g. after a Morningstar prefill), reset to the first available code.
  $effect(() => {
    const codes = availableCodes
    if (codes.length > 0 && !codes.includes(addCountryCode)) {
      addCountryCode = codes[0]
    }
  })

  function removeCountry(code: string): void {
    countriesEdit = countriesEdit.filter((r) => r.name !== code)
  }

  function addCountry(): void {
    if (!addCountryCode || presentCodes.has(addCountryCode)) return
    countriesEdit = [...countriesEdit, { name: addCountryCode, weight: '0' }]
    // Reset selection to next available
    const next = CANONICAL_COUNTRIES.find((c) => !countriesEdit.some((r) => r.name === c))
    addCountryCode = next ?? ''
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
    <div
      class="relative mx-4 max-h-[90vh] w-full max-w-5xl overflow-y-auto rounded-xl bg-white p-6 shadow-xl"
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
        <!-- Regions -->
        <div class="flex flex-col rounded-xl border border-gray-200 bg-gray-50 p-4">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="font-medium">Distribuzione geografica</h3>
            <div class="flex items-center gap-1.5">
              <button
                onclick={deriveRegionsFromCountries}
                disabled={derivingRegions}
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
                disabled={fetchingMorningstar || assetType !== 'etf'}
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
          <div class="flex flex-1 flex-col gap-4 md:flex-row">
            <div class="flex-1">
            <table class="w-full text-left text-sm">
              <thead>
                <tr class="border-b text-gray-500">
                  <th class="pb-2">Area geografica</th>
                  <th class="pb-2 text-right">Peso %</th>
                </tr>
              </thead>
              <tbody>
                {#each regionsEdit as r (r.name)}
                  <tr class="border-b last:border-0">
                    <td class="py-2">
                      <span class="flex items-center gap-2">
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
                        value={r.weight}
                        oninput={(e) => (r.weight = e.currentTarget.value)}
                        class="w-24 rounded-lg border px-3 py-1.5 text-right text-sm"
                      />
                    </td>
                  </tr>
                {/each}
                <tr class="border-t font-semibold">
                  <td class="py-2">Totale</td>
                  <td class="py-2 text-right {regionsValid ? 'text-green-600' : 'text-red-600'}">
                    {sumRegions.toFixed(2)}%
                  </td>
                </tr>
              </tbody>
            </table>
            {#if !regionsValid}
              <p class="mt-2 text-sm text-red-600">
                La somma dei pesi deve essere 100 (±0.5) — attuale: {sumRegions.toFixed(2)}%
              </p>
            {/if}
          </div>
          <div class="w-48 shrink-0 md:w-56">
            <ExposurePie data={regionsEdit} title="Distribuzione geografica" mute />
          </div>
        </div>
        <div class="mt-auto flex justify-end pt-4">
          <button
            onclick={saveRegions}
            disabled={!regionsValid || savingRegions}
            class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {savingRegions ? 'Salvataggio...' : 'Salva'}
          </button>
        </div>
      </div>

        <!-- Countries -->
        <div class="flex flex-col rounded-xl border border-gray-200 bg-gray-50 p-4">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="font-medium">Distribuzione paesi</h3>
            <div class="flex items-center gap-1.5">
              <button
                onclick={prefillCountriesFromETF}
                disabled={fetchingETF || assetType !== 'etf'}
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
                disabled={fetchingMorningstar || assetType !== 'etf'}
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
          <div class="flex flex-1 flex-col gap-4 md:flex-row">
            <div class="flex-1">
              <div class="max-h-[400px] overflow-y-auto">
                <table class="w-full text-left text-sm">
                  <thead class="sticky top-0 bg-gray-50">
                    <tr class="border-b text-gray-500">
                      <th class="pb-2">Paese</th>
                      <th class="pb-2 text-right">Peso %</th>
                      <th class="w-10"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each countriesEdit as r (r.name)}
                      <tr class="border-b last:border-0">
                        <td class="py-2">
                          <span class="flex items-center gap-2">
                            <span
                              class="inline-block h-3 w-3 shrink-0 rounded"
                              style="background-color: {colorForRow(r, countriesEdit)};"
                            ></span>
                            <span class="text-xs font-medium text-gray-400">{r.name}</span>
                            {countryDisplayName(r.name)}
                          </span>
                        </td>
                        <td class="py-2 text-right">
                          <input
                            type="number"
                            min="0"
                            max="100"
                            step="0.1"
                            value={r.weight}
                            oninput={(e) => (r.weight = e.currentTarget.value)}
                            class="w-20 rounded-lg border px-3 py-1.5 text-right text-sm"
                          />
                        </td>
                        <td class="py-2 text-right">
                          <button
                            onclick={() => removeCountry(r.name)}
                            class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-500"
                            title="Rimuovi {countryDisplayName(r.name)}"
                            aria-label="Rimuovi {countryDisplayName(r.name)}"
                          >
                            <Trash2 class="h-3.5 w-3.5" />
                          </button>
                        </td>
                      </tr>
                    {/each}
                    <tr class="border-t font-semibold">
                      <td class="py-2">Totale</td>
                      <td class="py-2 text-right {countriesValid ? 'text-green-600' : 'text-amber-600'}">
                        {sumCountries.toFixed(2)}%
                      </td>
                      <td></td>
                    </tr>
                  </tbody>
                </table>
              </div>
              {#if !countriesValid}
                <p class="mt-2 text-sm text-amber-600">
                  Somma dei paesi: {sumCountries.toFixed(2)}% — può non raggiungere 100
                  (la quota non attribuita a un paese confluisce nella regione
                  «Other / Not Classified»).
                </p>
              {/if}
              {#if availableCodes.length > 0}
                <div class="mt-3 flex items-center gap-2">
                  <select
                    bind:value={addCountryCode}
                    class="flex-1 rounded-lg border px-3 py-1.5 text-sm"
                  >
                    {#each availableCodes as code (code)}
                      <option value={code}>{code} — {countryDisplayName(code)}</option>
                    {/each}
                  </select>
                  <button
                    onclick={addCountry}
                    disabled={!addCountryCode}
                    class="flex shrink-0 items-center gap-1 rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
                  >
                    <Plus class="h-4 w-4" />
                    Aggiungi
                  </button>
                </div>
              {/if}
            </div>
            <div class="w-48 shrink-0 md:w-56">
              <ExposurePie data={countriesEdit} title="Distribuzione paesi" mute />
            </div>
          </div>
          <div class="mt-auto flex justify-end pt-4">
            <button
              onclick={saveCountries}
              disabled={savingCountries}
              class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {savingCountries ? 'Salvataggio...' : 'Salva'}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
