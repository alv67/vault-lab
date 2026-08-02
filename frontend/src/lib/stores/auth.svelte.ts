import { authApi, type User } from '$lib/services/api'

export const auth = $state({
  user: null as User | null,
  isLoading: true,
})

export async function initAuth(): Promise<void> {
  const token = localStorage.getItem('access_token')
  if (!token) {
    auth.isLoading = false
    return
  }
  try {
    const res = await authApi.me()
    auth.user = res
  } catch {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  } finally {
    auth.isLoading = false
  }
}

export async function login(email: string, password: string): Promise<void> {
  const res = await authApi.login(email, password)
  localStorage.setItem('access_token', res.access_token)
  localStorage.setItem('refresh_token', res.refresh_token)
  auth.user = res.user
  window.location.replace('/')
}

export async function register(email: string, name: string, password: string): Promise<void> {
  await authApi.register(email, name, password)
}

export async function updateProfile(name: string, email: string): Promise<void> {
  const user = await authApi.updateProfile({ name, email })
  auth.user = user
}

export function logout(): void {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  auth.user = null
  window.location.replace('/login')
}
