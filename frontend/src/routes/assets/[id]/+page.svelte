<script module lang="ts">
  let sessionRefreshed = false
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { toast } from '$lib/stores/toast.svelte'
  import {
    assetApi,
    pricesApi,
    type Asset,
    type AssetExposure,
    type AssetPatch,
    type AssetQuote,
    type ExposureRow,
    type Price,
    type SplitInfo,
  } from '$lib/services/api'
  import { formatCurrency, formatPercent, ASSET_CLASS_LABELS, PRICE_SOURCE_LABELS } from '$lib/format'
  import PriceChart from '$lib/components/PriceChart.svelte'
  import ExposurePie from '$lib/components/ExposurePie.svelte'
  import { EllipsisVertical, Loader2, Pencil } from 'lucide-svelte'
  import ExposureModal from '$lib/components/ExposureModal.svelte'

  const id = $derived(page.params.id as string | undefined)

  const LEGEND_PALETTE = [
    '#2563eb', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
    '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16',
    '#06b6d4', '#a855f7',
  ]

  const ASSET_TYPES = [
    { value: 'stock', label: 'Stock' },
    { value: 'etf', label: 'ETF' },
    { value: 'bond', label: 'Bond' },
    { value: 'mutual_fund', label: 'Mutual fund' },
    { value: 'crypto', label: 'Crypto' },
    { value: 'commodity', label: 'Commodity' },
  ]

  const METRICS = [
    { key: 'change_1d', label: '1G' },
    { key: 'change_1w', label: '1S' },
    { key: 'change_1m', label: '1M' },
    { key: 'change_1y', label: '1Y' },
    { key: 'change_ytd', label: 'YTD' },
  ] as const

  const RANGES = [
    { key: '1M', days: 30 },
    { key: '3M', days: 90 },
    { key: '1Y', days: 365 },
    { key: 'YTD', days: -1 },
    { key: 'MAX', days: Infinity },
  ] as const
  type RangeKey = (typeof RANGES)[number]['key']

  let loading = $state(true)
  let asset = $state<Asset | null>(null)
  let quote = $state<AssetQuote | null>(null)
  let prices = $state<Price[]>([])
  let splits = $state<SplitInfo[]>([])
  let exposure = $state<AssetExposure | null>(null)
  let regionsEdit = $state<ExposureRow[]>([])
  let sectorsEdit = $state<ExposureRow[]>([])
  let savingRegions = $state(false)
  let savingSectors = $state(false)
  let saving = $state(false)
  let prefilling = $state(false)
  let fetchingETF = $state(false)
  let refreshingMeta = $state(false)
  let backfillingHistory = $state(false)
  let metaMenuOpen = $state(false)
  let exposureModalOpen = $state(false)
  let range = $state<RangeKey | null>('1Y')
  let programmaticallyZooming = $state(false)

  let form = $state({
    ticker: '',
    isin: '',
    name: '',
    type: 'stock',
    currency: 'USD',
    exchange: '',
    asset_class: 'other',
    price_source: 'yahoo',
  })

  const currency = $derived(asset?.currency || 'USD')

  const exposureApplicable = $derived(
    asset
      ? asset.type === 'stock' ||
          ((asset.type === 'etf' || asset.type === 'mutual_fund') &&
            (asset.asset_class === 'equity' || asset.asset_class === 'real_estate'))
      : false,
  )

  const hasChanges = $derived.by(() => {
    if (!asset) return false
    return (
      form.ticker !== asset.ticker ||
      form.isin !== (asset.isin || '') ||
      form.name !== asset.name ||
      form.type !== asset.type ||
      form.currency !== asset.currency ||
      form.exchange !== (asset.exchange || '') ||
      form.asset_class !== (asset.asset_class || 'other') ||
      form.price_source !== (asset.price_source || 'yahoo')
    )
  })

  const chartSeries = $derived.by(() => {
    return [...prices]
      .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
      .map((p) => ({ date: p.date, close: p.close }))
  })

  const zoomStart = $derived.by(() => {
    if (!range) return null
    if (range === 'MAX') return 'MAX' as const
    if (range === 'YTD') {
      const now = new Date()
      const ytd = new Date(now.getFullYear(), 0, 1)
      return ytd.toISOString().slice(0, 10)
    }
    const days = RANGES.find((r) => r.key === range)?.days ?? 365
    const cutoff = Date.now() - days * 24 * 60 * 60 * 1000
    return new Date(cutoff).toISOString().slice(0, 10)
  })

  const sumRegions = $derived(
    regionsEdit.reduce((acc, r) => acc + (Number(r.weight) || 0), 0),
  )
  const sumSectors = $derived(
    sectorsEdit.reduce((acc, r) => acc + (Number(r.weight) || 0), 0),
  )
  const regionsValid = $derived(Math.abs(sumRegions - 100) <= 0.5)
  const sectorsValid = $derived(Math.abs(sumSectors - 100) <= 0.5)

  function changeClass(value: string | number | undefined): string {
    const n = Number(value ?? 0)
    if (n === 0) return 'text-gray-500'
    return n > 0 ? 'text-green-600' : 'text-red-600'
  }

  function selectRange(r: RangeKey): void {
    programmaticallyZooming = true
    range = r
    setTimeout(() => (programmaticallyZooming = false), 300)
  }

  function handleChartZoom(): void {
    if (!programmaticallyZooming) {
      range = null
    }
  }

  function fillForm(a: Asset): void {
    form = {
      ticker: a.ticker,
      isin: a.isin || '',
      name: a.name,
      type: a.type,
      currency: a.currency,
      exchange: a.exchange || '',
      asset_class: a.asset_class || 'other',
      price_source: a.price_source || 'yahoo',
    }
  }

  onMount(load)

  async function load(): Promise<void> {
    if (!id) return
    try {
      const [a, q, ps, ex, sp] = await Promise.all([
        assetApi.get(id),
        assetApi.quote(id),
        pricesApi.byAsset(id),
        assetApi.exposure(id),
        assetApi.splits(id),
      ])
      asset = a
      quote = q
      prices = ps
      exposure = ex
      splits = sp
      regionsEdit = ex.regions.map((r) => ({ ...r }))
      sectorsEdit = ex.sectors.map((r) => ({ ...r }))
      fillForm(a)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load asset'
      toast.error(message)
      const status = (err as { status?: number } | null)?.status
      if (status === 404) {
        goto(resolve('/assets'), { replaceState: true })
        return
      }
    } finally {
      loading = false
    }

    // Refresh prezzi una volta per sessione: la pagina può essere aperta come
    // deep-link senza passare dalla dashboard, che normalmente fa il refresh.
    if (!sessionRefreshed) {
      sessionRefreshed = true
      pricesApi.refresh()
        .then((report) => {
          if (report.rate_limited) {
            toast.warning('Yahoo Finance ha limitato le richieste: alcuni prezzi non aggiornati')
          } else if (report.issues.length > 0) {
            toast.warning(`${report.issues.length} aggiornamenti prezzi non riusciti (Yahoo)`)
          }
          return Promise.all([assetApi.quote(id), pricesApi.byAsset(id)])
        })
        .then(([freshQuote, freshPrices]) => {
          quote = freshQuote
          prices = freshPrices
        })
        .catch(() => { /* keep current data */ })
    }
  }

  async function saveAsset(): Promise<void> {
    if (!id || !asset) return
    if (!form.ticker.trim() || !form.name.trim() || !form.currency.trim()) {
      toast.error('Ticker, Name e Currency sono obbligatori')
      return
    }
    saving = true
    const patch: AssetPatch = {
      ticker: form.ticker.trim(),
      isin: form.isin.trim(),
      name: form.name.trim(),
      type: form.type,
      currency: form.currency.trim(),
      exchange: form.exchange.trim(),
      asset_class: form.asset_class,
      price_source: form.price_source,
    }
    try {
      const updated = await assetApi.update(id, patch)
      asset = updated
      fillForm(updated)
      toast.success('Asset aggiornato')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed'
      toast.error(message)
    } finally {
      saving = false
    }
  }

  async function refreshFromYahoo(): Promise<void> {
    if (!asset) return
    refreshingMeta = true
    metaMenuOpen = false
    try {
      // La classe arriva da meta; l'override manuale deve vincere: se l'asset
      // ha già una classe diversa da "other"/vuota, il refresh non la sovrascrive.
      const meta = await assetApi.meta(asset.ticker)
      form = {
        ...form,
        name: meta.name || form.name,
        type: meta.type || form.type,
        currency: meta.currency || form.currency,
        exchange: meta.exchange || form.exchange,
        asset_class:
          !asset.asset_class || asset.asset_class === 'other'
            ? meta.asset_class || form.asset_class
            : form.asset_class,
      }
      toast.success('Campi aggiornati da Yahoo')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Aggiornamento fallito'
      toast.error(message)
    } finally {
      refreshingMeta = false
    }
  }

  // Backfill sincrono: il POST può impiegare alcuni secondi, ma al 200 la cache
  // GET del client è già stata invalidata (ogni non-GET svuota la cache), quindi
  // il refetch di byAsset ritorna lo storico completo e aggiornato.
  async function backfillHistory(): Promise<void> {
    if (!id) return
    backfillingHistory = true
    metaMenuOpen = false
    try {
      await assetApi.backfillHistory(id)
      prices = await pricesApi.byAsset(id)
      toast.success('Storico prezzi aggiornato')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Backfill fallito'
      toast.error(message)
    } finally {
      backfillingHistory = false
    }
  }

  // Prefill da JustETF: popola SOLO la distribuzione geografica (regioni).
  async function prefillRegionsFromETF(): Promise<void> {
    if (!id || !asset) return
    fetchingETF = true
    try {
      const saved = await assetApi.fetchETFExposure(id)
      exposure = saved
      regionsEdit = saved.regions.map((r) => ({ ...r }))
      if (saved.isin) form.isin = saved.isin
      toast.success('Distribuzione geografica precompilata da JustETF')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Download fallito'
      toast.error(message)
    } finally {
      fetchingETF = false
    }
  }

  // Prefill da JustETF: popola SOLO la distribuzione settoriale.
  async function prefillSectorsFromETF(): Promise<void> {
    if (!id || !asset) return
    fetchingETF = true
    try {
      const saved = await assetApi.fetchETFExposure(id)
      exposure = saved
      sectorsEdit = saved.sectors.map((r) => ({ ...r }))
      if (saved.isin) form.isin = saved.isin
      toast.success('Distribuzione settoriale precompilata da JustETF')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Download fallito'
      toast.error(message)
    } finally {
      fetchingETF = false
    }
  }

  // Prefill da Yahoo: popola SOLO la distribuzione settoriale (topHoldings). Per
  // le azioni singole ricade sul settore unico al 100% (assetProfile).
  async function prefillSectorsFromYahoo(): Promise<void> {
    if (!id) return
    prefilling = true
    try {
      const saved = await assetApi.fetchExposure(id)
      exposure = saved
      sectorsEdit = saved.sectors.map((r) => ({ ...r }))
      toast.success('Distribuzione settoriale precompilata da Yahoo')
    } catch (err: unknown) {
      const status = (err as { status?: number } | null)?.status
      const message =
        status === 502
          ? 'Yahoo non ha risposto'
          : err instanceof Error
            ? err.message
            : 'Prefill fallito'
      toast.error(message)
    } finally {
      prefilling = false
    }
  }

  // Il backend salva ogni dimensione indipendentemente: ogni sezione invia
  // SOLO la propria dimensione (l'altra viene omessa dal JSON → nil → non
  // toccata). Dopo il successo ricarichiamo entrambe le tabelle dalla risposta
  // canonica (regioni + settori memorizzati).
  async function saveRegions(): Promise<void> {
    if (!id || !exposure || !regionsValid) return
    savingRegions = true
    try {
      const saved = await assetApi.saveExposure(id, {
        regions: regionsEdit,
      })
      exposure = saved
      regionsEdit = saved.regions.map((r) => ({ ...r }))
      sectorsEdit = saved.sectors.map((r) => ({ ...r }))
      toast.success('Distribuzione geografica salvata')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed'
      toast.error(message)
    } finally {
      savingRegions = false
    }
  }

  async function saveSectors(): Promise<void> {
    if (!id || !exposure || !sectorsValid) return
    savingSectors = true
    try {
      const saved = await assetApi.saveExposure(id, {
        sectors: sectorsEdit,
      })
      exposure = saved
      regionsEdit = saved.regions.map((r) => ({ ...r }))
      sectorsEdit = saved.sectors.map((r) => ({ ...r }))
      toast.success('Distribuzione settoriale salvata')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed'
      toast.error(message)
    } finally {
      savingSectors = false
    }
  }
