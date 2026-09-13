<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { api } from '$lib/services/api'
  import SettingsTabs from '$lib/components/domain/SettingsTabs.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte'
  import Spinner from '$lib/components/ui/Spinner.svelte'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'

  interface HealthSummary {
    successes: number
    failures: number
    success_rate: number
    rate_limited: number
    has_data: boolean
    period?: string
  }

  interface HealthEvent {
    id: string
    asset_id: string
    event_type: string
    status: string
    code: string
    message: string
    duration_ms: number
    error_code: string
    created_at: string
  }

  const PAGE_SIZE = 50

  const periodItems = [
    { value: 'today', label: 'Today' },
    { value: '24h', label: 'Last 24h' },
    { value: '100', label: 'Last 100' },
  ]

  let period = $state('today')
  let offset = $state(0)
  let summary = $state<HealthSummary | null>(null)
  let events = $state<HealthEvent[]>([])
  let eventsTotal = $state(0)
  let loading = $state(true)

  const periodLabel = $derived(periodItems.find((item) => item.value === period)?.label ?? period)
  const rangeLabel = $derived(
    events.length === 0
      ? `0 of ${eventsTotal}`
      : `${offset + 1}–${offset + events.length} of ${eventsTotal}`,
  )

  async function fetchHealth() {
    loading = true
    try {
      const data = (await api.get(
        `/health/prices?period=${period}&limit=${PAGE_SIZE}&offset=${offset}`,
      )) as {
        summary: HealthSummary
        events: HealthEvent[]
        events_total: number
      }
      summary = data.summary
      events = data.events ?? []
      eventsTotal = data.events_total ?? 0
    } catch {
      toast.error('Failed to fetch health data')
    } finally {
      loading = false
    }
  }

  $effect(() => {
    fetchHealth()
  })

  function getPeriod() {
    return period
  }

  function setPeriod(value: string) {
    period = value
    offset = 0
  }

  function formatRate(val: number | null | undefined) {
    if (val === null || val === undefined || Number.isNaN(val)) return 'N/A'
    return (val * 100).toFixed(1) + '%'
  }

  function getStatusColor(status: string) {
    if (status === 'success') return 'text-positive bg-positive/10 border-positive/20'
    if (status === 'rate_limited' || status.includes('429')) return 'text-warning bg-warning/10 border-warning/20'
    return 'text-negative bg-negative/10 border-negative/20'
  }
</script>

<div class="mx-auto max-w-6xl p-6">
  <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
    <div>
      <h1 class="text-2xl font-bold text-foreground">Price Sync Health</h1>
      <p class="text-muted-foreground">Monitoring Yahoo Finance API connectivity and performance</p>
    </div>
    <div class="flex flex-wrap items-center gap-3">
      <SegmentedControl items={periodItems} bind:value={getPeriod, setPeriod} ariaLabel="Health period" />
      <Button variant="secondary" onclick={fetchHealth} disabled={loading}>
        {loading ? 'Refreshing...' : 'Refresh Now'}
      </Button>
    </div>
  </div>

  <SettingsTabs class="mb-8" />

  {#if loading}
    <div class="flex justify-center py-12">
      <Spinner size="lg" class="text-accent-text" />
    </div>
  {:else if !summary}
    <div class="py-12 text-center text-muted-foreground">
      No health data available.
    </div>
  {:else}
    <div class="mb-3 text-sm text-muted-foreground">Period: {periodLabel}</div>
    <div class="mb-8 grid grid-cols-1 gap-4 md:grid-cols-4">
      <div class="rounded-card border border-border bg-surface p-4 shadow-card">
        <div class="mb-1 text-sm text-muted-foreground">Success Rate</div>
        {#if !summary.has_data}
          <div class="text-2xl font-bold tabular-nums text-muted-foreground">N/A</div>
        {:else}
          <div class="text-2xl font-bold tabular-nums {summary.success_rate > 0.9 ? 'text-positive' : 'text-warning'}">
            {formatRate(summary.success_rate)}
          </div>
        {/if}
      </div>
      <div class="rounded-card border border-border bg-surface p-4 shadow-card">
        <div class="mb-1 text-sm text-muted-foreground">Total Successes</div>
        <div class="text-2xl font-bold tabular-nums text-foreground">{summary.successes}</div>
      </div>
      <div class="rounded-card border border-border bg-surface p-4 shadow-card">
        <div class="mb-1 text-sm text-muted-foreground">Total Failures</div>
        <div class="text-2xl font-bold tabular-nums text-negative">{summary.failures}</div>
      </div>
      <div class="rounded-card border border-border bg-surface p-4 shadow-card">
        <div class="mb-1 text-sm text-muted-foreground">Rate Limited</div>
        <div class="text-2xl font-bold tabular-nums text-warning">{summary.rate_limited}</div>
      </div>
    </div>

    <div class="overflow-hidden rounded-card border border-border bg-surface shadow-card">
      <div class="border-b border-border bg-muted px-6 py-4">
        <h2 class="font-semibold text-foreground">Recent Events</h2>
      </div>
      <div class="overflow-x-auto px-6 pb-4">
        <Table aria-label="Recent events">
          <THead>
            <Tr>
              <Th>Timestamp</Th>
              <Th>Type</Th>
              <Th>Status</Th>
              <Th>Code</Th>
              <Th>Message</Th>
              <Th align="right">Duration</Th>
            </Tr>
          </THead>
          <TBody>
            {#each events as event (event.id)}
              <Tr class="hover:bg-muted">
                <Td class="text-muted-foreground">{new Date(event.created_at).toLocaleString()}</Td>
                <Td class="font-medium">{event.event_type}</Td>
                <Td>
                  <span class="rounded-full border px-2 py-1 text-xs font-medium {getStatusColor(event.status)}">
                    {event.status}
                  </span>
                </Td>
                <Td class="font-mono text-xs">{event.code || '—'}</Td>
                <Td class="max-w-xs truncate text-muted-foreground">{event.message || '—'}</Td>
                <Td align="right" class="text-muted-foreground">{event.duration_ms}ms</Td>
              </Tr>
            {/each}
          </TBody>
        </Table>
      </div>
      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-border px-6 py-4">
        <span class="text-sm text-muted-foreground tabular-nums">{rangeLabel}</span>
        <div class="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            disabled={offset === 0}
            onclick={() => (offset = Math.max(0, offset - PAGE_SIZE))}
          >
            Previous
          </Button>
          <Button
            variant="secondary"
            size="sm"
            disabled={offset + PAGE_SIZE >= eventsTotal}
            onclick={() => (offset += PAGE_SIZE)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  {/if}
</div>
