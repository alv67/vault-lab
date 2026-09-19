<script lang="ts">
  import { Banknote, PanelLeft } from 'lucide-svelte'
  import { resolve } from '$app/paths'
  import { t } from '$lib/i18n/index.svelte'
  import Button from '../ui/Button.svelte'
  import ThemeToggle from './ThemeToggle.svelte'
  import UserMenu from './UserMenu.svelte'
  import { cx } from '../ui/utils'

  /**
   * Sticky top bar (EPIC D.3, adaptive since K.2). Left: the sidebar
   * collapse toggle — `lg`+ only, since phones have no sidebar (bottom nav
   * + More sheet, decision D2) and tablets are a forced icon rail — plus the
   * brand link shown below `lg` (the rail/drawer carry it otherwise).
   * Right: the theme toggle at every size (the rail and the More sheet have
   * no theme control of their own) and the user menu, phone-only
   * (`sm:hidden`): from `sm` up the sidebar rail's footer hosts one already
   * so the header must not duplicate it, while on phones — where the More
   * sheet is navigation only — account and theme stay in the header (spec
   * §5.2). The control
   * aria-labels go through `t()` (EPIC K.1b, decision D1); the "VaultLab"
   * brand is a proper noun and stays as-is.
   *
   * Condensing (spec §5.1): the shell measures its main scroll container
   * and flips `condensed` once scrolled past a small threshold; the bar
   * shrinks 56px → 44px with a plain CSS height transition, which the global
   * `prefers-reduced-motion` rule in app.css already neutralises.
   *
   * z-20: same tier as the dropdowns it hosts (header and its popovers must
   * both stay under the drawer/sheet tier, z-30).
   */
  let {
    collapsed,
    condensed = false,
    ontogglecollapse,
  }: {
    /** Sidebar rail state (drives the collapse toggle's label/pressed). */
    collapsed: boolean
    /** Main scroll container scrolled past the condense threshold. */
    condensed?: boolean
    ontogglecollapse: () => void
  } = $props()
</script>

<header
  class={cx(
    'sticky top-0 z-20 flex shrink-0 items-center justify-between gap-2 border-b border-border bg-surface/90 px-3 backdrop-blur transition-[height] duration-base ease-standard lg:px-4',
    condensed ? 'h-11' : 'h-14',
  )}
>
  <div class="flex min-w-0 items-center gap-1">
    <Button
      variant="ghost"
      size="icon"
      class="hidden lg:inline-flex"
      aria-label={collapsed ? t('header.expandSidebar') : t('header.collapseSidebar')}
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
    <UserMenu compact direction="down" align="end" class="sm:hidden" />
  </div>
</header>
