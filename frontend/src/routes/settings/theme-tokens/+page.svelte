<script lang="ts">
  // Token smoke-test (EPIC D.1a): visual evidence that the semantic token
  // layer renders correctly in both themes. New file by design — it is
  // token-native from the start, so commit 2 does not migrate it.
  import { Sun, Moon, Monitor } from 'lucide-svelte'
  import { DEFAULT_MODE, resolved, setThemeMode, theme, type ThemeMode } from '$lib/stores/theme.svelte'
  import { chartThemeName } from '$lib/chartTheme'
  import {
    CHART_MUTED,
    CHART_SEMANTIC_COLORS,
    resolveChartMuted,
    resolvePalette,
  } from '$lib/chartPalette'
  import type { ResolvedTheme } from '$lib/stores/theme.svelte'

  const themeModes: ResolvedTheme[] = ['light', 'dark']
  const semanticKeys = Object.keys(CHART_SEMANTIC_COLORS.light) as (keyof typeof CHART_SEMANTIC_COLORS.light)[]

  const modes: { value: ThemeMode; label: string; icon: typeof Sun }[] = [
    { value: 'light', label: 'Light', icon: Sun },
    { value: 'dark', label: 'Dark', icon: Moon },
    { value: 'system', label: 'System', icon: Monitor },
  ]

  // Literal class strings so Tailwind's scanner can see them.
  const surfaces: { token: string; swatch: string; text: string }[] = [
    { token: 'background', swatch: 'bg-background', text: 'text-foreground' },
    { token: 'foreground', swatch: 'bg-foreground', text: 'text-background' },
    { token: 'surface', swatch: 'bg-surface', text: 'text-foreground' },
    { token: 'surface-raised', swatch: 'bg-surface-raised', text: 'text-foreground' },
    { token: 'muted', swatch: 'bg-muted', text: 'text-muted-foreground' },
    { token: 'muted-foreground', swatch: 'bg-muted-foreground', text: 'text-surface' },
    { token: 'border', swatch: 'bg-border', text: 'text-foreground' },
    { token: 'input', swatch: 'bg-input', text: 'text-foreground' },
    { token: 'ring', swatch: 'bg-ring', text: 'text-surface' },
    { token: 'accent', swatch: 'bg-accent', text: 'text-accent-foreground' },
    { token: 'accent-hover', swatch: 'bg-accent-hover', text: 'text-accent-foreground' },
    { token: 'accent-text', swatch: 'bg-accent-text', text: 'text-surface' },
    { token: 'positive', swatch: 'bg-positive', text: 'text-surface' },
    { token: 'negative', swatch: 'bg-negative', text: 'text-surface' },
    { token: 'warning', swatch: 'bg-warning', text: 'text-surface' },
    { token: 'overlay', swatch: 'bg-overlay', text: 'text-surface' },
  ]

  const alphaChips: { label: string; cls: string }[] = [
    { label: 'bg-accent/10 text-accent-text', cls: 'bg-accent/10 text-accent-text' },
    { label: 'bg-positive/10 text-positive', cls: 'bg-positive/10 text-positive' },
    { label: 'bg-negative/10 text-negative', cls: 'bg-negative/10 text-negative' },
    { label: 'bg-warning/10 text-warning', cls: 'bg-warning/10 text-warning' },
    { label: 'bg-overlay/40 text-surface', cls: 'bg-overlay/40 text-surface' },
  ]
</script>

