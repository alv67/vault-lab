export type ToastType = 'success' | 'error' | 'warning'

/** Inline action button (EPIC K.4c, decision D11): rendered inside the toast
 * card as a real `<button>` (keyboard-reachable), invokes `onclick` and
 * dismisses the toast when clicked. Used by the "undo" delete toasts. */
export interface ToastAction {
  label: string
  onclick: () => void
}

interface ToastItem {
  id: number
  type: ToastType
  message: string
  action?: ToastAction
}

/** Per-toast overrides accepted by the `toast.*` helpers. */
export interface ToastOptions {
  /** Auto-dismiss delay in ms (defaults: 3500, 4500 for warnings). */
  duration?: number
  /** Undo-style action rendered next to the message. */
  action?: ToastAction
}

const DEFAULT_DURATIONS: Record<ToastType, number> = {
  success: 3500,
  error: 3500,
  warning: 4500,
}

const items = $state<ToastItem[]>([])

let nextId = 1

function show(type: ToastType, message: string, options: ToastOptions = {}): void {
  const id = nextId++
  items.push({ id, type, message, action: options.action })
  setTimeout(() => dismiss(id), options.duration ?? DEFAULT_DURATIONS[type])
}

function dismiss(id: number): void {
  const index = items.findIndex((t) => t.id === id)
  if (index !== -1) items.splice(index, 1)
}

export const toasts = items

export const toast = {
  success: (message: string, options?: ToastOptions) => show('success', message, options),
  error: (message: string, options?: ToastOptions) => show('error', message, options),
  warning: (message: string, options?: ToastOptions) => show('warning', message, options),
  /** Public dismissal used by the Toaster's close buttons (EPIC D.2) and by
   * the action buttons, which dismiss the toast after running their click
   * handler (EPIC K.4c). */
  dismiss,
}
