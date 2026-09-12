<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /**
   * Dashboard/KPI tile (EPIC D.2): a `Card`-styled box with a muted label, a
   * big tabular value and an optional signed delta line.
   *
   * The delta is colored from the numeric `deltaValue` (positive →
   * `text-positive`, negative → `text-negative`, zero/absent → muted) rather
   * than from the already-formatted `delta` string. `invertColor` flips the
   * mapping for "lower is better" metrics (e.g. costs): a positive
   * `deltaValue` then renders negative.
   */
  let {
    label,
    value,
    delta = undefined,
    deltaValue = undefined,
    invertColor = false,
    class: className = '',
    ...rest
  }: {
    label: string
    /** Pre-formatted main value (use `format.ts`). */
    value: string
    /** Pre-formatted delta line; rendered only when provided. */
    delta?: string
    /** Raw signed number driving the delta color. */
    deltaValue?: number
    invertColor?: boolean
  } & HTMLAttributes<HTMLDivElement> = $props()

  const deltaClass = $derived.by(() => {
    const n = Number(deltaValue ?? 0)
    if (!Number.isFinite(n) || n === 0) return 'text-muted-foreground'
    const good = n > 0 ? !invertColor : invertColor
    return good ? 'text-positive' : 'text-negative'
  })
</script>

<div class={cx('rounded-card border border-border bg-surface p-4 shadow-card', className)} {...rest}>
  <p class="text-sm text-muted-foreground">{label}</p>
  <p class="text-xl font-bold tabular-nums text-foreground">{value}</p>
  {#if delta !== undefined}
    <p class="mt-0.5 text-xs font-medium tabular-nums {deltaClass}">{delta}</p>
  {/if}
</div>
