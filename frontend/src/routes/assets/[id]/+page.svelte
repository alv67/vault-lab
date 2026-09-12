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
  import { pnlColorClass } from '$lib/ui-colors'
  import { resolvePalette } from '$lib/chartPalette'
  import { resolved } from '$lib/stores/theme.svelte'
  import { countryDisplayName } from '$lib/countryNames'
  import PriceChart from '$lib/components/PriceChart.svelte'
  import ExposurePie from '$lib/components/ExposurePie.svelte'
  import { EllipsisVertical, Loader2, Pencil } from 'lucide-svelte'
  import ExposureGeoModal from '$lib/components/ExposureGeoModal.svelte'
  import ExposureSectorModal from '$lib/components/ExposureSectorModal.svelte'

  const id = $derived(page.params.id as string | undefined)

  // Resolved chart palette: keeps the country bars and the geo/sector legend
  // swatches in sync with the donut colors across theme flips.
  const palette = $derived(resolvePalette(resolved()))

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
  // Working copy of the modals. Prefill handlers write provider data here as a
  // non-persisted preview: the page cards never read these lists, they render
  // the stored `exposure` (see the display* derivations below), so an unsaved
  // preview only lives inside the modal until the user presses Save. The
  // pending copy is scoped to a single modal session: `openGeoModal`/
  // `openSectorModal` re-hydrate these lists (and their provenance state)
  // from the saved `exposure` before every open, so edits left unsaved when
  // the modal was last closed are discarded on reopen.
  let regionsEdit = $state<ExposureRow[]>([])
  let sectorsEdit = $state<ExposureRow[]>([])
  let countriesEdit = $state<ExposureRow[]>([])
  // Data provenance for the geo/sector modal badges: which source currently
  // owns each dimension ('manual' once the user edits it) and when it was
  // last persisted. Hydrated from `ex.provenance` on load and re-hydrated on
  // every modal reopen (see `openGeoModal`/`openSectorModal`), set by the
  // prefill/dirty handlers and confirmed by every save. A null source means
  // unknown (no badge); a null updatedAt means the shown source is not
  // persisted yet (unsaved preview or fresh manual edit), so the badge shows
  // the label without a date until the next successful save.
  let countriesSource = $state<string | null>(null)
  let regionsSource = $state<string | null>(null)
  let sectorsSource = $state<string | null>(null)
  let countriesUpdatedAt = $state<string | null>(null)
  let regionsUpdatedAt = $state<string | null>(null)
  let sectorsUpdatedAt = $state<string | null>(null)
  let savingRegions = $state(false)
  let savingSectors = $state(false)
  let savingCountries = $state(false)
  let saving = $state(false)
  let prefilling = $state(false)
  let fetchingETF = $state(false)
  let fetchingMorningstar = $state(false)
  let derivingRegions = $state(false)
  let refreshingMeta = $state(false)
  let backfillingHistory = $state(false)
  let metaMenuOpen = $state(false)
  let geoModalOpen = $state(false)
  let sectorModalOpen = $state(false)
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

  // Totals are computed on the 2-decimal values that actually enter the edit
  // lists (also mirrored in the modal footer): this keeps the displayed sum,
  // the save validation and the persisted weights identical, so a provider
  // total at float precision (e.g. 100.004) never trips the > 100 guard.
  const sumRegions = $derived(
    Math.round(regionsEdit.reduce((acc, r) => acc + (Number(r.weight) || 0), 0) * 100) / 100,
  )
  const sumSectors = $derived(
    sectorsEdit.reduce((acc, r) => acc + (Number(r.weight) || 0), 0),
  )
  const sumCountries = $derived(
    Math.round(
      countriesEdit.reduce((acc, r) => acc + (Number(r.weight) || 0), 0) * 100,
    ) / 100,
  )
  // Countries and regions may legitimately sum below 100: for regions the
  // backend folds the residual into «Other / Not Classified» at persist time;
  // for countries the residual just stays unattributed (saving countries no
  // longer re-derives the regions — that happens only via «Calcola da paesi»).
  // Only a sum above 100 (± float epsilon) blocks saving. Sectors still
  // require 100 ±0.5.
  const regionsValid = $derived(sumRegions <= 100 + 1e-9)
  const sectorsValid = $derived(Math.abs(sumSectors - 100) <= 0.5)
  const countriesValid = $derived(sumCountries <= 100 + 1e-9)

  /** The hidden fallback region injected server-side at persist time; the UI
   * never displays or edits it. */
  const OTHER_REGION = 'Other / Not Classified'

  /**
   * Round a persisted weight to 2 decimals (the precision the inputs allow and
   * the UI shows): providers return greedy floats (21.26815...) whose raw sum
   * can trip the > 100 guard by fractions of a percent even when the display
   * reads 100.00%.
   */
  function roundWeight(w: string | number): string {
    const n = Number(w)
    return Number.isFinite(n) ? String(Math.round(n * 100) / 100) : '0'
  }

  /**
   * Normalise a provider import whose 2-decimal weights slightly exceed 100:
   * some providers (e.g. JustETF on LYSX.DE) publish pre-rounded weights that
   * sum to 100.01; the backend accepts up to 100.5 (`weightSumMax100`) but the
   * UI save guard blocks anything > 100, so the import would be unsavable.
   * Totals in (100, 100.5] are shaved down to exactly 100 by subtracting the
   * excess from the single heaviest row (first row wins ties, weight kept as a
   * 2-decimal string via `roundWeight`). This is an IMPORT-time fix only: it
   * runs at the bottom of `positiveCountries`/`withoutOther`, which every
   * provider assignment and canonical reload passes through; manual edits
   * bypass these helpers and stay blocked by the guard when they exceed 100.
   * Totals ≤ 100 are returned unchanged (no-op for normal load/save/display),
   * and totals > 100.5 are a genuine provider anomaly the guard must keep
   * surfacing, so they are also left untouched. Negative weights are never
   * introduced: if the excess exceeds the heaviest row (extreme case) the
   * rows are returned unchanged.
   */
  function capAtHundred(rows: ExposureRow[]): ExposureRow[] {
    const total =
      Math.round(rows.reduce((acc, r) => acc + (Number(r.weight) || 0), 0) * 100) / 100
    if (total <= 100 || total > 100.5) return rows
    const excess = Math.round((total - 100) * 100) / 100
    let maxIdx = 0
    for (let i = 1; i < rows.length; i++) {
      if ((Number(rows[i].weight) || 0) > (Number(rows[maxIdx].weight) || 0)) maxIdx = i
    }
    const maxWeight = Number(rows[maxIdx]?.weight) || 0
    if (excess > maxWeight) return rows
    return rows.map((r, i) =>
      i === maxIdx ? { ...r, weight: roundWeight(maxWeight - excess) } : r,
    )
  }

  /**
   * Copy backend/provider country rows keeping only positive weights, rounded
   * to 2 decimals and normalised to ≤ 100 via `capAtHundred` (slightly-over
   * provider totals are shaved at import, manual over-100 edits are not). The
   * exposure endpoints answer with the full canonical zero-filled list, while
   * the edit list must stay minimal (the backend drops non-positive rows on
   * save, so sending only the > 0 rows is lossless).
   */
  function positiveCountries(rows: ExposureRow[]): ExposureRow[] {
    return capAtHundred(
      rows
        .filter((r) => Number(r.weight) > 0)
        .map((r) => ({ ...r, weight: roundWeight(r.weight) })),
    )
  }

  /**
   * Copy backend/provider region rows in canonical order, dropping the
   * «Other / Not Classified» residual, rounding weights and normalising
   * slightly-over-100 provider totals via `capAtHundred`: sums, donut and
   * table then work on visible rows only (the page re-adds the open-donut
   * gap via complete={false}).
   */
  function withoutOther(rows: ExposureRow[]): ExposureRow[] {
    return capAtHundred(
      rows
        .filter((r) => r.name !== OTHER_REGION)
        .map((r) => ({ ...r, weight: roundWeight(r.weight) })),
    )
  }

  /** Copy provider/backend sector rows rounding weights to 2 decimals and
   *  normalising slightly-over-100 totals via capAtHundred (import-time only;
   *  manual edits bypass it and stay governed by the sector guard). */
  function sectorsList(rows: ExposureRow[]): ExposureRow[] {
    return capAtHundred(rows.map((r) => ({ ...r, weight: roundWeight(r.weight) })))
  }

  // ---------------------------------------------------------------------------
  // Display vs edit split: the cards always render the STORED exposure (the
  // `GET /assets/{id}/exposure` response, refreshed by `load` and by every
  // successful save). The edit lists are the modals' working copy and may hold
  // unsaved manual edits or provider prefill previews while the modal is open;
  // they never leak into the cards. Each "Modifica" button first restores its
  // lists from the saved exposure via `openGeoModal`/`openSectorModal` (the
  // same hydration `load` does), so reopening a modal after closing without
  // saving discards the pending changes and starts from persisted data. After
  // a save the canonical response updates `exposure` and re-syncs the saved
  // dimension's edit list, so card and modal become consistent again (and the
  // next reopen restores from `exposure` anyway).
  // ---------------------------------------------------------------------------
  const displayCountries = $derived(
    (exposure?.countries ?? []).filter((c) => Number(c.weight) > 0),
  )
  // Stored regions without the «Other / Not Classified» residual (same filter
  // the edit list applies at assignment time, here re-used for display).
  const displayRegions = $derived(exposure ? withoutOther(exposure.regions) : [])
  const displaySectors = $derived(exposure?.sectors ?? [])

  // Top 15 stored countries by weight (desc, > 0) for the geographic card bar
  // list. The copy-then-sort keeps the derived displayCountries array pristine.
  const topCountries = $derived.by(() =>
    [...displayCountries]
      .sort((a, b) => Number(b.weight) - Number(a.weight))
      .slice(0, 15),
  )
  // Largest visible weight: bars are scaled proportionally against it.
  const maxCountryWeight = $derived(
    topCountries.reduce((max, c) => Math.max(max, Number(c.weight) || 0), 0),
  )

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
      regionsEdit = withoutOther(ex.regions)
      sectorsEdit = sectorsList(ex.sectors)
      countriesEdit = positiveCountries(ex.countries)
      // Hydrate the persisted provenance per dimension (source + last-update
      // date). Dimensions never persisted carry no provenance entry: badge
      // hidden (source null).
      countriesSource = ex.provenance?.countries?.source ?? null
      countriesUpdatedAt = ex.provenance?.countries?.updated_at ?? null
      regionsSource = ex.provenance?.regions?.source ?? null
      regionsUpdatedAt = ex.provenance?.regions?.updated_at ?? null
      sectorsSource = ex.provenance?.sectors?.source ?? null
      sectorsUpdatedAt = ex.provenance?.sectors?.updated_at ?? null
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

  // Prefill da JustETF: popola SOLO la lista paesi della modale (raw JustETF).
  // È un'anteprima NON persistita: `exposure` (e quindi le card) resta ai dati
  // salvati; solo «Salva paesi» invierà la lista al backend. L'ISIN arriva già
  // persistito dal server, quindi lo sincronizziamo nel form.
  async function prefillCountriesFromETF(): Promise<void> {
    if (!id || !asset) return
    fetchingETF = true
    try {
      const preview = await assetApi.fetchETFExposure(id)
      countriesEdit = positiveCountries(preview.countries)
      countriesSource = 'justetf'
      // Unsaved preview: no persisted date yet (badge shows the label only).
      countriesUpdatedAt = null
      if (preview.isin) form.isin = preview.isin
      toast.success('Paesi precompilati da JustETF')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Download fallito'
      toast.error(message)
    } finally {
      fetchingETF = false
    }
  }

  // Deriva le regioni canoniche dai paesi correnti (preview, non persistito).
  async function deriveRegionsFromCountries(): Promise<void> {
    if (!id || !asset) return
    if (!countriesEdit.some((c) => Number(c.weight) > 0)) {
      toast.error('Nessun paese con peso: aggiungi paesi prima')
      return
    }
    derivingRegions = true
    try {
      const result = await assetApi.deriveRegions(id, countriesEdit)
      regionsEdit = withoutOther(result.regions)
      // The residual «Other / Not Classified» row is filtered out: the modal's
      // totals line already explains what is left unattributed.
      regionsSource = countriesSource === 'justetf' ? 'derived-etf' : 'derived'
      // Preview only: drop any previously persisted date until it is saved.
      regionsUpdatedAt = null
      toast.success('Regioni ricalcolate dai paesi')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Calcolo fallito'
      toast.error(message)
    } finally {
      derivingRegions = false
    }
  }

  // Prefill da Morningstar: popola SOLO la lista regioni della modale con le
  // regioni ufficiali Morningstar. Anteprima NON persistita: la card regioni
  // continua a mostrare le regioni salvate finché non si preme «Salva regioni».
  async function prefillRegionsFromMorningstar(): Promise<void> {
    if (!id || !asset) return
    fetchingMorningstar = true
    try {
      const preview = await assetApi.fetchMorningstarExposure(id)
      regionsEdit = withoutOther(preview.regions)
      regionsSource = 'morningstar-regions'
      regionsUpdatedAt = null
      if (preview.isin) form.isin = preview.isin
      toast.success('Regioni precompilate da Morningstar')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Download fallito'
      toast.error(message)
    } finally {
      fetchingMorningstar = false
    }
  }

  // Prefill da JustETF: popola SOLO la lista settori della modale. Anteprima
  // NON persistita: la card settori resta ai dati salvati finché non si salva.
  async function prefillSectorsFromETF(): Promise<void> {
    if (!id || !asset) return
    fetchingETF = true
    try {
      const preview = await assetApi.fetchETFExposure(id)
      sectorsEdit = sectorsList(preview.sectors)
      sectorsSource = 'justetf'
      sectorsUpdatedAt = null
      if (preview.isin) form.isin = preview.isin
      toast.success('Distribuzione settoriale precompilata da JustETF')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Download fallito'
      toast.error(message)
    } finally {
      fetchingETF = false
    }
  }

  // Prefill da Yahoo: popola SOLO la lista settori della modale (topHoldings).
  // Per le azioni singole ricade sul settore unico al 100% (assetProfile).
  // Anteprima NON persistita: la card settori mostra i dati salvati finché non
  // si preme «Salva».
  async function prefillSectorsFromYahoo(): Promise<void> {
    if (!id) return
    prefilling = true
    try {
      const preview = await assetApi.fetchExposure(id)
      sectorsEdit = sectorsList(preview.sectors)
      sectorsSource = 'yahoo'
      sectorsUpdatedAt = null
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

  // Prefill da Morningstar: popola SOLO la lista settori della modale. Anteprima
  // NON persistita: la card settori resta ai dati salvati finché non si salva
  // (stesso pattern dei prefill settori JustETF/Yahoo). L'endpoint è cachato per
  // ISIN lato backend: se paesi/regioni sono già stati letti la chiamata è
  // immediata.
  async function prefillSectorsFromMorningstar(): Promise<void> {
    if (!id || !asset) return
    fetchingMorningstar = true
    try {
      const preview = await assetApi.fetchMorningstarExposure(id)
      sectorsEdit = sectorsList(preview.sectors)
      sectorsSource = 'morningstar'
      sectorsUpdatedAt = null
      if (preview.isin) form.isin = preview.isin
      toast.success('Distribuzione settoriale precompilata da Morningstar')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Download fallito'
      toast.error(message)
    } finally {
      fetchingMorningstar = false
    }
  }

  // Il backend salva ogni dimensione indipendentemente: ogni sezione invia
  // SOLO la propria dimensione (l'altra viene omessa dal JSON → nil → non
  // toccata) insieme alla sua fonte di provenienza (`*_source`; senza fonte
  // il backend usa 'manual'). Dopo il successo la risposta canonica rinfresca
  // `exposure` (quindi le card), risincronizza SOLO la lista di edit della
  // dimensione salvata e ne riporta la provenance persistita (source +
  // data): le altre liste di edit sono la working copy della modale e
  // possono contenere modifiche pendenti non salvate, quindi non vanno
  // mai toccate qui.
  async function saveRegions(): Promise<void> {
    if (!id || !exposure || !regionsValid) return
    savingRegions = true
    const sentSource = regionsSource ?? 'manual'
    try {
      const saved = await assetApi.saveExposure(id, {
        regions: regionsEdit,
        regions_source: sentSource,
      })
      exposure = saved
      regionsEdit = withoutOther(saved.regions)
      // Provenance confirmed by the canonical response; fall back to the
      // sent source if the backend did not echo it, and drop the date when
      // absent so the badge never shows a stale timestamp.
      regionsSource = saved.provenance?.regions?.source ?? sentSource
      regionsUpdatedAt = saved.provenance?.regions?.updated_at ?? null
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
    const sentSource = sectorsSource ?? 'manual'
    try {
      const saved = await assetApi.saveExposure(id, {
        sectors: sectorsEdit,
        sectors_source: sentSource,
      })
      exposure = saved
      sectorsEdit = sectorsList(saved.sectors)
      sectorsSource = saved.provenance?.sectors?.source ?? sentSource
      sectorsUpdatedAt = saved.provenance?.sectors?.updated_at ?? null
      toast.success('Distribuzione settoriale salvata')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed'
      toast.error(message)
    } finally {
      savingSectors = false
    }
  }

  // Prefill da Morningstar: popola paesi e settori della modale (ETF only).
  // Anteprima NON persistita: le card restano ai dati salvati; ogni dimensione
  // va salvata con il proprio pulsante Salva.
  async function prefillCountriesFromMorningstar(): Promise<void> {
    if (!id || !asset) return
    fetchingMorningstar = true
    try {
      const preview = await assetApi.fetchMorningstarExposure(id)
      countriesEdit = positiveCountries(preview.countries)
      countriesSource = 'morningstar'
      countriesUpdatedAt = null
      sectorsEdit = sectorsList(preview.sectors)
      sectorsSource = 'morningstar'
      sectorsUpdatedAt = null
      if (preview.isin) form.isin = preview.isin
      toast.success('Paesi e settori precompilati da Morningstar')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Download fallito'
      toast.error(message)
    } finally {
      fetchingMorningstar = false
    }
  }

  async function saveCountries(): Promise<void> {
    if (!id || !exposure || !countriesValid) return
    savingCountries = true
    const sentSource = countriesSource ?? 'manual'
    try {
      const saved = await assetApi.saveExposure(id, {
        countries: countriesEdit,
        countries_source: sentSource,
      })
      exposure = saved
      countriesEdit = positiveCountries(saved.countries)
      countriesSource = saved.provenance?.countries?.source ?? sentSource
      countriesUpdatedAt = saved.provenance?.countries?.updated_at ?? null
      // Regions and sectors are deliberately NOT touched here: saving one
      // dimension must never clobber the other dimensions' edit lists, which
      // are the modal's working copy and may hold unsaved pending edits
      // (manual or from a prefill). The stored data still refreshes via
      // `exposure` (cards), and regionsSource is left alone so the regions
      // provenance badge keeps reflecting its real source.
      toast.success('Distribuzione paesi salvata')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed'
      toast.error(message)
    } finally {
      savingCountries = false
    }
  }

  // Provenance flips to 'manual' on the first user mutation of a dimension
  // (invoked by the geo/sector modals at every add/remove/weight-edit point).
  // The persisted date is dropped as well: the badge shows a date only once
  // the new state has actually been saved.
  function markCountriesManual(): void {
    countriesSource = 'manual'
    countriesUpdatedAt = null
  }

  function markRegionsManual(): void {
    regionsSource = 'manual'
    regionsUpdatedAt = null
  }

  function markSectorsManual(): void {
    sectorsSource = 'manual'
    sectorsUpdatedAt = null
  }

  // Open the geo modal after resetting its working copy from the SAVED
  // exposure: the same hydration `load` performs (`withoutOther`/
  // `positiveCountries` normalisation + persisted provenance with the `?? null`
  // fallbacks). Any unsaved manual edits or prefill previews left over from
  // the previous modal session — including a 'manual' provenance flip that was
  // never persisted — are discarded here, so reopening always shows the
  // stored data.
  function openGeoModal(): void {
    if (!exposure) return
    regionsEdit = withoutOther(exposure.regions)
    countriesEdit = positiveCountries(exposure.countries)
    countriesSource = exposure.provenance?.countries?.source ?? null
    countriesUpdatedAt = exposure.provenance?.countries?.updated_at ?? null
    regionsSource = exposure.provenance?.regions?.source ?? null
    regionsUpdatedAt = exposure.provenance?.regions?.updated_at ?? null
    geoModalOpen = true
  }

  // Sector-modal counterpart of `openGeoModal`: restores `sectorsEdit` and the
  // sectors provenance from the saved exposure before opening, so unsaved
  // sector edits/previews never resurface on reopen.
  function openSectorModal(): void {
    if (!exposure) return
    sectorsEdit = sectorsList(exposure.sectors)
    sectorsSource = exposure.provenance?.sectors?.source ?? null
    sectorsUpdatedAt = exposure.provenance?.sectors?.updated_at ?? null
    sectorModalOpen = true
  }
</script>

<div class="p-6">
  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else if asset}
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">{asset.name}</h1>
        <p class="text-sm text-muted-foreground">{asset.ticker}</p>
        {#if asset.price_source && asset.price_source !== 'yahoo'}
          <span class="mt-1 inline-block rounded-full bg-warning/10 px-2 py-0.5 text-xs text-warning">
            {PRICE_SOURCE_LABELS[asset.price_source] ?? asset.price_source} — nessun sync automatico
          </span>
        {/if}
      </div>
      <div class="flex items-center gap-2">
        <button
          onclick={saveAsset}
          disabled={!hasChanges || saving}
          class="flex items-center gap-2 rounded-control bg-accent px-4 py-2 text-sm text-accent-foreground hover:bg-accent-hover disabled:opacity-50"
        >
          {#if saving}
            <Loader2 class="h-4 w-4 animate-spin" />
          {/if}
          {saving ? 'Salvataggio...' : 'Salva modifiche'}
        </button>
        <button
          onclick={() => goto(resolve('/assets'))}
          class="rounded-control border border-border px-4 py-2 text-sm text-foreground hover:bg-muted"
        >
          Back
        </button>
      </div>
    </div>

    <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="font-semibold">Caratteristiche</h2>
        <div class="relative">
          <button
            onclick={() => (metaMenuOpen = !metaMenuOpen)}
            class="rounded-control p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
            title="Aggiorna da Yahoo"
          >
            <EllipsisVertical class="h-5 w-5" />
          </button>
          {#if metaMenuOpen}
            <div class="absolute right-0 z-10 mt-1 w-56 rounded-card border border-border bg-surface py-1 shadow-raised">
              <button
                onclick={refreshFromYahoo}
                disabled={refreshingMeta}
                class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-foreground hover:bg-muted disabled:opacity-50"
              >
                {#if refreshingMeta}
                  <Loader2 class="h-4 w-4 animate-spin" />
                {/if}
                Aggiorna da Yahoo
              </button>
              <button
                onclick={backfillHistory}
                disabled={backfillingHistory}
                class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-foreground hover:bg-muted disabled:opacity-50"
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
          <label for="asset-ticker" class="mb-1 block text-xs font-medium text-muted-foreground">Ticker</label>
          <input
            id="asset-ticker"
            type="text"
            bind:value={form.ticker}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-isin" class="mb-1 block text-xs font-medium text-muted-foreground">ISIN</label>
          <input
            id="asset-isin"
            type="text"
            bind:value={form.isin}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-name" class="mb-1 block text-xs font-medium text-muted-foreground">Name</label>
          <input
            id="asset-name"
            type="text"
            bind:value={form.name}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-type" class="mb-1 block text-xs font-medium text-muted-foreground">Type</label>
          <select
            id="asset-type"
            bind:value={form.type}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          >
            {#each ASSET_TYPES as t (t.value)}
              <option value={t.value}>{t.label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label for="asset-currency" class="mb-1 block text-xs font-medium text-muted-foreground">Currency</label>
          <input
            id="asset-currency"
            type="text"
            bind:value={form.currency}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-exchange" class="mb-1 block text-xs font-medium text-muted-foreground">Exchange</label>
          <input
            id="asset-exchange"
            type="text"
            bind:value={form.exchange}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label for="asset-class" class="mb-1 block text-xs font-medium text-muted-foreground">Classe</label>
          <select
            id="asset-class"
            bind:value={form.asset_class}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          >
            {#each Object.entries(ASSET_CLASS_LABELS) as [value, label] (value)}
              <option value={value}>{label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label for="asset-price-source" class="mb-1 block text-xs font-medium text-muted-foreground">Fonte prezzo</label>
          <select
            id="asset-price-source"
            bind:value={form.price_source}
            class="w-full rounded-control border border-input px-3 py-2 text-sm"
          >
            <option value="yahoo">Yahoo Finance</option>
            <option value="manual">Prezzo manuale</option>
            <option value="none">Nessun prezzo</option>
          </select>
        </div>
      </div>
    </div>

    <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
      <h2 class="mb-4 font-semibold">Metriche quote</h2>
      {#if quote?.has_data}
        <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-6">
          <div class="rounded-card border-border bg-surface p-4 shadow-card">
            <p class="text-sm text-muted-foreground">Ultima chiusura</p>
            <p class="text-xl font-bold tabular-nums">{formatCurrency(quote.last_close, quote.currency)}</p>
          </div>
          {#each METRICS as m (m.key)}
            <div class="rounded-card border-border bg-surface p-4 shadow-card">
              <p class="text-sm text-muted-foreground">{m.label}</p>
              <p class="text-xl font-bold tabular-nums {pnlColorClass(quote[m.key], 'text-muted-foreground')}">
                {formatPercent(quote[m.key])}
              </p>
            </div>
          {/each}
        </div>
        <p class="mt-3 text-xs text-muted-foreground">
          Aggiornato il {new Date(quote.last_date).toLocaleDateString()}
        </p>
      {:else}
        <p class="text-sm text-muted-foreground">Nessun dato prezzo</p>
      {/if}
    </div>

    <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h2 class="font-semibold">Storico prezzo</h2>
        <div class="flex gap-1">
          {#each RANGES as r (r.key)}
            <button
              onclick={() => selectRange(r.key)}
              class="rounded-control px-3 py-1.5 text-sm {range === r.key
                ? 'bg-accent text-accent-foreground'
                : 'text-muted-foreground hover:bg-muted'}"
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
      <!-- Geographic distribution card (stored exposure): countries bar list +
           regions pie -->
      <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 class="font-semibold">Distribuzione geografica</h2>
          <button
            onclick={openGeoModal}
            aria-label="Modifica distribuzione geografica"
            class="flex items-center gap-2 rounded-control border border-border px-3 py-1.5 text-sm text-foreground hover:bg-muted"
          >
            <Pencil class="h-4 w-4" />
            Modifica
          </button>
        </div>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div class="rounded-card border border-border bg-muted p-4">
            <h3 class="mb-2 font-medium">Paesi</h3>
            {#if topCountries.length === 0}
              <div class="flex h-[240px] w-full items-center justify-center text-sm text-muted-foreground">
                Nessuna distribuzione
              </div>
            {:else}
              <div class="space-y-2 py-1">
                {#each topCountries as c, i (c.name)}
                  {@const weight = Number(c.weight) || 0}
                  {@const barPct = maxCountryWeight > 0 ? (weight / maxCountryWeight) * 100 : 0}
                  <div class="flex items-center gap-2 text-xs">
                    <span
                      class="w-28 shrink-0 truncate sm:w-36"
                      title={c.name + ' — ' + countryDisplayName(c.name)}
                    >{countryDisplayName(c.name)}</span>
                    <div class="h-2.5 min-w-0 flex-1 overflow-hidden rounded-full bg-input">
                      <div
                        class="h-full rounded-full"
                        style="width: {barPct.toFixed(1)}%; background-color: {palette[i % palette.length]};"
                      ></div>
                    </div>
                    <span class="w-14 shrink-0 text-right text-muted-foreground tabular-nums">{formatPercent(weight)}</span>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
          <div class="rounded-card border border-border bg-muted p-4">
            <h3 class="mb-2 font-medium">Regioni</h3>
            <!-- displayRegions never carries the «Other / Not Classified» row
                 (withoutOther filters it out of the stored exposure), so the
                 donut renders open: complete={false} adds the transparent
                 residual gap. -->
            <ExposurePie data={displayRegions} title="Distribuzione geografica" complete={false} />
            <div class="mt-3 grid grid-cols-2 gap-x-3 gap-y-1">
              {#each displayRegions.filter((r) => Number(r.weight) > 0) as r, i (r.name)}
                <div class="flex items-center gap-1.5 text-xs">
                  <span
                    class="inline-block h-2.5 w-2.5 shrink-0 rounded-sm"
                    style="background-color: {palette[i % palette.length]};"
                  ></span>
                  <span class="truncate">{r.name}</span>
                  <span class="ml-auto text-muted-foreground tabular-nums">{formatPercent(Number(r.weight))}</span>
                </div>
              {/each}
            </div>
          </div>
        </div>
      </div>

      <!-- Sector distribution card -->
      <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 class="font-semibold">Distribuzione settoriale</h2>
          <button
            onclick={openSectorModal}
            aria-label="Modifica distribuzione settoriale"
            class="flex items-center gap-2 rounded-control border border-border px-3 py-1.5 text-sm text-foreground hover:bg-muted"
          >
            <Pencil class="h-4 w-4" />
            Modifica
          </button>
        </div>
        <div class="rounded-card border border-border bg-muted p-4">
          <h3 class="mb-2 font-medium">Settori</h3>
          <ExposurePie data={displaySectors} title="Distribuzione settoriale" />
          <div class="mt-3 grid grid-cols-2 gap-x-3 gap-y-1">
            {#each displaySectors.filter((r) => Number(r.weight) > 0) as s, i (s.name)}
              <div class="flex items-center gap-1.5 text-xs">
                <span
                  class="inline-block h-2.5 w-2.5 shrink-0 rounded-sm"
                  style="background-color: {palette[i % palette.length]};"
                ></span>
                <span class="truncate">{s.name}</span>
                <span class="ml-auto text-muted-foreground tabular-nums">{formatPercent(Number(s.weight))}</span>
              </div>
            {/each}
          </div>
        </div>
      </div>

      <ExposureGeoModal
        bind:open={geoModalOpen}
        onClose={() => (geoModalOpen = false)}
        bind:regionsEdit
        bind:countriesEdit
        {sumRegions}
        {sumCountries}
        {savingRegions}
        {savingCountries}
        {saveRegions}
        {saveCountries}
        {fetchingETF}
        {fetchingMorningstar}
        {derivingRegions}
        {prefillCountriesFromETF}
        {deriveRegionsFromCountries}
        {prefillRegionsFromMorningstar}
        {prefillCountriesFromMorningstar}
        {countriesSource}
        {regionsSource}
        {countriesUpdatedAt}
        {regionsUpdatedAt}
        onCountriesDirty={markCountriesManual}
        onRegionsDirty={markRegionsManual}
        assetType={asset.type}
      />

      <ExposureSectorModal
        bind:open={sectorModalOpen}
        onClose={() => (sectorModalOpen = false)}
        bind:sectorsEdit
        {sumSectors}
        {sectorsValid}
        {savingSectors}
        {saveSectors}
        {prefilling}
        {fetchingETF}
        {fetchingMorningstar}
        {prefillSectorsFromETF}
        {prefillSectorsFromYahoo}
        {prefillSectorsFromMorningstar}
        {sectorsSource}
        {sectorsUpdatedAt}
        onSectorsDirty={markSectorsManual}
        assetType={asset.type}
      />
    {:else if exposureApplicable === false && asset}
      <div class="mb-6 rounded-card border-border bg-surface p-4 shadow-card">
        <h2 class="mb-2 font-semibold">Distribuzione geografica e settoriale</h2>
        <p class="text-sm text-muted-foreground">
          Questa distribuzione si applica solo agli asset azionari (azioni ed ETF/fondi di classe equity).
        </p>
        {#if asset.type !== 'stock'}
          <p class="mt-2 text-sm text-muted-foreground">
            Imposta la classe 'Azioni' o 'Immobiliare' nelle Caratteristiche per attivarla.
          </p>
        {/if}
      </div>
    {/if}
  {/if}
</div>