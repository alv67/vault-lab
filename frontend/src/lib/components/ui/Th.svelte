<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLThAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Header cell (EPIC D.2). `align="right"` right-aligns numeric columns. */
  let {
    align = 'left',
    class: className = '',
    children,
    ...rest
  }: {
    align?: 'left' | 'right'
    children?: Snippet
  } & HTMLThAttributes = $props()
</script>

<th
  class={cx('pb-2 font-medium', align === 'right' && 'text-right', className)}
  {...rest}
>
  {@render children?.()}
</th>
