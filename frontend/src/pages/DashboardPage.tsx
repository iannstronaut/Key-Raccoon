import { useEffect, useState } from 'react'
import { Users, Link2, Shield, Activity, Database, Server } from 'lucide-react'
import { api } from '../services/api'
import { useAuth } from '../contexts/AuthContext'
import type { User, Channel, Proxy, HealthStatus } from '../types'

interface Stats {
  totalUsers: number
  totalChannels: number
  activeProxies: number
  tokenUsage: number
}

export default function DashboardPage() {
  const { hasPermission } = useAuth()
  const isAdmin = hasPermission('view:users')

  const [stats, setStats] = useState<Stats>({
    totalUsers: 0,
    totalChannels: 0,
    activeProxies: 0,
    tokenUsage: 0,
  })
  const [health, setHealth] = useState<HealthStatus | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadDashboardData()
    const interval = setInterval(loadDashboardData, 30000)
    return () => clearInterval(interval)
  }, [])

  async function loadDashboardData() {
    try {
      const promises: Promise<unknown>[] = [api.getHealth()]

      if (isAdmin) {
        promises.push(api.getUsers(1), api.getChannels(1), api.getProxies(1))
      } else {
        promises.push(api.getMyChannels())
      }

      const results = await Promise.all(promises)
      const healthRes = results[0]

      if (isAdmin) {
        const usersData = results[1] as { users?: User[]; total?: number }
        const channelsData = results[2] as { channels?: Channel[]; total?: number }
        const proxiesData = results[3] as { proxies?: Proxy[]; total?: number }

        const proxies = proxiesData?.proxies ?? []
        const activeProxies = proxies.filter((p) => p.status === 'healthy').length

        setStats({
          totalUsers: usersData?.total ?? 0,
          totalChannels: channelsData?.total ?? 0,
          activeProxies,
          tokenUsage: 0,
        })
      } else {
        const channelsData = results[1] as { channels?: Channel[]; total?: number }
        setStats({
          totalUsers: 0,
          totalChannels: channelsData?.total ?? (channelsData?.channels?.length ?? 0),
          activeProxies: 0,
          tokenUsage: 0,
        })
      }

      if (healthRes && typeof healthRes === 'object' && 'database_ok' in (healthRes as Record<string, unknown>)) {
        setHealth(healthRes as HealthStatus)
      }
    } catch (err) {
      console.error('Failed to load dashboard stats:', err)
    } finally {
      setLoading(false)
    }
  }

  const adminStatCards = [
    { label: 'Total Users', value: stats.totalUsers, icon: Users },
    { label: 'Total Channels', value: stats.totalChannels, icon: Link2 },
    { label: 'Active Proxies', value: stats.activeProxies, icon: Shield },
    { label: 'Token Usage Today', value: stats.tokenUsage, icon: Activity },
  ]

  const userStatCards = [
    { label: 'My Channels', value: stats.totalChannels, icon: Link2 },
    { label: 'Token Usage Today', value: stats.tokenUsage, icon: Activity },
  ]

  const statCards = isAdmin ? adminStatCards : userStatCards
  const isHealthy = health?.database_ok && health?.redis_ok

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[20px] font-medium text-white tracking-body">
          Dashboard
        </h2>
        <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
          Overview of your KeyRaccoon instance
        </p>
      </div>

      {/* Stats Grid */}
      <div className={`grid grid-cols-1 sm:grid-cols-2 ${isAdmin ? 'lg:grid-cols-4' : 'lg:grid-cols-2'} gap-3`}>
        {statCards.map((card) => (
          <div key={card.label} className="card-elevated p-4 hover:shadow-lg transition-all duration-300">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-[12px] font-medium text-text-muted tracking-body">
                  {card.label}
                </p>
                <p className="text-[24px] font-semibold text-white mt-1.5 tracking-tight">
                  {loading ? '—' : card.value}
                </p>
              </div>
              <div className="w-8 h-8 rounded-lg bg-white/5 backdrop-blur-sm flex items-center justify-center border border-white/[0.05]">
                <card.icon className="w-4 h-4 text-text-muted" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* System Health */}
      <div className="card-elevated p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-[16px] font-medium text-white tracking-body">
            System Health
          </h3>
          <span
            className={`badge ${isHealthy ? 'badge-success' : 'badge-warning'}`}
          >
            {isHealthy ? 'Healthy' : 'Degraded'}
          </span>
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between p-3 glass-subtle rounded-lg hover:bg-white/[0.04] transition-all">
            <div className="flex items-center gap-2.5">
              <Database className="w-4 h-4 text-text-muted" />
              <span className="text-[14px] font-medium text-text-secondary tracking-body">
                Database
              </span>
            </div>
            <span
              className={`text-[12px] font-medium tracking-body ${
                health?.database_ok ? 'text-raycast-green' : 'text-raycast-red'
              }`}
            >
              {loading
                ? 'Checking...'
                : health?.database_ok
                ? 'Connected'
                : 'Disconnected'}
            </span>
          </div>

          <div className="flex items-center justify-between p-3 glass-subtle rounded-lg hover:bg-white/[0.04] transition-all">
            <div className="flex items-center gap-2.5">
              <Server className="w-4 h-4 text-text-muted" />
              <span className="text-[14px] font-medium text-text-secondary tracking-body">
                Redis
              </span>
            </div>
            <span
              className={`text-[12px] font-medium tracking-body ${
                health?.redis_ok ? 'text-raycast-green' : 'text-raycast-red'
              }`}
            >
              {loading
                ? 'Checking...'
                : health?.redis_ok
                ? 'Connected'
                : 'Disconnected'}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
