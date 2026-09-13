<script lang="ts">
  import { Pencil } from 'lucide-svelte'
  import { formatCurrency } from '$lib/format'
  import type { Transaction } from '$lib/services/api'
  import Table from '$lib/components/ui/Table.svelte'
  import THead from '$lib/components/ui/THead.svelte'
  import TBody from '$lib/components/ui/TBody.svelte'
  import Tr from '$lib/components/ui/Tr.svelte'
  import Th from '$lib/components/ui/Th.svelte'
  import Td from '$lib/components/ui/Td.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Button from '$lib/components/ui/Button.svelte'

  /**
   * Transactions table (EPIC E.2): extracted from the portfolio detail page.
   * Monetary columns use the shared right-aligned + `tabular-nums` recipe;
   * `onedit` hands the row back to the caller (usually an edit modal).
   */
  let {
    transactions = [],
    currency = 'USD',
    onedit,
  }: {
    transactions?: Transaction[]
    currency?: string
    onedit: (tx: Transaction) => void
  } = $props()

  function typeVariant(type: Transaction['type']): 'positive' | 'negative' | 'accent' | 'neutral' {
    if (type === 'buy') return 'positive'
    if (type === 'sell') return 'negative'
    if (type === 'dividend') return 'accent'
    return 'neutral'
  }
</script>

<div class="overflow-x-auto">
  <Table aria-label="Transactions">
    <THead>
      <Tr>
        <Th>Date</Th>
        <Th>Asset</Th>
        <Th>Type</Th>
        <Th align="right">Qty</Th>
        <Th align="right">Price</Th>
        <Th align="right">Total</Th>
        <Th align="right">Actions</Th>
      </Tr>
    </THead>
    <TBody>
      {#each transactions as tx (tx.id)}
        <Tr>
          <Td>{new Date(tx.date).toLocaleDateString()}</Td>
          <Td>
            <span class="font-medium">{tx.asset_ticker}</span>
            <span class="ml-1 text-xs text-muted-foreground">{tx.asset_name}</span>
          </Td>
          <Td>
            <Badge variant={typeVariant(tx.type)}>{tx.type}</Badge>
          </Td>
          <Td align="right">{tx.quantity}</Td>
          <Td align="right">{formatCurrency(tx.price, currency)}</Td>
          <Td align="right">
            {formatCurrency(Number(tx.quantity) * Number(tx.price), currency)}
          </Td>
          <Td align="right">
            <Button
              variant="ghost"
              size="icon"
              aria-label="Edit transaction"
              onclick={() => onedit(tx)}
            >
              <Pencil class="h-4 w-4" />
            </Button>
          </Td>
        </Tr>
      {/each}
    </TBody>
  </Table>
</div>
