import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, Trash2, Key, Cpu, X, RefreshCw } from 'lucide-react'
import { api } from '../services/api'
import { useAuth } from '../contexts/AuthContext'

interface Channel {
  id: number
  name: string
  type: string
  endpoint?: string
  description: string
  is_active: boolean
  api_keys?: APIKey[]
  models?: Model[]
}

interface APIKey {
  id: number
  api_key: string
  is_active: boolean
  created_at: string
}

interface Model {
  id: number
  name: string
  display_name: string
  is_active: boolean
  token_price: number
  system_prompt: string
}

interface DiscoveredModel {
  id: string
  object: string
  created?: number
  owned_by?: string
}

export default function ChannelDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { hasPermission } = useAuth()
  const [channel, setChannel] = useState<Channel | null>(null)
  const [loading, setLoading] = useState(true)
  
  // API Keys state
  const [newApiKeys, setNewApiKeys] = useState('')
  const [savingApiKeys, setSavingApiKeys] = useState(false)
  
  // Models state
  const [showModelForm, setShowModelForm] = useState(false)
  const [modelForm, setModelForm] = useState({
    name: '',
    display_name: '',
    token_price: 0,
    system_prompt: '',
  })
  const [savingModel, setSavingModel] = useState(false)
  
  // Autocheck Models state
  const [showAutocheckModal, setShowAutocheckModal] = useState(false)
  const [fetchingModels, setFetchingModels] = useState(false)
  const [availableModels, setAvailableModels] = useState<DiscoveredModel[]>([])
  const [selectedModels, setSelectedModels] = useState<Set<string>>(new Set())
  const [savingBulkModels, setSavingBulkModels] = useState(false)

  const canEdit = hasPermission('edit:channels')
  const canDelete = hasPermission('delete:channels')

  useEffect(() => {
    loadChannel()
  }, [id])

  async function loadChannel() {
    if (!id) return
    
    setLoading(true)
    try {
      const response = await api.getChannel(parseInt(id))
      setChannel(response as Channel)
    } catch (err) {
      console.error('Failed to load channel:', err)
    } finally {
      setLoading(false)
    }
  }

  async function handleAddApiKeys() {
    if (!id || !newApiKeys.trim()) return
    
    setSavingApiKeys(true)
    try {
      const keys = newApiKeys
        .split('\n')
        .map(k => k.trim())
        .filter(k => k.length > 0)
      
      for (const key of keys) {
        await api.addChannelAPIKey(parseInt(id), key)
      }
      
      setNewApiKeys('')
      loadChannel()
    } catch (err) {
      console.error('Failed to add API keys:', err)
      alert('Failed to add API keys')
    } finally {
      setSavingApiKeys(false)
    }
  }

  async function handleDeleteApiKey(keyId: number) {
    if (!id || !confirm('Are you sure you want to delete this API key?')) return
    
    try {
      await api.deleteChannelAPIKey(parseInt(id), keyId)
      loadChannel()
    } catch (err) {
      console.error('Failed to delete API key:', err)
      alert('Failed to delete API key')
    }
  }

  async function handleAddModel() {
    if (!id || !modelForm.name.trim()) return
    
    setSavingModel(true)
    try {
      await api.addChannelModel(parseInt(id), modelForm)
      setModelForm({
        name: '',
        display_name: '',
        token_price: 0,
        system_prompt: '',
      })
      setShowModelForm(false)
      loadChannel()
    } catch (err) {
      console.error('Failed to add model:', err)
      alert('Failed to add model')
    } finally {
      setSavingModel(false)
    }
  }

  async function handleDeleteModel(modelId: number) {
    if (!id || !confirm('Are you sure you want to delete this model?')) return
    
    try {
      await api.deleteChannelModel(parseInt(id), modelId)
      loadChannel()
    } catch (err) {
      console.error('Failed to delete model:', err)
      alert('Failed to delete model')
    }
  }

  async function handleAutocheckModels() {
    if (!channel?.endpoint) {
      alert('No endpoint configured for this channel')
      return
    }

    // Check if channel has active API keys
    const activeKey = channel.api_keys?.find(k => k.is_active)
    if (!activeKey) {
      alert('No active API key found. Please add an API key first.')
      return
    }

    setFetchingModels(true)
    setShowAutocheckModal(true)
    
    try {
      const endpoint = channel.endpoint.replace(/\/$/, '')
      const response = await fetch(`${endpoint}/models`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${activeKey.api_key}`,
        },
      })

      if (!response.ok) {
        throw new Error('Failed to fetch models from endpoint')
      }

      const data = await response.json()
      
      let models = []
      if (data.data && Array.isArray(data.data)) {
        models = data.data
      } else if (Array.isArray(data)) {
        models = data
      }

      setAvailableModels(models)
      setSelectedModels(new Set())
    } catch (err) {
      console.error('Failed to fetch models:', err)
      alert('Failed to fetch models from endpoint. Make sure the endpoint is correct and accessible.')
      setShowAutocheckModal(false)
    } finally {
      setFetchingModels(false)
    }
  }

  function selectAllModels() {
    setSelectedModels(new Set(availableModels.map(m => m.id)))
  }

  function deselectAllModels() {
    setSelectedModels(new Set())
  }

  async function handleSaveBulkModels() {
    if (!id || selectedModels.size === 0) return

    setSavingBulkModels(true)
    try {
      for (const modelId of selectedModels) {
        const model = availableModels.find(m => m.id === modelId)
        if (model) {
          await api.addChannelModel(parseInt(id), {
            name: model.id,
            display_name: model.id,
            token_price: 0,
            system_prompt: '',
          })
        }
      }

      setShowAutocheckModal(false)
      setAvailableModels([])
      setSelectedModels(new Set())
      loadChannel()
    } catch (err) {
      console.error('Failed to save models:', err)
      alert('Failed to save some models')
    } finally {
      setSavingBulkModels(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-text-muted">Loading...</p>
      </div>
    )
  }

  if (!channel) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-text-muted">Channel not found</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/channels')}
          className="p-2 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/[0.05]"
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div className="flex-1">
          <h2 className="text-[20px] font-medium text-white tracking-body">
            {channel.name}
          </h2>
          <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
            {channel.description || 'No description'}
          </p>
        </div>
        <span className={`badge ${channel.is_active ? 'badge-success' : 'badge-danger'}`}>
          {channel.is_active ? 'Active' : 'Inactive'}
        </span>
      </div>

      {/* Channel Info */}
      <div className="card-elevated p-4">
        <h3 className="text-[16px] font-medium text-white tracking-body mb-3 flex items-center gap-2">
          <Key className="w-4 h-4" />
          Channel Information
        </h3>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <p className="text-[12px] text-text-muted tracking-body">Type</p>
            <p className="text-[14px] text-text-secondary tracking-body mt-1">
              <span className="badge">{channel.type}</span>
            </p>
          </div>
          <div>
            <p className="text-[12px] text-text-muted tracking-body">Status</p>
            <p className="text-[14px] text-text-secondary tracking-body mt-1">
              <span className={`badge ${channel.is_active ? 'badge-success' : 'badge-danger'}`}>
                {channel.is_active ? 'Active' : 'Inactive'}
              </span>
            </p>
          </div>
          {channel.endpoint && (
            <div className="col-span-2">
              <p className="text-[12px] text-text-muted tracking-body">Endpoint</p>
              <p className="text-[12px] font-mono text-text-secondary tracking-body mt-1 break-all">
                {channel.endpoint}
              </p>
            </div>
          )}
          <div>
            <p className="text-[12px] text-text-muted tracking-body">API Keys</p>
            <p className="text-[14px] text-text-secondary tracking-body mt-1">
              {channel.api_keys?.length || 0} keys
            </p>
          </div>
          <div>
            <p className="text-[12px] text-text-muted tracking-body">Models</p>
            <p className="text-[14px] text-text-secondary tracking-body mt-1">
              {channel.models?.length || 0} models
            </p>
          </div>
        </div>
      </div>

      {/* API Keys Section */}
      <div className="card-elevated p-4">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-[16px] font-medium text-white tracking-body flex items-center gap-2">
            <Key className="w-4 h-4" />
            API Keys
          </h3>
        </div>

        {canEdit && (
          <div className="mb-4 p-3 glass-subtle rounded-lg">
            <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
              Add API Keys (one per line)
            </label>
            <textarea
              value={newApiKeys}
              onChange={(e) => setNewApiKeys(e.target.value)}
              placeholder="sk-1234567890abcdef&#10;sk-0987654321fedcba"
              rows={4}
              className="input-dark resize-none font-mono text-[12px]"
            />
            <div className="flex justify-end gap-2 mt-2">
              <button
                onClick={handleAddApiKeys}
                disabled={!newApiKeys.trim() || savingApiKeys}
                className="btn-primary text-[12px] px-3 py-1.5 disabled:opacity-50"
              >
                {savingApiKeys ? 'Adding...' : 'Add Keys'}
              </button>
            </div>
          </div>
        )}

        <div className="space-y-2">
          {channel.api_keys && channel.api_keys.length > 0 ? (
            channel.api_keys.map((key) => (
              <div
                key={key.id}
                className="flex items-center justify-between p-3 glass-subtle rounded-lg hover:bg-white/[0.04] transition-all"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-[12px] font-mono text-text-secondary tracking-body truncate">
                    {key.api_key.substring(0, 20)}...{key.api_key.substring(key.api_key.length - 4)}
                  </p>
                  <p className="text-[10px] text-text-dim tracking-body mt-0.5">
                    Added {new Date(key.created_at).toLocaleDateString()}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`badge ${key.is_active ? 'badge-success' : 'badge-danger'}`}>
                    {key.is_active ? 'Active' : 'Inactive'}
                  </span>
                  {canDelete && (
                    <button
                      onClick={() => handleDeleteApiKey(key.id)}
                      className="p-1.5 text-text-muted hover:text-raycast-red transition-colors rounded-lg hover:bg-white/[0.05]"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  )}
                </div>
              </div>
            ))
          ) : (
            <p className="text-[12px] text-text-muted text-center py-4">No API keys added yet</p>
          )}
        </div>
      </div>

      {/* Models Section */}
      <div className="card-elevated p-4">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-[16px] font-medium text-white tracking-body flex items-center gap-2">
            <Cpu className="w-4 h-4" />
            Models
          </h3>
          {canEdit && (
            <div className="flex items-center gap-2">
              {channel.endpoint && (
                <button
                  onClick={handleAutocheckModels}
                  disabled={fetchingModels}
                  className="btn-secondary text-[12px] px-3 py-1.5 flex items-center gap-1"
                >
                  <RefreshCw className={`w-3 h-3 ${fetchingModels ? 'animate-spin' : ''}`} />
                  {fetchingModels ? 'Checking...' : 'Autocheck Models'}
                </button>
              )}
              {!showModelForm && (
                <button
                  onClick={() => setShowModelForm(true)}
                  className="btn-primary text-[12px] px-3 py-1.5 flex items-center gap-1"
                >
                  <Plus className="w-3 h-3" />
                  Add Model
                </button>
              )}
            </div>
          )}
        </div>

        {canEdit && showModelForm && (
          <div className="mb-4 p-3 glass-subtle rounded-lg">
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-[14px] font-medium text-white tracking-body">Add New Model</h4>
              <button
                onClick={() => setShowModelForm(false)}
                className="p-1 text-text-muted hover:text-white transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                    Model Name *
                  </label>
                  <input
                    type="text"
                    value={modelForm.name}
                    onChange={(e) => setModelForm({ ...modelForm, name: e.target.value })}
                    placeholder="gpt-4"
                    className="input-dark"
                  />
                </div>
                <div>
                  <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                    Display Name
                  </label>
                  <input
                    type="text"
                    value={modelForm.display_name}
                    onChange={(e) => setModelForm({ ...modelForm, display_name: e.target.value })}
                    placeholder="GPT-4"
                    className="input-dark"
                  />
                </div>
              </div>
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Token Price (per 1K tokens)
                </label>
                <input
                  type="number"
                  step="0.0001"
                  value={modelForm.token_price}
                  onChange={(e) => setModelForm({ ...modelForm, token_price: parseFloat(e.target.value) || 0 })}
                  placeholder="0.03"
                  className="input-dark"
                />
              </div>
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  System Prompt (optional)
                </label>
                <textarea
                  value={modelForm.system_prompt}
                  onChange={(e) => setModelForm({ ...modelForm, system_prompt: e.target.value })}
                  placeholder="You are a helpful assistant..."
                  rows={3}
                  className="input-dark resize-none"
                />
              </div>
              <div className="flex justify-end gap-2">
                <button
                  onClick={() => setShowModelForm(false)}
                  className="btn-secondary text-[12px] px-3 py-1.5"
                >
                  Cancel
                </button>
                <button
                  onClick={handleAddModel}
                  disabled={!modelForm.name.trim() || savingModel}
                  className="btn-primary text-[12px] px-3 py-1.5 disabled:opacity-50"
                >
                  {savingModel ? 'Adding...' : 'Add Model'}
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="space-y-2">
          {channel.models && channel.models.length > 0 ? (
            channel.models.map((model) => (
              <div
                key={model.id}
                className="p-3 glass-subtle rounded-lg hover:bg-white/[0.04] transition-all"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <p className="text-[14px] font-medium text-text-secondary tracking-body">
                        {model.display_name || model.name}
                      </p>
                      <span className={`badge ${model.is_active ? 'badge-success' : 'badge-danger'}`}>
                        {model.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </div>
                    <p className="text-[12px] font-mono text-text-muted tracking-body mt-0.5">
                      {model.name}
                    </p>
                    {model.token_price > 0 && (
                      <p className="text-[12px] text-text-dim tracking-body mt-1">
                        ${model.token_price.toFixed(4)} per 1K tokens
                      </p>
                    )}
                    {model.system_prompt && (
                      <p className="text-[12px] text-text-dim tracking-body mt-1 line-clamp-2">
                        {model.system_prompt}
                      </p>
                    )}
                  </div>
                  {canDelete && (
                    <button
                      onClick={() => handleDeleteModel(model.id)}
                      className="p-1.5 text-text-muted hover:text-raycast-red transition-colors rounded-lg hover:bg-white/[0.05]"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  )}
                </div>
              </div>
            ))
          ) : (
            <p className="text-[12px] text-text-muted text-center py-4">No models added yet</p>
          )}
        </div>
      </div>

      {/* Autocheck Models Modal */}
      {showAutocheckModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 glass-overlay">
          <div className="glass-strong w-full max-w-2xl max-h-[80vh] flex flex-col relative rounded-xl shadow-2xl border border-white/[0.1]">
            <div className="p-5 border-b border-white/[0.08]">
              <div className="flex items-center justify-between">
                <h3 className="text-[18px] font-medium text-white tracking-body">
                  Available Models from Endpoint
                </h3>
                <button
                  onClick={() => {
                    setShowAutocheckModal(false)
                    setSelectedModels(new Set())
                  }}
                  className="p-1.5 text-text-muted hover:text-white transition-all rounded-lg hover:bg-white/[0.1]"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
              <p className="text-[12px] text-text-muted mt-1">
                Select models to add to this channel
              </p>
              {availableModels.length > 0 && (
                <div className="flex gap-2 mt-3">
                  <button
                    onClick={selectAllModels}
                    className="text-[12px] text-raycast-blue hover:underline"
                  >
                    Select All
                  </button>
                  <span className="text-text-dim">|</span>
                  <button
                    onClick={deselectAllModels}
                    className="text-[12px] text-raycast-blue hover:underline"
                  >
                    Deselect All
                  </button>
                </div>
              )}
            </div>

            <div className="flex-1 overflow-y-auto p-5">
              {fetchingModels ? (
                <div className="flex items-center justify-center py-8">
                  <RefreshCw className="w-6 h-6 text-raycast-blue animate-spin" />
                  <p className="text-[14px] text-text-muted ml-3">Fetching models...</p>
                </div>
              ) : availableModels.length === 0 ? (
                <p className="text-[14px] text-text-muted text-center py-8">
                  No models found from endpoint
                </p>
              ) : (
                <div className="space-y-2">
                  {availableModels.map((model) => (
                    <label
                      key={model.id}
                      className="flex items-start gap-3 p-3 glass-subtle rounded-lg hover:bg-white/[0.04] transition-all cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={selectedModels.has(model.id)}
                        onChange={(e) => {
                          const newSelected = new Set(selectedModels)
                          if (e.target.checked) {
                            newSelected.add(model.id)
                          } else {
                            newSelected.delete(model.id)
                          }
                          setSelectedModels(newSelected)
                        }}
                        className="mt-1 w-4 h-4 rounded border-white/[0.2] bg-white/[0.05] text-raycast-blue focus:ring-raycast-blue focus:ring-offset-0"
                      />
                      <div className="flex-1">
                        <p className="text-[14px] font-medium text-text-secondary tracking-body">
                          {model.id}
                        </p>
                        {model.owned_by && (
                          <p className="text-[12px] text-text-muted tracking-body mt-0.5">
                            Owned by: {model.owned_by}
                          </p>
                        )}
                        {model.created && (
                          <p className="text-[10px] text-text-dim tracking-body mt-0.5">
                            Created: {new Date(model.created * 1000).toLocaleDateString()}
                          </p>
                        )}
                      </div>
                    </label>
                  ))}
                </div>
              )}
            </div>

            <div className="p-5 border-t border-white/[0.08] flex items-center justify-between">
              <p className="text-[12px] text-text-muted">
                {selectedModels.size} model{selectedModels.size !== 1 ? 's' : ''} selected
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => {
                    setShowAutocheckModal(false)
                    setSelectedModels(new Set())
                  }}
                  className="btn-secondary text-[12px] px-3 py-1.5"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSaveBulkModels}
                  disabled={selectedModels.size === 0 || savingBulkModels}
                  className="btn-primary text-[12px] px-3 py-1.5 disabled:opacity-50"
                >
                  {savingBulkModels ? 'Saving...' : `Add ${selectedModels.size} Model${selectedModels.size !== 1 ? 's' : ''}`}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
