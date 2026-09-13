<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements'
  import { cx } from './utils'

  /** Pulsing loading placeholder (EPIC D.2). Size it with `class` (h- and w- utilities). */
  let { class: className = '', ...rest }: HTMLAttributes<HTMLDivElement> = $props()
</script>

<div class={cx('animate-pulse rounded-md bg-muted', className)} {...rest}></div>
