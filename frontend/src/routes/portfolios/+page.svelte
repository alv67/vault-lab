<script lang="ts">
  import { onMount } from 'svelte'
  import { resolve } from '$app/paths'
  import { toast } from '$lib/stores/toast.svelte'
  import { portfolioApi, type Portfolio } from '$lib/services/api'
  import { Plus, ExternalLink } from 'lucide-svelte'

  let showCreate = $state(false)
  let name = $state('')
  let description = $state('')
  let currency = $state('USD')
  let portfolios = $state<Portfolio[] | null>(null)
  let loading = $state(true)
  let creating = $state(false)

  onMount(async () => {
    try {
      portfolios = await portfolioApi.list()
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
</script>

<div class="p-6">
  <div class="mb-6 flex items-center justify-between">
    <h1 class="text-2xl font-bold">Portfolios</h1>
    <button
      onclick={() => (showCreate = !showCreate)}
      class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
    >
      <Plus class="h-4 w-4" />
      New Portfolio
    </button>
  </div>

  {#if showCreate}
    <div class="mb-6 rounded-xl border bg-white p-4">
      <h2 class="mb-4 font-semibold">Create Portfolio</h2>
      <div class="space-y-3">
        <input
          type="text"
          placeholder="Portfolio name"
          bind:value={name}
          class="w-full rounded-lg border px-3 py-2 text-sm"
        />
        <textarea
          placeholder="Description (optional)"
          bind:value={description}
          class="w-full rounded-lg border px-3 py-2 text-sm"
        ></textarea>
        <select
          bind:value={currency}
          class="w-full rounded-lg border px-3 py-2 text-sm"
        >
          <option value="USD">USD</option>
          <option value="EUR">EUR</option>
          <option value="GBP">GBP</option>
          <option value="CHF">CHF</option>
        </select>
        <button
          onclick={createPortfolio}
          disabled={!name || creating}
          class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {creating ? 'Creating...' : 'Create'}
        </button>
      </div>
    </div>
  {/if}

  {#if loading}
    <p class="text-gray-500">Loading...</p>
  {:else}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {#each portfolios ?? [] as p (p.id)}
        <div class="rounded-xl border bg-white p-4 shadow-sm">
          <div class="mb-2 flex items-start justify-between">
            <div>
              <h3 class="font-semibold">{p.name}</h3>
              <p class="text-xs text-gray-500">{p.currency}</p>
            </div>
            <a href={resolve(`/portfolios/${p.id}`)} class="text-blue-600 hover:text-blue-800">
              <ExternalLink class="h-4 w-4" />
            </a>
          </div>
          {#if p.description}
            <p class="mb-3 text-sm text-gray-600">{p.description}</p>
          {/if}
          <button
            onclick={() => deletePortfolio(p.id)}
            class="text-xs text-red-500 hover:underline"
          >
            Delete
          </button>
        </div>
      {/each}
      {#if (portfolios ?? []).length === 0}
        <p class="col-span-full text-center text-gray-400">
          No portfolios yet. Create one!
        </p>
      {/if}
    </div>
  {/if}
</div>
