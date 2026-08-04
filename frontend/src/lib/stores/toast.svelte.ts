export type ToastType = 'success' | 'error' | 'warning'

interface ToastItem {
  id: number
  type: ToastType
  message: string
}

const items = $state<ToastItem[]>([])

let nextId = 1

function show(type: ToastType, message: string, duration = 3500): void {
  const id = nextId++
  items.push({ id, type, message })
  setTimeout(() => dismiss(id), duration)
}

function dismiss(id: number): void {
  const index = items.findIndex((t) => t.id === id)
  if (index !== -1) items.splice(index, 1)
}

export const toasts = items

export const toast = {
  success: (message: string) => show('success', message),
  error: (message: string) => show('error', message),
  warning: (message: string) => show('warning', message, 4500),
}
