<script lang="ts">
  import '../app.css'
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import { resolve } from '$app/paths'
  import { goto } from '$app/navigation'
  import { auth, initAuth } from '$lib/stores/auth.svelte'
  import '$lib/stores/theme.svelte' // side effect: theme listeners + <html class="dark"> sync
  import { assetApi } from '$lib/services/api'
  import AppShell from '$lib/components/layout/AppShell.svelte'
  import Spinner from '$lib/components/ui/Spinner.svelte'
  import Toaster from '$lib/components/Toaster.svelte'

  let { children } = $props()

  let synced = $state(false)

  onMount(() => {
    initAuth()
  })

  $effect(() => {
    if (auth.user && !synced) {
      synced = true
      assetApi.sync().catch(() => {
        // keep going; individual pages backfill what they need
      })
    }
  })

  $effect(() => {
    if (auth.isLoading) return
    if (!auth.user && page.url.pathname !== '/login') {
      goto(resolve('/login'), { replaceState: true })
    } else if (auth.user && page.url.pathname === '/login') {
      goto(resolve('/'), { replaceState: true })
    }
  })
</script>

{#if auth.isLoading}
  <div class="grid h-dvh place-items-center bg-background">
    <Spinner size="lg" class="text-accent-text" />
  </div>
{:else if auth.user}
  <AppShell>{@render children()}</AppShell>
{:else}
  {@render children()}
{/if}

<Toaster />
