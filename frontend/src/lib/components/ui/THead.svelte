<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Table head: header rows inherit the muted foreground tone (EPIC D.2). */
  let { class: className = '', children, ...rest }: { children?: Snippet } & HTMLAttributes<HTMLTableSectionElement> =
    $props()
</script>

<thead class={cx('text-muted-foreground', className)} {...rest}>
  {@render children?.()}
</thead>
