<script lang="ts">
  import type { Snippet } from 'svelte'
  import { cx } from './utils'

  /**
   * Label + hint + error wrapper (EPIC D.2).
   *
   * Without `for` the control is expected inside the children slot, wrapped by
   * an implicit `<label>` (no id plumbing needed). With `for` the label points
   * at an external control id. Hint and error render below the control; the
   * error wins over the hint when both are present.
   */
  let {
    label,
    hint = undefined,
    error = undefined,
    for: htmlFor = undefined,
    class: className = '',
    children,
  }: {
    label: string
    hint?: string
    error?: string
    for?: string
    class?: string
    children: Snippet
  } = $props()
</script>

<div class={cx('space-y-1.5', className)}>
  {#if htmlFor}
    <label for={htmlFor} class="block text-sm font-medium text-foreground">{label}</label>
    {@render children()}
  {:else}
    <label class="block space-y-1.5">
      <span class="block text-sm font-medium text-foreground">{label}</span>
      {@render children()}
    </label>
  {/if}

  {#if error}
    <p class="text-xs font-medium text-negative">{error}</p>
  {:else if hint}
    <p class="text-xs text-muted-foreground">{hint}</p>
  {/if}
</div>
