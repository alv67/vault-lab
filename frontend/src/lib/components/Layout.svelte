<script lang="ts">
  import type { Snippet } from 'svelte'
  import { resolve } from '$app/paths'
  import { LayoutDashboard, Briefcase, Banknote, LogOut, Settings, ChevronUp } from 'lucide-svelte'
  import { auth, logout } from '$lib/stores/auth.svelte'

  type IconType = typeof LayoutDashboard

  let { children }: { children: Snippet } = $props()

  let menuOpen = $state(false)

  const navItems = [
    { to: '/', label: 'Dashboard', icon: LayoutDashboard as IconType },
    { to: '/portfolios', label: 'Portfolios', icon: Briefcase as IconType },
    { to: '/assets', label: 'Assets', icon: Banknote as IconType },
  ] as const

  function initials(): string {
    const name = auth.user?.name?.trim()
    if (!name) return '?'
    const parts = name.split(/\s+/)
    return (parts[0][0] + (parts[1]?.[0] ?? '')).toUpperCase()
  }
</script>

<div class="flex h-screen">
  <aside class="flex w-64 flex-col border-r border-border bg-surface">
    <div class="flex items-center gap-2 border-b border-border px-6 py-4">
      <Banknote class="h-6 w-6 text-accent-text" />
      <span class="text-lg font-bold">VaultLab</span>
    </div>
    <nav class="flex-1 space-y-1 p-4">
      {#each navItems as item (item.to)}
        {@const Icon = item.icon}
        <a
          href={resolve(item.to)}
          class="flex items-center gap-3 rounded-control px-3 py-2 text-sm font-medium text-foreground hover:bg-muted"
        >
          <Icon class="h-5 w-5" />
          {item.label}
        </a>
      {/each}
    </nav>
    <div class="relative border-t border-border p-4">
      {#if menuOpen}
        <button
          class="fixed inset-0 z-10 cursor-default"
          aria-label="Close menu"
          onclick={() => (menuOpen = false)}
        ></button>
        <div
          class="absolute bottom-full left-4 z-20 mb-2 w-56 overflow-hidden rounded-card border border-border bg-surface shadow-raised"
        >
          <div class="border-b border-border px-4 py-3">
            <div class="text-sm font-medium text-foreground">{auth.user?.name || 'User'}</div>
            <div class="truncate text-xs text-muted-foreground">{auth.user?.email}</div>
          </div>
          <a
            href={resolve('/settings')}
            onclick={() => (menuOpen = false)}
            class="flex items-center gap-2 px-4 py-2.5 text-sm text-foreground hover:bg-muted"
          >
            <Settings class="h-4 w-4" />
            Settings
          </a>
          <button
            onclick={logout}
            class="flex w-full items-center gap-2 px-4 py-2.5 text-sm text-negative hover:bg-negative/10"
          >
            <LogOut class="h-4 w-4" />
            Sign out
          </button>
        </div>
      {/if}
      <button
        onclick={() => (menuOpen = !menuOpen)}
        class="flex w-full items-center gap-2 rounded-control px-3 py-2 text-sm text-foreground hover:bg-muted"
      >
        <span
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-accent text-xs font-semibold text-accent-foreground"
        >
          {initials()}
        </span>
        <span class="flex-1 truncate text-left font-medium">{auth.user?.name || auth.user?.email || 'User'}</span>
        <ChevronUp class="h-4 w-4 text-muted-foreground" />
      </button>
    </div>
  </aside>
  <main class="flex-1 overflow-auto">
    {@render children()}
  </main>
</div>
