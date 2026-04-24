import { useEffect, useState, useCallback } from 'react'
import { BarChart3, TrendingUp, Clock, Activity, Zap, DollarSign, CheckCircle, XCircle, RefreshCw } from 'lucide-react'
import { api } from '../services/api'
import { useAuth } from '../contexts/AuthContext'
import type { RequestLog, UsageStats } from '../types'

const REFRESH_INTERVAL = 60_000 // 1 minute

export default function AnalyticsPage() {
  const { hasPermission, user } = useAuth()
  const isAdmin = hasPermission('view:users')

  const [stats, setStats] = useState<UsageStats | null>(null)
  const [recentLogs, setRecentLogs] = useState<RequestLog[]>([])
  const [loading, setLoading] = useState(true)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const loadAnalytics = useCallback(async (showLoading = true) => {
    if (showLoading) setLoading(true)
    try {
      // Stats: backend auto-scopes to user's own data for non-admin
      // Logs: admin gets all, user gets own
      const [statsRes, logsRes] = await Promise.all([
        api.getLogStats(),
        isAdmin
          ? api.getLogs({ limit: 10 })
          : api.getUserLogs(user?.id || 0, 10, 0),
      ])

      const statsData = statsRes as UsageStats | null
      if (statsData && 'total_requests' in statsData) {
        setStats(statsData)
      }

      const logsData = logsRes as { logs?: RequestLog[] } | null
      setRecentLogs(logsData?.logs || [])
      setLastRefresh(new Date())
    } catch (err) {
      console.error('Failed to load analytics:', err)
    } finally {
      setLoading(false)
    }
  }, [isAdmin, user?.id])

  // Initial load
  useEffect(() => {
    loadAnalytics()
  }, [loadAnalytics])

  // Auto-refresh every 60 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      loadAnalytics(false) // silent refresh, no loading spinner
    }, REFRESH_INTERVAL)
    return () => clearInterval(interval)
  }, [loadAnalytics])

  function formatTimeAgo(dateStr: string): string {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffSec = Math.floor(diffMs / 1000)
    const diffMin = Math.floor(diffSec / 60)
    const diffHour = Math.floor(diffMin / 60)
    const diffDay = Math.floor(diffHour / 24)

    if (diffSec < 60) return `${diffSec}s ago`
    if (diffMin < 60) return `${diffMin}m ago`
    if (diffHour < 24) return `${diffHour}h ago`
    return `${diffDay}d ago`
  }

  const statCards = [
    {
      label: 'Total Requests',
      value: stats?.total_requests ?? 0,
      icon: Activity,
    },
    {
      label: 'Total Tokens',
      value: stats?.total_tokens ?? 0,
      icon: Zap,
    },
    {
      label: 'Total Cost',
      value: `$${(stats?.total_cost ?? 0).toFixed(4)}`,
      icon: DollarSign,
    },
    {
      label: 'Avg Latency',
      value: `${Math.round(stats?.avg_latency_ms ?? 0)}ms`,
      icon: Clock,
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[20px] font-medium text-white tracking-body">
            Analytics
          </h2>
          <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
            {isAdmin ? 'System-wide usage statistics' : 'Your usage statistics'}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-[11px] text-text-dim tracking-body">
            Updated {formatTimeAgo(lastRefresh.toISOString())}
          </span>
          <button
            onClick={() => loadAnalytics(false)}
            className="p-2 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/[0.05]"
            title="Refresh now"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        {statCards.map((card) => (
          <div key={card.label} className="card-elevated p-4 hover:shadow-lg transition-all duration-300">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-[12px] font-medium text-text-muted tracking-body">
                  {card.label}
                </p>
                <p className="text-[24px] font-semibold text-white mt-1.5 tracking-tight">
                  {loading ? '—' : typeof card.value === 'number' ? card.value.toLocaleString() : card.value}
                </p>
              </div>
              <div className="w-8 h-8 rounded-lg bg-white/5 backdrop-blur-sm flex items-center justify-center border border-white/[0.05]">
                <card.icon className="w-4 h-4 text-text-muted" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Success/Failed breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <div className="card-elevated p-4">
          <div className="flex items-center gap-2.5 mb-4">
            <div className="w-8 h-8 rounded-lg bg-raycast-green/15 flex items-center justify-center backdrop-blur-sm border border-raycast-green/20">
              <CheckCircle className="w-4 h-4 text-raycast-green" />
            </div>
            <div>
              <h3 className="text-[16px] font-medium text-white tracking-body">
                Success Rate
              </h3>
              <p className="text-[12px] text-text-muted tracking-body">
                Request success vs failure
              </p>
            </div>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between p-3 glass-subtle rounded-lg">
              <div className="flex items-center gap-2">
                <CheckCircle className="w-4 h-4 text-raycast-green" />
                <span className="text-[14px] text-text-secondary tracking-body">Successful</span>
              </div>
              <span className="text-[14px] font-medium text-white">
                {loading ? '—' : (stats?.success_count ?? 0).toLocaleString()}
              </span>
            </div>
            <div className="flex items-center justify-between p-3 glass-subtle rounded-lg">
              <div className="flex items-center gap-2">
                <XCircle className="w-4 h-4 text-raycast-red" />
                <span className="text-[14px] text-text-secondary tracking-body">Failed</span>
              </div>
              <span className="text-[14px] font-medium text-white">
                {loading ? '—' : (stats?.failed_count ?? 0).toLocaleString()}
              </span>
            </div>
            {!loading && stats && stats.total_requests > 0 && (
              <div className="pt-2">
                <div className="w-full h-2 bg-white/[0.05] rounded-full overflow-hidden">
                  <div
                    className="h-full bg-raycast-green rounded-full transition-all"
                    style={{ width: `${(stats.success_count / stats.total_requests) * 100}%` }}
                  />
                </div>
                <p className="text-[11px] text-text-dim mt-1.5 tracking-body">
                  {((stats.success_count / stats.total_requests) * 100).toFixed(1)}% success rate
                </p>
              </div>
            )}
          </div>
        </div>

        <div className="card-elevated p-4">
          <div className="flex items-center gap-2.5 mb-4">
            <div className="w-8 h-8 rounded-lg bg-raycast-blue/15 flex items-center justify-center backdrop-blur-sm border border-raycast-blue/20">
              <TrendingUp className="w-4 h-4 text-raycast-blue" />
            </div>
            <div>
              <h3 className="text-[16px] font-medium text-white tracking-body">
                Token Usage
              </h3>
              <p className="text-[12px] text-text-muted tracking-body">
                Token consumption overview
              </p>
            </div>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between p-3 glass-subtle rounded-lg">
              <span className="text-[14px] text-text-secondary tracking-body">Total Tokens Used</span>
              <span className="text-[14px] font-medium text-white">
                {loading ? '—' : (stats?.total_tokens ?? 0).toLocaleString()}
              </span>
            </div>
            <div className="flex items-center justify-between p-3 glass-subtle rounded-lg">
              <span className="text-[14px] text-text-secondary tracking-body">Total Cost</span>
              <span className="text-[14px] font-medium text-white">
                {loading ? '—' : `$${(stats?.total_cost ?? 0).toFixed(4)}`}
              </span>
            </div>
            <div className="flex items-center justify-between p-3 glass-subtle rounded-lg">
              <span className="text-[14px] text-text-secondary tracking-body">Avg Tokens/Request</span>
              <span className="text-[14px] font-medium text-white">
                {loading ? '—' : stats && stats.total_requests > 0
                  ? Math.round(stats.total_tokens / stats.total_requests).toLocaleString()
                  : '0'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="card-elevated p-4">
        <div className="flex items-center gap-2.5 mb-4">
          <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center backdrop-blur-sm border border-white/[0.05]">
            <BarChart3 className="w-4 h-4 text-text-muted" />
          </div>
          <h3 className="text-[16px] font-medium text-white tracking-body">
            Recent Activity
          </h3>
        </div>
        <div className="space-y-2">
          {loading ? (
            <p className="text-[14px] text-text-muted text-center py-6 tracking-body">Loading...</p>
          ) : recentLogs.length === 0 ? (
            <div className="text-center py-6">
              <BarChart3 className="w-8 h-8 text-text-dark mx-auto mb-2" />
              <p className="text-[14px] text-text-muted tracking-body">
                No activity yet
              </p>
              <p className="text-[12px] text-text-dim mt-1 tracking-body">
                Logs will appear here after API requests are made
              </p>
            </div>
          ) : (
            recentLogs.map((log) => (
              <div
                key={log.id}
                className="flex items-center gap-3 p-3 glass-subtle rounded-lg hover:bg-white/[0.04] transition-all"
              >
                <div className={`w-2 h-2 rounded-full ${
                  log.status === 'success' ? 'bg-raycast-green' :
                  log.status === 'failed' ? 'bg-raycast-red' :
                  'bg-raycast-orange'
                }`} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="text-[14px] text-text-secondary tracking-body">
                      <span className="font-mono text-[12px]">{log.model_name}</span>
                    </p>
                    {log.channel_name && (
                      <span className="text-[11px] text-text-dim tracking-body">
                        via {log.channel_name}
                      </span>
                    )}
                    <span className={`badge ${
                      log.status === 'success' ? 'badge-success' :
                      log.status === 'failed' ? 'badge-danger' :
                      'badge-warning'
                    }`}>
                      {log.status}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 mt-0.5">
                    <span className="text-[12px] text-text-dim tracking-body">
                      {log.total_tokens.toLocaleString()} tokens
                    </span>
                    {log.token_price > 0 && (
                      <span className="text-[12px] text-text-dim tracking-body">
                        @${log.token_price.toFixed(4)}/1K
                      </span>
                    )}
                    {log.cost > 0 && (
                      <span className="text-[12px] text-text-dim tracking-body">
                        = ${log.cost.toFixed(4)}
                      </span>
                    )}
                    <span className="text-[12px] text-text-dim tracking-body">
                      {log.latency_ms}ms
                    </span>
                    {isAdmin && (
                      <span className="text-[12px] text-text-dim tracking-body">
                        {log.user_email || `User #${log.user_id}`}
                      </span>
                    )}
                  </div>
                </div>
                <span className="text-[12px] text-text-dim tracking-body whitespace-nowrap">
                  {formatTimeAgo(log.created_at)}
                </span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
