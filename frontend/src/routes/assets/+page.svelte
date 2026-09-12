<script lang="ts">
  import { onMount } from 'svelte'
  import { resolve } from '$app/paths'
  import { toast } from '$lib/stores/toast.svelte'
  import { assetApi, settingsApi, type Asset, type AssetLookupResult, type Currency } from '$lib/services/api'
  import { Plus, Loader2, Search, Trash2 } from 'lucide-svelte'

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

  let showCreate = $state(false)
  let form = $state(defaultForm())
  let assets = $state<Asset[] | null>(null)
  let loading = $state(true)
  let creating = $state(false)
  let currencies = $state<Currency[]>([])

  let lookupResults = $state<AssetLookupResult[] | null>(null)
  let lookupLoading = $state(false)
  let showSuggestions = $state(false)
  let selectedTicker = $state(false)
  let debounceTimer: ReturnType<typeof setTimeout> | undefined

  onMount(async () => {
    try {
      const [assetList, curList] = await Promise.all([assetApi.list(), settingsApi.listCurrencies()])
      assets = assetList
      currencies = curList.currencies
      if (curList.currencies.length > 0) {
        const preferred = curList.currencies.find((c) => c.code === 'USD') ?? curList.currencies[0]
        form.currency = preferred.code
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load assets'
      toast.error(message)
    } finally {
      loading = false
    }
  })

  async function handleTickerChange(value: string): Promise<void> {
    const ticker = value.toUpperCase()
    form.ticker = ticker
    selectedTicker = false
    showSuggestions = ticker.length >= 2

    clearTimeout(debounceTimer)
    if (ticker.length < 2) {
      lookupResults = null
      lookupLoading = false
      return
    }

    lookupLoading = true
    debounceTimer = setTimeout(async () => {
      try {
        lookupResults = await assetApi.lookup(ticker)
      } catch {
        lookupResults = []
      } finally {
        lookupLoading = false
      }
    }, 350)
  }

  async function selectSuggestion(result: AssetLookupResult): Promise<void> {
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
    selectedTicker = true
    showSuggestions = false

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

  async function createAsset(): Promise<void> {
    creating = true
    try {
      await assetApi.create(form)
      assets = await assetApi.list()
      closeCreate()
      form = defaultForm()
      toast.success('Asset created')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Create failed'
      toast.error(message)
    } finally {
      creating = false
    }
  }

  async function deleteAsset(id: string, ticker: string): Promise<void> {
    if (!confirm(`Delete ${ticker}?`)) return
    try {
      await assetApi.remove(id)
      assets = await assetApi.list()
      toast.success('Asset deleted')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Delete failed'
      toast.error(message)
    }
  }

  function closeCreate(): void {
    showCreate = false
    lookupResults = null
    showSuggestions = false
    selectedTicker = false
  }

  function toggleCreate(): void {
    if (showCreate) {
      closeCreate()
    } else {
      showCreate = true
    }
  }
</script>

<div class="p-6">
  <div class="mb-6 flex items-center justify-between">
    <h1 class="text-2xl font-bold">Assets</h1>
    <button
      onclick={toggleCreate}
      class="flex items-center gap-2 rounded-control bg-accent px-4 py-2 text-sm text-accent-foreground hover:bg-accent-hover"
    >
      <Plus class="h-4 w-4" />
      Add Asset
    </button>
  </div>

  {#if showCreate}
    <div class="mb-6 rounded-card border border-border bg-surface p-4">
      <h2 class="mb-3 font-semibold">New Asset</h2>
      <div class="grid grid-cols-2 gap-3">
        <div class="relative">
          <input
            type="text"
            placeholder="Ticker (e.g. AAPL)"
            value={form.ticker}
            oninput={(e) => handleTickerChange(e.currentTarget.value)}
            onfocus={() => form.ticker.length >= 2 && (showSuggestions = true)}
            onblur={() => setTimeout(() => (showSuggestions = false), 200)}
            class="w-full rounded-control border border-input px-3 py-2 pr-8 text-sm"
          />
          {#if lookupLoading}
            <Loader2 class="absolute right-2 top-2.5 h-4 w-4 animate-spin text-muted-foreground" />
          {:else if form.ticker && !selectedTicker}
            <Search class="absolute right-2 top-2.5 h-4 w-4 text-muted-foreground" />
          {/if}

          {#if showSuggestions && lookupResults && lookupResults.length > 0 && !selectedTicker}
            <div class="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-card border border-border bg-surface shadow-raised">
              {#each lookupResults as r (r.ticker)}
                <button
                  type="button"
                  onmousedown={() => selectSuggestion(r)}
                  class="flex w-full items-center gap-3 px-3 py-2 text-left text-sm hover:bg-muted"
                >
                  <span class="font-medium">{r.ticker}</span>
                  <span class="flex-1 truncate text-muted-foreground">{r.name}</span>
                  <span class="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">{r.exchange}</span>
                  <span class="text-xs text-muted-foreground">{r.type}</span>
                </button>
              {/each}
            </div>
          {/if}

          {#if showSuggestions && lookupResults && lookupResults.length === 0 && !selectedTicker}
            <div class="absolute z-10 mt-1 w-full rounded-card border border-border bg-surface p-3 text-center text-sm text-muted-foreground shadow-raised">
              No results found
            </div>
          {/if}
        </div>

        <input
          type="text"
          placeholder="Name"
          bind:value={form.name}
          class="rounded-control border border-input px-3 py-2 text-sm"
        />
        <input
          type="text"
          placeholder="ISIN (optional)"
          bind:value={form.isin}
          class="rounded-control border border-input px-3 py-2 text-sm"
        />
        <select
          bind:value={form.type}
          class="rounded-control border border-input px-3 py-2 text-sm"
        >
          <option value="stock">Stock</option>
          <option value="etf">ETF</option>
          <option value="bond">Bond</option>
          <option value="mutual_fund">Mutual fund</option>
          <option value="crypto">Crypto</option>
          <option value="commodity">Commodity</option>
        </select>
        <select
          bind:value={form.currency}
          class="rounded-control border border-input px-3 py-2 text-sm"
        >
          {#each currencies as c (c.code)}
            <option value={c.code}>{c.code}</option>
          {/each}
        </select>
        <input
          type="text"
          placeholder="Exchange"
          bind:value={form.exchange}
          class="rounded-control border border-input px-3 py-2 text-sm"
        />
        <select
          bind:value={form.price_source}
          class="rounded-control border border-input px-3 py-2 text-sm"
        >
          <option value="yahoo">Yahoo Finance</option>
          <option value="manual">Prezzo manuale</option>
          <option value="none">Nessun prezzo</option>
        </select>
      </div>
      <button
        onclick={createAsset}
        disabled={!form.ticker || !form.name || creating}
        class="mt-3 rounded-control bg-accent px-4 py-2 text-sm text-accent-foreground hover:bg-accent-hover disabled:opacity-50"
      >
        {creating ? 'Saving...' : 'Save'}
      </button>
    </div>
  {/if}

  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else}
    <table class="w-full text-left text-sm">
      <thead>
        <tr class="border-b border-border text-muted-foreground">
          <th class="pb-2">Ticker</th>
          <th class="pb-2">Name</th>
          <th class="pb-2">Type</th>
          <th class="pb-2">Currency</th>
          <th class="pb-2">Country</th>
          <th class="pb-2 text-right">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each assets ?? [] as a (a.id)}
          <tr class="border-b border-border last:border-0">
            <td class="py-2 font-medium">
              <a href={resolve(`/assets/${a.id}`)} class="text-accent-text hover:underline">{a.ticker}</a>
            </td>
            <td class="py-2 text-muted-foreground">{a.name}</td>
            <td class="py-2">
              <span class="rounded-full bg-muted px-2 py-0.5 text-xs">
                {a.type}
              </span>
            </td>
            <td class="py-2">{a.currency}</td>
            <td class="py-2">{a.country || '-'}</td>
            <td class="py-2 text-right">
              <button
                onclick={() => deleteAsset(a.id, a.ticker)}
                class="rounded-control p-1.5 text-muted-foreground hover:bg-negative/10 hover:text-negative"
                title="Delete asset"
              >
                <Trash2 class="h-4 w-4" />
              </button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
