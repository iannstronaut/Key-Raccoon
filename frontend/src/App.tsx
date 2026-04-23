import { Routes, Route, Navigate } from 'react-router-dom'
import { api } from './services/api'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/DashboardPage'
import UsersPage from './pages/UsersPage'
import ChannelsPage from './pages/ChannelsPage'
import ProxiesPage from './pages/ProxiesPage'
import AnalyticsPage from './pages/AnalyticsPage'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  if (!api.isAuthenticated()) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="users" element={<UsersPage />} />
        <Route path="channels" element={<ChannelsPage />} />
        <Route path="proxies" element={<ProxiesPage />} />
        <Route path="analytics" element={<AnalyticsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}

export default App
