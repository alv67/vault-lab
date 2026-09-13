<script lang="ts">
  import { cx } from './utils'

  /**
   * Mutually-exclusive choice rendered as a segmented pill row (EPIC D.2) —
   * the light/dark/system switch, table view toggles, …
   *
   * Keyboard: every segment is a real button, so tabbing + Enter/Space work
   * out of the box; `aria-selected` exposes the active one.
   */
  let {
    items,
    value = $bindable(''),
    ariaLabel = undefined,
    class: className = '',
  }: {
    items: { value: string; label: string }[]
    value?: string
    /** Accessible name for the whole control (e.g. "Theme"). */
    ariaLabel?: string
    class?: string
  } = $props()
</script>

<div
  role="tablist"
  aria-label={ariaLabel}
  class={cx(
    'inline-flex items-center gap-1 rounded-control border border-border bg-muted p-1',
    className,
  )}
>
  {#each items as item (item.value)}
    <button
      type="button"
      role="tab"
      aria-selected={item.value === value}
      onclick={() => (value = item.value)}
      class={cx(
        'focus-ring flex-1 whitespace-nowrap rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
        item.value === value
          ? 'bg-surface text-foreground shadow-card'
          : 'text-muted-foreground hover:text-foreground',
      )}
    >
      {item.label}
    </button>
  {/each}
</div>
