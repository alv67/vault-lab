/**
 * Shared focus-trap action (EPIC K.1c), extracted from the recipe established
 * by `ui/Modal.svelte` and `layout/MobileDrawer.svelte`: on activation the
 * previously focused element is captured and focus moves into the panel (so
 * screen readers announce the dialog name); Tab is cycled within the panel;
 * on deactivation/destroy the original focus is restored (when still in the
 * document). `Drawer.svelte` and `Sheet.svelte` consume it; Modal/MobileDrawer
 * keep their inline copies untouched in this phase (zero-regression rule).
 *
 * The node passed in must be the dialog panel itself and carry
 * `tabindex="-1"` so it can receive focus programmatically.
 */

/** Everything the user can Tab to, in DOM order (same selector as Modal). */
export const FOCUSABLE =
  'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'

export function focusTrap(node: HTMLElement, enabled = true) {
  let lastFocused: HTMLElement | null = null
  let active = false

  function start(): void {
    if (active) return
    active = true
    lastFocused =
      document.activeElement instanceof HTMLElement ? document.activeElement : null
    // The panel (not a child) receives focus: its role/aria-labelledby is
    // announced immediately, matching the MobileDrawer behaviour.
    node.focus()
    document.addEventListener('keydown', handleKeydown, true)
  }

  function stop(): void {
    if (!active) return
    active = false
    document.removeEventListener('keydown', handleKeydown, true)
    // Restore only when the trigger survived (page navigations can unmount
    // it while this panel is still open).
    if (lastFocused?.isConnected) lastFocused.focus()
    lastFocused = null
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Tab') return
    const focusables = Array.from(node.querySelectorAll<HTMLElement>(FOCUSABLE))
    if (focusables.length === 0) {
      event.preventDefault()
      node.focus()
      return
    }
    const first = focusables[0]
    const last = focusables[focusables.length - 1]
    const current = document.activeElement
    if (event.shiftKey && (current === first || current === node)) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && current === last) {
      event.preventDefault()
      first.focus()
    }
  }

  if (enabled) start()

  return {
    update(next: boolean): void {
      if (next) start()
      else stop()
    },
    destroy(): void {
      stop()
    },
  }
}
