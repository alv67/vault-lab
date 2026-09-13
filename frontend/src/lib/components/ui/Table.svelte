<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Data-table root (EPIC D.2): full width, left aligned, small text. */
  let { class: className = '', children, ...rest }: { children?: Snippet } & HTMLAttributes<HTMLTableElement> =
    $props()
</script>

<table class={cx('w-full text-left text-sm', className)} {...rest}>
  {@render children?.()}
</table>
