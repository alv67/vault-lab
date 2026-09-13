<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { transactionApi, type Asset, type Transaction } from '$lib/services/api'
  import { formatCurrency } from '$lib/format'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import Textarea from '$lib/components/ui/Textarea.svelte'
  import AssetCombobox from './AssetCombobox.svelte'

  /**
   * Add/edit transaction dialog (EPIC E.2): owns the form state, inline
   * validation, API calls, toasts and the delete confirmation so the
   * portfolio page only toggles `open` and refreshes via `onsuccess`.
   */
  let {
    open = $bindable(false),
    portfolioId,
    assets = [],
    currency = 'USD',
    editing = null,
    onsuccess = undefined,
  }: {
    open?: boolean
    portfolioId: string
    assets?: Asset[]
    currency?: string
    editing?: Transaction | null
    onsuccess?: () => void
  } = $props()

  const uid = $props.id()

  interface TxForm {
    asset_id: string
    /** Kept as `string` for the `ui/Select` binding; cast back on submit. */
    type: string
    quantity: string
    price: string
    date: string
    fees: string
    notes: string
  }

  interface ErrorBag {
    asset: string
    quantity: string
    price: string
    amount: string
    date: string
  }

  const defaultForm = (): TxForm => ({
    asset_id: '',
    type: 'buy',
    quantity: '',
    price: '',
    date: new Date().toISOString().split('T')[0],
    fees: '0',
    notes: '',
  })

  const emptyErrors = (): ErrorBag => ({
    asset: '',
    quantity: '',
    price: '',
    amount: '',
    date: '',
  })

  let form = $state(defaultForm())
  let dividendAmount = $state('')
  let errors = $state(emptyErrors())
  let saving = $state(false)
  let deleting = $state(false)
  let confirmDeleteOpen = $state(false)

  const isDividend = $derived(form.type === 'dividend')
  const total = $derived(
    isDividend ? Number(dividendAmount) : Number(form.quantity) * Number(form.price),
  )
  const totalEntered = $derived(
    isDividend
      ? dividendAmount.trim() !== '' && Number.isFinite(total)
      : form.quantity.trim() !== '' && form.price.trim() !== '' && Number.isFinite(total),
  )
  const totalText = $derived(totalEntered ? formatCurrency(total, currency) : '—')

  // Prefill on open, clear the draft on close (and while never opened).
  $effect(() => {
    if (open && editing) {
      form = {
        asset_id: editing.asset_id,
        type: editing.type,
        quantity: editing.quantity,
        price: editing.price,
        date: new Date(editing.date).toISOString().split('T')[0],
        fees: editing.fees || '0',
        notes: editing.notes || '',
      }
      dividendAmount = editing.type === 'dividend' ? String(Number(editing.price) * Number(editing.quantity)) : ''
    } else {
      form = defaultForm()
      dividendAmount = ''
    }
    errors = emptyErrors()
    if (!open) confirmDeleteOpen = false
  })

  // The combobox selection is the "input" for the asset field.
  $effect(() => {
    if (form.asset_id && errors.asset) errors.asset = ''
  })

  function validate(): ErrorBag {
    const next = emptyErrors()
    if (!form.asset_id) next.asset = 'Select an asset'
    if (!form.date) next.date = 'Date is required'
    if (isDividend) {
      if (!(Number(dividendAmount) > 0)) next.amount = 'Amount must be greater than 0'
    } else {
      if (!(Number(form.quantity) > 0)) next.quantity = 'Quantity must be greater than 0'
      if (form.price.trim() === '' || !(Number(form.price) >= 0)) {
        next.price = 'Price must be 0 or greater'
      }
    }
    return next
  }

  async function submit(): Promise<void> {
    errors = validate()
    if (Object.values(errors).some(Boolean)) return
    saving = true
    const payload: Partial<Transaction> = {
      asset_id: form.asset_id,
      type: form.type as Transaction['type'],
      quantity: isDividend ? '1' : form.quantity,
      price: isDividend ? dividendAmount : form.price,
      date: new Date(form.date).toISOString(),
      fees: form.fees,
      notes: form.notes,
    }
    try {
      if (editing) {
        await transactionApi.update(editing.id, payload)
      } else {
        await transactionApi.create(portfolioId, payload)
      }
      toast.success(editing ? 'Transaction updated' : 'Transaction added')
      open = false
      onsuccess?.()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed'
      toast.error(message)
    } finally {
      saving = false
    }
  }

  async function confirmDelete(): Promise<void> {
    if (!editing) return
    deleting = true
    try {
      await transactionApi.remove(editing.id)
      toast.success('Transaction deleted')
      open = false
      onsuccess?.()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Delete failed'
      toast.error(message)
    } finally {
      deleting = false
    }
  }
