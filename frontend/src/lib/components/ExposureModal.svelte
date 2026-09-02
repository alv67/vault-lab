<script lang="ts">
  import { Loader2, X } from 'lucide-svelte'
  import type { ExposureRow } from '$lib/services/api'
  import ExposurePie from './ExposurePie.svelte'

  let {
    open = $bindable(false),
    onClose,
    regionsEdit = $bindable([] as ExposureRow[]),
    sectorsEdit = $bindable([] as ExposureRow[]),
    sumRegions = 0,
    sumSectors = 0,
    regionsValid = false,
    sectorsValid = false,
    savingRegions = false,
    savingSectors = false,
    saveRegions,
    saveSectors,
    prefilling = false,
    fetchingETF = false,
    prefillExposure,
    fetchETFExposure,
    assetType = 'stock',
  }: {
    open: boolean
    onClose: () => void
    regionsEdit: ExposureRow[]
    sectorsEdit: ExposureRow[]
    sumRegions: number
    sumSectors: number
    regionsValid: boolean
    sectorsValid: boolean
    savingRegions: boolean
    savingSectors: boolean
    saveRegions: () => void
    saveSectors: () => void
    prefilling: boolean
    fetchingETF: boolean
    prefillExposure: () => void
    fetchETFExposure: () => void
    assetType: string
  } = $props()

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
    aria-label="Modifica distribuzione"
    tabindex="-1"
  >
    <div
      class="relative mx-4 max-h-[90vh] w-full max-w-4xl overflow-y-auto rounded-xl bg-white p-6 shadow-xl"
    >
      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-lg font-semibold">Modifica distribuzione</h2>
        <div class="flex items-center gap-2">
          <button
            onclick={fetchETFExposure}
            disabled={fetchingETF || assetType !== 'etf'}
            class="flex items-center gap-2 rounded-lg bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {#if fetchingETF}
              <Loader2 class="h-4 w-4 animate-spin" />
            {/if}
            Carica da JustETF
          </button>
          <button
            onclick={prefillExposure}
            disabled={prefilling}
            class="flex items-center gap-2 rounded-lg bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {#if prefilling}
              <Loader2 class="h-4 w-4 animate-spin" />
            {/if}
            Prefill da Yahoo
          </button>
          <button
            onclick={onClose}
            class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            aria-label="Chiudi"
          >
            <X class="h-5 w-5" />
          </button>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <!-- Regions -->
        <div>
          <h3 class="mb-3 font-medium">Distribuzione geografica</h3>
          <div class="flex flex-col gap-4 md:flex-row">
            <div class="flex-1 overflow-x-auto">
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
                      <td class="py-2">{r.name}</td>
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
              <button
                onclick={saveRegions}
                disabled={!regionsValid || savingRegions}
                class="mt-3 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {savingRegions ? 'Salvataggio...' : 'Salva'}
              </button>
            </div>
            <div class="w-full md:w-1/3">
              <ExposurePie data={regionsEdit} title="Distribuzione geografica" />
            </div>
          </div>
        </div>

        <!-- Sectors -->
        <div>
          <h3 class="mb-3 font-medium">Distribuzione settoriale</h3>
          <div class="flex flex-col gap-4 md:flex-row">
            <div class="flex-1 overflow-x-auto">
              <table class="w-full text-left text-sm">
                <thead>
                  <tr class="border-b text-gray-500">
                    <th class="pb-2">Settore GICS</th>
                    <th class="pb-2 text-right">Peso %</th>
                  </tr>
                </thead>
                <tbody>
                  {#each sectorsEdit as s (s.name)}
                    <tr class="border-b last:border-0">
                      <td class="py-2">{s.name}</td>
                      <td class="py-2 text-right">
                        <input
                          type="number"
                          min="0"
                          max="100"
                          step="0.1"
                          value={s.weight}
                          oninput={(e) => (s.weight = e.currentTarget.value)}
                          class="w-24 rounded-lg border px-3 py-1.5 text-right text-sm"
                        />
                      </td>
                    </tr>
                  {/each}
                  <tr class="border-t font-semibold">
                    <td class="py-2">Totale</td>
                    <td class="py-2 text-right {sectorsValid ? 'text-green-600' : 'text-red-600'}">
                      {sumSectors.toFixed(2)}%
                    </td>
                  </tr>
                </tbody>
              </table>
              {#if !sectorsValid}
                <p class="mt-2 text-sm text-red-600">
                  La somma dei pesi deve essere 100 (±0.5) — attuale: {sumSectors.toFixed(2)}%
                </p>
              {/if}
              <button
                onclick={saveSectors}
                disabled={!sectorsValid || savingSectors}
                class="mt-3 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {savingSectors ? 'Salvataggio...' : 'Salva'}
              </button>
            </div>
            <div class="w-full md:w-1/3">
              <ExposurePie data={sectorsEdit} title="Distribuzione settoriale" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
