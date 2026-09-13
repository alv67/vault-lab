import type { ClassValue } from 'svelte/elements'

/**
 * Shared helpers for the internal design system (EPIC D.2).
 */

/**
 * Joins class names, dropping falsy values (`undefined`, `null`, `false`, `''`)
 * and flattening arrays/dictionaries so callers can pass anything that is
 * assignable to the native `class` prop (`ClassValue` in Svelte 5 =
 * `string | array | dictionary`).
 *
 * Purely concatenative: later classes do NOT override earlier ones the way
 * `tailwind-merge` would, so a caller-provided `class` must not fight the
 * variant classes (same pattern the pages already follow).
 */
export function cx(...classes: Array<ClassValue | boolean | null | undefined>): string {
  const out: string[] = []
  const walk = (value: ClassValue | boolean | null | undefined): void => {
    if (!value || value === true) return
    if (typeof value === 'string') {
      out.push(value)
    } else if (Array.isArray(value)) {
      for (const item of value) walk(item)
    } else {
      for (const [key, enabled] of Object.entries(value)) {
        if (enabled) out.push(key)
      }
    }
  }
  for (const value of classes) walk(value)
  return out.join(' ')
}

/**
 * Z-index scale — every fixed/overlay layer must pick one of these four so
 * stacked surfaces never fight each other:
 *
 *   dropdown / popover  z-20   (menus, autocomplete lists)
 *   drawer              z-30   (off-canvas panels)
 *   modal               z-40   (Modal.svelte)
 *   toast               z-50   (Toaster.svelte)
 *
 * NOTE: the pre-D exposure modals still sit on z-50 and are migrated to
 * `Modal` in EPIC E; until then they coexist with the toasts by design.
 */
