<script lang="ts">
  import { onMount } from 'svelte'
  import { resolve } from '$app/paths'
  import { toast } from '$lib/stores/toast.svelte'
  import { assetApi, settingsApi, type Asset, type Currency } from '$lib/services/api'
  import { Plus, Trash2 } from 'lucide-svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'
  import CreateAssetModal from '$lib/components/domain/CreateAssetModal.svelte'

  let showCreate = $state(false)
  let assets = $state<Asset[] | null>(null)
  let loading = $state(true)
  let currencies = $state<Currency[]>([])

  // Delete-confirmation dialog (D.2): replaces the native confirm().
  let showDeleteDialog = $state(false)
  let deleting = $state(false)
  let assetToDelete = $state<{ id: string; ticker: string } | null>(null)

  onMount(async () => {
    try {
      const [assetList, curList] = await Promise.all([assetApi.list(), settingsApi.listCurrencies()])
      assets = assetList
      currencies = curList.currencies
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load assets'
      toast.error(message)
    } finally {
      loading = false
    }
  })

  async function reloadAssets(): Promise<void> {
    assets = await assetApi.list()
  }

  function requestDeleteAsset(id: string, ticker: string): void {
    assetToDelete = { id, ticker }
    showDeleteDialog = true
  }

  async function deleteAsset(): Promise<void> {
    if (!assetToDelete) return
    const { id } = assetToDelete
    deleting = true
    try {
      await assetApi.remove(id)
      assets = await assetApi.list()
      toast.success('Asset deleted')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Delete failed'
      toast.error(message)
    } finally {
      deleting = false
    }
  }
</script>

<div class="p-6">
  <div class="mb-6 flex items-center justify-between">
    <h1 class="text-2xl font-bold">Assets</h1>
    <Button onclick={() => (showCreate = true)}>
      <Plus class="h-4 w-4" />
      Add Asset
    </Button>
  </div>

  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else}
    <Table aria-label="Assets">
      <THead>
        <Tr>
          <Th>Ticker</Th>
          <Th>Name</Th>
          <Th>Type</Th>
          <Th>Currency</Th>
          <Th>Country</Th>
          <Th align="right">Actions</Th>
        </Tr>
      </THead>
      <TBody>
        {#each assets ?? [] as a (a.id)}
          <Tr>
            <Td class="font-medium">
              <a href={resolve(`/assets/${a.id}`)} class="text-accent-text hover:underline">{a.ticker}</a>
            </Td>
            <Td class="text-muted-foreground">{a.name}</Td>
            <Td><Badge>{a.type}</Badge></Td>
            <Td>{a.currency}</Td>
            <Td>{a.country || '-'}</Td>
            <Td align="right">
              <Button
                variant="ghost"
                size="icon"
                aria-label={`Delete ${a.ticker}`}
                onclick={() => requestDeleteAsset(a.id, a.ticker)}
              >
                <Trash2 class="h-4 w-4" />
              </Button>
            </Td>
          </Tr>
        {/each}
      </TBody>
    </Table>
  {/if}
</div>

<CreateAssetModal bind:open={showCreate} {currencies} onsuccess={reloadAssets} />

<ConfirmDialog
  bind:open={showDeleteDialog}
  variant="danger"
  title="Delete asset"
  message={`Delete ${assetToDelete?.ticker ?? ''}?`}
  confirmLabel="Delete"
  cancelLabel="Cancel"
  loading={deleting}
  onconfirm={deleteAsset}
/>
