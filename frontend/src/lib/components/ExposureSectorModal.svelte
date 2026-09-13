<script lang="ts">
  import { Loader2, X } from 'lucide-svelte'
  import type { ExposureRow } from '$lib/services/api'
  import ExposurePie from './ExposurePie.svelte'
  import ProvenanceBadge from './ProvenanceBadge.svelte'
  import { colorForRow, resolvePalette } from '$lib/chartPalette'
  import { resolved } from '$lib/stores/theme.svelte'

  let {
    open = $bindable(false),
    onClose,
    sectorsEdit = $bindable([] as ExposureRow[]),
    sumSectors = 0,
    sectorsValid = false,
    savingSectors = false,
    saveSectors,
    prefilling = false,
    fetchingETF = false,
    fetchingMorningstar = false,
    prefillSectorsFromETF,
    prefillSectorsFromYahoo,
    prefillSectorsFromMorningstar,
    sectorsSource = null as string | null,
    sectorsUpdatedAt = null as string | null,
    onSectorsDirty,
    assetType = 'stock',
  }: {
    open: boolean
    onClose: () => void
    sectorsEdit: ExposureRow[]
    sumSectors: number
    sectorsValid: boolean
    savingSectors: boolean
    saveSectors: () => void
    prefilling: boolean
    fetchingETF: boolean
    fetchingMorningstar: boolean
    prefillSectorsFromETF: () => void
    prefillSectorsFromYahoo: () => void
    prefillSectorsFromMorningstar: () => void
    sectorsSource: string | null
    sectorsUpdatedAt: string | null
    onSectorsDirty: () => void
    assetType: string
  } = $props()

  // Resolved chart palette: keeps the table swatches in sync with the donut
  // colors and re-evaluates on theme flips.
  const palette = $derived(resolvePalette(resolved()))

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
    class="fixed inset-0 z-50 flex items-center justify-center bg-overlay/50"
    onclick={handleBackdropClick}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
    role="dialog"
    aria-modal="true"
    aria-label="Modifica distribuzione settoriale"
    tabindex="-1"
  >
    <div
      class="relative mx-4 max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-card border-border bg-surface p-6 shadow-raised"
    >
      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-lg font-semibold">Modifica distribuzione settoriale</h2>
        <button
          onclick={onClose}
          class="rounded-control p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
          aria-label="Chiudi"
        >
          <X class="h-5 w-5" />
        </button>
      </div>

      <!-- Sectors -->
      <div class="flex flex-col rounded-card border border-border bg-muted p-4">
        <div class="mb-3 flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <h3 class="font-medium">Distribuzione settoriale</h3>
            <ProvenanceBadge source={sectorsSource} updatedAt={sectorsUpdatedAt} />
          </div>
          <div class="flex items-center gap-1.5">
            <button
              onclick={prefillSectorsFromETF}
              disabled={fetchingETF || assetType !== 'etf'}
              title="Prefill da JustETF"
              aria-label="Prefill settori da JustETF"
              class="rounded-control border border-input bg-surface p-1.5 shadow-card hover:bg-muted disabled:cursor-not-allowed disabled:opacity-40"
            >
              {#if fetchingETF}
                <Loader2 class="h-5 w-5 animate-spin text-muted-foreground" />
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
              class="rounded-control border border-input bg-surface p-1.5 shadow-card hover:bg-muted disabled:cursor-not-allowed disabled:opacity-40"
            >
              {#if prefilling}
                <Loader2 class="h-5 w-5 animate-spin text-muted-foreground" />
              {:else}
                <img
                  class="h-5 w-5 rounded"
                  src="https://www.google.com/s2/favicons?domain=finance.yahoo.com&sz=32"
                  alt="Yahoo"
                />
              {/if}
            </button>
            <button
              onclick={prefillSectorsFromMorningstar}
              disabled={fetchingMorningstar || assetType !== 'etf'}
              title="Prefill da Morningstar"
              aria-label="Prefill settori da Morningstar"
              class="rounded-control border border-input bg-surface p-1.5 shadow-card hover:bg-muted disabled:cursor-not-allowed disabled:opacity-40"
            >
              {#if fetchingMorningstar}
                <Loader2 class="h-5 w-5 animate-spin text-muted-foreground" />
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
                <tr class="border-b border-border text-muted-foreground">
                  <th class="pb-2">Settore GICS</th>
                  <th class="pb-2 text-right">Peso %</th>
                </tr>
              </thead>
              <tbody>
                {#each sectorsEdit as s (s.name)}
                  <tr class="border-b border-border last:border-0">
                    <td class="py-2">
                      <span class="flex items-center gap-2">
                        <span
                          class="inline-block h-3 w-3 shrink-0 rounded"
                          style="background-color: {colorForRow(s, sectorsEdit, palette)};"
                        ></span>
                        {s.name}
                      </span>
                    </td>
                    <td class="py-2 text-right">
                      <input
                        type="number"
                        min="0"
                        max="100"
                        step="0.01"
                        value={s.weight}
                        oninput={(e) => {
                          s.weight = e.currentTarget.value
                          onSectorsDirty()
                        }}
                        class="no-spinner w-24 rounded-control border border-input px-3 py-1.5 text-right text-sm tabular-nums"
                      />
                    </td>
                  </tr>
                {/each}
                <tr class="border-t border-border font-semibold">
                  <td class="py-2">Totale</td>
                  <td class="py-2 text-right tabular-nums {sectorsValid ? 'text-positive' : 'text-negative'}">
                    {sumSectors.toFixed(2)}%
                  </td>
                </tr>
              </tbody>
            </table>
            {#if !sectorsValid}
              <p class="mt-2 text-sm text-negative">
                La somma dei pesi deve essere 100 (±0.5) — attuale: {sumSectors.toFixed(2)}%
              </p>
            {/if}
          </div>
          <div class="w-48 shrink-0 md:w-56">
            <ExposurePie data={sectorsEdit} title="Distribuzione settoriale" mute />
          </div>
        </div>
        <div class="mt-auto flex justify-end pt-4">
          <button
            onclick={saveSectors}
            disabled={!sectorsValid || savingSectors}
            class="rounded-control bg-accent px-4 py-2 text-sm text-accent-foreground hover:bg-accent-hover disabled:opacity-50"
          >
            {savingSectors ? 'Salvataggio...' : 'Salva'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
