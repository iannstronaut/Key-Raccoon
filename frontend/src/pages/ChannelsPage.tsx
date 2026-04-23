import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, X, Eye } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import { useAuth } from '../contexts/AuthContext'
import type { Channel } from '../types'

export default function ChannelsPage() {
  const navigate = useNavigate()
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    type: 'openai',
    endpoint: '',
    description: '',
  })
  const { hasPermission } = useAuth()
  
  const canEdit = hasPermission('edit:channels')
  const canDelete = hasPermission('delete:channels')

  useEffect(() => {
    loadChannels()
  }, [])

  async function loadChannels() {
    setLoading(true)
    try {
      const response = await api.getChannels()
      const data = response as { channels?: Channel[] }
      setChannels(data?.channels ?? [])
    } catch (err) {
      console.error('Failed to load channels:', err)
    } finally {
      setLoading(false)
    }
  }

  async function handleCreateChannel(e: React.FormEvent) {
    e.preventDefault()
    try {
      const response = await api.createChannel(formData)
      if (response && typeof response === 'object' && !('error' in response)) {
        setModalOpen(false)
        setFormData({ name: '', type: 'openai', description: '' })
        loadChannels()
      } else {
        alert('Error: ' + ((response as { error?: string })?.error || 'Failed to create channel'))
      }
    } catch {
      alert('Error creating channel')
    }
  }

  async function handleDeleteChannel(id: number) {
    if (!confirm('Are you sure you want to delete this channel?')) return
    try {
      await api.deleteChannel(id)
      loadChannels()
    } catch {
      alert('Error deleting channel')
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[20px] font-medium text-white tracking-body">
            Channels
          </h2>
          <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
            {canEdit ? 'Manage AI provider channels' : 'View AI provider channels'}
          </p>
        </div>
        {canEdit && (
          <button onClick={() => setModalOpen(true)} className="btn-primary flex items-center gap-2">
            <Plus className="w-4 h-4" />
            Add Channel
          </button>
        )}
      </div>

      <div className="card-elevated overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border-medium">
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Name
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Type
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Status
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Models
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  API Keys
                </th>
                {(canEdit || canDelete) && (
                  <th className="text-right px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                    Actions
                  </th>
                )}
              </tr>
            </thead>
            <tbody className="divide-y divide-border-medium">
              {loading ? (
                <tr>
                  <td colSpan={canEdit || canDelete ? 6 : 5} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    Loading...
                  </td>
                </tr>
              ) : channels.length === 0 ? (
                <tr>
                  <td colSpan={canEdit || canDelete ? 6 : 5} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    No channels found
                  </td>
                </tr>
              ) : (
                channels.map((channel) => (
                  <tr key={channel.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-4 py-3 text-[14px] text-text-secondary tracking-body">
                      {channel.name}
                    </td>
                    <td className="px-4 py-3">
                      <span className="badge">{channel.type}</span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`badge ${channel.is_active ? 'badge-success' : 'badge-danger'}`}>
                        {channel.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-[14px] text-text-muted tracking-body">
                      {channel.models?.length || 0}
                    </td>
                    <td className="px-4 py-3 text-[14px] text-text-muted tracking-body">
                      {channel.api_keys?.length || 0}
                    </td>
                    {(canEdit || canDelete) && (
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            onClick={() => navigate(`/channels/${channel.id}`)}
                            className="p-1.5 text-text-muted hover:text-raycast-blue transition-colors rounded-lg hover:bg-white/[0.05]"
                            title="View Details"
                          >
                            <Eye className="w-4 h-4" />
                          </button>
                          {canEdit && (
                            <button className="p-1.5 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/[0.05]">
                              <Pencil className="w-4 h-4" />
                            </button>
                          )}
                          {canDelete && (
                            <button
                              onClick={() => handleDeleteChannel(channel.id)}
                              className="p-1.5 text-text-muted hover:text-raycast-red transition-colors rounded-lg hover:bg-white/[0.05]"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          )}
                        </div>
                      </td>
                    )}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Channel Modal - Only show if user can edit */}
      {canEdit && modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 glass-overlay">
          <div className="glass-strong w-full max-w-md p-5 relative rounded-xl shadow-2xl border border-white/[0.1]">
            <button
              onClick={() => setModalOpen(false)}
              className="absolute top-3 right-3 p-1.5 text-text-muted hover:text-white transition-all rounded-lg hover:bg-white/[0.1]"
            >
              <X className="w-4 h-4" />
            </button>
            <h3 className="text-[18px] font-medium text-white tracking-body mb-4">
              Add Channel
            </h3>
            <form onSubmit={handleCreateChannel} className="space-y-3">
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="0penAI Production"
                  required
                  className="input-dark"
                />
              </div>
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Type
                </label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                  className="input-dark"
                >
                  <option value="openai">0penAI</option>
                  <option value="anthr0pic">Anthr0pic</option>
                  <option value="custom">Custom (0penAI SDK)</option>
                </select>
              </div>
              {formData.type === 'custom' && (
                <div>
                  <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                    Endpoint URL *
                  </label>
                  <input
                    type="url"
                    value={formData.endpoint}
                    onChange={(e) => setFormData({ ...formData, endpoint: e.target.value })}
                    placeholder="https://api.example.com/v1"
                    required={formData.type === 'custom'}
                    className="input-dark"
                  />
                  <p className="text-[10px] text-text-dim mt-1 tracking-body">
                    Custom endpoint compatible with 0penAI SDK
                  </p>
                </div>
              )}
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Description
                </label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Optional description"
                  className="input-dark"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setModalOpen(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Add Channel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
