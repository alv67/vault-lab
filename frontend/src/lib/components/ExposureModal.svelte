<script lang="ts">
  import { Loader2, X } from 'lucide-svelte'
  import type { ExposureRow } from '$lib/services/api'

  const PALETTE = [
    '#2563eb', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
    '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16',
    '#06b6d4', '#a855f7',
  ]

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
    prefillRegionsFromETF,
    prefillSectorsFromETF,
    prefillSectorsFromYahoo,
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
    prefillRegionsFromETF: () => void
    prefillSectorsFromETF: () => void
    prefillSectorsFromYahoo: () => void
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
        <div>
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="font-medium">Distribuzione geografica</h3>
            <button
              onclick={prefillRegionsFromETF}
              disabled={fetchingETF || assetType !== 'etf'}
              title="Prefill da JustETF"
              aria-label="Prefill geografico da JustETF"
              class="rounded-lg p-1.5 hover:bg-gray-100 disabled:opacity-40"
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
          </div>
          <div class="flex-1">
            <table class="w-full text-left text-sm">
              <thead>
                <tr class="border-b text-gray-500">
                  <th class="pb-2">Area geografica</th>
                  <th class="pb-2 text-right">Peso %</th>
                </tr>
              </thead>
              <tbody>
                {#each regionsEdit as r, i (r.name)}
                  <tr class="border-b last:border-0">
                    <td class="py-2">
                      <span class="flex items-center gap-2">
                        <span
                          class="inline-block h-3 w-3 shrink-0 rounded"
                          style="background-color: {PALETTE[i % PALETTE.length]};"
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
            <button
              onclick={saveRegions}
              disabled={!regionsValid || savingRegions}
              class="mt-3 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {savingRegions ? 'Salvataggio...' : 'Salva'}
            </button>
          </div>
        </div>

        <!-- Sectors -->
        <div>
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="font-medium">Distribuzione settoriale</h3>
            <div class="flex items-center gap-1.5">
              <button
                onclick={prefillSectorsFromETF}
                disabled={fetchingETF || assetType !== 'etf'}
                title="Prefill da JustETF"
                aria-label="Prefill settori da JustETF"
                class="rounded-lg p-1.5 hover:bg-gray-100 disabled:opacity-40"
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
                onclick={prefillSectorsFromYahoo}
                disabled={prefilling}
                title="Prefill da Yahoo"
                aria-label="Prefill settori da Yahoo"
                class="rounded-lg p-1.5 hover:bg-gray-100 disabled:opacity-40"
              >
                {#if prefilling}
                  <Loader2 class="h-5 w-5 animate-spin text-gray-500" />
                {:else}
                  <img
                    class="h-5 w-5 rounded"
                    src="https://www.google.com/s2/favicons?domain=finance.yahoo.com&sz=32"
                    alt="Yahoo"
                  />
                {/if}
              </button>
            </div>
          </div>
          <div class="flex-1">
            <table class="w-full text-left text-sm">
              <thead>
                <tr class="border-b text-gray-500">
                  <th class="pb-2">Settore GICS</th>
                  <th class="pb-2 text-right">Peso %</th>
                </tr>
              </thead>
              <tbody>
                {#each sectorsEdit as s, i (s.name)}
                  <tr class="border-b last:border-0">
                    <td class="py-2">
                      <span class="flex items-center gap-2">
                        <span
                          class="inline-block h-3 w-3 shrink-0 rounded"
                          style="background-color: {PALETTE[i % PALETTE.length]};"
                        ></span>
                        {s.name}
                      </span>
                    </td>
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
        </div>
      </div>
    </div>
  </div>
{/if}
