import { useState } from "react";
import { Outlet, NavLink, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  Users,
  Link2,
  Shield,
  BarChart3,
  LogOut,
  Menu,
  Key,
  FileText,
  Layers,
} from "lucide-react";
import { useAuth } from "../contexts/AuthContext";

const navItems = [
  { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard, permission: "view:dashboard" },
  { path: "/users", label: "Users", icon: Users, permission: "view:users" },
  { path: "/channels", label: "Channels", icon: Link2, permission: "view:users" },
  { path: "/my-channels", label: "My Channels", icon: Layers, permission: "view:my-channels", hideForAdmin: true },
  { path: "/proxies", label: "Proxies", icon: Shield, permission: "view:proxies" },
  { path: "/api-keys", label: "API Keys", icon: Key, permission: "view:dashboard" },
  { path: "/logs", label: "Logs", icon: FileText, permission: "view:logs" },
  { path: "/analytics", label: "Analytics", icon: BarChart3, permission: "view:analytics" },
];

export default function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const navigate = useNavigate();
  const { logout, hasPermission, isAdmin: userIsAdmin } = useAuth();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  // Filter nav items based on permissions
  const visibleNavItems = navItems.filter(item => {
    if (!hasPermission(item.permission)) return false;
    if ((item as { hideForAdmin?: boolean }).hideForAdmin && userIsAdmin) return false;
    return true;
  });

  return (
    <div className="flex h-screen overflow-hidden bg-bg-deep">
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 glass-overlay z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`fixed lg:static inset-y-0 left-0 z-50 w-56 glass-strong border-r border-white/[0.08] transform transition-transform duration-300 lg:transform-none ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex flex-col h-full">
          <div className="px-4 py-5">
            <div className="flex items-center gap-2.5">
              <div className="w-10 h-10 rounded-lg flex items-center justify-center backdrop-blur-sm overflow-hidden">
                <img
                  src="/keyraccoon_icon.png"
                  alt="KeyRaccoon"
                  className="w-10 h-10 object-contain"
                />
              </div>
              <h1 className="text-lg font-semibold tracking-tight text-white">
                KeyRaccoon
              </h1>
            </div>
          </div>

          <nav className="flex-1 px-3">
            <ul className="space-y-0.5">
              {visibleNavItems.map((item) => (
                <li key={item.path}>
                  <NavLink
                    to={item.path}
                    onClick={() => setSidebarOpen(false)}
                    className={({ isActive }) =>
                      `flex items-center gap-2.5 px-3 py-2 rounded-lg text-[14px] font-medium tracking-body transition-all duration-200 ${
                        isActive
                          ? "text-white bg-white/[0.08] backdrop-blur-sm border border-white/[0.1] shadow-sm"
                          : "text-text-muted hover:text-white hover:bg-white/[0.05] hover:backdrop-blur-sm"
                      }`
                    }
                  >
                    <item.icon className="w-4 h-4" />
                    {item.label}
                  </NavLink>
                </li>
              ))}
            </ul>
          </nav>

          <div className="p-3 border-t border-white/[0.08]">
            <button
              onClick={handleLogout}
              className="flex items-center gap-2.5 w-full px-3 py-2 text-[14px] font-medium tracking-body text-text-dim hover:text-raycast-red transition-all duration-200 rounded-lg hover:bg-white/[0.05] hover:backdrop-blur-sm"
            >
              <LogOut className="w-4 h-4" />
              Logout
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top bar */}
        <header className="shrink-0 z-30 glass-strong border-b border-white/[0.08]">
          <div className="flex items-center justify-between px-4 py-3">
            <button
              onClick={() => setSidebarOpen(true)}
              className="lg:hidden p-2 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/[0.05]"
            >
              <Menu className="w-5 h-5" />
            </button>
            <div className="lg:hidden" />
          </div>
        </header>

        {/* Page content — only this area scrolls */}
        <main className="flex-1 p-4 lg:p-6 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
