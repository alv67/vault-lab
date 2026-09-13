/**
 * Shared semantic text tones for numeric values (EPIC D / D.1b).
 *
 * Pages used to duplicate their own "green when >= 0, red otherwise" text
 * color helpers (`glClass`, `pnlClass`, `changeClass`, `totalColorClass`).
 * They all collapse into the map below, which returns design-token classes so
 * the colors flip correctly with the theme.
 */

/**
 * Text color class for a signed P/L-style value.
 *
 * Positive → `text-positive`, negative → `text-negative`. The zero case is
 * caller-selectable because the legacy call sites disagreed: portfolio and
 * dashboard tables counted 0 as a gain (historical `>= 0` formula → keep the
 * default), while the asset quote deltas rendered 0 in a muted grey.
 */
export function pnlColorClass(
  value: string | number | null | undefined,
  zeroClass: string = 'text-positive',
): string {
  const n = Number(value ?? 0)
  if (n > 0) return 'text-positive'
  if (n < 0) return 'text-negative'
  return zeroClass
}

/**
 * Text color class for the exposure-modal totals: over 100 is an error
 * (`text-negative`), a complete distribution (≥ 99.5) is a success
 * (`text-positive`) and anything below is plain foreground.
 */
export function totalColorClass(sum: number, over: boolean): string {
  if (over) return 'text-negative'
  if (sum >= 99.5) return 'text-positive'
  return 'text-foreground'
}
