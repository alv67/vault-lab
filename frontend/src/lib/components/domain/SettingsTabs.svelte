<script lang="ts">
  import { page } from '$app/state'
  import { resolve } from '$app/paths'
  import { t } from '$lib/i18n/index.svelte'
  import { cx } from '$lib/components/ui/utils'

  /**
   * Link-based tab bar for the Settings subroutes (EPIC E.4): real anchors so
   * switching sections is plain navigation, with the current route marked by
   * `aria-current="page"`. Pills reuse the SegmentedControl recipe. Labels
   * are translated (`t()` reads the reactive locale, EPIC K.1b / D1); the
   * tab order follows the UX-redesign spec sections (Profile · Security ·
   * Preferences · Currencies).
   */
  let { class: className = '' }: { class?: string } = $props()

  // `as const` keeps `to` as literal route types so the typed `resolve()`
  // accepts them (and the label keys against the `MessageKey` union).
  const tabs = [
    { to: '/settings', labelKey: 'settingsTabs.profile' },
    { to: '/settings/password', labelKey: 'settingsTabs.password' },
    { to: '/settings/preferences', labelKey: 'settingsTabs.preferences' },
    { to: '/settings/currencies', labelKey: 'settingsTabs.currencies' },
  ] as const
</script>

<nav aria-label={t('settingsTabs.sections')} class={cx('flex flex-wrap', className)}>
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
        {t(tab.labelKey)}
      </a>
    {/each}
  </div>
</nav>
