<script lang="ts">
  import type { Snippet } from 'svelte'
  import { t } from '$lib/i18n/index.svelte'
  import { viewport } from '$lib/stores/viewport.svelte'
  import AppHeader from './AppHeader.svelte'
  import BottomNav from './BottomNav.svelte'
  import CommandPalette from './CommandPalette.svelte'
  import Fab from './Fab.svelte'
  import MobileDrawer from './MobileDrawer.svelte'
  import Sidebar from './Sidebar.svelte'
  import UserMenu from './UserMenu.svelte'

  /**
   * Responsive application shell (EPIC D.3, adaptive since EPIC K.2) —
   * replaces the old fixed-width Layout.svelte.
   *
   * Three device classes (spec §5.1, decision D2):
   * - phone (< `sm`): no sidebar at all — a fixed `BottomNav` (Overview,
   *   Portfolios, Assets, More) and the quick-actions `Fab` mount over the
   *   content column, and `<main>` gets an extra bottom pad so nothing hides
   *   behind the bar (56px + safe area + breathing room). The hamburger is
   *   gone: `MobileDrawer` survives as the "More" sheet, opened only by the
   *   bottom nav;
   * - tablet (`sm`–`lg`): the sidebar is always the 64px icon rail — the
   *   persisted expand preference deliberately only applies at `lg`+;
   * - desktop (`lg`+): unchanged — expandable/collapsible sidebar, sticky
   *   header with the collapse toggle, user menu in the sidebar footer.
   *
   * State owned here and pushed down:
   * - `collapsed` — persisted under `vaultlab-sidebar` so the rail survives
   *   reloads (the app is SPA-only: `export const ssr = false`, so reading
   *   localStorage at init never runs on the server);
    * - `moreOpen` — the phone More sheet (the drawer handles its own
    *   close-on-navigation + focus trap);
    * - `condensed` — measured on the main scroll container past a small
    *   threshold and handed to the sticky header (§5.1 "condenses on
    *   scroll");
    * - `paletteOpen` — the K.5a command palette (⌘K/Ctrl+K chord, the header
    *   search trigger and Esc all share this one bindable state).
    */
  const SIDEBAR_STORAGE_KEY = 'vaultlab-sidebar'

  /** Scroll distance (px) after which the sticky header condenses. */
  const CONDENSE_THRESHOLD = 16

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
  let moreOpen = $state(false)
  let condensed = $state(false)
  let paletteOpen = $state(false)
  let scrollContainer = $state<HTMLElement | null>(null)

  // Tablet forces the rail (spec §5.1: on the `sm`–`lg` classes the sidebar
  // *is* the rail); desktop keeps the persisted preference. Below `sm` the
  // sidebar is not rendered at all.
  const railCollapsed = $derived(viewport.isDesktop ? collapsed : true)

  function toggleCollapsed(): void {
    collapsed = !collapsed
    try {
      localStorage.setItem(SIDEBAR_STORAGE_KEY, String(collapsed))
    } catch {
      // Storage unavailable: the state still applies for this visit.
    }
  }

  // The scrollable element is this shell column (not the window), so the
  // condensing state must be measured here; passive listener because the
  // handler never cancels the scroll (K.2 header).
  $effect(() => {
    const el = scrollContainer
    if (!el) return
    const update = () => (condensed = el.scrollTop > CONDENSE_THRESHOLD)
    update()
    el.addEventListener('scroll', update, { passive: true })
    return () => el.removeEventListener('scroll', update)
  })
</script>

<a
  href="#content"
  class="focus-ring sr-only rounded-control bg-surface px-4 py-2 text-sm font-medium text-foreground focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50"
>
  {t('nav.skipToContent')}
</a>

<div class="flex h-dvh overflow-hidden bg-background text-foreground">
  <aside class="hidden shrink-0 sm:flex">
    <Sidebar collapsed={railCollapsed}>
      {#snippet footer()}
        <UserMenu compact={railCollapsed} />
      {/snippet}
    </Sidebar>
  </aside>

  <div
    bind:this={scrollContainer}
    class="flex min-w-0 flex-1 flex-col overflow-y-auto"
  >
    <AppHeader
      {collapsed}
      {condensed}
      ontogglecollapse={toggleCollapsed}
      onopenpalette={() => (paletteOpen = true)}
    />
    <!-- Longhand padding utilities only: `p-*` shorthand would fight the
         phone-only `pb-[…]` clearance below the fixed bottom nav. -->
    <main
      id="content"
      class="min-w-0 flex-1 px-4 pt-4 pb-[calc(6rem_+_env(safe-area-inset-bottom))] outline-none sm:pb-4 lg:px-6 lg:pt-6 lg:pb-6"
      tabindex="-1"
    >
      {@render children()}
    </main>
  </div>
</div>

{#if viewport.isPhone}
  <BottomNav {moreOpen} onopenmore={() => (moreOpen = true)} />
  <Fab />
{/if}

<MobileDrawer bind:open={moreOpen}>
  <Sidebar collapsed={false} />
</MobileDrawer>

<!-- Global command palette (EPIC K.5a, spec §8.1): mounted once here — the
     ⌘K/Ctrl+K chord lives in the component, the header search button opens
     it, and it borrows `toggleCollapsed` for its Toggle-sidebar action. -->
<CommandPalette bind:open={paletteOpen} ontogglesidebar={toggleCollapsed} />
