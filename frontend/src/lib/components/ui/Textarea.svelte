<script lang="ts">
  import type { HTMLTextareaAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Design-system textarea (EPIC D.2). Same recipe as `Input`. */
  let {
    value = $bindable(''),
    error = undefined,
    disabled = false,
    placeholder = undefined,
    class: className = '',
    ...rest
  }: {
    value?: string
    error?: string
  } & Omit<HTMLTextareaAttributes, 'value'> = $props()
</script>

<textarea
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
></textarea>
