<script lang="ts">
  import { RefreshCw } from 'lucide-svelte'
  import { t } from '$lib/i18n/index.svelte'
  import { cx } from '../ui/utils'

  /**
   * "Prices as of HH:MM" stamp (EPIC K.3a, redesign spec §8.5): surfaces the
   * session `pricesApi.refresh()` outcome right next to the hero figure, so
   * freshness is answered where the number lives instead of in a transient
   * toast (the toasts stay for rate-limit warnings). Purely display — the
   * once-per-session refresh and its refetch logic live in the page.
   *
   * While the session refresh is in flight `refreshing` renders a muted
   * spinning glyph with an sr-only label; a completed-but-partial outcome
   * (rate-limited or failed issues) tints the stamp with the `--info` token
   * instead of failing silently.
   */
  let {
    /** `RefreshReport.finished_at`; empty/blank renders nothing. */
    finishedAt = '',
    refreshing = false,
    /** Rate-limited or ≥1 failed fetch: the timestamp stands, but partial. */
    partial = false,
    class: className = '',
  }: {
    finishedAt?: string
    refreshing?: boolean
    partial?: boolean
    class?: string
  } = $props()

  const time = $derived.by(() => {
    if (!finishedAt) return ''
    const d = new Date(finishedAt)
    return Number.isNaN(d.getTime()) ? '' : d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })

  const title = $derived(partial ? t('freshness.partialHint', { time }) : t('freshness.hint'))
</script>

{#if refreshing}
  <span
    class={cx('inline-flex items-center gap-1 text-micro text-muted-foreground', className)}
    role="status"
  >
    <RefreshCw class="h-3 w-3 animate-spin" aria-hidden="true" />
    {t('freshness.refreshing')}
  </span>
{:else if time}
  <span
    class={cx(
      'inline-flex items-center gap-1 text-micro tabular-nums',
      partial ? 'text-info' : 'text-muted-foreground',
      className,
    )}
    title={title}
  >
    <RefreshCw class="h-3 w-3" aria-hidden="true" />
    {t('freshness.asOf', { time })}
  </span>
{/if}
