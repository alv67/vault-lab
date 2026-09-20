<script lang="ts">
  import TransactionTable from '$lib/components/domain/TransactionTable.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import { getPortfolioPage } from '../context'

  /**
   * Activity tab (EPIC K.4a): the paginated transactions table exactly as
   * it lived on the old single page (EPIC I.9, #88) — same columns, same
   * edit-pencil flow into the shell-owned modal, same Previous/Next footer
   * with the "1–20 of 137" range label. The pagination state and its
   * monotonic-request-id guarded fetches live in the layout, so flipping
   * pages refetches only the window (and survives tab switches). Filters
   * and sheet-based editing land later in K.4c.
   */
  const ctx = getPortfolioPage()
  const currency = $derived(ctx.currency)
</script>

<div class="rounded-card border-border bg-surface p-4 shadow-card">
  <h2 class="mb-4 font-semibold">Transactions</h2>
  <TransactionTable
    transactions={ctx.transactions ?? []}
    {currency}
    onedit={ctx.editTransaction}
  />
  <!-- EPIC I.9 (#88): pagination footer with the same "1–20 of 137" range
       label and Previous/Next layout used by the admin health page. -->
  <div class="mt-3 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-3">
    <span class="text-sm tabular-nums text-muted-foreground">{ctx.txRangeLabel}</span>
    <div class="flex items-center gap-2">
      <Button
        variant="secondary"
        size="sm"
        disabled={ctx.txOffset === 0 || ctx.txLoading}
        onclick={() => ctx.gotoTxPage(ctx.txPage - 1)}
      >
        Previous
      </Button>
      <Button
        variant="secondary"
        size="sm"
        disabled={ctx.txOffset + ctx.txLimit >= ctx.txTotal || ctx.txLoading}
        onclick={() => ctx.gotoTxPage(ctx.txPage + 1)}
      >
        Next
      </Button>
    </div>
  </div>
</div>
