<script lang="ts">
  import { Upload } from 'lucide-svelte'
  import { toast } from '$lib/stores/toast.svelte'
  import {
    assetApi,
    portfolioApi,
    type Portfolio,
    type PortfolioExportDocument,
  } from '$lib/services/api'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'

  /** JSON portfolio-import dialog (EPIC E.3), extracted from the portfolios page. */
  let {
    open = $bindable(false),
    portfolios = [],
    onsuccess = undefined,
  }: {
    open?: boolean
    portfolios?: Portfolio[]
    onsuccess?: () => void
  } = $props()

  let fileInput = $state<HTMLInputElement | null>(null)
  let importDoc = $state<PortfolioExportDocument | null>(null)
  let importMode = $state<'new' | 'overwrite'>('new')
  let importName = $state('')
  let importTarget = $state('')
  let importError = $state('')
  let importing = $state(false)

  // Closing (or first mount while still closed) restarts the flow from the
  // file-picker step.
  $effect(() => {
    if (!open) {
      importDoc = null
      importMode = 'new'
      importName = ''
      importTarget = ''
      importError = ''
    }
  })

  async function onFileSelected(e: Event): Promise<void> {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    importError = ''
    if (!file) return
    try {
      const text = await file.text()
      const doc = JSON.parse(text) as PortfolioExportDocument
      if (!doc || doc.version !== 1 || !doc.portfolio?.name) {
        throw new Error('File non valido: formato di export non riconosciuto')
      }
      importDoc = doc
      importName = doc.portfolio.name
      importMode = 'new'
      importTarget = ''
    } catch (err: unknown) {
      importDoc = null
      importError = err instanceof Error ? err.message : 'Impossibile leggere il file'
      toast.error(importError)
    }
  }

  function importRange(doc: PortfolioExportDocument): string {
    const dates = (doc.transactions ?? []).map((t) => new Date(t.date).getTime()).filter(Number.isFinite)
    if (dates.length === 0) return '—'
    const fmt = (t: number) => new Date(t).toISOString().split('T')[0]
    return `${fmt(Math.min(...dates))} → ${fmt(Math.max(...dates))}`
  }

  async function confirmImport(): Promise<void> {
    if (!importDoc) return
    importing = true
    importError = ''
    try {
      await portfolioApi.importDoc({
        document: importDoc,
        mode: importMode,
        name: importMode === 'new' ? importName : undefined,
        target_portfolio_id: importMode === 'overwrite' ? importTarget : undefined,
      })
      // Imported assets have no market data yet: trigger the backfill
      // (history + splits) right away instead of waiting for next app load.
      assetApi.sync().catch(() => {})
      toast.success('Portfolio imported')
      open = false
      onsuccess?.()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Import failed'
      importError = message
      toast.error(message)
    } finally {
      importing = false
    }
  }
</script>

{#snippet footer()}
  <Button variant="secondary" onclick={() => (open = false)} disabled={importing}>Cancel</Button>
  {#if importDoc}
    <Button
      variant="ghost"
      disabled={importing}
      onclick={() => {
        importDoc = null
        fileInput?.click()
      }}
    >
      Change file
    </Button>
    <Button
      onclick={confirmImport}
      loading={importing}
      disabled={
        (importMode === 'new' && !importName.trim()) ||
        (importMode === 'overwrite' && !importTarget)
      }
    >
      Import
    </Button>
  {/if}
{/snippet}

<Modal bind:open title="Import portfolio" {footer} dismissible={!importing}>
  {#if importError}
    <div class="mb-4 rounded-control border border-negative/20 bg-negative/10 px-4 py-2 text-sm text-negative">
      {importError}
    </div>
  {/if}

  {#if importDoc}
    <div class="mb-4 grid grid-cols-2 gap-3 text-sm">
      <div>
        <p class="text-xs text-muted-foreground">Name</p>
        <p class="font-medium">{importDoc.portfolio.name}</p>
      </div>
      <div>
        <p class="text-xs text-muted-foreground">Currency</p>
        <p class="font-medium">{importDoc.portfolio.currency || '—'}</p>
      </div>
      <div>
        <p class="text-xs text-muted-foreground">Transactions</p>
        <p class="font-medium">{importDoc.transactions?.length ?? 0}</p>
      </div>
      <div>
        <p class="text-xs text-muted-foreground">Date range</p>
        <p class="font-medium">{importRange(importDoc)}</p>
      </div>
    </div>
    <div class="space-y-3">
      <label class="flex items-center gap-2 text-sm">
        <input type="radio" bind:group={importMode} value="new" />
        Create as new portfolio
      </label>
      {#if importMode === 'new'}
        <Input bind:value={importName} placeholder="Portfolio name" />
      {/if}
      <label class="flex items-center gap-2 text-sm">
        <input type="radio" bind:group={importMode} value="overwrite" />
        Overwrite existing portfolio
      </label>
      {#if importMode === 'overwrite'}
        <Field label="Target portfolio">
          <Select bind:value={importTarget}>
            <option value="" disabled>Select portfolio to overwrite</option>
            {#each portfolios as p (p.id)}
              <option value={p.id}>{p.name}</option>
            {/each}
          </Select>
        </Field>
      {/if}
    </div>
  {:else}
    <div class="flex flex-col items-center gap-3 py-6 text-center">
      <p class="text-sm text-muted-foreground">
        Choose a VaultLab portfolio export (.json) to import.
      </p>
      <Button variant="secondary" onclick={() => fileInput?.click()}>
        <Upload class="h-4 w-4" />
        Choose file
      </Button>
    </div>
  {/if}

  <input
    type="file"
    accept=".json,application/json"
    class="hidden"
    bind:this={fileInput}
    onchange={onFileSelected}
  />
</Modal>
