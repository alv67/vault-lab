<script lang="ts">
  /**
   * Small pill that tells the user where the weights shown in a box come from.
   * Unknown or null sources render nothing (e.g. right after page load, when
   * the data provenance is not known to the UI).
   */
  let { source = null as string | null }: { source?: string | null } = $props()

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
</script>

{#if info}
  <span
    class="inline-flex items-center gap-1.5 rounded-full border border-gray-300 bg-white px-2 py-0.5 text-[11px] font-medium text-gray-600"
    title={info.description}
    aria-label={`${info.label}: ${info.description}`}
  >
    <span class="h-1.5 w-1.5 rounded-full" style={`background-color: ${info.dot}`}></span>
    {info.label}
  </span>
{/if}
