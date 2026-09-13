<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { auth, updateProfile } from '$lib/stores/auth.svelte'
  import SettingsTabs from '$lib/components/domain/SettingsTabs.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'

  let name = $state(auth.user?.name ?? '')
  let email = $state(auth.user?.email ?? '')
  let savingProfile = $state(false)

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
</script>

<div class="p-6">
  <h1 class="mb-6 text-2xl font-bold">Settings</h1>

  <SettingsTabs class="mb-6" />

  <Card class="max-w-lg p-6">
    <h2 class="mb-4 font-semibold">Profile</h2>
    <div class="space-y-4">
      <Field label="Name">
        <Input type="text" bind:value={name} autocomplete="name" />
      </Field>
      <Field label="Email">
        <Input type="email" bind:value={email} autocomplete="email" />
      </Field>
      <Button onclick={saveProfile} disabled={savingProfile || !name || !email}>
        {savingProfile ? 'Saving...' : 'Save'}
      </Button>
    </div>
  </Card>
</div>
