<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { authApi } from '$lib/services/api'
  import SettingsTabs from '$lib/components/domain/SettingsTabs.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'

  let currentPassword = $state('')
  let newPassword = $state('')
  let confirmPassword = $state('')
  let savingPassword = $state(false)

  let currentError = $state<string | undefined>(undefined)
  let newPasswordError = $state<string | undefined>(undefined)
  let confirmPasswordError = $state<string | undefined>(undefined)

  function errorStatus(err: unknown): number | undefined {
    return err instanceof Error && 'status' in err ? (err as Error & { status: number }).status : undefined
  }

  function validate(): boolean {
    if (!currentPassword) currentError = 'Current password is required'
    if (newPassword.length < 8) newPasswordError = 'Password must be at least 8 characters'
    if (confirmPassword !== newPassword) confirmPasswordError = 'Passwords do not match'
    return !currentError && !newPasswordError && !confirmPasswordError
  }

  async function savePassword(): Promise<void> {
    currentError = undefined
    newPasswordError = undefined
    confirmPasswordError = undefined
    if (!validate()) return
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
      const status = errorStatus(err)
      const message = err instanceof Error ? err.message : 'Change failed'
      if (status === 401) {
        currentError = 'Current password is incorrect'
      } else if (status === 400) {
        newPasswordError = message
      } else {
        toast.error(message)
      }
    } finally {
      savingPassword = false
    }
  }
</script>

<div class="p-6">
  <h1 class="mb-6 text-2xl font-bold">Settings</h1>

  <SettingsTabs class="mb-6" />

  <Card class="max-w-lg p-6">
    <h2 class="mb-4 font-semibold">Change password</h2>
    <div class="space-y-4">
      <Field label="Current password" error={currentError}>
        <Input
          type="password"
          bind:value={currentPassword}
          error={currentError}
          autocomplete="current-password"
          oninput={() => { currentError = undefined }}
        />
      </Field>
      <Field label="New password" error={newPasswordError} hint="At least 8 characters">
        <Input
          type="password"
          bind:value={newPassword}
          error={newPasswordError}
          autocomplete="new-password"
          oninput={() => { newPasswordError = undefined }}
        />
      </Field>
      <Field label="Confirm new password" error={confirmPasswordError}>
        <Input
          type="password"
          bind:value={confirmPassword}
          error={confirmPasswordError}
          autocomplete="new-password"
          oninput={() => { confirmPasswordError = undefined }}
        />
      </Field>
      <Button
        onclick={savePassword}
        disabled={savingPassword || !currentPassword || !newPassword || !confirmPassword}
      >
        {savingPassword ? 'Saving...' : 'Change password'}
      </Button>
    </div>
  </Card>
</div>
