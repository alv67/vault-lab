<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { portfolioApi, type Currency } from '$lib/services/api'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Textarea from '$lib/components/ui/Textarea.svelte'
  import CurrencySelect from './CurrencySelect.svelte'

  /** Portfolio creation dialog (EPIC E.3): owns the form, API call and toast. */
  let {
    open = $bindable(false),
    currencies = [],
    onsuccess = undefined,
  }: {
    open?: boolean
    currencies?: Currency[]
    onsuccess?: () => void
  } = $props()

  let name = $state('')
  let description = $state('')
  let currency = $state('USD')
  let creating = $state(false)

  const defaultCurrency = $derived(
    currencies.find((c) => c.code === 'USD')?.code ?? currencies[0]?.code ?? 'USD',
  )

  // Closing (or first mount while still closed) clears the draft form.
  $effect(() => {
    if (!open) {
      name = ''
      description = ''
      currency = defaultCurrency
    }
  })

  async function submit(): Promise<void> {
    creating = true
    try {
      await portfolioApi.create({ name, description, currency })
      toast.success('Portfolio created')
      open = false
      onsuccess?.()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to create portfolio'
      toast.error(message)
    } finally {
      creating = false
    }
  }
</script>

{#snippet footer()}
  <Button variant="secondary" onclick={() => (open = false)} disabled={creating}>Cancel</Button>
  <Button onclick={submit} disabled={!name} loading={creating}>Create</Button>
{/snippet}

<Modal bind:open title="Create Portfolio" {footer} dismissible={!creating}>
  <div class="space-y-4">
    <Field label="Name">
      <Input bind:value={name} placeholder="Portfolio name" />
    </Field>
    <Field label="Description" hint="optional">
      <Textarea bind:value={description} rows={3} />
    </Field>
    <Field label="Currency">
      <CurrencySelect bind:value={currency} {currencies} />
    </Field>
  </div>
</Modal>
