<script lang="ts">
  import { toast, toasts } from '$lib/stores/toast.svelte'
  import { CheckCircle2, XCircle, AlertTriangle, X } from 'lucide-svelte'

  /**
   * Toast viewport (EPIC D.2): surface-raised cards with a semantic icon
   * color instead of saturated fills. The wrapper is always mounted with
   * `aria-live="polite"` (live regions must exist before content arrives);
   * error toasts additionally carry `role="alert"` so they interrupt.
   *
   * Z-index: toasts sit at the top of the scale in $lib/components/ui/utils.ts
   * (dropdown 20 / drawer 30 / modal 40 / toast 50).
   */
  const iconFor = {
    success: CheckCircle2,
    warning: AlertTriangle,
    error: XCircle,
  } as const

  const iconColor = {
    success: 'text-positive',
    warning: 'text-warning',
    error: 'text-negative',
  } as const
</script>

<div class="pointer-events-none fixed right-4 top-4 z-50 flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2" aria-live="polite">
  {#each toasts as t (t.id)}
    {@const Icon = iconFor[t.type]}
    <div
      class="pointer-events-auto flex items-start gap-2 rounded-control border border-border bg-surface-raised px-4 py-3 text-sm font-medium text-foreground shadow-raised"
      role={t.type === 'error' ? 'alert' : undefined}
    >
      <Icon class="mt-0.5 h-4 w-4 shrink-0 {iconColor[t.type]}" />
      <span class="min-w-0 flex-1">{t.message}</span>
      <button
        type="button"
        onclick={() => toast.dismiss(t.id)}
        aria-label="Dismiss notification"
        class="focus-ring -mr-1 shrink-0 rounded-control p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        <X class="h-3.5 w-3.5" />
      </button>
    </div>
  {/each}
</div>
