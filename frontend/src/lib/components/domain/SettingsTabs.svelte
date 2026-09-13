<script lang="ts">
  import { page } from '$app/state'
  import { resolve } from '$app/paths'
  import { cx } from '$lib/components/ui/utils'

  /**
   * Link-based tab bar for the Settings subroutes (EPIC E.4): real anchors so
   * switching sections is plain navigation, with the current route marked by
   * `aria-current="page"`. Pills reuse the SegmentedControl recipe.
   */
  let { class: className = '' }: { class?: string } = $props()

  // `as const` keeps `to` as literal route types so the typed `resolve()` accepts them.
  const tabs = [
    { to: '/settings', label: 'Profile' },
    { to: '/settings/password', label: 'Password' },
    { to: '/settings/currencies', label: 'Currencies' },
    { to: '/settings/health', label: 'Health' },
  ] as const
</script>

<nav aria-label="Settings sections" class={cx('flex flex-wrap', className)}>
  <div class="inline-flex items-center gap-1 rounded-control border border-border bg-muted p-1">
    {#each tabs as tab (tab.to)}
      {@const href = resolve(tab.to)}
      {@const active = page.url.pathname === href}
      <a
        {href}
        class={cx(
          'focus-ring whitespace-nowrap rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
          active ? 'bg-surface text-foreground shadow-card' : 'text-muted-foreground hover:text-foreground',
        )}
        aria-current={active ? 'page' : undefined}
      >
        {tab.label}
      </a>
    {/each}
  </div>
</nav>
