<script lang="ts">
  import { toast } from '$lib/stores/toast.svelte'
  import { login, register } from '$lib/stores/auth.svelte'

  let email = $state('')
  let password = $state('')
  let name = $state('')
  let isRegister = $state(false)
  let submitting = $state(false)

  async function handleSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault()
    submitting = true
    try {
      if (isRegister) {
        await register(email, name, password)
        toast.success('Registered! You can now log in.')
        isRegister = false
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

<div class="flex min-h-screen items-center justify-center bg-background">
  <div class="w-full max-w-sm rounded-card border-border bg-surface p-8 shadow-raised">
    <h1 class="mb-6 text-2xl font-bold text-foreground">VaultLab</h1>
    <p class="mb-6 text-sm text-muted-foreground">
      {isRegister ? 'Create an account' : 'Sign in to your account'}
    </p>
    <form onsubmit={handleSubmit} class="space-y-4">
      {#if isRegister}
        <input
          type="text"
          placeholder="Name"
          bind:value={name}
          class="w-full rounded-control border border-input px-3 py-2 text-sm"
          required
        />
      {/if}
      <input
        type="email"
        placeholder="Email"
        bind:value={email}
        class="w-full rounded-control border border-input px-3 py-2 text-sm"
        required
      />
      <input
        type="password"
        placeholder="Password"
        bind:value={password}
        class="w-full rounded-control border border-input px-3 py-2 text-sm"
        required
      />
      <button
        type="submit"
        disabled={submitting}
        class="w-full rounded-control bg-accent px-4 py-2 text-sm font-medium text-accent-foreground hover:bg-accent-hover disabled:opacity-50"
      >
        {isRegister ? 'Register' : 'Sign in'}
      </button>
    </form>
    <button
      onclick={() => (isRegister = !isRegister)}
      class="mt-4 text-sm text-accent-text hover:underline"
    >
      {isRegister ? 'Already have an account? Sign in' : "Don't have an account? Register"}
    </button>
  </div>
</div>
