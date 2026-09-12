<script lang="ts">
  import type { Snippet } from 'svelte'
  import AppHeader from './AppHeader.svelte'
  import MobileDrawer from './MobileDrawer.svelte'
  import Sidebar from './Sidebar.svelte'
  import UserMenu from './UserMenu.svelte'

  /**
   * Responsive application shell (EPIC D.3) — replaces the old fixed-width
   * Layout.svelte.
   *
   * Layout: a `h-dvh` row with the desktop sidebar (`hidden lg:flex`, the
   * 64px/16px collapsible rail) and a scrollable main column carrying the
   * sticky AppHeader above the routed content. Below `lg` the sidebar is
   * gone: the hamburger opens the MobileDrawer instead.
   *
   * State owned here and pushed down:
   * - `collapsed` — persisted under `vaultlab-sidebar` so the rail survives
   *   reloads (the app is SPA-only: `export const ssr = false`, so reading
   *   localStorage at init never runs on the server);
   * - `drawerOpen` — shared by the header hamburger (aria-expanded) and the
   *   drawer itself (which handles its own close-on-navigation + focus trap).
   */
  const SIDEBAR_STORAGE_KEY = 'vaultlab-sidebar'

  let { children }: { children: Snippet } = $props()

  function readCollapsed(): boolean {
    try {
      return localStorage.getItem(SIDEBAR_STORAGE_KEY) === 'true'
    } catch {
      // Storage unavailable (e.g. private mode): always start expanded.
      return false
    }
  }

  let collapsed = $state(readCollapsed())
  let drawerOpen = $state(false)

  function toggleCollapsed(): void {
    collapsed = !collapsed
    try {
      localStorage.setItem(SIDEBAR_STORAGE_KEY, String(collapsed))
    } catch {
      // Storage unavailable: the state still applies for this visit.
    }
  }
</script>

<a
  href="#content"
  class="focus-ring sr-only rounded-control bg-surface px-4 py-2 text-sm font-medium text-foreground focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50"
>
  Skip to content
</a>

<div class="flex h-dvh overflow-hidden bg-background text-foreground">
  <aside class="hidden shrink-0 lg:flex">
    <Sidebar {collapsed}>
      {#snippet footer()}
        <UserMenu compact={collapsed} />
      {/snippet}
    </Sidebar>
  </aside>

  <div class="flex min-w-0 flex-1 flex-col overflow-y-auto">
    <AppHeader
      {collapsed}
      {drawerOpen}
      ontogglecollapse={toggleCollapsed}
      ontoggledrawer={() => (drawerOpen = !drawerOpen)}
    />
    <main id="content" class="min-w-0 flex-1 p-4 outline-none lg:p-6" tabindex="-1">
      {@render children()}
    </main>
  </div>
</div>

<MobileDrawer bind:open={drawerOpen}>
  <Sidebar collapsed={false} />
</MobileDrawer>
