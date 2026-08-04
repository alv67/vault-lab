<script lang="ts">
  import { toasts } from '$lib/stores/toast.svelte'
  import { CheckCircle2, XCircle, AlertTriangle } from 'lucide-svelte'
</script>

{#if toasts.length > 0}
  <div class="pointer-events-none fixed right-4 top-4 z-50 flex flex-col gap-2">
    {#each toasts as t (t.id)}
      <div
        class="pointer-events-auto flex w-72 items-center gap-2 rounded-lg px-4 py-3 text-sm font-medium shadow-lg {t.type === 'success'
          ? 'bg-green-600 text-white'
          : t.type === 'warning'
            ? 'bg-amber-500 text-white'
            : 'bg-red-600 text-white'}"
      >
        {#if t.type === 'success'}
          <CheckCircle2 class="h-4 w-4 shrink-0" />
        {:else if t.type === 'warning'}
          <AlertTriangle class="h-4 w-4 shrink-0" />
        {:else}
          <XCircle class="h-4 w-4 shrink-0" />
        {/if}
        {t.message}
      </div>
    {/each}
  </div>
{/if}
