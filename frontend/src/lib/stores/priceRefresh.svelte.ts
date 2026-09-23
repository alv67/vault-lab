/**
 * Session price-refresh outcome (EPIC K.3a, spec §8.8), module-scoped so it
 * survives SPA navigations: the shell triggers the refresh at most once per
 * session on any landing page and every consumer reads the result back here
 * (otherwise the freshness stamp / quality strip vanish after navigating
 * away and back, issue #119). Reactive so an in-flight refresh that resolves
 * after a remount still lands in the UI.
 *
 * `refreshPrices` is the SINGLE refresh path for the whole app: the automatic
 * session trigger, the header button, the Fab and the command palette all go
 * through it, so one shared `revision` counter lets price-derived page data
 * refetch on completion and concurrent triggers de-duplicate into the one
 * in-flight POST (never a double Yahoo hit while a refresh is already
 * running).
 */
import { t } from '$lib/i18n/index.svelte'
import { pricesApi, type RefreshReport } from '$lib/services/api'
import { toast } from '$lib/stores/toast.svelte'

export const priceRefresh = $state({
  /** Once-per-session guard: true as soon as the refresh has been triggered. */
  started: false,
  refreshing: false,
  /** `RefreshReport.finished_at`; empty renders nothing. */
  finishedAt: '',
  rateLimited: false,
  issueCount: 0,
  failed: false,
  /** Bumped on every completed refresh (success or failure): pages watch it
   * to refetch price-derived data after a refresh they did not trigger. */
  revision: 0,
})

/**
 * Record a completed report on the shared state: the freshness stamp, the
 * rate-limit/issues chips and a previously surfaced failure all reset from
 * the fresh outcome (`failed = false` — a new success supersedes it).
 */
export function applyRefreshReport(report: RefreshReport): void {
  priceRefresh.finishedAt = report.finished_at
  priceRefresh.rateLimited = report.rate_limited
  priceRefresh.issueCount = report.issues.length
  priceRefresh.failed = false
}

/** Shared in-flight POST (see the de-duplication note above). */
let inFlight: Promise<RefreshReport | null> | null = null

/**
 * Trigger a price refresh through the shared path and publish the outcome on
 * `priceRefresh`. Concurrent calls return the one in-flight promise. Toast
 * policy: warnings (rate-limited / partial) always fire — the user must know
 * the numbers may be stale — while the plain success toast is opt-in (manual
 * triggers) so the automatic session refresh stays silent, and the error
 * toast can be silenced by the shell's background trigger (the persistent
 * strip surfaces the failure instead). Resolves with the report, or `null`
 * when the POST itself failed.
 */
export function refreshPrices(opts?: {
  portfolioId?: string
  announceSuccess?: boolean
  announceError?: boolean
}): Promise<RefreshReport | null> {
  if (inFlight) return inFlight

  const run = async (): Promise<RefreshReport | null> => {
    priceRefresh.started = true
    priceRefresh.refreshing = true
    try {
      const report = await pricesApi.refresh(opts?.portfolioId)
      applyRefreshReport(report)
      if (report.rate_limited) {
        toast.warning(t('quickActions.refreshRateLimited'))
      } else if (report.issues.length > 0) {
        toast.warning(t('quickActions.refreshIssues', { count: report.issues.length }))
      } else if (opts?.announceSuccess) {
        toast.success(t('quickActions.refreshSuccess'))
      }
      return report
    } catch {
      priceRefresh.failed = true
      if (opts?.announceError !== false) toast.error(t('quickActions.refreshError'))
      return null
    } finally {
      priceRefresh.refreshing = false
      priceRefresh.revision += 1
    }
  }

  const promise = run()
  inFlight = promise
  // Identity-guarded release: a stale settled promise must never block the
  // next trigger (an already-resolved `.finally` still lands as a microtask,
  // so this also covers a body that completes without suspending).
  void promise.finally(() => {
    if (inFlight === promise) inFlight = null
  })
  return promise
}
