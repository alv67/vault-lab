<script lang="ts">
  import type { Snippet } from 'svelte'
  import { afterNavigate } from '$app/navigation'
  import { fade, slide } from 'svelte/transition'

  /**
   * Off-canvas navigation drawer below `lg` (EPIC D.3): the same `Sidebar`
   * the desktop rail shows, rendered over a dimmed backdrop on the drawer
   * tier (z-30) of the z-index scale (../ui/utils.ts).
   *
   * Follows the focus recipe established by ui/Modal.svelte: focus moves
   * into the panel on open, Tab is trapped inside it, and focus returns to
   * the hamburger on close — except when the close was caused by a route
   * change (the user clicked a nav link, so the new page should own focus).
   * Esc and backdrop clicks also close it; `open` is `$bindable`.
   */
  let {
    open = $bindable(false),
    children,
  }: {
    open?: boolean
    children?: Snippet
  } = $props()

  let panel = $state<HTMLDivElement | null>(null)
  let lastFocused: HTMLElement | null = null
  let suppressRestore = false

  const FOCUSABLE =
    'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'

  function close(): void {
    open = false
  }

  function handleBackdropClick(event: MouseEvent): void {
    // Only clicks that land on the overlay itself (not bubbled from the panel).
    if (event.target === event.currentTarget) close()
  }

  // Handled on the dialog root (not `<svelte:window>`): focus is moved into
  // the panel on open and Tab-trapped below, so every keypress while the
  // drawer is open bubbles through here.
  function handleKeydown(event: KeyboardEvent): void {
    if (!open) return
    if (event.key === 'Escape') {
      close()
      return
    }
    if (event.key !== 'Tab' || !panel) return
    const focusables = Array.from(panel.querySelectorAll<HTMLElement>(FOCUSABLE))
    if (focusables.length === 0) {
      event.preventDefault()
      panel.focus()
      return
    }
    const first = focusables[0]
    const last = focusables[focusables.length - 1]
    const active = document.activeElement
    if (event.shiftKey && (active === first || active === panel)) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && active === last) {
      event.preventDefault()
      first.focus()
    }
  }

  // Route change (nav link clicked inside the drawer): close without
  // restoring focus to the hamburger — see `suppressRestore` above.
  afterNavigate(() => {
    if (!open) return
    suppressRestore = true
    open = false
  })

  // Capture/restore keyed on `open`; the cleanup form also covers the drawer
  // being destroyed while open (e.g. logout unmounts the shell).
  $effect(() => {
    if (!open) return
    lastFocused =
      document.activeElement instanceof HTMLElement ? document.activeElement : null
    panel?.focus()
    return () => {
      if (!suppressRestore) lastFocused?.focus()
      suppressRestore = false
      lastFocused = null
    }
  })
</script>

{#if open}
  <div
    class="fixed inset-0 z-30 bg-overlay/50 lg:hidden"
    role="dialog"
    aria-modal="true"
    aria-label="Navigation"
    tabindex="-1"
    transition:fade={{ duration: 150 }}
    onclick={handleBackdropClick}
    onkeydown={handleKeydown}
  >
    <div
      bind:this={panel}
      class="absolute inset-y-0 left-0 flex max-w-[85vw] overflow-hidden bg-surface shadow-raised outline-none"
      tabindex="-1"
      transition:slide={{ axis: 'x' }}
    >
      {@render children?.()}
    </div>
  </div>
{/if}
