import { useState } from 'react'
import { useAuth } from '@/hooks/useAuth'
import toast from 'react-hot-toast'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [isRegister, setIsRegister] = useState(false)
  const { login, register } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (isRegister) {
        await register(email, name, password)
        toast.success('Registered! You can now log in.')
        setIsRegister(false)
      } else {
        await login(email, password)
      }
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Something went wrong')
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-sm rounded-xl bg-white p-8 shadow-lg">
        <h1 className="mb-6 text-2xl font-bold text-gray-900">VaultLab</h1>
        <p className="mb-6 text-sm text-gray-500">
          {isRegister ? 'Create an account' : 'Sign in to your account'}
        </p>
        <form onSubmit={handleSubmit} className="space-y-4">
          {isRegister && (
            <input
              type="text"
              placeholder="Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-lg border px-3 py-2 text-sm"
              required
            />
          )}
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-lg border px-3 py-2 text-sm"
            required
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-lg border px-3 py-2 text-sm"
            required
          />
          <button
            type="submit"
            className="w-full rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
          >
            {isRegister ? 'Register' : 'Sign in'}
          </button>
        </form>
        <button
          onClick={() => setIsRegister(!isRegister)}
          className="mt-4 text-sm text-blue-600 hover:underline"
        >
          {isRegister ? 'Already have an account? Sign in' : "Don't have an account? Register"}
        </button>
      </div>
    </div>
  )
}
