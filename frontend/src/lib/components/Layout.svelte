<script lang="ts">
  import type { Snippet } from 'svelte'
  import { resolve } from '$app/paths'
  import { LayoutDashboard, Briefcase, Banknote, LogOut } from 'lucide-svelte'
  import { auth, logout } from '$lib/stores/auth.svelte'

  type IconType = typeof LayoutDashboard

  let { children }: { children: Snippet } = $props()

  const navItems = [
    { to: '/', label: 'Dashboard', icon: LayoutDashboard as IconType },
    { to: '/portfolios', label: 'Portfolios', icon: Briefcase as IconType },
    { to: '/assets', label: 'Assets', icon: Banknote as IconType },
  ] as const
</script>

<div class="flex h-screen">
  <aside class="flex w-64 flex-col border-r bg-white">
    <div class="flex items-center gap-2 border-b px-6 py-4">
      <Banknote class="h-6 w-6 text-blue-600" />
      <span class="text-lg font-bold">VaultLab</span>
    </div>
    <nav class="flex-1 space-y-1 p-4">
      {#each navItems as item (item.to)}
        {@const Icon = item.icon}
        <a
          href={resolve(item.to)}
          class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
        >
          <Icon class="h-5 w-5" />
          {item.label}
        </a>
      {/each}
    </nav>
    <div class="border-t p-4">
      <div class="mb-2 text-sm text-gray-500">{auth.user?.email}</div>
      <button
        onclick={logout}
        class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-red-600 hover:bg-red-50"
      >
        <LogOut class="h-5 w-5" />
        Sign out
      </button>
    </div>
  </aside>
  <main class="flex-1 overflow-auto">
    {@render children()}
  </main>
</div>
