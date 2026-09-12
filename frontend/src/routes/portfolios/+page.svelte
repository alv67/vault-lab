<script lang="ts">
  import { onMount } from 'svelte'
  import { resolve } from '$app/paths'
  import { toast } from '$lib/stores/toast.svelte'
  import {
    portfolioApi,
    settingsApi,
    assetApi,
    type Portfolio,
    type Currency,
    type PortfolioExportDocument,
  } from '$lib/services/api'
  import { Plus, ExternalLink, Upload, X } from 'lucide-svelte'

  let showCreate = $state(false)
  let name = $state('')
  let description = $state('')
  let currency = $state('USD')
  let currencies = $state<Currency[]>([])
  let portfolios = $state<Portfolio[] | null>(null)
  let loading = $state(true)
  let creating = $state(false)

  let fileInput = $state<HTMLInputElement | null>(null)
  let importDoc = $state<PortfolioExportDocument | null>(null)
  let importMode = $state<'new' | 'overwrite'>('new')
  let importName = $state('')
  let importTarget = $state('')
  let importError = $state('')
  let importing = $state(false)

onMount(async () => {
    try {
      const [portfolioList, curList] = await Promise.all([portfolioApi.list(), settingsApi.listCurrencies()])
      portfolios = portfolioList
      currencies = curList.currencies
      if (curList.currencies.length > 0) {
        const preferred = curList.currencies.find((c) => c.code === 'USD') ?? curList.currencies[0]
        currency = preferred.code
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load portfolios'
      toast.error(message)
    } finally {
      loading = false
    }
  })

  async function createPortfolio(): Promise<void> {
    creating = true
    try {
      await portfolioApi.create({ name, description, currency })
      portfolios = await portfolioApi.list()
      showCreate = false
      name = ''
      description = ''
      toast.success('Portfolio created')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to create portfolio'
      toast.error(message)
    } finally {
      creating = false
    }
  }

  async function deletePortfolio(id: string): Promise<void> {
    if (!confirm('Delete this portfolio?')) return
    try {
      await portfolioApi.delete(id)
      portfolios = await portfolioApi.list()
      toast.success('Portfolio deleted')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Delete failed'
      toast.error(message)
    }
  }

  function openFilePicker(): void {
    fileInput?.click()
  }

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
      portfolios = await portfolioApi.list()
      importDoc = null
      // Imported assets have no market data yet: trigger the backfill
      // (history + splits) right away instead of waiting for next app load.
      assetApi.sync().catch(() => {})
      toast.success('Portfolio imported')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Import failed'
      importError = message
      toast.error(message)
    } finally {
      importing = false
    }
  }
</script>

<div class="p-6">
  <div class="mb-6 flex items-center justify-between">
    <h1 class="text-2xl font-bold">Portfolios</h1>
    <div class="flex items-center gap-2">
      <input
        type="file"
        accept=".json,application/json"
        class="hidden"
        bind:this={fileInput}
        onchange={onFileSelected}
      />
      <button
        onclick={openFilePicker}
        class="flex items-center gap-2 rounded-control border border-border px-4 py-2 text-sm text-foreground hover:bg-muted"
      >
        <Upload class="h-4 w-4" />
        Import
      </button>
      <button
        onclick={() => (showCreate = !showCreate)}
        class="flex items-center gap-2 rounded-control bg-accent px-4 py-2 text-sm text-accent-foreground hover:bg-accent-hover"
      >
        <Plus class="h-4 w-4" />
        New Portfolio
      </button>
    </div>
  </div>

  {#if importError}
    <div class="mb-4 rounded-control border border-negative/20 bg-negative/10 px-4 py-2 text-sm text-negative">
      {importError}
    </div>
  {/if}

  {#if importDoc}
    <div class="mb-6 rounded-card border border-border bg-surface p-4">
      <div class="mb-3 flex items-center justify-between">
        <h2 class="font-semibold">Import portfolio</h2>
        <button
          onclick={() => (importDoc = null)}
          class="rounded-control p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
          aria-label="Close import"
        >
          <X class="h-4 w-4" />
        </button>
      </div>
      <div class="mb-4 grid grid-cols-2 gap-3 text-sm md:grid-cols-4">
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
      <div class="mb-4 space-y-3">
        <label class="flex items-center gap-2 text-sm">
          <input type="radio" bind:group={importMode} value="new" />
          Create as new portfolio
        </label>
        {#if importMode === 'new'}
          <input
            type="text"
            placeholder="Portfolio name"
            bind:value={importName}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          />
        {/if}
        <label class="flex items-center gap-2 text-sm">
          <input type="radio" bind:group={importMode} value="overwrite" />
          Overwrite existing portfolio
        </label>
        {#if importMode === 'overwrite'}
          <select
            bind:value={importTarget}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          >
            <option value="" disabled>Select portfolio to overwrite</option>
            {#each portfolios ?? [] as p (p.id)}
              <option value={p.id}>{p.name}</option>
            {/each}
          </select>
        {/if}
      </div>
      <button
        onclick={confirmImport}
        disabled={
          importing ||
          (importMode === 'new' && !importName.trim()) ||
          (importMode === 'overwrite' && !importTarget)
        }
        class="rounded-control bg-accent px-4 py-2 text-sm text-accent-foreground hover:bg-accent-hover disabled:opacity-50"
      >
        {importing ? 'Importing...' : 'Import'}
      </button>
    </div>
  {/if}

  {#if showCreate}
    <div class="mb-6 rounded-card border border-border bg-surface p-4">
      <h2 class="mb-4 font-semibold">Create Portfolio</h2>
      <div class="space-y-3">
        <input
          type="text"
          placeholder="Portfolio name"
          bind:value={name}
          class="w-full rounded-control border border-input px-3 py-2 text-sm"
        />
        <textarea
          placeholder="Description (optional)"
          bind:value={description}
          class="w-full rounded-control border border-input px-3 py-2 text-sm"
        ></textarea>
        <select
          bind:value={currency}
          class="w-full rounded-control border border-input px-3 py-2 text-sm"
        >
          {#each currencies as c (c.code)}
            <option value={c.code}>{c.code}</option>
          {/each}
        </select>
        <button
          onclick={createPortfolio}
          disabled={!name || creating}
          class="rounded-control bg-accent px-4 py-2 text-sm text-accent-foreground hover:bg-accent-hover disabled:opacity-50"
        >
          {creating ? 'Creating...' : 'Create'}
        </button>
      </div>
    </div>
  {/if}

  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {#each portfolios ?? [] as p (p.id)}
        <div class="rounded-card border border-border bg-surface p-4 shadow-card">
          <div class="mb-2 flex items-start justify-between">
            <div>
              <h3 class="font-semibold">{p.name}</h3>
              <p class="text-xs text-muted-foreground">{p.currency}</p>
            </div>
            <a href={resolve(`/portfolios/${p.id}`)} class="text-accent-text hover:underline">
              <ExternalLink class="h-4 w-4" />
            </a>
          </div>
          {#if p.description}
            <p class="mb-3 text-sm text-muted-foreground">{p.description}</p>
          {/if}
          <button
            onclick={() => deletePortfolio(p.id)}
            class="text-xs text-negative hover:underline"
          >
            Delete
          </button>
        </div>
      {/each}
      {#if (portfolios ?? []).length === 0}
        <p class="col-span-full text-center text-muted-foreground">
          No portfolios yet. Create one!
        </p>
      {/if}
    </div>
  {/if}
</div>
