import type { EChartsType } from 'echarts/core'

/**
 * Hide an ECharts tooltip when a pointer goes down outside the node it is
 * attached to (touch has no hover/leave, so a tap otherwise leaves the
 * tooltip pinned). The chart instance is reactive: `update` keeps it current.
 */
export function dismissTooltipOutside(node: HTMLElement, chart: EChartsType | null = null) {
  let instance = chart
  function handlePointerDown(event: PointerEvent): void {
    if (instance && event.target instanceof Node && !node.contains(event.target)) {
      instance.dispatchAction({ type: 'hideTip' })
    }
  }
  window.addEventListener('pointerdown', handlePointerDown, true)
  return {
    update(next: EChartsType | null): void {
      instance = next
    },
    destroy(): void {
      window.removeEventListener('pointerdown', handlePointerDown, true)
    },
  }
}
