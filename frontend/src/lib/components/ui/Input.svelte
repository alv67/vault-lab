<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /**
   * Design-system text/number input (EPIC D.2).
   *
   * Extra classes land on the control itself, so existing recipes keep
   * working (e.g. `class="no-spinner"` on number inputs, `uppercase`, widths).
   */
  let {
    value = $bindable(''),
    error = undefined,
    disabled = false,
    placeholder = undefined,
    type = 'text',
    class: className = '',
    ...rest
  }: {
    value?: string
    /** Validation message: switches the border to `negative` and sets `aria-invalid`. */
    error?: string
  } & Omit<HTMLInputAttributes, 'value'> = $props()
</script>

<input
  {type}
  bind:value
  {disabled}
  {placeholder}
  aria-invalid={error ? true : undefined}
  class={cx(
    'w-full rounded-control border bg-surface px-3 py-2 text-sm text-foreground',
    'placeholder:text-muted-foreground focus-ring',
    'disabled:cursor-not-allowed disabled:opacity-50',
    error ? 'border-negative' : 'border-input',
    className,
  )}
  {...rest}
/>
