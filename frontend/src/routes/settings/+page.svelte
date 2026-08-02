<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { auth, updateProfile } from '$lib/stores/auth.svelte'
  import { authApi } from '$lib/services/api'

  let name = $state(auth.user?.name ?? '')
  let email = $state(auth.user?.email ?? '')
  let savingProfile = $state(false)

  let currentPassword = $state('')
  let newPassword = $state('')
  let confirmPassword = $state('')
  let savingPassword = $state(false)

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
</div>
