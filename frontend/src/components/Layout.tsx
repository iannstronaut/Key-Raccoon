import { useState } from 'react'
import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { LayoutDashboard, Users, Link2, Shield, BarChart3, LogOut, Menu } from 'lucide-react'
import { api } from '../services/api'

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/users', label: 'Users', icon: Users },
  { path: '/channels', label: 'Channels', icon: Link2 },
  { path: '/proxies', label: 'Proxies', icon: Shield },
  { path: '/analytics', label: 'Analytics', icon: BarChart3 },
]

export default function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const navigate = useNavigate()

  const handleLogout = () => {
    api.clearToken()
    navigate('/login')
  }

  return (
    <div className="flex min-h-screen bg-bg-deep">
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`fixed lg:static inset-y-0 left-0 z-50 w-64 bg-bg-deep border-r border-border-medium transform transition-transform duration-300 lg:transform-none ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex flex-col h-full">
          <div className="p-6">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-raycast-red/15 flex items-center justify-center">
                <Shield className="w-5 h-5 text-raycast-red" />
              </div>
              <h1 className="text-xl font-semibold tracking-tight text-white">
                KeyRaccoon
              </h1>
            </div>
          </div>

          <nav className="flex-1 px-4">
            <ul className="space-y-1">
              {navItems.map((item) => (
                <li key={item.path}>
                  <NavLink
                    to={item.path}
                    onClick={() => setSidebarOpen(false)}
                    className={({ isActive }) =>
                      `flex items-center gap-3 px-4 py-2.5 rounded-lg text-[16px] font-medium tracking-body transition-colors duration-200 ${
                        isActive
                          ? 'text-white bg-white/5'
                          : 'text-text-muted hover:text-white hover:bg-white/5'
                      }`
                    }
                  >
                    <item.icon className="w-5 h-5" />
                    {item.label}
                  </NavLink>
                </li>
              ))}
            </ul>
          </nav>

          <div className="p-4 border-t border-border-medium">
            <button
              onClick={handleLogout}
              className="flex items-center gap-3 w-full px-4 py-2.5 text-[16px] font-medium tracking-body text-text-dim hover:text-raycast-red transition-colors duration-200 rounded-lg hover:bg-white/5"
            >
              <LogOut className="w-5 h-5" />
              Logout
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top bar */}
        <header className="sticky top-0 z-30 bg-bg-deep/80 backdrop-blur-md border-b border-border-medium">
          <div className="flex items-center justify-between px-6 py-4">
            <button
              onClick={() => setSidebarOpen(true)}
              className="lg:hidden p-2 text-text-muted hover:text-white transition-colors"
            >
              <Menu className="w-6 h-6" />
            </button>
            <div className="lg:hidden" />
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 p-6 lg:p-8 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
