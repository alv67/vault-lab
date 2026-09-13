<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Table row: hairline separator, last row unbordered (EPIC D.2). */
  let { class: className = '', children, ...rest }: { children?: Snippet } & HTMLAttributes<HTMLTableRowElement> =
    $props()
</script>

<tr class={cx('border-b border-border last:border-0', className)} {...rest}>
  {@render children?.()}
</tr>
