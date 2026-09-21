<script lang="ts">
  import { cx } from './utils'

  /**
   * Mutually-exclusive choice rendered as a segmented pill row (EPIC D.2) —
   * the light/dark/system switch, table view toggles, …
   *
   * Overflow-safe (EPIC K bug-fix): `max-w-full` + `flex-wrap` keep the pill
   * inside its container at phone widths — content-based flex-basis
   * (`flex-auto`, not `flex-1`, whose 0% basis would defeat wrapping) and
   * wrappable labels mean long strings wrap *inside* the row instead of
   * pushing a horizontal scroll. Desktop is unchanged: the inline-flex
   * container still shrink-wraps the pills to their single-line content.
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
    'inline-flex max-w-full flex-wrap items-center gap-1 rounded-control border border-border bg-muted p-1',
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
        'focus-ring flex-auto break-words rounded-md px-3 py-1.5 text-center text-sm font-medium transition-colors',
        item.value === value
          ? 'bg-surface text-foreground shadow-card'
          : 'text-muted-foreground hover:text-foreground',
      )}
    >
      {item.label}
    </button>
  {/each}
</div>
