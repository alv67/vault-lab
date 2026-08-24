<script lang="ts">
  import { onMount } from 'svelte'
  import { toast } from '$lib/stores/toast.svelte'
  import { auth, updateProfile } from '$lib/stores/auth.svelte'
  import { authApi, settingsApi, type Currency } from '$lib/services/api'
  import { currencySymbol } from '$lib/format'
  import { Trash2, Activity } from 'lucide-svelte'
  import { page } from '$app/stores'
  import { goto } from '$app/navigation'
  
  let name = $state(auth.user?.name ?? '')
  let email = $state(auth.user?.email ?? '')
  let savingProfile = $state(false)
  
  let currentPassword = $state('')
  let newPassword = $state('')
  let confirmPassword = $state('')
  let savingPassword = $state(false)
  
  let currencies = $state<Currency[]>([])
  let currenciesLoading = $state(true)
  let newCode = $state('')
  let newName = $state('')

  let addingCurrency = $state(false)
  let removingCode = $state('')

  onMount(async () => {
    try {
      currencies = await listCurrencies()
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load currencies'
      toast.error(message)
    } finally {
      currenciesLoading = false
    }
  })

  async function listCurrencies(): Promise<Currency[]> {
    const res = await settingsApi.listCurrencies()
    return res.currencies
  }

  function errorStatus(err: unknown): number | undefined {
    return err instanceof Error && 'status' in err ? (err as Error & { status: number }).status : undefined
  }

  async function addCurrency(): Promise<void> {
    const code = newCode.trim().toUpperCase()
    if (!/^[A-Z]{3}$/.test(code)) {
      toast.error('Code must be 3 uppercase letters (e.g. GBP)')
      return
    }
    addingCurrency = true
    try {
      await settingsApi.addCurrency(code, newName.trim() || undefined)
      currencies = await listCurrencies()
      newCode = ''
      newName = ''
      toast.success(`Currency ${code} added`)
    } catch (err: unknown) {
      const status = errorStatus(err)
      if (status === 422) {
        toast.error(`USD->${code} conversion not available; currency not manageable`)
      } else if (status === 409) {
        toast.error('Currency already present')
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to add currency')
      }
    } finally {
      addingCurrency = false
    }
  }

  async function removeCurrency(code: string): Promise<void> {
    if (!confirm(`Delete currency ${code}?`)) return
    removingCode = code
    try {
      await settingsApi.deleteCurrency(code)
      currencies = await listCurrencies()
      toast.success(`Currency ${code} removed`)
    } catch (err: unknown) {
      const status = errorStatus(err)
      if (status === 409) {
        toast.error('Currency in use or protected')
      } else {
        toast.error(err instanceof Error ? err.message : 'Failed to remove currency')
      }
    } finally {
      removingCode = ''
    }
  }

  async function saveProfile(): Promise<void> {
    savingProfile = true
    try {
      await updateProfile(name, email)
      toast.success('Profile updated')
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Update failed'
      toast.error(message)
    } finally {
      savingProfile = false
    }
  }

  async function savePassword(): Promise<void> {
    if (newPassword !== confirmPassword) {
      toast.error('New passwords do not match')
      return
    }
    savingPassword = true
    try {
      await authApi.changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      })
      toast.success('Password changed')
      currentPassword = ''
      newPassword = ''
      confirmPassword = ''
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Change failed'
      toast.error(message)
    } finally {
      savingPassword = false
    }
  }
</script>