</script>

<div class="p-6">
  {#if loading}
    <p class="text-gray-500">Loading...</p>
  {:else if asset}
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">{asset.name}</h1>
        <p class="text-sm text-gray-500">{asset.ticker}</p>
        {#if asset.price_source && asset.price_source !== 'yahoo'}
          <span class="mt-1 inline-block rounded-full bg-amber-100 px-2 py-0.5 text-xs text-amber-700">
            {PRICE_SOURCE_LABELS[asset.price_source] ?? asset.price_source} — nessun sync automatico
          </span>
        {/if}
      </div>
      <div class="flex items-center gap-2">
        <button
          onclick={saveAsset}
          disabled={!hasChanges || saving}
          class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {#if saving}
            <Loader2 class="h-4 w-4 animate-spin" />
          {/if}
          {saving ? 'Salvataggio...' : 'Salva modifiche'}
        </button>
        <button
          onclick={() => goto(resolve('/assets'))}
          class="rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
        >
          Back
        </button>
      </div>
    </div>

    <div class="mb-6 rounded-xl bg-white p-4 shadow">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="font-semibold">Caratteristiche</h2>
        <div class="relative">
          <button
            onclick={() => (metaMenuOpen = !metaMenuOpen)}
            class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            title="Aggiorna da Yahoo"
          >
            <EllipsisVertical class="h-5 w-5" />
          </button>
          {#if metaMenuOpen}
            <div class="absolute right-0 z-10 mt-1 w-56 rounded-lg border bg-white py-1 shadow-lg">
              <button
                onclick={refreshFromYahoo}
                disabled={refreshingMeta}
                class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
              >
                {#if refreshingMeta}
                  <Loader2 class="h-4 w-4 animate-spin" />
                {/if}
                Aggiorna da Yahoo
              </button>
              <button
                onclick={backfillHistory}
                disabled={backfillingHistory}
                class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
              >
                {#if backfillingHistory}
                  <Loader2 class="h-4 w-4 animate-spin" />
                {/if}
                Backfill storico completo
              </button>
            </div>
          {/if}
        </div>
      </div>
      <div class="grid grid-cols-2 gap-3 md:grid-cols-3">
        <div>
          <label for="asset-ticker" class="mb-1 block text-xs font-medium text-gray-500">Ticker</label>
          <input
            id="asset-ticker"
            type="text"
            bind:value={form.ticker}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-isin" class="mb-1 block text-xs font-medium text-gray-500">ISIN</label>
          <input
            id="asset-isin"
            type="text"
            bind:value={form.isin}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-name" class="mb-1 block text-xs font-medium text-gray-500">Name</label>
          <input
            id="asset-name"
            type="text"
            bind:value={form.name}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-type" class="mb-1 block text-xs font-medium text-gray-500">Type</label>
          <select
            id="asset-type"
            bind:value={form.type}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          >
            {#each ASSET_TYPES as t (t.value)}
              <option value={t.value}>{t.label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label for="asset-currency" class="mb-1 block text-xs font-medium text-gray-500">Currency</label>
          <input
            id="asset-currency"
            type="text"
            bind:value={form.currency}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-exchange" class="mb-1 block text-xs font-medium text-gray-500">Exchange</label>
          <input
            id="asset-exchange"
            type="text"
            bind:value={form.exchange}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-class" class="mb-1 block text-xs font-medium text-gray-500">Classe</label>
          <select
            id="asset-class"
            bind:value={form.asset_class}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          >
            {#each Object.entries(ASSET_CLASS_LABELS) as [value, label] (value)}
              <option value={value}>{label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label for="asset-price-source" class="mb-1 block text-xs font-medium text-gray-500">Fonte prezzo</label>
          <select
            id="asset-price-source"
            bind:value={form.price_source}
            class="w-full rounded-lg border px-3 py-2 text-sm"
          >
            <option value="yahoo">Yahoo Finance</option>
            <option value="manual">Prezzo manuale</option>
            <option value="none">Nessun prezzo</option>
          </select>
        </div>
      </div>
    </div>

    <div class="mb-6 rounded-xl bg-white p-4 shadow">
      <h2 class="mb-4 font-semibold">Metriche quote</h2>
      {#if quote?.has_data}
        <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-6">
          <div class="rounded-xl bg-white p-4 shadow">
            <p class="text-sm text-gray-500">Ultima chiusura</p>
            <p class="text-xl font-bold">{formatCurrency(quote.last_close, quote.currency)}</p>
          </div>
          {#each METRICS as m (m.key)}
            <div class="rounded-xl bg-white p-4 shadow">
              <p class="text-sm text-gray-500">{m.label}</p>
              <p class="text-xl font-bold {changeClass(quote[m.key])}">
                {formatPercent(quote[m.key])}
              </p>
            </div>
          {/each}
        </div>
        <p class="mt-3 text-xs text-gray-500">
          Aggiornato il {new Date(quote.last_date).toLocaleDateString()}
        </p>
      {:else}
        <p class="text-sm text-gray-400">Nessun dato prezzo</p>
      {/if}
    </div>

    <div class="mb-6 rounded-xl bg-white p-4 shadow">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h2 class="font-semibold">Storico prezzo</h2>
        <div class="flex gap-1">
          {#each RANGES as r (r.key)}
            <button
              onclick={() => selectRange(r.key)}
              class="rounded-lg px-3 py-1.5 text-sm {range === r.key
                ? 'bg-blue-600 text-white'
                : 'text-gray-600 hover:bg-gray-100'}"
            >
              {r.key}
            </button>
          {/each}
        </div>
      </div>
      <PriceChart
        series={chartSeries}
        {currency}
        zoomStart={zoomStart}
        splits={splits}
        onDataZoom={handleChartZoom}
      />
    </div>

    {#if exposureApplicable && exposure}
      <div class="mb-6 rounded-xl bg-white p-4 shadow">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 class="font-semibold">Distribuzione</h2>
          <div class="flex items-center gap-2">
            <button
              onclick={() => (exposureModalOpen = true)}
              aria-label="Modifica distribuzione"
              class="flex items-center gap-2 rounded-lg border px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-100"
            >
              <Pencil class="h-4 w-4" />
              Modifica
            </button>
          </div>
        </div>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <ExposurePie data={regionsEdit} title="Distribuzione geografica" />
          </div>
          <div>
            <ExposurePie data={sectorsEdit} title="Distribuzione settoriale" />
          </div>
        </div>
        <div class="mt-4 grid grid-cols-2 gap-x-6 gap-y-3 rounded-xl border bg-gray-50 p-4">
          <div>
            <p class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-gray-500">
              Distribuzione geografica
            </p>
            <div class="grid grid-cols-2 gap-x-3 gap-y-1">
              {#each regionsEdit.filter((r) => Number(r.weight) > 0) as r, i (r.name)}
                <div class="flex items-center gap-1.5 text-xs">
                  <span
                    class="inline-block h-2.5 w-2.5 shrink-0 rounded-sm"
                    style="background-color: {LEGEND_PALETTE[i % LEGEND_PALETTE.length]};"
                  ></span>
                  <span class="truncate">{r.name}</span>
                  <span class="ml-auto text-gray-500">{formatPercent(Number(r.weight))}</span>
                </div>
              {/each}
            </div>
          </div>
          <div>
            <p class="mb-1.5 text-xs font-semibold uppercase tracking-wide text-gray-500">
              Distribuzione settoriale
            </p>
            <div class="grid grid-cols-2 gap-x-3 gap-y-1">
              {#each sectorsEdit.filter((r) => Number(r.weight) > 0) as s, i (s.name)}
                <div class="flex items-center gap-1.5 text-xs">
                  <span
                    class="inline-block h-2.5 w-2.5 shrink-0 rounded-sm"
                    style="background-color: {LEGEND_PALETTE[i % LEGEND_PALETTE.length]};"
                  ></span>
                  <span class="truncate">{s.name}</span>
                  <span class="ml-auto text-gray-500">{formatPercent(Number(s.weight))}</span>
                </div>
              {/each}
            </div>
          </div>
        </div>
      </div>

      <ExposureModal
        bind:open={exposureModalOpen}
        onClose={() => (exposureModalOpen = false)}
        bind:regionsEdit
        bind:sectorsEdit
        {sumRegions}
        {sumSectors}
        {regionsValid}
        {sectorsValid}
        {savingRegions}
        {savingSectors}
        {saveRegions}
        {saveSectors}
        {prefilling}
        {fetchingETF}
        {prefillRegionsFromETF}
        {prefillSectorsFromETF}
        {prefillSectorsFromYahoo}
        assetType={asset.type}
      />
    {:else if exposureApplicable === false && asset}
      <div class="mb-6 rounded-xl bg-white p-4 shadow">
        <h2 class="mb-2 font-semibold">Distribuzione geografica e settoriale</h2>
        <p class="text-sm text-gray-500">
          Questa distribuzione si applica solo agli asset azionari (azioni ed ETF/fondi di classe equity).
        </p>
        {#if asset.type !== 'stock'}
          <p class="mt-2 text-sm text-gray-400">
            Imposta la classe 'Azioni' o 'Immobiliare' nelle Caratteristiche per attivarla.
          </p>
        {/if}
      </div>
    {/if}
  {/if}
</div>