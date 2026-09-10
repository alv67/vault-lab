export const CHART_PALETTE = [
  '#2563eb', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
  '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16',
  '#06b6d4', '#a855f7',
]

export interface ChartRow {
  name: string
  weight: string | number
}

// Restituisce il colore che ECharts assegna a una riga in un pie chart:
// i colori sono dati per indice sulle sole righe con weight > 0 (ordine della
// palette), quindi il quadratino in tabella combacia sempre con la fetta.
export function colorForRow(row: ChartRow, rows: ChartRow[]): string {
  const visible = rows.filter((r) => Number(r.weight) > 0)
  const index = visible.findIndex((r) => r.name === row.name)
  if (index < 0) {
    return '#9ca3af'
  }
  return CHART_PALETTE[index % CHART_PALETTE.length]
}
