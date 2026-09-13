<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { assetApi, type AssetLookupResult, type Currency } from '$lib/services/api'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import AssetSearchAutocomplete from './AssetSearchAutocomplete.svelte'
  import CurrencySelect from './CurrencySelect.svelte'

  /** Asset creation dialog (EPIC E.3): owns the form, the API call and the toast. */
  let {
    open = $bindable(false),
    currencies = [],
    onsuccess = undefined,
  }: {
    open?: boolean
    currencies?: Currency[]
    onsuccess?: () => void
  } = $props()

  const uid = $props.id()

  const defaultForm = () => ({
    ticker: '',
    isin: '',
    name: '',
    type: 'stock' as string,
    currency: 'USD',
    country: '',
    exchange: '',
    asset_class: '',
    price_source: 'yahoo',
  })

  let form = $state(defaultForm())
  let saving = $state(false)

  const defaultCurrency = $derived(
    currencies.find((c) => c.code === 'USD')?.code ?? currencies[0]?.code ?? 'USD',
  )

  function resetForm(): void {
    form = { ...defaultForm(), currency: defaultCurrency }
  }

  // Closing (or first mount while still closed) clears the draft form.
  $effect(() => {
    if (!open) resetForm()
  })

  async function handleSelect(result: AssetLookupResult): Promise<void> {
    form = {
      ticker: result.ticker,
      isin: form.isin,
      name: result.name || '',
      type: result.type || 'stock',
      currency: result.currency || 'USD',
      country: '',
      exchange: result.exchange || '',
      asset_class: '',
      price_source: 'yahoo',
    }

    try {
      const meta = await assetApi.meta(result.ticker)
      form = {
        ...form,
        name: meta.name || form.name,
        type: meta.type || form.type,
        currency: meta.currency || form.currency,
        country: meta.country || form.country,
        exchange: meta.exchange || form.exchange,
        asset_class: meta.asset_class || '',
      }
    } catch {
      // keep lookup defaults; currency/type remain editable
    }
  }

  async function submit(): Promise<void> {
    saving = true
    try {
      await assetApi.create(form)
      toast.success('Asset created')
      open = false
      onsuccess?.()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Create failed'
      toast.error(message)
    } finally {
      saving = false
    }
  }
</script>

{#snippet footer()}
  <Button variant="secondary" onclick={() => (open = false)} disabled={saving}>Cancel</Button>
  <Button onclick={submit} disabled={!form.ticker || !form.name} loading={saving}>Save</Button>
{/snippet}

<Modal bind:open title="New Asset" description="Look up a ticker to prefill the details." size="lg" {footer} dismissible={!saving}>
  <div class="grid gap-3 sm:grid-cols-2">
    <Field label="Ticker" for={`${uid}-ticker`}>
      <AssetSearchAutocomplete
        inputId={`${uid}-ticker`}
        bind:ticker={form.ticker}
        onselect={handleSelect}
      />
    </Field>
    <Field label="Name">
      <Input bind:value={form.name} placeholder="Name" />
    </Field>
    <Field label="ISIN" hint="optional">
      <Input bind:value={form.isin} />
    </Field>
    <Field label="Type">
      <Select bind:value={form.type}>
        <option value="stock">Stock</option>
        <option value="etf">ETF</option>
        <option value="bond">Bond</option>
        <option value="mutual_fund">Mutual fund</option>
        <option value="crypto">Crypto</option>
        <option value="commodity">Commodity</option>
      </Select>
    </Field>
    <Field label="Currency">
      <CurrencySelect bind:value={form.currency} {currencies} />
    </Field>
    <Field label="Exchange">
      <Input bind:value={form.exchange} placeholder="Exchange" />
    </Field>
    <Field label="Price source">
      <Select bind:value={form.price_source}>
        <option value="yahoo">Yahoo Finance</option>
        <option value="manual">Prezzo manuale</option>
        <option value="none">Nessun prezzo</option>
      </Select>
    </Field>
  </div>
</Modal>
