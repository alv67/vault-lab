<script lang="ts">
  import { Banknote, Menu, PanelLeft } from 'lucide-svelte'
  import { resolve } from '$app/paths'
  import Button from '../ui/Button.svelte'
  import ThemeToggle from './ThemeToggle.svelte'
  import UserMenu from './UserMenu.svelte'

  /**
   * Sticky top bar (EPIC D.3). Left: hamburger below `lg` (opens the mobile
   * drawer) / sidebar collapse toggle from `lg` up, plus a mobile-only brand
   * (the sidebar is hidden there). Right: theme toggle and — on mobile,
   * where the sidebar's user menu is gone — the user menu too.
   *
   * z-20: same tier as the dropdowns it hosts (header and its popovers must
   * both stay under the drawer, z-30).
   */
  let {
    collapsed,
    drawerOpen,
    ontogglecollapse,
    ontoggledrawer,
  }: {
    /** Sidebar rail state (drives the collapse toggle's label/pressed). */
    collapsed: boolean
    /** Drawer state, surfaced on the hamburger via `aria-expanded`. */
    drawerOpen: boolean
    ontogglecollapse: () => void
    /** Opens the drawer (and closes it again if it is already open). */
    ontoggledrawer: () => void
  } = $props()
</script>

<header
  class="sticky top-0 z-20 flex h-14 shrink-0 items-center justify-between gap-2 border-b border-border bg-surface/90 px-3 backdrop-blur lg:px-4"
>
  <div class="flex min-w-0 items-center gap-1">
    <Button
      variant="ghost"
      size="icon"
      class="lg:hidden"
      aria-label={drawerOpen ? 'Close navigation menu' : 'Open navigation menu'}
      aria-haspopup="dialog"
      aria-expanded={drawerOpen}
      onclick={ontoggledrawer}
    >
      <Menu class="h-5 w-5" />
    </Button>
    <Button
      variant="ghost"
      size="icon"
      class="hidden lg:inline-flex"
      aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
      aria-pressed={collapsed}
      onclick={ontogglecollapse}
    >
      <PanelLeft class="h-5 w-5" />
    </Button>
    <a
      href={resolve('/')}
      class="focus-ring flex items-center gap-2 rounded-control px-1 lg:hidden"
    >
      <Banknote class="h-5 w-5 shrink-0 text-accent-text" aria-hidden="true" />
      <span class="truncate text-base font-bold">VaultLab</span>
    </a>
  </div>

  <div class="flex shrink-0 items-center gap-1">
    <ThemeToggle />
    <UserMenu compact direction="down" align="end" class="lg:hidden" />
  </div>
</header>
