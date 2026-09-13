<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLTdAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /**
   * Data cell (EPIC D.2). `align="right"` (the numeric-column convention)
   * also applies `tabular-nums` so figures stay aligned across rows.
   */
  let {
    align = 'left',
    class: className = '',
    children,
    ...rest
  }: {
    align?: 'left' | 'right'
    children?: Snippet
  } & HTMLTdAttributes = $props()
</script>

<td
  class={cx(
    'py-2',
    align === 'right' && 'text-right tabular-nums',
    className,
  )}
  {...rest}
>
  {@render children?.()}
</td>
