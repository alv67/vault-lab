<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Resting panel surface: token border + card radius/shadow (EPIC D.2). */
  let { class: className = '', children, ...rest }: { children?: Snippet } & HTMLAttributes<HTMLElement> =
    $props()
</script>

<div
  class={cx('rounded-card border border-border bg-surface shadow-card', className)}
  {...rest}
>
  {@render children?.()}
</div>
