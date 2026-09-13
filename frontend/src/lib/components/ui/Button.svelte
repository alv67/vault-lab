<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAnchorAttributes, HTMLButtonAttributes } from 'svelte/elements'
  import Spinner from './Spinner.svelte'
  import { cx } from './utils'

  /**
   * Design-system button (EPIC D.2). Renders a `<button>`, or an `<a>` with the
   * same visual when `href` is set (SvelteKit enhances the navigation).
   *
   * Danger note: the danger fill uses `text-surface` instead of
   * `text-accent-foreground` (white) because the `negative` token lightens in
   * dark mode, where white text would fall below AA contrast; `surface`
   * inverts with the theme and stays AA in both.
   */
  let {
    variant = 'primary',
    size = 'md',
    disabled = false,
    loading = false,
    href = undefined,
    type = 'button',
    class: className = '',
    children = undefined,
    onclick,
    ...rest
  }: {
    variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger' | 'link'
    size?: 'sm' | 'md' | 'icon'
    loading?: boolean
    /** When set the component renders an `<a>` (pair with `aria-label` if icon-only). */
    href?: string
    children?: Snippet
  } & Omit<HTMLButtonAttributes, 'href'> = $props()

  const variants = {
    primary: 'bg-accent text-accent-foreground hover:bg-accent-hover',
    secondary: 'border border-input bg-surface text-foreground hover:bg-muted',
    outline: 'border border-input bg-transparent text-foreground hover:bg-muted',
    ghost: 'border border-transparent bg-transparent text-foreground hover:bg-muted',
    danger: 'bg-negative text-surface hover:bg-negative/90',
    link: 'text-accent-text underline decoration-1 hover:opacity-80',
  } as const

  // `link` is inline text: no fixed box, only the font step of the size.
  const sizes = {
    sm: 'h-8 px-3 text-xs',
    md: 'h-9 px-4 text-sm',
    icon: 'h-9 w-9 p-0',
  } as const

  const classes = $derived(
    cx(
      'inline-flex items-center justify-center gap-2 rounded-control font-medium transition-colors focus-ring',
      'disabled:pointer-events-none disabled:opacity-50',
      variants[variant],
      variant === 'link'
        ? size === 'sm'
          ? 'text-xs'
          : 'text-sm'
        : sizes[size],
      className,
    ),
  )

  // `loading` blocks interaction as well as signalling it (aria-busy below).
  const blocked = $derived(disabled || loading)
</script>

{#if href !== undefined}
  <!-- Callers pass an href they resolved themselves ($app/paths `resolve()`),
       per the repo convention the rule below enforces on raw `<a href>`. -->
  <!-- eslint-disable svelte/no-navigation-without-resolve -->
  <a
    {href}
    class={classes}
    aria-disabled={blocked || undefined}
    aria-busy={loading || undefined}
    onclick={onclick as HTMLAnchorAttributes['onclick']}
    {...(rest as unknown as HTMLAnchorAttributes)}
  >
    {#if loading}
      <Spinner size="sm" />
    {/if}
    {@render children?.()}
  </a>
  <!-- eslint-enable svelte/no-navigation-without-resolve -->
{:else}
  <button
    {type}
    disabled={blocked}
    class={classes}
    aria-busy={loading || undefined}
    {onclick}
    {...rest}
  >
    {#if loading}
      <Spinner size="sm" />
    {/if}
    {@render children?.()}
  </button>
{/if}
