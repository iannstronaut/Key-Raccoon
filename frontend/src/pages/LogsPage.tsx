import { useState, useEffect } from 'react'
import { FileText, Filter, ChevronLeft, ChevronRight } from 'lucide-react'
import { api } from '../services/api'
import type { RequestLog } from '../types'

export default function LogsPage() {
  const [logs, setLogs] = useState<RequestLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [limit] = useState(25)
  const [offset, setOffset] = useState(0)
  const [filters, setFilters] = useState({
    status: '',
    model: '',
    date_from: '',
    date_to: '',
  })
  const [showFilters, setShowFilters] = useState(false)

  useEffect(() => {
    loadLogs()
  }, [offset, filters])

  async function loadLogs() {
    setLoading(true)
    try {
      const params: Record<string, string | number> = { limit, offset }
      if (filters.status) params.status = filters.status
      if (filters.model) params.model = filters.model
      if (filters.date_from) params.date_from = new Date(filters.date_from).toISOString()
      if (filters.date_to) params.date_to = new Date(filters.date_to).toISOString()

      const res = await api.getLogs(params as any)
      const data = res as { logs?: RequestLog[]; total?: number }
      setLogs(data?.logs || [])
      setTotal(data?.total || 0)
    } catch (err) {
      console.error('Failed to load logs:', err)
    } finally {
      setLoading(false)
    }
  }

  function handleFilterChange(key: string, value: string) {
    setFilters(prev => ({ ...prev, [key]: value }))
    setOffset(0)
  }

  function clearFilters() {
    setFilters({ status: '', model: '', date_from: '', date_to: '' })
    setOffset(0)
  }

  const totalPages = Math.ceil(total / limit)
  const currentPage = Math.floor(offset / limit) + 1

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[20px] font-medium text-white tracking-body">
            Request Logs
          </h2>
          <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
            View all API request logs
          </p>
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className="btn-secondary flex items-center gap-2"
        >
          <Filter className="w-4 h-4" />
          Filters
        </button>
      </div>

      {showFilters && (
        <div className="card-elevated p-4">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <div>
              <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                Status
              </label>
              <select
                value={filters.status}
                onChange={(e) => handleFilterChange('status', e.target.value)}
                className="input-dark"
              >
                <option value="">All</option>
                <option value="success">Success</option>
                <option value="failed">Failed</option>
                <option value="pending">Pending</option>
              </select>
            </div>
            <div>
              <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                Model
              </label>
              <input
                type="text"
                value={filters.model}
                onChange={(e) => handleFilterChange('model', e.target.value)}
                placeholder="e.g. gpt-4"
                className="input-dark"
              />
            </div>
            <div>
              <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                Date From
              </label>
              <input
                type="datetime-local"
                value={filters.date_from}
                onChange={(e) => handleFilterChange('date_from', e.target.value)}
                className="input-dark"
              />
            </div>
            <div>
              <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                Date To
              </label>
              <input
                type="datetime-local"
                value={filters.date_to}
                onChange={(e) => handleFilterChange('date_to', e.target.value)}
                className="input-dark"
              />
            </div>
          </div>
          <div className="flex justify-end mt-3">
            <button onClick={clearFilters} className="btn-secondary text-[12px] px-3 py-1.5">
              Clear Filters
            </button>
          </div>
        </div>
      )}

      <div className="card-elevated overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border-medium">
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Time</th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">User</th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Channel</th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Model</th>
                <th className="text-right px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Tokens</th>
                <th className="text-right px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Price/1K</th>
                <th className="text-right px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Cost</th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Status</th>
                <th className="text-right px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">Latency</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-medium">
              {loading ? (
                <tr>
                  <td colSpan={9} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    Loading...
                  </td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={9} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    <FileText className="w-8 h-8 mx-auto mb-2 opacity-50" />
                    No logs found
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-4 py-3 text-[12px] text-text-muted tracking-body whitespace-nowrap">
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-[12px] text-text-secondary tracking-body">
                      {log.user_email || `#${log.user_id}`}
                    </td>
                    <td className="px-4 py-3 text-[12px] text-text-secondary tracking-body">
                      {log.channel_name || `#${log.channel_id}`}
                    </td>
                    <td className="px-4 py-3 text-[12px] font-mono text-text-secondary tracking-body">
                      {log.model_name}
                    </td>
                    <td className="px-4 py-3 text-[12px] text-text-secondary tracking-body text-right">
                      {log.total_tokens.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-[12px] text-text-muted tracking-body text-right">
                      {log.token_price > 0 ? `$${log.token_price.toFixed(4)}` : '—'}
                    </td>
                    <td className="px-4 py-3 text-[12px] text-text-secondary tracking-body text-right">
                      ${log.cost.toFixed(4)}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`badge ${
                        log.status === 'success' ? 'badge-success' :
                        log.status === 'failed' ? 'badge-danger' :
                        'badge-warning'
                      }`}>
                        {log.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-[12px] text-text-muted tracking-body text-right">
                      {log.latency_ms}ms
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {total > limit && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-border-medium">
            <p className="text-[12px] text-text-muted tracking-body">
              Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setOffset(Math.max(0, offset - limit))}
                disabled={offset === 0}
                className="p-1.5 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/[0.05] disabled:opacity-30"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <span className="text-[12px] text-text-muted">
                Page {currentPage} of {totalPages}
              </span>
              <button
                onClick={() => setOffset(offset + limit)}
                disabled={offset + limit >= total}
                className="p-1.5 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/[0.05] disabled:opacity-30"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
