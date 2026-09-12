<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLSelectAttributes } from 'svelte/elements'
  import { ChevronDown } from 'lucide-svelte'
  import { cx } from './utils'

  /**
   * Native `<select>` with the shared control recipe and a lucide chevron
   * (EPIC D.2). Options are passed as children so the whole native option
   * API (groups, disabled options…) stays available.
   */
  let {
    value = $bindable(''),
    error = undefined,
    disabled = false,
    class: className = '',
    children = undefined,
    ...rest
  }: {
    value?: string
    error?: string
    children?: Snippet
  } & Omit<HTMLSelectAttributes, 'value'> = $props()
</script>

<div class="relative">
  <select
    bind:value
    {disabled}
    aria-invalid={error ? true : undefined}
    class={cx(
      'w-full appearance-none rounded-control border bg-surface px-3 py-2 pr-9 text-sm text-foreground',
      'focus-ring disabled:cursor-not-allowed disabled:opacity-50',
      error ? 'border-negative' : 'border-input',
      className,
    )}
    {...rest}
  >
    {@render children?.()}
  </select>
  <ChevronDown
    class="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
  />
</div>
