<script lang="ts">
  import type { Snippet } from 'svelte'
  import { cx } from './utils'

  /**
   * Card header row: title/subtitle on the left, optional `actions` snippet
   * (buttons, badges…) pushed to the right (EPIC D.2).
   */
  let {
    title,
    subtitle = undefined,
    actions = undefined,
    class: className = '',
  }: {
    title: string
    subtitle?: string
    actions?: Snippet
    class?: string
  } = $props()
</script>

<div class={cx('flex flex-wrap items-start justify-between gap-3 px-6 pt-6', className)}>
  <div class="min-w-0">
    <h3 class="truncate text-base font-semibold text-foreground">{title}</h3>
    {#if subtitle}
      <p class="mt-0.5 text-sm text-muted-foreground">{subtitle}</p>
    {/if}
  </div>
  {#if actions}
    <div class="flex shrink-0 items-center gap-2">{@render actions()}</div>
  {/if}
</div>
