/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  // The theme is driven by the `.dark` class on <html> (set pre-paint by the
  // bootstrap script in app.html and kept in sync by lib/stores/theme.svelte.ts).
  darkMode: 'class',
  theme: {
    extend: {
      // Semantic design tokens (EPIC D). HSL triples are defined in app.css
      // (:root / .dark) and composed here so alpha modifiers work:
      // `bg-accent/10`, `border-border`, `text-muted-foreground`, ...
      colors: {
        background: 'hsl(var(--background) / <alpha-value>)',
        foreground: 'hsl(var(--foreground) / <alpha-value>)',
        // 4-step elevation ladder (EPIC K.1a): surface-0 = app bg,
        // surface-1 = card, surface-2 = raised/drawer, surface-3 =
        // popover/tooltip. `DEFAULT`/`raised` keep the pre-K.1a classes
        // (`bg-surface`, `bg-surface-raised`) working unchanged.
        surface: {
          DEFAULT: 'hsl(var(--surface-1) / <alpha-value>)',
          0: 'hsl(var(--surface-0) / <alpha-value>)',
          1: 'hsl(var(--surface-1) / <alpha-value>)',
          2: 'hsl(var(--surface-2) / <alpha-value>)',
          3: 'hsl(var(--surface-3) / <alpha-value>)',
          raised: 'hsl(var(--surface-2) / <alpha-value>)',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted) / <alpha-value>)',
          foreground: 'hsl(var(--muted-foreground) / <alpha-value>)',
        },
        border: 'hsl(var(--border) / <alpha-value>)',
        input: 'hsl(var(--input) / <alpha-value>)',
        ring: 'hsl(var(--ring) / <alpha-value>)',
        accent: {
          DEFAULT: 'hsl(var(--accent) / <alpha-value>)',
          hover: 'hsl(var(--accent-hover) / <alpha-value>)',
          foreground: 'hsl(var(--accent-foreground) / <alpha-value>)',
          text: 'hsl(var(--accent-text) / <alpha-value>)',
        },
        positive: 'hsl(var(--positive) / <alpha-value>)',
        negative: 'hsl(var(--negative) / <alpha-value>)',
        warning: 'hsl(var(--warning) / <alpha-value>)',
        info: {
          DEFAULT: 'hsl(var(--info) / <alpha-value>)',
          foreground: 'hsl(var(--info-foreground) / <alpha-value>)',
        },
        overlay: 'hsl(var(--overlay) / <alpha-value>)',
        // Raw hex tokens consumed by both CSS (`bg-chart-1`) and the ECharts
        // canvas (lib/chartTheme.ts / lib/chartPalette.ts) — no <alpha-value>
        // here on purpose: the vars hold literal colors, not HSL triples.
        chart: {
          1: 'var(--chart-1)',
          2: 'var(--chart-2)',
          3: 'var(--chart-3)',
          4: 'var(--chart-4)',
          5: 'var(--chart-5)',
          6: 'var(--chart-6)',
          7: 'var(--chart-7)',
          8: 'var(--chart-8)',
          9: 'var(--chart-9)',
          10: 'var(--chart-10)',
          11: 'var(--chart-11)',
          12: 'var(--chart-12)',
          muted: 'var(--chart-muted)',
          grid: 'var(--chart-grid)',
        },
      },
      borderRadius: {
        // `rounded-control` for inputs/buttons, `rounded-card` for panels.
        control: '0.5rem',
        card: '0.75rem',
      },
      boxShadow: {
        // Elevation ramp for the surface ladder (EPIC K.1a): `shadow-card`
        // for resting surfaces (surface-1), `shadow-raised` for drawers/
        // modals (surface-2), `shadow-popover` for tooltips/popovers
        // (surface-3). In light mode the ramp is what separates the steps.
        card: '0 1px 2px 0 rgb(0 0 0 / 0.05), 0 1px 3px 0 rgb(0 0 0 / 0.1)',
        raised: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
        popover: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
      },
      fontFamily: {
        // D5 (EPIC K.1a): Inter is self-hosted via @fontsource and leads the
        // UI stack; the rest stays Tailwind's default system fallback.
        sans: [
          'Inter',
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'Roboto',
          '"Helvetica Neue"',
          'Arial',
          '"Noto Sans"',
          'sans-serif',
          '"Apple Color Emoji"',
          '"Segoe UI Emoji"',
          '"Segoe UI Symbol"',
          '"Noto Color Emoji"',
        ],
        // JetBrains Mono (self-hosted, D5) for tickers, ISINs and codes.
        mono: [
          '"JetBrains Mono"',
          'ui-monospace',
          'SFMono-Regular',
          'Menlo',
          'Consolas',
          '"Liberation Mono"',
          'monospace',
        ],
      },
      fontSize: {
        // Redesign type scale (EPIC K.1a): the named steps used by the spec
        // (hero 40px semibold tabular, micro 11px labels); defaults cover the
        // rest (h1≈text-2xl, h2≈text-lg, body≈text-sm, caption≈text-xs).
        hero: ['2.5rem', { lineHeight: '1.15', fontWeight: '600' }],
        micro: ['0.6875rem', { lineHeight: '1rem' }],
      },
      transitionDuration: {
        // Motion tokens (EPIC K.1a): 120ms confirmations, 200ms surfaces,
        // 320ms drawers/sheets. Neutralised by prefers-reduced-motion in app.css.
        fast: '120ms',
        base: '200ms',
        slow: '320ms',
      },
      transitionTimingFunction: {
        // One standard ease-out curve for all entrances/state changes.
        standard: 'cubic-bezier(0.16, 1, 0.3, 1)',
      },
    },
  },
  plugins: [],
}