</script>

{#snippet footer()}
  <Button variant="secondary" onclick={() => (open = false)} disabled={saving || deleting}>
    Cancel
  </Button>
  {#if editing}
    <Button
      variant="danger"
      onclick={() => (confirmDeleteOpen = true)}
      disabled={saving || deleting}
    >
      Delete
    </Button>
  {/if}
  <Button onclick={submit} loading={saving} disabled={deleting}>
    {editing ? 'Save Changes' : 'Save'}
  </Button>
{/snippet}

<Modal
  bind:open
  title={editing ? 'Edit Transaction' : 'New Transaction'}
  size="lg"
  {footer}
  dismissible={!saving && !deleting}
>
  <div class="grid gap-3 sm:grid-cols-2">
    <Field label="Asset" for={`${uid}-asset`} error={errors.asset || undefined}>
      <AssetCombobox
        inputId={`${uid}-asset`}
        bind:value={form.asset_id}
        {assets}
        error={errors.asset || undefined}
        disabled={saving || deleting}
      />
    </Field>
    <Field label="Type">
      <Select
        bind:value={form.type}
        oninput={() => {
          errors.quantity = ''
          errors.price = ''
          errors.amount = ''
        }}
      >
        <option value="buy">Buy</option>
        <option value="sell">Sell</option>
        <option value="dividend">Dividend</option>
      </Select>
    </Field>
    {#if isDividend}
      <Field label="Amount" error={errors.amount || undefined}>
        <Input
          type="number"
          step="0.01"
          min="0"
          placeholder="Amount"
          bind:value={dividendAmount}
          error={errors.amount || undefined}
          oninput={() => (errors.amount = '')}
        />
      </Field>
    {:else}
      <Field label="Quantity" error={errors.quantity || undefined}>
        <Input
          type="number"
          step="any"
          placeholder="Quantity"
          bind:value={form.quantity}
          error={errors.quantity || undefined}
          oninput={() => (errors.quantity = '')}
        />
      </Field>
      <Field label="Price" error={errors.price || undefined}>
        <Input
          type="number"
          step="0.01"
          placeholder="Price"
          bind:value={form.price}
          error={errors.price || undefined}
          oninput={() => (errors.price = '')}
        />
      </Field>
    {/if}
    <Field label="Date" error={errors.date || undefined}>
      <Input
        type="date"
        bind:value={form.date}
        error={errors.date || undefined}
        oninput={() => (errors.date = '')}
      />
    </Field>
    <Field label="Fees">
      <Input type="number" step="0.01" placeholder="Fees" bind:value={form.fees} />
    </Field>
    <Field label="Notes" class="sm:col-span-2">
      <Textarea placeholder="Notes" rows={2} bind:value={form.notes} />
    </Field>
  </div>

  <div class="mt-4 flex items-baseline justify-end gap-2 text-sm">
    <span class="text-muted-foreground">Total</span>
    <span class="font-semibold tabular-nums text-foreground">{totalText}</span>
  </div>
</Modal>

<ConfirmDialog
  bind:open={confirmDeleteOpen}
  variant="danger"
  title="Delete transaction"
  message="Delete this transaction?"
  confirmLabel="Delete"
  cancelLabel="Cancel"
  loading={deleting}
  onconfirm={confirmDelete}
/>
