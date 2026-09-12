<script lang="ts">
  import type { Snippet } from 'svelte'
  import { cx } from './utils'

  /** Padded body section of a `Card` (EPIC D.2). */
  let {
    class: className = '',
    children,
  }: { class?: string; children: Snippet } = $props()
</script>

<div class={cx('p-6', className)}>
  {@render children()}
</div>
