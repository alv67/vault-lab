<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { login, register } from '$lib/stores/auth.svelte'
  import { t } from '$lib/i18n/index.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Field from '$lib/components/ui/Field.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte'

  const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  const modeItems = [
    { value: 'signin', label: 'Sign in' },
    { value: 'register', label: 'Register' },
  ]

  let mode = $state('signin')
  let email = $state('')
  let password = $state('')
  let confirmPassword = $state('')
  let name = $state('')
  let submitting = $state(false)

  let emailError = $state<string | undefined>(undefined)
  let passwordError = $state<string | undefined>(undefined)
  let confirmError = $state<string | undefined>(undefined)
  let nameError = $state<string | undefined>(undefined)

  const isRegister = $derived(mode === 'register')

  // The two modes have different password/name constraints, so validation
  // state never carries over a mode switch.
  $effect(() => {
    if (mode) {
      emailError = undefined
      passwordError = undefined
      confirmError = undefined
      nameError = undefined
    }
  })

  function validate(): boolean {
    let valid = true
    if (!email.trim()) {
      emailError = 'Email is required'
      valid = false
    } else if (!EMAIL_REGEX.test(email)) {
      emailError = 'Enter a valid email address'
      valid = false
    }
    // Length is only enforced on register: legacy accounts may have shorter passwords.
    if (!password) {
      passwordError = 'Password is required'
      valid = false
    } else if (isRegister && password.length < 8) {
      passwordError = 'Password must be at least 8 characters'
      valid = false
    }
    if (isRegister && (!confirmPassword || confirmPassword !== password)) {
      confirmError = 'Passwords do not match'
      valid = false
    }
    if (isRegister && !name.trim()) {
      nameError = 'Name is required'
      valid = false
    }
    return valid
  }

  async function handleSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault()
    if (!validate()) return
    submitting = true
    try {
      if (isRegister) {
        await register(email, name, password)
        toast.success('Registered! You can now log in.')
        mode = 'signin'
      } else {
        await login(email, password)
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Something went wrong'
      toast.error(message)
    } finally {
      submitting = false
    }
  }
</script>

<div class="grid min-h-dvh place-items-center bg-background p-4">
  <div class="w-full max-w-sm rounded-card border border-border bg-surface p-6 shadow-raised sm:p-8">
    <div class="mb-6 flex flex-col items-center gap-3">
      <img src="/peculium.svg" alt="" class="h-12 w-12" />
      <h1 class="text-2xl font-bold text-foreground">Peculium</h1>
      <p class="text-center text-sm text-muted-foreground">{t('login.tagline')}</p>
    </div>
    <p class="mb-6 text-center text-sm text-muted-foreground">
      {isRegister ? 'Create an account' : 'Sign in to your account'}
    </p>
    <SegmentedControl items={modeItems} bind:value={mode} ariaLabel="Authentication mode" class="mb-6 w-full" />
    <!-- `novalidate` keeps the submit on our inline validation; `required` stays for a11y. -->
    <form onsubmit={handleSubmit} class="space-y-4" novalidate>
      {#if isRegister}
        <Field label="Name" error={nameError}>
          <Input
            bind:value={name}
            type="text"
            autocomplete="name"
            required
            error={nameError}
            oninput={() => (nameError = undefined)}
          />
        </Field>
      {/if}
      <Field label="Email" error={emailError}>
        <Input
          bind:value={email}
          type="email"
          autocomplete="email"
          required
          error={emailError}
          oninput={() => (emailError = undefined)}
        />
      </Field>
      <Field label="Password" error={passwordError} hint={isRegister ? 'At least 8 characters' : undefined}>
        <Input
          bind:value={password}
          type="password"
          autocomplete={isRegister ? 'new-password' : 'current-password'}
          required
          error={passwordError}
          oninput={() => (passwordError = undefined)}
        />
      </Field>
      {#if isRegister}
        <Field label="Confirm password" error={confirmError}>
          <Input
            bind:value={confirmPassword}
            type="password"
            autocomplete="new-password"
            required
            error={confirmError}
            oninput={() => (confirmError = undefined)}
          />
        </Field>
      {/if}
      <Button type="submit" class="w-full" loading={submitting}>
        {isRegister ? 'Register' : 'Sign in'}
      </Button>
    </form>
  </div>
</div>
