<script lang="ts">
  import { onMount } from 'svelte'
  import { resolve } from '$app/paths'
  import { toast } from '$lib/stores/toast.svelte'
  import { portfolioApi, settingsApi, type Portfolio, type Currency } from '$lib/services/api'
  import { Plus, ExternalLink, Upload } from 'lucide-svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte'
  import EmptyState from '$lib/components/ui/EmptyState.svelte'
  import CreatePortfolioModal from '$lib/components/domain/CreatePortfolioModal.svelte'
  import ImportPortfolioModal from '$lib/components/domain/ImportPortfolioModal.svelte'

  let showCreate = $state(false)
  let showImport = $state(false)
  let currencies = $state<Currency[]>([])
  let portfolios = $state<Portfolio[] | null>(null)
  let loading = $state(true)

  // Delete-confirmation dialog (D.2): replaces the native confirm().
  let showDeleteDialog = $state(false)
  let deleting = $state(false)
  let portfolioToDelete = $state('')

  onMount(async () => {
    try {
      const [portfolioList, curList] = await Promise.all([portfolioApi.list(), settingsApi.listCurrencies()])
      portfolios = portfolioList
      currencies = curList.currencies
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load portfolios'
      toast.error(message)
    } finally {
      loading = false
    }
  })

  async function reloadPortfolios(): Promise<void> {
    portfolios = await portfolioApi.list()
  }

  function requestDeletePortfolio(id: string): void {
    portfolioToDelete = id
    showDeleteDialog = true
  }

  async function deletePortfolio(): Promise<void> {
    if (!portfolioToDelete) return
    deleting = true
    try {
      await portfolioApi.delete(portfolioToDelete)
      portfolios = await portfolioApi.list()
      toast.success('Portfolio deleted')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Delete failed'
      toast.error(message)
    } finally {
      deleting = false
    }
  }
</script>

{#snippet createAction()}
  <Button onclick={() => (showCreate = true)}>
    <Plus class="h-4 w-4" />
    New Portfolio
  </Button>
{/snippet}

<div class="p-6">
  <div class="mb-6 flex items-center justify-between">
    <h1 class="text-2xl font-bold">Portfolios</h1>
    <div class="flex items-center gap-2">
      <Button variant="secondary" onclick={() => (showImport = true)}>
        <Upload class="h-4 w-4" />
        Import
      </Button>
      <Button onclick={() => (showCreate = true)}>
        <Plus class="h-4 w-4" />
        New Portfolio
      </Button>
    </div>
  </div>

  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {#each portfolios ?? [] as p (p.id)}
        <Card class="p-4">
          <div class="mb-2 flex items-start justify-between">
            <div>
              <h3 class="font-semibold">{p.name}</h3>
              <p class="text-xs text-muted-foreground">{p.currency}</p>
            </div>
            <a
              href={resolve(`/portfolios/${p.id}`)}
              aria-label={`Open ${p.name}`}
              class="text-accent-text hover:underline"
            >
              <ExternalLink class="h-4 w-4" />
            </a>
          </div>
          {#if p.description}
            <p class="mb-3 text-sm text-muted-foreground">{p.description}</p>
          {/if}
          <button
            onclick={() => requestDeletePortfolio(p.id)}
            class="text-xs text-negative hover:underline"
          >
            Delete
          </button>
        </Card>
      {/each}
      {#if (portfolios ?? []).length === 0}
        <EmptyState
          class="col-span-full"
          title="No portfolios yet"
          description="Create one to get started"
          action={createAction}
        />
      {/if}
    </div>
  {/if}
</div>

<CreatePortfolioModal bind:open={showCreate} {currencies} onsuccess={reloadPortfolios} />

<ImportPortfolioModal
  bind:open={showImport}
  portfolios={portfolios ?? []}
  onsuccess={reloadPortfolios}
/>

<ConfirmDialog
  bind:open={showDeleteDialog}
  variant="danger"
  title="Delete portfolio"
  message="Delete this portfolio?"
  confirmLabel="Delete"
  cancelLabel="Cancel"
  loading={deleting}
  onconfirm={deletePortfolio}
/>
