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
          const data = await api.get<{ summary: HealthSummary; events: HealthEvent[] }>('/health/prices');
          summary = data.summary;
          events = data.events ?? [];
      } catch (err) {
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
      if (status === 'success') return 'text-green-600 bg-green-50 border-green-200';
      if (status === 'rate_limited' || status.includes('429')) return 'text-orange-600 bg-orange-50 border-orange-200';
      return 'text-red-600 bg-red-50 border-red-200';
  }
</script>

<div class="p-6 max-w-6xl mx-auto">
    <div class="flex items-center justify-between mb-8">
        <div>
            <h1 class="text-2xl font-bold text-gray-900">Price Sync Health</h1>
            <p class="text-gray-500">Monitoring Yahoo Finance API connectivity and performance</p>
        </div>
        <button 
            onclick={fetchHealth} 
            disabled={loading}
            class="px-4 py-2 bg-white border rounded-lg hover:bg-gray-50 disabled:opacity-50 text-sm font-medium"
        >
            {loading ? 'Refreshing...' : 'Refresh Now'}
        </button>
    </div>

    {#if loading}
        <div class="flex justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        </div>
    {:else if !summary}
        <div class="text-center py-12 text-gray-500">
            No health data available.
        </div>
    {:else}
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
            <div class="p-4 bg-white border rounded-xl shadow-sm">
                <div class="text-sm text-gray-500 mb-1">Success Rate</div>
                <div class="text-2xl font-bold {summary.successRate > 0.9 ? 'text-green-600' : 'text-orange-600'}">
                    {formatRate(summary.successRate)}
                </div>
            </div>
            <div class="p-4 bg-white border rounded-xl shadow-sm">
                <div class="text-sm text-gray-500 mb-1">Total Successes</div>
                <div class="text-2xl font-bold text-gray-900">{summary.successes}</div>
            </div>
            <div class="p-4 bg-white border rounded-xl shadow-sm">
                <div class="text-sm text-gray-500 mb-1">Total Failures</div>
                <div class="text-2xl font-bold text-red-600">{summary.failures}</div>
            </div>
            <div class="p-4 bg-white border rounded-xl shadow-sm">
                <div class="text-sm text-gray-500 mb-1">Rate Limited</div>
                <div class="text-2xl font-bold text-orange-600">{summary.rateLimited}</div>
            </div>
        </div>

        <div class="bg-white border rounded-xl shadow-sm overflow-hidden">
            <div class="px-6 py-4 border-b bg-gray-50">
                <h2 class="font-semibold text-gray-800">Recent Events</h2>
            </div>
            <div class="overflow-x-auto">
                <table class="w-full text-left text-sm">
                    <thead>
                        <tr class="bg-gray-50 border-b">
                            <th class="px-6 py-3 font-medium text-gray-500">Timestamp</th>
                            <th class="px-6 py-3 font-medium text-gray-500">Type</th>
                            <th class="px-6 py-3 font-medium text-gray-500">Status</th>
                            <th class="px-6 py-3 font-medium text-gray-500">Code</th>
                            <th class="px-6 py-3 font-medium text-gray-500">Message</th>
                            <th class="px-6 py-3 font-medium text-gray-500 text-right">Duration</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each events as event}
                            <tr class="border-b last:border-0 hover:bg-gray-50">
                                <td class="px-6 py-4 text-gray-600">
                                    {new Date(event.created_at).toLocaleString()}
                                </td>
                                <td class="px-6 py-4 font-medium">{event.event_type}</td>
                                <td class="px-6 py-4">
                                    <span class="px-2 py-1 rounded-full text-xs font-medium border {getStatusColor(event.status)}">
                                        {event.status}
                                    </span>
                                </td>
                                <td class="px-6 py-4 font-mono text-xs">{event.code || '—'}</td>
                                <td class="px-6 py-4 text-gray-600 truncate max-w-xs">{event.message || '—'}</td>
                                <td class="px-6 py-4 text-right text-gray-500">{event.duration_ms}ms</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}
</div>
