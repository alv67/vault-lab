<script lang="ts">
  import '../app.css'
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import { resolve } from '$app/paths'
  import { goto } from '$app/navigation'
  import { auth, initAuth } from '$lib/stores/auth.svelte'
  import { assetApi } from '$lib/services/api'
  import Layout from '$lib/components/Layout.svelte'
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
  <div class="flex h-screen items-center justify-center">
    <div class="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent"></div>
  </div>
{:else if auth.user}
  <Layout>{@render children()}</Layout>
{:else}
  {@render children()}
{/if}

<Toaster />
