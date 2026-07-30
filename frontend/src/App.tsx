import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/DashboardPage'
import PortfoliosPage from '@/pages/PortfoliosPage'
import PortfolioDetailPage from '@/pages/PortfolioDetailPage'
import AssetsPage from '@/pages/AssetsPage'
import Layout from '@/components/Layout'

export default function App() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    )
  }

  if (!user) {
    return (
      <Routes>
        <Route path="*" element={<Navigate to="/login" />} />
        <Route path="/login" element={<LoginPage />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route path="/login" element={<Navigate to="/" />} />
      <Route element={<Layout />}>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/portfolios" element={<PortfoliosPage />} />
        <Route path="/portfolios/:id" element={<PortfolioDetailPage />} />
        <Route path="/assets" element={<AssetsPage />} />
      </Route>
    </Routes>
  )
}
