<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Table body (EPIC D.2). */
  let { class: className = '', children, ...rest }: { children?: Snippet } & HTMLAttributes<HTMLTableSectionElement> =
    $props()
</script>

<tbody class={cx(className)} {...rest}>
  {@render children?.()}
</tbody>
