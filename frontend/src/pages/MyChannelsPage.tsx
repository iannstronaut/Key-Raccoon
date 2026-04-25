import { useState, useEffect } from 'react'
import { Link2, Cpu, ChevronDown, ChevronRight, DollarSign } from 'lucide-react'
import { api } from '../services/api'
import type { Channel, Model } from '../types'

export default function MyChannelsPage() {
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedChannel, setExpandedChannel] = useState<number | null>(null)

  useEffect(() => {
    loadChannels()
  }, [])

  async function loadChannels() {
    setLoading(true)
    try {
      const res = await api.getMyChannels()
      const data = res as { channels?: Channel[] }
      setChannels(data?.channels || [])
    } catch (err) {
      console.error('Failed to load channels:', err)
    } finally {
      setLoading(false)
    }
  }

  function toggleChannel(channelId: number) {
    setExpandedChannel(expandedChannel === channelId ? null : channelId)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-text-muted">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-[20px] font-medium text-white tracking-body">
          My Channels
        </h2>
        <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
          Channels assigned to you by admin
        </p>
      </div>

      {channels.length === 0 ? (
        <div className="card-elevated p-8 text-center">
          <Link2 className="w-10 h-10 mx-auto mb-3 text-text-muted opacity-50" />
          <p className="text-[14px] text-text-muted tracking-body">
            No channels assigned to you yet
          </p>
          <p className="text-[12px] text-text-dim tracking-body mt-1">
            Contact your admin to get access to channels
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {channels.map((channel) => (
            <div key={channel.id} className="card-elevated overflow-hidden">
              <button
                onClick={() => toggleChannel(channel.id)}
                className="w-full p-4 flex items-center justify-between hover:bg-white/[0.02] transition-colors"
              >
                <div className="flex items-center gap-3">
                  <Link2 className="w-5 h-5 text-text-muted" />
                  <div className="text-left">
                    <p className="text-[14px] font-medium text-white tracking-body">
                      {channel.name}
                    </p>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className="badge">{channel.type}</span>
                      <span className={`badge ${channel.is_active ? 'badge-success' : 'badge-danger'}`}>
                        {channel.is_active ? 'Active' : 'Inactive'}
                      </span>
                      <span className="text-[12px] text-text-dim">
                        {(channel.models as Model[] || []).length} models
                      </span>
                      <span className="text-[12px] text-text-dim">
                        {channel.budget <= 0
                          ? 'Unlimited budget'
                          : channel.budget_type === 'token'
                            ? `${channel.budget_used.toLocaleString()} / ${channel.budget.toLocaleString()} tokens`
                            : `$${channel.budget_used.toFixed(2)} / $${channel.budget.toFixed(2)}`}
                      </span>
                    </div>
                  </div>
                </div>
                {expandedChannel === channel.id ? (
                  <ChevronDown className="w-4 h-4 text-text-muted" />
                ) : (
                  <ChevronRight className="w-4 h-4 text-text-muted" />
                )}
              </button>

              {expandedChannel === channel.id && (
                <div className="border-t border-white/[0.08] p-4 space-y-4">
                  {channel.description && (
                    <p className="text-[12px] text-text-muted tracking-body">
                      {channel.description}
                    </p>
                  )}

                  {/* Budget */}
                  <div>
                    <h4 className="text-[12px] font-medium text-text-tertiary tracking-body mb-2 flex items-center gap-1.5">
                      <DollarSign className="w-3.5 h-3.5" />
                      Budget
                    </h4>
                    <div className="p-2.5 glass-subtle rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-[12px] text-text-muted tracking-body">
                          {channel.budget <= 0
                            ? 'Unlimited'
                            : channel.budget_type === 'token'
                              ? `${channel.budget_used.toLocaleString()} used of ${channel.budget.toLocaleString()} tokens`
                              : `$${channel.budget_used.toFixed(4)} used of $${channel.budget.toFixed(4)}`}
                        </span>
                        {channel.budget > 0 && (
                          <span className={`text-[11px] font-medium ${
                            channel.budget_used / channel.budget > 0.9 ? 'text-raycast-red' :
                            channel.budget_used / channel.budget > 0.7 ? 'text-raycast-orange' :
                            'text-raycast-green'
                          }`}>
                            {channel.budget_type === 'token'
                              ? `${(channel.budget - channel.budget_used).toLocaleString()} tokens remaining`
                              : `$${(channel.budget - channel.budget_used).toFixed(4)} remaining`}
                          </span>
                        )}
                      </div>
                      {channel.budget > 0 && (
                        <div className="w-full h-1.5 bg-white/[0.05] rounded-full overflow-hidden mt-2">
                          <div
                            className={`h-full rounded-full transition-all ${
                              channel.budget_used / channel.budget > 0.9 ? 'bg-raycast-red' :
                              channel.budget_used / channel.budget > 0.7 ? 'bg-raycast-orange' :
                              'bg-raycast-green'
                            }`}
                            style={{ width: `${Math.min((channel.budget_used / channel.budget) * 100, 100)}%` }}
                          />
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Models */}
                  <div>
                  <h4 className="text-[12px] font-medium text-text-tertiary tracking-body mb-2 flex items-center gap-1.5">
                    <Cpu className="w-3.5 h-3.5" />
                    Available Models
                  </h4>
                  {(channel.models as Model[] || []).length === 0 ? (
                    <p className="text-[12px] text-text-dim tracking-body">No models available</p>
                  ) : (
                    <div className="space-y-1.5">
                      {(channel.models as Model[] || []).map((model) => (
                        <div
                          key={model.id}
                          className="flex items-center justify-between p-2.5 glass-subtle rounded-lg"
                        >
                          <div>
                            <p className="text-[13px] font-medium text-text-secondary tracking-body">
                              {model.display_name || model.name}
                            </p>
                            <p className="text-[11px] font-mono text-text-muted tracking-body">
                              {model.name}
                            </p>
                          </div>
                          <div className="flex items-center gap-2">
                            {(model.token_price ?? 0) > 0 && (
                              <span className="text-[11px] text-text-dim">
                                ${(model.token_price ?? 0).toFixed(4)}/1K
                              </span>
                            )}
                            <span className={`badge ${model.is_active ? 'badge-success' : 'badge-danger'}`}>
                              {model.is_active ? 'Active' : 'Inactive'}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
