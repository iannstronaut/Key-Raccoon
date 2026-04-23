import { useEffect, useState } from 'react'
import { Users, Link2, Shield, Activity, Database, Server } from 'lucide-react'
import { api } from '../services/api'
import type { User, Channel, Proxy, HealthStatus } from '../types'

interface Stats {
  totalUsers: number
  totalChannels: number
  activeProxies: number
  tokenUsage: number
}

export default function DashboardPage() {
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
      const [usersRes, channelsRes, proxiesRes, healthRes] = await Promise.all([
        api.getUsers(1),
        api.getChannels(1),
        api.getProxies(1),
        api.getHealth(),
      ])

      const usersData = usersRes as { users?: User[]; total?: number }
      const channelsData = channelsRes as { channels?: Channel[]; total?: number }
      const proxiesData = proxiesRes as { proxies?: Proxy[]; total?: number }

      const users = usersData?.users ?? []
      const channels = channelsData?.channels ?? []
      const proxies = proxiesData?.proxies ?? []

      const activeProxies = proxies.filter((p) => p.status === 'healthy').length

      setStats({
        totalUsers: usersData?.total ?? users.length,
        totalChannels: channelsData?.total ?? channels.length,
        activeProxies,
        tokenUsage: 0,
      })

      if (healthRes && 'database_ok' in healthRes) {
        setHealth(healthRes as HealthStatus)
      }
    } catch (err) {
      console.error('Failed to load dashboard stats:', err)
    } finally {
      setLoading(false)
    }
  }

  const statCards = [
    { label: 'Total Users', value: stats.totalUsers, icon: Users },
    { label: 'Total Channels', value: stats.totalChannels, icon: Link2 },
    { label: 'Active Proxies', value: stats.activeProxies, icon: Shield },
    { label: 'Token Usage Today', value: stats.tokenUsage, icon: Activity },
  ]

  const isHealthy = health?.database_ok && health?.redis_ok

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-[24px] font-medium text-white tracking-body">
          Dashboard
        </h2>
        <p className="text-[16px] text-text-muted mt-1 tracking-body">
          Overview of your KeyRaccoon instance
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((card) => (
          <div key={card.label} className="card-elevated p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-[14px] font-medium text-text-muted tracking-body">
                  {card.label}
                </p>
                <p className="text-[32px] font-semibold text-white mt-2 tracking-tight">
                  {loading ? '—' : card.value}
                </p>
              </div>
              <div className="w-10 h-10 rounded-lg bg-white/5 flex items-center justify-center">
                <card.icon className="w-5 h-5 text-text-muted" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* System Health */}
      <div className="card-elevated p-6">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-[20px] font-medium text-white tracking-body">
            System Health
          </h3>
          <span
            className={`badge ${isHealthy ? 'badge-success' : 'badge-warning'}`}
          >
            {isHealthy ? 'Healthy' : 'Degraded'}
          </span>
        </div>

        <div className="space-y-4">
          <div className="flex items-center justify-between p-4 bg-white/[0.02] rounded-lg">
            <div className="flex items-center gap-3">
              <Database className="w-5 h-5 text-text-muted" />
              <span className="text-[16px] font-medium text-text-secondary tracking-body">
                Database
              </span>
            </div>
            <span
              className={`text-[14px] font-medium tracking-body ${
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

          <div className="flex items-center justify-between p-4 bg-white/[0.02] rounded-lg">
            <div className="flex items-center gap-3">
              <Server className="w-5 h-5 text-text-muted" />
              <span className="text-[16px] font-medium text-text-secondary tracking-body">
                Redis
              </span>
            </div>
            <span
              className={`text-[14px] font-medium tracking-body ${
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