<div class="mx-auto max-w-5xl p-6">
  <h1 class="text-2xl font-bold">Design tokens (D.1a smoke test)</h1>
  <p class="mt-1 text-sm text-muted-foreground">
    Resolved theme: <code class="rounded-control bg-muted px-1.5 py-0.5 font-mono text-xs">{resolved()}</code>
    — default mode is <code class="rounded-control bg-muted px-1.5 py-0.5 font-mono text-xs">{DEFAULT_MODE}</code>
    (flipped to dark in the final EPIC&nbsp;D commit).
  </p>

  <!-- 3-mode control (the real settings UI ships in a later commit) -->
  <div class="mt-4 inline-flex gap-1 rounded-card border border-border bg-surface p-1 shadow-card">
    {#each modes as { value, label, icon: Icon } (value)}
      <button
        type="button"
        onclick={() => setThemeMode(value)}
        class="focus-ring inline-flex items-center gap-1.5 rounded-control px-3 py-1.5 text-sm font-medium {theme.mode === value
          ? 'bg-accent text-accent-foreground'
          : 'text-muted-foreground hover:bg-muted'}"
      >
        <Icon class="h-4 w-4" />
        {label}
      </button>
    {/each}
  </div>

  <h2 class="mt-8 text-lg font-semibold">Semantic tokens</h2>
  <div class="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
    {#each surfaces as { token, swatch, text } (token)}
      <div class="rounded-card border border-border bg-surface p-3 shadow-card">
        <div class="{swatch} {text} flex h-12 items-center justify-center rounded-control text-xs font-medium">
          Aa
        </div>
        <div class="mt-2 font-mono text-xs">--{token}</div>
      </div>
    {/each}
  </div>

  <h2 class="mt-8 text-lg font-semibold">Alpha modifiers</h2>
  <div class="mt-3 flex flex-wrap gap-2">
    {#each alphaChips as { label, cls } (label)}
      <span class="{cls} rounded-control px-3 py-1.5 text-xs font-medium">{label}</span>
    {/each}
  </div>

  <h2 class="mt-8 text-lg font-semibold">Shape &amp; elevation</h2>
  <div class="mt-3 flex flex-wrap items-center gap-4">
    <div class="rounded-control bg-surface px-4 py-3 text-sm shadow-card ring-1 ring-border">rounded-control · shadow-card</div>
    <div class="rounded-card bg-surface px-4 py-3 text-sm shadow-raised ring-1 ring-border">rounded-card · shadow-raised</div>
    <button type="button" class="focus-ring rounded-control bg-accent px-4 py-2 text-sm font-medium text-accent-foreground">
      focus-ring (Tab here)
    </button>
  </div>

  <h2 class="mt-8 text-lg font-semibold">Chart tokens (live from CSS vars)</h2>
  <p class="mt-1 text-sm text-muted-foreground">
    ECharts theme name: <code class="rounded-control bg-muted px-1.5 py-0.5 font-mono text-xs">{chartThemeName(resolved())}</code>
  </p>
  <div class="mt-3 flex flex-wrap gap-2">
    {#each resolvePalette(resolved()) as color, i (i)}
      <div class="flex flex-col items-center gap-1">
        <span class="h-10 w-10 rounded-control ring-1 ring-border" style="background-color: {color};"></span>
        <span class="font-mono text-[10px] text-muted-foreground">--chart-{i + 1}</span>
      </div>
    {/each}
    <div class="flex flex-col items-center gap-1">
      <span class="h-10 w-10 rounded-control ring-1 ring-border" style="background-color: {resolveChartMuted(resolved())};"></span>
      <span class="font-mono text-[10px] text-muted-foreground">--chart-muted</span>
    </div>
  </div>

  <h3 class="mt-6 text-sm font-semibold">Semantic chart colors (static mirror)</h3>
  <div class="mt-2 overflow-hidden rounded-card border border-border bg-surface shadow-card">
    <table class="w-full text-left text-sm">
      <thead>
        <tr class="border-b border-border bg-muted/50 text-muted-foreground">
          <th class="px-4 py-2 font-medium">Role</th>
          <th class="px-4 py-2 font-medium">Light</th>
          <th class="px-4 py-2 font-medium">Dark</th>
        </tr>
      </thead>
      <tbody>
        {#each semanticKeys as k (k)}
          <tr class="border-b border-border last:border-0">
            <td class="px-4 py-2">{k}</td>
            {#each themeModes as mode (mode)}
              <td class="px-4 py-2">
                <span class="inline-flex items-center gap-2">
                  <span
                    class="h-4 w-4 rounded-sm ring-1 ring-border"
                    style="background-color: {CHART_SEMANTIC_COLORS[mode][k]};"
                  ></span>
                  <code class="font-mono text-xs">{CHART_SEMANTIC_COLORS[mode][k]}</code>
                </span>
              </td>
            {/each}
          </tr>
        {/each}
        <tr>
          <td class="px-4 py-2">chart-muted (CSS token)</td>
          <td class="px-4 py-2" colspan="2">
            <span class="inline-flex items-center gap-2">
              <span class="h-4 w-4 rounded-sm ring-1 ring-border" style="background-color: {CHART_MUTED.light};"></span>
              <code class="font-mono text-xs">{CHART_MUTED.light}</code>
              <span class="h-4 w-4 rounded-sm ring-1 ring-border" style="background-color: {CHART_MUTED.dark};"></span>
              <code class="font-mono text-xs">{CHART_MUTED.dark}</code>
            </span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <p class="mt-6 text-xs text-muted-foreground">
    Page-level styling uses tokens only; existing pages keep their legacy colors until commit D.1b.
  </p>
</div>
