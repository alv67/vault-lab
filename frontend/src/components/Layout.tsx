import { Link, Outlet } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import {
  LayoutDashboard,
  Briefcase,
  Banknote,
  LogOut,
} from 'lucide-react'

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/portfolios', label: 'Portfolios', icon: Briefcase },
  { to: '/assets', label: 'Assets', icon: Banknote },
]

export default function Layout() {
  const { user, logout } = useAuth()

  const handleLogout = () => {
    logout()
  }

  return (
    <div className="flex h-screen">
      <aside className="flex w-64 flex-col border-r bg-white">
        <div className="flex items-center gap-2 border-b px-6 py-4">
          <Banknote className="h-6 w-6 text-blue-600" />
          <span className="text-lg font-bold">VaultLab</span>
        </div>
        <nav className="flex-1 space-y-1 p-4">
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
            >
              <item.icon className="h-5 w-5" />
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="border-t p-4">
          <div className="mb-2 text-sm text-gray-500">{user?.email}</div>
          <button
            onClick={handleLogout}
            className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-red-600 hover:bg-red-50"
          >
            <LogOut className="h-5 w-5" />
            Sign out
          </button>
        </div>
      </aside>
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
