<script lang="ts">
  import type { Snippet } from 'svelte'
  import { Banknote } from 'lucide-svelte'
  import SidebarNav from './SidebarNav.svelte'
  import { cx } from '../ui/utils'

  /**
   * Sidebar column (EPIC D.3): brand + nav, with an optional footer slot
   * (the shell passes the desktop `UserMenu` into it). Used twice:
   * - desktop: inside AppShell's `hidden lg:flex` aside, width driven by the
   *   persisted `collapsed` rail state;
   * - mobile drawer: always full width (`collapsed={false}`).
   */
  let {
    collapsed = false,
    footer = undefined,
  }: {
    collapsed?: boolean
    /** Bottom region (user menu). Hidden in the icon rail unless provided. */
    footer?: Snippet
  } = $props()
</script>

<div
  class={cx(
    'flex h-full flex-col border-r border-border bg-surface transition-[width] duration-200',
    collapsed ? 'w-16' : 'w-64',
  )}
>
  <div
    class={cx(
      'flex h-14 shrink-0 items-center gap-2 border-b border-border',
      collapsed ? 'justify-center' : 'px-4',
    )}
  >
    <Banknote class="h-6 w-6 shrink-0 text-accent-text" aria-hidden="true" />
    {#if !collapsed}
      <span class="truncate text-lg font-bold">VaultLab</span>
    {/if}
  </div>

  <SidebarNav {collapsed} />

  {#if footer}
    <div class={cx('shrink-0 border-t border-border p-3', collapsed && 'flex justify-center')}>
      {@render footer()}
    </div>
  {/if}
</div>
