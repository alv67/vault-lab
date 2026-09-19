import { cubicOut } from 'svelte/easing'
import type { TransitionConfig } from 'svelte/transition'

/**
 * Overlay transitions wired to the K.1a motion tokens (tailwind.config.js:
 * `duration-base` 200 ms for surfaces/backdrops, `duration-slow` 320 ms for
 * drawers/sheets, one ease-out curve).
 *
 * They read the numbers here because the `prefers-reduced-motion` block in
 * `app.css` only neutralises *CSS* transitions/animations: Svelte JS
 * transitions run on the frame loop and would ignore it, so the media query
 * is checked at (re)start instead — the reduced preference is honoured per
 * open, without re-creating the components when the setting changes.
 */

const DURATION_BASE = 200
const DURATION_SLOW = 320

function prefersReducedMotion(): boolean {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

/** Fade params for overlay backdrops (200 ms `duration-base`, 0 when reduced). */
export function backdropFade(): { duration: number } {
  return { duration: prefersReducedMotion() ? 0 : DURATION_BASE }
}

/**
 * Panel transition: slides in from the anchored edge and back out on exit.
 * `right` suits `Drawer.svelte`, `bottom` suits `Sheet.svelte`.
 */
export function edgeSlide(
  _node: HTMLElement,
  { edge = 'right' }: { edge?: 'right' | 'bottom' } = {},
): TransitionConfig {
  const axis = edge === 'right' ? 'X' : 'Y'
  return {
    duration: prefersReducedMotion() ? 0 : DURATION_SLOW,
    easing: cubicOut,
    css: (t, u) => `opacity: ${t}; transform: translate${axis}(${u * 100}%)`,
  }
}
