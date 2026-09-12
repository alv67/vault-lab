<script lang="ts">
  import { Activity, Banknote, Briefcase, LayoutDashboard, Settings } from 'lucide-svelte'
  import { page } from '$app/state'
  import { resolve } from '$app/paths'
  import { cx } from '../ui/utils'

  /**
   * Navigation inside the sidebar (EPIC D.3): main section on top, Settings
   * section pinned to the bottom. `collapsed` renders the icon-rail variant
   * (64px): labels disappear, so each link carries an `aria-label`/`title`.
   *
   * The active entry is the item whose path is the *longest* prefix of the
   * current URL (with `/` matching exactly): only one link ever gets
   * `aria-current="page"`, so `/settings/health` highlights Health alone,
   * while the dev-only `/settings/theme-tokens` highlights Settings.
   */
  type IconType = typeof LayoutDashboard

  let { collapsed = false }: { collapsed?: boolean } = $props()

  // `as const` keeps `to` as literal route types so the typed `resolve()`
  // accepts them; the icon cast unifies the component type per list.
  const mainItems = [
    { to: '/', label: 'Dashboard', icon: LayoutDashboard as IconType },
    { to: '/portfolios', label: 'Portfolios', icon: Briefcase as IconType },
    { to: '/assets', label: 'Assets', icon: Banknote as IconType },
  ] as const

  const settingsItems = [
    { to: '/settings', label: 'Settings', icon: Settings as IconType },
    { to: '/settings/health', label: 'Health', icon: Activity as IconType },
  ] as const

  function matches(to: string, pathname: string): boolean {
    if (to === '/') return pathname === '/'
    return pathname === to || pathname.startsWith(`${to}/`)
  }

  const activeTo = $derived.by(() => {
    let best: string | null = null
    for (const item of [...mainItems, ...settingsItems]) {
      if (
        matches(item.to, page.url.pathname) &&
        (best === null || item.to.length > best.length)
      ) {
        best = item.to
      }
    }
    return best
  })

  function itemClasses(to: string): string {
    return cx(
      'focus-ring flex items-center gap-3 rounded-control py-2 text-sm transition-colors',
      collapsed ? 'justify-center px-0' : 'px-3',
      to === activeTo
        ? 'bg-accent/10 font-medium text-accent-text'
        : 'text-muted-foreground hover:bg-muted hover:text-foreground',
    )
  }
</script>

<nav aria-label="Main" class="flex min-h-0 flex-1 flex-col">
  <div class="flex-1 space-y-1 overflow-y-auto p-3">
    {#each mainItems as item (item.to)}
      {@const Icon = item.icon}
      <a
        href={resolve(item.to)}
        class={itemClasses(item.to)}
        aria-current={item.to === activeTo ? 'page' : undefined}
        aria-label={collapsed ? item.label : undefined}
        title={collapsed ? item.label : undefined}
      >
        <Icon class="h-5 w-5 shrink-0" />
        {#if !collapsed}{item.label}{/if}
      </a>
    {/each}
  </div>

  <div class="border-t border-border p-3">
    {#if !collapsed}
      <div class="px-3 pb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        Settings
      </div>
    {/if}
    <div class="space-y-1">
      {#each settingsItems as item (item.to)}
        {@const Icon = item.icon}
        <a
          href={resolve(item.to)}
          class={itemClasses(item.to)}
          aria-current={item.to === activeTo ? 'page' : undefined}
          aria-label={collapsed ? item.label : undefined}
          title={collapsed ? item.label : undefined}
        >
          <Icon class="h-5 w-5 shrink-0" />
          {#if !collapsed}{item.label}{/if}
        </a>
      {/each}
    </div>
  </div>
</nav>
