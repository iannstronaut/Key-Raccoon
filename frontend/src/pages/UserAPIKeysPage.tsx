import { useState, useEffect } from 'react'
import { Plus, Trash2, X, Copy, Eye, EyeOff, Key } from 'lucide-react'
import { api } from '../services/api'
import { useAuth } from '../contexts/AuthContext'
import type { UserAPIKey, User, Channel, Model } from '../types'

export default function UserAPIKeysPage() {
  const [apiKeys, setApiKeys] = useState<UserAPIKey[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [channels, setChannels] = useState<Channel[]>([])
  const [allModels, setAllModels] = useState<Model[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [showKey, setShowKey] = useState<{ [key: number]: boolean }>({})
  const [formData, setFormData] = useState({
    user_id: 0,
    name: '',
    usage_limit: 0,
    expires_at: '',
    channel_ids: [] as number[],
    model_ids: [] as number[],
  })
  const { hasPermission, user } = useAuth()
  
  const canEdit = hasPermission('edit:users')
  const isAdmin = hasPermission('view:users')

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    setLoading(true)
    try {
      const [apiKeysRes, usersRes, channelsRes] = await Promise.all([
        isAdmin ? api.getUserAPIKeys() : api.getUserAPIKeysByUser(user?.id || 0),
        isAdmin ? api.getUsers() : Promise.resolve({ users: [] }),
        api.getChannels(),
      ])

      const apiKeysData = apiKeysRes as { api_keys?: UserAPIKey[] }
      const usersData = usersRes as { users?: User[] }
      const channelsData = channelsRes as { channels?: Channel[] }

      setApiKeys(apiKeysData?.api_keys || [])
      setUsers(usersData?.users || [])
      setChannels(channelsData?.channels || [])

      // Load all models from all channels
      const models: Model[] = []
      for (const channel of channelsData?.channels || []) {
        try {
          const modelsRes = await api.getChannelModels(channel.id)
          const modelsData = modelsRes as { models?: Model[] }
          if (modelsData?.models) {
            models.push(...modelsData.models)
          }
        } catch (err) {
          console.error(`Failed to load models for channel ${channel.id}:`, err)
        }
      }
      setAllModels(models)
    } catch (err) {
      console.error('Failed to load data:', err)
    } finally {
      setLoading(false)
    }
  }

  async function handleCreateAPIKey(e: React.FormEvent) {
    e.preventDefault()
    try {
      const response = await api.createUserAPIKey({
        ...formData,
        expires_at: formData.expires_at || undefined,
      })
      if (response && typeof response === 'object' && !('error' in response)) {
        setModalOpen(false)
        setFormData({
          user_id: 0,
          name: '',
          usage_limit: 0,
          expires_at: '',
          channel_ids: [],
          model_ids: [],
        })
        loadData()
      } else {
        alert('Error: ' + ((response as { error?: string })?.error || 'Failed to create API key'))
      }
    } catch (err) {
      alert('Error creating API key')
    }
  }

  async function handleDeleteAPIKey(id: number) {
    if (!confirm('Are you sure you want to delete this API key?')) return
    try {
      await api.deleteUserAPIKey(id)
      loadData()
    } catch (err) {
      alert('Error deleting API key')
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text)
    alert('API key copied to clipboard!')
  }

  function toggleShowKey(id: number) {
    setShowKey(prev => ({ ...prev, [id]: !prev[id] }))
  }

  function formatDate(date?: string) {
    if (!date) return 'Never'
    return new Date(date).toLocaleString()
  }

  function isExpired(expiresAt?: string) {
    if (!expiresAt) return false
    return new Date(expiresAt) < new Date()
  }

  function isLimitReached(apiKey: UserAPIKey) {
    if (apiKey.usage_limit === 0) return false
    return apiKey.usage_count >= apiKey.usage_limit
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[20px] font-medium text-white tracking-body">
            User API Keys
          </h2>
          <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
            Manage user API keys for accessing the system
          </p>
        </div>
        {canEdit && (
          <button onClick={() => setModalOpen(true)} className="btn-primary flex items-center gap-2">
            <Plus className="w-4 h-4" />
            Create API Key
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
                {isAdmin && (
                  <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                    User
                  </th>
                )}
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  API Key
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Status
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Usage
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Expires
                </th>
                <th className="text-right px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-medium">
              {loading ? (
                <tr>
                  <td colSpan={isAdmin ? 7 : 6} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    Loading...
                  </td>
                </tr>
              ) : apiKeys.length === 0 ? (
                <tr>
                  <td colSpan={isAdmin ? 7 : 6} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    No API keys found
                  </td>
                </tr>
              ) : (
                apiKeys.map((apiKey) => (
                  <tr key={apiKey.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-4 py-3 text-[14px] text-text-secondary tracking-body">
                      <div className="flex items-center gap-2">
                        <Key className="w-4 h-4 text-text-muted" />
                        {apiKey.name}
                      </div>
                    </td>
                    {isAdmin && (
                      <td className="px-4 py-3 text-[14px] text-text-secondary tracking-body">
                        {apiKey.user?.email || 'Unknown'}
                      </td>
                    )}
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <code className="text-[12px] font-mono text-text-secondary">
                          {showKey[apiKey.id] ? apiKey.key : '••••••••••••••••'}
                        </code>
                        <button
                          onClick={() => toggleShowKey(apiKey.id)}
                          className="p-1 text-text-muted hover:text-white transition-colors"
                        >
                          {showKey[apiKey.id] ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                        </button>
                        <button
                          onClick={() => copyToClipboard(apiKey.key)}
                          className="p-1 text-text-muted hover:text-white transition-colors"
                        >
                          <Copy className="w-3 h-3" />
                        </button>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      {isExpired(apiKey.expires_at) ? (
                        <span className="badge badge-danger">Expired</span>
                      ) : isLimitReached(apiKey) ? (
                        <span className="badge badge-warning">Limit Reached</span>
                      ) : apiKey.is_active ? (
                        <span className="badge badge-success">Active</span>
                      ) : (
                        <span className="badge badge-danger">Inactive</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-[14px] text-text-muted tracking-body">
                      {apiKey.usage_count} / {apiKey.usage_limit === 0 ? '∞' : apiKey.usage_limit}
                    </td>
                    <td className="px-4 py-3 text-[12px] text-text-muted tracking-body">
                      {apiKey.expires_at ? formatDate(apiKey.expires_at) : 'Never'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        {canEdit && (
                          <button
                            onClick={() => handleDeleteAPIKey(apiKey.id)}
                            className="p-1.5 text-text-muted hover:text-raycast-red transition-colors rounded-lg hover:bg-white/[0.05]"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create API Key Modal */}
      {canEdit && modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 glass-overlay">
          <div className="glass-strong w-full max-w-2xl p-5 relative rounded-xl shadow-2xl border border-white/[0.1] max-h-[90vh] overflow-y-auto">
            <button
              onClick={() => setModalOpen(false)}
              className="absolute top-3 right-3 p-1.5 text-text-muted hover:text-white transition-all rounded-lg hover:bg-white/[0.1]"
            >
              <X className="w-4 h-4" />
            </button>
            <h3 className="text-[18px] font-medium text-white tracking-body mb-4">
              Create User API Key
            </h3>
            <form onSubmit={handleCreateAPIKey} className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                    User *
                  </label>
                  <select
                    value={formData.user_id}
                    onChange={(e) => setFormData({ ...formData, user_id: parseInt(e.target.value) })}
                    required
                    className="input-dark"
                  >
                    <option value={0}>Select user</option>
                    {users.map((user) => (
                      <option key={user.id} value={user.id}>
                        {user.email} ({user.name})
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                    Name *
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="Production API Key"
                    required
                    className="input-dark"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                    Usage Limit (0 = unlimited)
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.usage_limit}
                    onChange={(e) => setFormData({ ...formData, usage_limit: parseInt(e.target.value) || 0 })}
                    placeholder="0"
                    className="input-dark"
                  />
                </div>
                <div>
                  <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                    Expires At (optional)
                  </label>
                  <input
                    type="datetime-local"
                    value={formData.expires_at}
                    onChange={(e) => setFormData({ ...formData, expires_at: e.target.value })}
                    className="input-dark"
                  />
                </div>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Allowed Channels (empty = all channels)
                </label>
                <div className="glass-subtle p-3 rounded-lg max-h-32 overflow-y-auto space-y-1">
                  {channels.map((channel) => (
                    <label key={channel.id} className="flex items-center gap-2 cursor-pointer hover:bg-white/[0.02] p-1 rounded">
                      <input
                        type="checkbox"
                        checked={formData.channel_ids.includes(channel.id)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setFormData({ ...formData, channel_ids: [...formData.channel_ids, channel.id] })
                          } else {
                            setFormData({ ...formData, channel_ids: formData.channel_ids.filter(id => id !== channel.id) })
                          }
                        }}
                        className="rounded"
                      />
                      <span className="text-[12px] text-text-secondary">{channel.name} ({channel.type})</span>
                    </label>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Allowed Models (empty = all models)
                </label>
                <div className="glass-subtle p-3 rounded-lg max-h-32 overflow-y-auto space-y-1">
                  {allModels.map((model) => (
                    <label key={model.id} className="flex items-center gap-2 cursor-pointer hover:bg-white/[0.02] p-1 rounded">
                      <input
                        type="checkbox"
                        checked={formData.model_ids.includes(model.id)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setFormData({ ...formData, model_ids: [...formData.model_ids, model.id] })
                          } else {
                            setFormData({ ...formData, model_ids: formData.model_ids.filter(id => id !== model.id) })
                          }
                        }}
                        className="rounded"
                      />
                      <span className="text-[12px] text-text-secondary">{model.display_name || model.name}</span>
                    </label>
                  ))}
                </div>
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
                  Create API Key
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
