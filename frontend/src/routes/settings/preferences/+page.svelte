<script lang="ts">
  import { isLocale, locale, setLocale, t, type Locale } from '$lib/i18n/index.svelte'
  import { setThemeMode, theme } from '$lib/stores/theme.svelte'
  import SettingsTabs from '$lib/components/domain/SettingsTabs.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte'
  import Select from '$lib/components/ui/Select.svelte'

  /**
   * Settings → Preferences (EPIC K.1b): the first fully translated surface,
   * proof that the i18n layer works end to end. Theme (Light/Dark/System —
   * the shared `theme.*` keys, default System per decision D9) is bound to
   * the existing theme store; interface language (Italiano/English, default
   * Italian per decision D1) to the locale store. Both apply immediately
   * and persist in `localStorage` (no save button — same behaviour as the
   * header theme toggle); the language control re-renders this page through
   * `t()` on change.
   */

  // The controls bind plain strings; the accessors keep the union types and
  // route writes through the stores (same recipe as the dashboard
  // granularity switch). Reading `theme.mode` / `locale.current` here also
  // keeps the controls in sync when the value changes elsewhere (header
  // toggle, another tab).
  function getMode(): string {
    return theme.mode
  }
  function setMode(value: string): void {
    if (value === 'light' || value === 'dark' || value === 'system') setThemeMode(value)
  }
  function getLanguage(): string {
    return locale.current
  }
  function setLanguage(value: string): void {
    if (isLocale(value)) setLocale(value)
  }

  const themeItems = $derived([
    { value: 'light', label: t('theme.light') },
    { value: 'dark', label: t('theme.dark') },
    { value: 'system', label: t('theme.system') },
  ])

  // Endonyms: each language is displayed in its own language, by convention
  // untranslated in both dictionaries.
  const languageOptions: { value: Locale; label: string }[] = [
    { value: 'it', label: 'Italiano' },
    { value: 'en', label: 'English' },
  ]
</script>

<div class="p-6">
  <h1 class="mb-6 text-2xl font-bold">{t('nav.settings')}</h1>

  <SettingsTabs class="mb-6" />

  <Card class="max-w-lg p-6">
    <h2 class="mb-4 font-semibold">{t('preferences.title')}</h2>
    <div class="space-y-4">
      <div class="space-y-1.5">
        <!-- Not a `Field`: a SegmentedControl is a tablist of buttons, so an
             implicit <label> wrapper would dangle; the control carries the
             accessible name instead. -->
        <span class="block text-sm font-medium text-foreground">{t('theme.group')}</span>
        <SegmentedControl
          items={themeItems}
          bind:value={getMode, setMode}
          ariaLabel={t('theme.group')}
        />
        <p class="text-xs text-muted-foreground">{t('preferences.themeHint')}</p>
      </div>
      <Field label={t('common.language')} hint={t('preferences.languageHint')} class="max-w-xs">
        <Select bind:value={getLanguage, setLanguage}>
          {#each languageOptions as option (option.value)}
            <option value={option.value}>{option.label}</option>
          {/each}
        </Select>
      </Field>
    </div>
  </Card>
</div>