<div class="p-6">
  <h1 class="mb-6 text-2xl font-bold">Settings</h1>

  <div class="mb-6 max-w-lg rounded-xl border bg-white p-6">
    <h2 class="mb-4 font-semibold">Profile</h2>
    <div class="space-y-4">
      <div>
        <label for="profile-name" class="mb-1 block text-sm text-gray-600">Name</label>
        <input
          id="profile-name"
          type="text"
          bind:value={name}
          class="w-full rounded-lg border px-3 py-2 text-sm"
        />
      </div>
      <div>
        <label for="profile-email" class="mb-1 block text-sm text-gray-600">Email</label>
        <input
          id="profile-email"
          type="email"
          bind:value={email}
          class="w-full rounded-lg border px-3 py-2 text-sm"
        />
      </div>
      <button
        onclick={saveProfile}
        disabled={savingProfile || !name || !email}
        class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {savingProfile ? 'Saving...' : 'Save'}
      </button>
    </div>
  </div>

  <div class="max-w-lg rounded-xl border bg-white p-6">
    <h2 class="mb-4 font-semibold">Change password</h2>
    <div class="space-y-4">
      <div>
        <label for="current-password" class="mb-1 block text-sm text-gray-600">Current password</label>
        <input
          id="current-password"
          type="password"
          bind:value={currentPassword}
          class="w-full rounded-lg border px-3 py-2 text-sm"
          autocomplete="current-password"
        />
      </div>
      <div>
        <label for="new-password" class="mb-1 block text-sm text-gray-600">New password</label>
        <input
          id="new-password"
          type="password"
          bind:value={newPassword}
          class="w-full rounded-lg border px-3 py-2 text-sm"
          autocomplete="new-password"
        />
      </div>
      <div>
        <label for="confirm-password" class="mb-1 block text-sm text-gray-600">Confirm new password</label>
        <input
          id="confirm-password"
          type="password"
          bind:value={confirmPassword}
          class="w-full rounded-lg border px-3 py-2 text-sm"
          autocomplete="new-password"
        />
      </div>
      <button
        onclick={savePassword}
        disabled={savingPassword || !currentPassword || !newPassword || !confirmPassword}
        class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {savingPassword ? 'Saving...' : 'Change password'}
      </button>
    </div>
  </div>

  <div class="mt-6 max-w-lg rounded-xl border bg-white p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="font-semibold">Infrastruttura</h2>
      <a
        href="/settings/health"
        class="flex items-center gap-1.5 rounded-lg bg-gray-100 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-200"
      >
        <Activity class="h-4 w-4" />
        Health Dashboard
      </a>
    </div>
  </div>

  <div class="mt-6 max-w-lg rounded-xl border bg-white p-6">
    <h2 class="mb-4 font-semibold">Valute gestite</h2>

    <div class="mb-4 flex gap-3">
      <input
        type="text"
        placeholder="Code (e.g. GBP)"
        bind:value={newCode}
        maxlength="3"
        oninput={(e) => { newCode = e.currentTarget.value.toUpperCase() }}
        class="w-28 rounded-lg border px-3 py-2 text-sm uppercase"
      />
      <input
        type="text"
        placeholder="Name (optional)"
        bind:value={newName}
        class="flex-1 rounded-lg border px-3 py-2 text-sm"
      />
      <button
        onclick={addCurrency}
        disabled={addingCurrency}
        class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {addingCurrency ? 'Adding...' : 'Add'}
      </button>
    </div>

    {#if currenciesLoading}
      <p class="text-gray-500">Loading...</p>
    {:else if currencies.length === 0}
      <p class="text-sm text-gray-400">No currencies found.</p>
    {:else}
      <table class="w-full text-left text-sm">
        <thead>
          <tr class="border-b text-gray-500">
            <th class="pb-2 font-medium">Code</th>
            <th class="pb-2 font-medium">Name</th>
            <th class="pb-2 font-medium text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each currencies as c (c.code)}
            <tr class="border-b last:border-0">
              <td class="py-2 font-medium">{c.code}</td>
              <td class="py-2 text-gray-600">{c.name || '—'} <span class="text-xs text-gray-400">{currencySymbol(c.code)}</span></td>
              <td class="py-2 text-right">
                <button
                  onclick={() => removeCurrency(c.code)}
                  disabled={removingCode === c.code}
                  class="rounded-lg p-1.5 text-gray-400 hover:text-red-600 disabled:opacity-50"
                  title="Remove currency"
                >
                  <Trash2 class="h-4 w-4" />
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>
