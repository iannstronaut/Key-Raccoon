import { useState, useEffect } from 'react'
import { Plus, RefreshCw, Trash2, X } from 'lucide-react'
import { api } from '../services/api'
import type { Proxy } from '../types'

export default function ProxiesPage() {
  const [proxies, setProxies] = useState<Proxy[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [formData, setFormData] = useState({
    proxy_url: '',
    type: 'http',
    username: '',
    password: '',
  })

  useEffect(() => {
    loadProxies()
  }, [])

  async function loadProxies() {
    setLoading(true)
    try {
      const response = await api.getProxies()
      const data = response as { proxies?: Proxy[] }
      setProxies(data?.proxies ?? [])
    } catch (err) {
      console.error('Failed to load proxies:', err)
    } finally {
      setLoading(false)
    }
  }

  async function handleAddProxy(e: React.FormEvent) {
    e.preventDefault()
    try {
      const payload = {
        ...formData,
        username: formData.username || undefined,
        password: formData.password || undefined,
      }
      const response = await api.addProxy(payload)
      if (response && typeof response === 'object' && !('error' in response)) {
        setModalOpen(false)
        setFormData({ proxy_url: '', type: 'http', username: '', password: '' })
        loadProxies()
      } else {
        alert('Error: ' + ((response as { error?: string })?.error || 'Failed to add proxy'))
      }
    } catch {
      alert('Error adding proxy')
    }
  }

  async function handleDeleteProxy(id: number) {
    if (!confirm('Are you sure you want to delete this proxy?')) return
    try {
      await api.deleteProxy(id)
      loadProxies()
    } catch {
      alert('Error deleting proxy')
    }
  }

  function getStatusBadge(status: string) {
    if (status === 'healthy') return 'badge-success'
    if (status === 'unhealthy') return 'badge-danger'
    return 'badge-warning'
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[24px] font-medium text-white tracking-body">
            Proxies
          </h2>
          <p className="text-[16px] text-text-muted mt-1 tracking-body">
            Manage proxy servers
          </p>
        </div>
        <button onClick={() => setModalOpen(true)} className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Add Proxy
        </button>
      </div>

      <div className="card-elevated overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border-medium">
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  URL
                </th>
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Type
                </th>
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Status
                </th>
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Last Check
                </th>
                <th className="text-right px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-medium">
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-8 text-center text-text-muted tracking-body">
                    Loading...
                  </td>
                </tr>
              ) : proxies.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-8 text-center text-text-muted tracking-body">
                    No proxies found
                  </td>
                </tr>
              ) : (
                proxies.map((proxy) => (
                  <tr key={proxy.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-6 py-4 text-[16px] text-text-secondary tracking-body font-mono text-[14px]">
                      {proxy.url}
                    </td>
                    <td className="px-6 py-4">
                      <span className="badge">{proxy.type}</span>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`badge ${getStatusBadge(proxy.status)}`}>
                        {proxy.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-[16px] text-text-muted tracking-body">
                      {proxy.last_check
                        ? new Date(proxy.last_check).toLocaleString()
                        : 'Never'}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <button className="p-2 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/5">
                          <RefreshCw className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteProxy(proxy.id)}
                          className="p-2 text-text-muted hover:text-raycast-red transition-colors rounded-lg hover:bg-white/5"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Add Proxy Modal */}
      {modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
          <div className="card w-full max-w-md p-6 relative">
            <button
              onClick={() => setModalOpen(false)}
              className="absolute top-4 right-4 p-1 text-text-muted hover:text-white transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
            <h3 className="text-[20px] font-medium text-white tracking-body mb-6">
              Add Proxy
            </h3>
            <form onSubmit={handleAddProxy} className="space-y-4">
              <div>
                <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
                  Proxy URL
                </label>
                <input
                  type="text"
                  value={formData.proxy_url}
                  onChange={(e) => setFormData({ ...formData, proxy_url: e.target.value })}
                  placeholder="http://proxy.example.com:8080"
                  required
                  className="input-dark"
                />
              </div>
              <div>
                <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
                  Type
                </label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                  className="input-dark"
                >
                  <option value="http">HTTP</option>
                  <option value="https">HTTPS</option>
                  <option value="socks5">SOCKS5</option>
                </select>
              </div>
              <div>
                <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
                  Username (optional)
                </label>
                <input
                  type="text"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  placeholder="Username"
                  className="input-dark"
                />
              </div>
              <div>
                <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
                  Password (optional)
                </label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Password"
                  className="input-dark"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setModalOpen(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Add Proxy
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
