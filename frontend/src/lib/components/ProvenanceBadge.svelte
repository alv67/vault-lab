<script lang="ts">
  /**
   * Small pill that tells the user where the weights shown in a box come from
   * and — when the dimension has actually been persisted — when it was last
   * updated (e.g. "da Morningstar (2026-09-05)"). Unknown or null sources
   * render nothing (e.g. a dimension never persisted and never prefilled).
   * A null/empty `updatedAt` means the shown source is not persisted yet
   * (unsaved prefill or fresh manual edit): the badge shows the label only,
   * and the date appears after the next successful save.
   */
  let {
    source = null as string | null,
    updatedAt = null as string | null,
  }: { source?: string | null; updatedAt?: string | null } = $props()

  interface Provenance {
    label: string
    dot: string
    /** Long explanation exposed via title/aria. */
    description: string
  }

  const PROVENANCE: Record<string, Provenance> = {
    manual: {
      label: 'manuale',
      dot: '#9ca3af',
      description: 'Dati modificati manualmente',
    },
    justetf: {
      label: 'da JustETF',
      dot: '#0ea5e9',
      description: 'Lista paesi importata da JustETF, non modificata manualmente',
    },
    morningstar: {
      label: 'da Morningstar',
      dot: '#f59e0b',
      description: 'Dati importati da Morningstar, non modificati manualmente',
    },
    'morningstar-regions': {
      label: 'da Morningstar (regioni ufficiali)',
      dot: '#f59e0b',
      description: 'Regioni ufficiali importate da Morningstar, non modificate manualmente',
    },
    yahoo: {
      label: 'da Yahoo',
      dot: '#720e9e',
      description: 'Settori importati da Yahoo, non modificati manualmente',
    },
    derived: {
      label: 'calcolato dai paesi',
      dot: '#8b5cf6',
      description: 'Regioni calcolate a partire dai pesi dei paesi',
    },
    'derived-etf': {
      label: 'da JustETF via paesi',
      dot: '#8b5cf6',
      description: 'Regioni calcolate dai paesi importati da JustETF',
    },
  }

  const info = $derived(source ? PROVENANCE[source] ?? null : null)

  /** YYYY-MM-DD part of the last persisted update; '' when not persisted yet. */
  const date = $derived(updatedAt ? updatedAt.slice(0, 10) : '')
  const label = $derived(info ? (date ? `${info.label} (${date})` : info.label) : '')
  const hint = $derived(
    info ? (date ? `${info.description} — aggiornato al ${date}` : info.description) : '',
  )
</script>

{#if info}
  <span
    class="inline-flex items-center gap-1.5 rounded-full border border-gray-300 bg-white px-2 py-0.5 text-[11px] font-medium text-gray-600"
    title={hint}
    aria-label={`${label}: ${hint}`}
  >
    <span class="h-1.5 w-1.5 rounded-full" style={`background-color: ${info.dot}`}></span>
    {label}
  </span>
{/if}
