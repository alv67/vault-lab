<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from '$lib/stores/toast.svelte';
  import { api } from '$lib/services/api';

  interface HealthSummary {
      successes: number;
      failures: number;
      successRate: number;
      rateLimited: number;
  }

  interface HealthEvent {
      id: string;
      asset_id: string;
      event_type: string;
      status: string;
      code: string;
      message: string;
      duration_ms: number;
      error_code: string;
      created_at: string;
  }

  let summary = $state<HealthSummary | null>(null);
  let events = $state<HealthEvent[]>([]);
  let loading = $state(true);

  async function fetchHealth() {
      loading = true;
      try {
          const data = (await api.get('/health/prices')) as {
              summary: HealthSummary;
              events: HealthEvent[];
          };
          summary = data.summary;
          events = data.events ?? [];
      } catch {
          toast.error('Failed to fetch health data');
      } finally {
          loading = false;
      }
  }

  onMount(() => {
      fetchHealth();
  });

  function formatRate(val: number) {
      return (val * 100).toFixed(1) + '%';
  }

  function getStatusColor(status: string) {
      if (status === 'success') return 'text-positive bg-positive/10 border-positive/20';
      if (status === 'rate_limited' || status.includes('429')) return 'text-warning bg-warning/10 border-warning/20';
      return 'text-negative bg-negative/10 border-negative/20';
  }
</script>

<div class="p-6 max-w-6xl mx-auto">
    <div class="flex items-center justify-between mb-8">
        <div>
            <h1 class="text-2xl font-bold text-foreground">Price Sync Health</h1>
            <p class="text-muted-foreground">Monitoring Yahoo Finance API connectivity and performance</p>
        </div>
        <button 
            onclick={fetchHealth} 
            disabled={loading}
            class="px-4 py-2 bg-surface border border-border rounded-control hover:bg-muted disabled:opacity-50 text-sm font-medium"
        >
            {loading ? 'Refreshing...' : 'Refresh Now'}
        </button>
    </div>

    {#if loading}
        <div class="flex justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-accent"></div>
        </div>
    {:else if !summary}
        <div class="text-center py-12 text-muted-foreground">
            No health data available.
        </div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
            <div class="p-4 bg-surface border border-border rounded-card shadow-card">
                <div class="text-sm text-muted-foreground mb-1">Success Rate</div>
                <div class="text-2xl font-bold tabular-nums {summary.successRate > 0.9 ? 'text-positive' : 'text-warning'}">
                    {formatRate(summary.successRate)}
                </div>
            </div>
            <div class="p-4 bg-surface border border-border rounded-card shadow-card">
                <div class="text-sm text-muted-foreground mb-1">Total Successes</div>
                <div class="text-2xl font-bold text-foreground tabular-nums">{summary.successes}</div>
            </div>
            <div class="p-4 bg-surface border border-border rounded-card shadow-card">
                <div class="text-sm text-muted-foreground mb-1">Total Failures</div>
                <div class="text-2xl font-bold text-negative tabular-nums">{summary.failures}</div>
            </div>
            <div class="p-4 bg-surface border border-border rounded-card shadow-card">
                <div class="text-sm text-muted-foreground mb-1">Rate Limited</div>
                <div class="text-2xl font-bold text-warning tabular-nums">{summary.rateLimited}</div>
            </div>
        </div>

        <div class="bg-surface border border-border rounded-card shadow-card overflow-hidden">
            <div class="px-6 py-4 border-b border-border bg-muted">
                <h2 class="font-semibold text-foreground">Recent Events</h2>
            </div>
            <div class="overflow-x-auto">
                <table class="w-full text-left text-sm">
                    <thead>
                        <tr class="bg-muted border-b border-border">
                            <th class="px-6 py-3 font-medium text-muted-foreground">Timestamp</th>
                            <th class="px-6 py-3 font-medium text-muted-foreground">Type</th>
                            <th class="px-6 py-3 font-medium text-muted-foreground">Status</th>
                            <th class="px-6 py-3 font-medium text-muted-foreground">Code</th>
                            <th class="px-6 py-3 font-medium text-muted-foreground">Message</th>
                            <th class="px-6 py-3 font-medium text-muted-foreground text-right">Duration</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each events as event (event.id)}
                            <tr class="border-b border-border last:border-0 hover:bg-muted">
                                <td class="px-6 py-4 text-muted-foreground">
                                    {new Date(event.created_at).toLocaleString()}
                                </td>
                                <td class="px-6 py-4 font-medium">{event.event_type}</td>
                                <td class="px-6 py-4">
                                    <span class="px-2 py-1 rounded-full text-xs font-medium border {getStatusColor(event.status)}">
                                        {event.status}
                                    </span>
                                </td>
                                <td class="px-6 py-4 font-mono text-xs">{event.code || '—'}</td>
                                <td class="px-6 py-4 text-muted-foreground truncate max-w-xs">{event.message || '—'}</td>
                                <td class="px-6 py-4 text-right text-muted-foreground tabular-nums">{event.duration_ms}ms</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}
</div>
