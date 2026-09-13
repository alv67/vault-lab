<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /**
   * Status pill (EPIC D.2). The accent/positive/negative/warning variants use
   * the shared `bg-{tone}/10 text-{tone}` formula already adopted by the
   * transaction-type pills (D.1b).
   *
   * `outline` deliberately carries no text color so callers can layer a
   * semantic text tone (see ProvenanceBadge).
   */
  let {
    variant = 'neutral',
    class: className = '',
    children,
    ...rest
  }: {
    variant?: 'neutral' | 'accent' | 'positive' | 'negative' | 'warning' | 'outline'
    children: Snippet
  } & HTMLAttributes<HTMLSpanElement> = $props()

  const variants = {
    neutral: 'bg-muted text-muted-foreground',
    accent: 'bg-accent/10 text-accent-text',
    positive: 'bg-positive/10 text-positive',
    negative: 'bg-negative/10 text-negative',
    warning: 'bg-warning/10 text-warning',
    outline: 'border border-border',
  } as const
</script>

<span
  class={cx(
    'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium',
    variants[variant],
    className,
  )}
  {...rest}
>
  {@render children()}
</span>
