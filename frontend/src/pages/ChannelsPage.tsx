import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, X } from 'lucide-react'
import { api } from '../services/api'
import type { Channel } from '../types'

export default function ChannelsPage() {
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    type: 'openai',
    description: '',
  })

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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[24px] font-medium text-white tracking-body">
            Channels
          </h2>
          <p className="text-[16px] text-text-muted mt-1 tracking-body">
            Manage AI provider channels
          </p>
        </div>
        <button onClick={() => setModalOpen(true)} className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Add Channel
        </button>
      </div>

      <div className="card-elevated overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border-medium">
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Name
                </th>
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Type
                </th>
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Status
                </th>
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Models
                </th>
                <th className="text-left px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  API Keys
                </th>
                <th className="text-right px-6 py-4 text-[14px] font-medium text-text-muted tracking-body">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-medium">
              {loading ? (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center text-text-muted tracking-body">
                    Loading...
                  </td>
                </tr>
              ) : channels.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center text-text-muted tracking-body">
                    No channels found
                  </td>
                </tr>
              ) : (
                channels.map((channel) => (
                  <tr key={channel.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-6 py-4 text-[16px] text-text-secondary tracking-body">
                      {channel.name}
                    </td>
                    <td className="px-6 py-4">
                      <span className="badge">{channel.type}</span>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`badge ${channel.is_active ? 'badge-success' : 'badge-danger'}`}>
                        {channel.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-[16px] text-text-muted tracking-body">
                      {channel.models?.length || 0}
                    </td>
                    <td className="px-6 py-4 text-[16px] text-text-muted tracking-body">
                      {channel.api_keys?.length || 0}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <button className="p-2 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/5">
                          <Pencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteChannel(channel.id)}
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

      {/* Create Channel Modal */}
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
              Add Channel
            </h3>
            <form onSubmit={handleCreateChannel} className="space-y-4">
              <div>
                <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
                  Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="OpenAI Production"
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
                  <option value="openai">OpenAI</option>
                  <option value="anthropic">Anthropic</option>
                  <option value="custom">Custom</option>
                </select>
              </div>
              <div>
                <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
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
              <div className="flex justify-end gap-3 pt-2">
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
