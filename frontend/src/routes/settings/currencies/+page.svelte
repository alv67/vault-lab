<script lang="ts">
  import { onMount } from 'svelte'
  import { toast } from '$lib/stores/toast.svelte'
  import { settingsApi, type Currency } from '$lib/services/api'
  import { CURRENCIES } from '$lib/currencies'
  import { currencySymbol } from '$lib/format'
  import { Trash2 } from 'lucide-svelte'
  import SettingsTabs from '$lib/components/domain/SettingsTabs.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'

  let currencies = $state<Currency[]>([])
  let currenciesLoading = $state(true)
  let newCode = $state('')
  let newName = $state('')

  let availableCurrencies = $derived(CURRENCIES.filter((c) => !currencies.some((m) => m.code === c.code)))

  let addingCurrency = $state(false)
  let removingCode = $state('')

  let showDeleteDialog = $state(false)
  let currencyToDelete = $state('')

  onMount(async () => {
    try {
      currencies = await listCurrencies()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load currencies'
      toast.error(message)
    } finally {
      currenciesLoading = false
    }
  })

  async function listCurrencies(): Promise<Currency[]> {
    const res = await settingsApi.listCurrencies()
    return res.currencies
  }

  function errorStatus(err: unknown): number | undefined {
    return err instanceof Error && 'status' in err ? (err as Error & { status: number }).status : undefined
  }

  function selectCurrency(code: string): void {
    newCode = code
    const match = CURRENCIES.find((c) => c.code === code)
    if (match) newName = match.name
  }

  async function addCurrency(): Promise<void> {
    const code = newCode
    if (!code) return
    addingCurrency = true
    try {
      await settingsApi.addCurrency(code, newName.trim() || undefined)
      currencies = await listCurrencies()
      newCode = ''
      newName = ''
      toast.success(`Currency ${code} added`)
    } catch (err: unknown) {
      const status = errorStatus(err)
      if (status === 422) {
        toast.error(`USD->${code} conversion not available; currency not manageable`)
      } else if (status === 409) {
        toast.error('Currency already present')
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to add currency')
      }
    } finally {
      addingCurrency = false
    }
  }

  function requestDeleteCurrency(code: string): void {
    currencyToDelete = code
    showDeleteDialog = true
  }

  async function removeCurrency(): Promise<void> {
    const code = currencyToDelete
    if (!code) return
    removingCode = code
    try {
      await settingsApi.deleteCurrency(code)
      currencies = await listCurrencies()
      toast.success(`Currency ${code} removed`)
    } catch (err: unknown) {
      const status = errorStatus(err)
      if (status === 409) {
        toast.error('Currency in use or protected')
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to remove currency')
      }
    } finally {
      removingCode = ''
    }
  }
</script>

<div class="p-6">
  <h1 class="mb-6 text-2xl font-bold">Settings</h1>

  <SettingsTabs class="mb-6" />

  <Card class="max-w-2xl p-6">
    <h2 class="mb-4 font-semibold">Valute gestite</h2>

    {#if availableCurrencies.length > 0}
      <div class="mb-4 flex items-end gap-3">
        <Field label="Code" class="w-64 shrink-0">
          <Select value={newCode} onchange={(e) => selectCurrency(e.currentTarget.value)}>
            <option value="" disabled>Select a currency</option>
            {#each availableCurrencies as c (c.code)}
              <option value={c.code}>{c.code} — {c.name}</option>
            {/each}
          </Select>
        </Field>
        <Field label="Name" class="flex-1">
          <Input placeholder="Optional" bind:value={newName} />
        </Field>
        <Button onclick={addCurrency} disabled={!newCode || addingCurrency}>
          {addingCurrency ? 'Adding...' : 'Add'}
        </Button>
      </div>
    {:else}
      <p class="mb-4 text-sm text-muted-foreground">All listed currencies are already managed.</p>
    {/if}

    {#if currenciesLoading}
      <p class="text-muted-foreground">Loading...</p>
    {:else if currencies.length === 0}
      <p class="text-sm text-muted-foreground">No currencies found.</p>
    {:else}
      <Table>
        <THead>
          <Tr>
            <Th>Code</Th>
            <Th>Name</Th>
            <Th align="right">Actions</Th>
          </Tr>
        </THead>
        <TBody>
          {#each currencies as c (c.code)}
            <Tr>
              <Td class="font-medium">{c.code}</Td>
              <Td class="text-muted-foreground">
                {c.name || '—'} <span class="text-xs text-muted-foreground">{currencySymbol(c.code)}</span>
              </Td>
              <Td align="right">
                <Button
                  variant="ghost"
                  size="icon"
                  class="text-muted-foreground hover:text-negative"
                  aria-label={`Remove currency ${c.code}`}
                  title="Remove currency"
                  disabled={removingCode === c.code}
                  onclick={() => requestDeleteCurrency(c.code)}
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </Td>
            </Tr>
          {/each}
        </TBody>
      </Table>
    {/if}
  </Card>
</div>

<ConfirmDialog
  bind:open={showDeleteDialog}
  variant="danger"
  title="Delete currency"
  message={`Delete currency ${currencyToDelete}?`}
  confirmLabel="Delete"
  cancelLabel="Cancel"
  loading={showDeleteDialog && removingCode === currencyToDelete}
  onconfirm={removeCurrency}
/>
