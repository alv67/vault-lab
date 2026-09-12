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
        surface: 'hsl(var(--surface) / <alpha-value>)',
        'surface-raised': 'hsl(var(--surface-raised) / <alpha-value>)',
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
        },
      },
      borderRadius: {
        // `rounded-control` for inputs/buttons, `rounded-card` for panels.
        control: '0.5rem',
        card: '0.75rem',
      },
      boxShadow: {
        // `shadow-card` for resting surfaces, `shadow-raised` for popovers/modals.
        card: '0 1px 2px 0 rgb(0 0 0 / 0.05), 0 1px 3px 0 rgb(0 0 0 / 0.1)',
        raised: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
      },
      fontFamily: {
        // Explicit system-ui stack (kept identical to Tailwind's default
        // `sans` so the token layer changes nothing visually in light mode).
        sans: [
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
      },
    },
  },
  plugins: [],
}
