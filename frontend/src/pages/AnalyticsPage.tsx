import { BarChart3, TrendingUp, Clock } from 'lucide-react'

export default function AnalyticsPage() {
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-[20px] font-medium text-white tracking-body">
          Analytics
        </h2>
        <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
          Usage statistics and insights
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <div className="card-elevated p-4">
          <div className="flex items-center gap-2.5 mb-4">
            <div className="w-8 h-8 rounded-lg bg-raycast-blue/15 flex items-center justify-center backdrop-blur-sm border border-raycast-blue/20">
              <TrendingUp className="w-4 h-4 text-raycast-blue" />
            </div>
            <div>
              <h3 className="text-[16px] font-medium text-white tracking-body">
                Token Usage Over Time
              </h3>
              <p className="text-[12px] text-text-muted tracking-body">
                Daily token consumption trends
              </p>
            </div>
          </div>
          <div className="flex items-center justify-center h-48 glass-subtle rounded-lg">
            <div className="text-center">
              <BarChart3 className="w-10 h-10 text-text-dark mx-auto mb-2" />
              <p className="text-[14px] text-text-muted tracking-body">
                Chart data will appear here
              </p>
              <p className="text-[12px] text-text-dim mt-1 tracking-body">
                Connect analytics backend to visualize
              </p>
            </div>
          </div>
        </div>

        <div className="card-elevated p-4">
          <div className="flex items-center gap-2.5 mb-4">
            <div className="w-8 h-8 rounded-lg bg-raycast-green/15 flex items-center justify-center backdrop-blur-sm border border-raycast-green/20">
              <Clock className="w-4 h-4 text-raycast-green" />
            </div>
            <div>
              <h3 className="text-[16px] font-medium text-white tracking-body">
                API Key Usage
              </h3>
              <p className="text-[12px] text-text-muted tracking-body">
                Usage breakdown by API key
              </p>
            </div>
          </div>
          <div className="flex items-center justify-center h-48 glass-subtle rounded-lg">
            <div className="text-center">
              <BarChart3 className="w-10 h-10 text-text-dark mx-auto mb-2" />
              <p className="text-[14px] text-text-muted tracking-body">
                Chart data will appear here
              </p>
              <p className="text-[12px] text-text-dim mt-1 tracking-body">
                Connect analytics backend to visualize
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="card-elevated p-4">
        <h3 className="text-[16px] font-medium text-white tracking-body mb-3">
          Recent Activity
        </h3>
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="flex items-center gap-3 p-3 glass-subtle rounded-lg hover:bg-white/[0.04] transition-all"
            >
              <div className="w-2 h-2 rounded-full bg-raycast-blue" />
              <div className="flex-1">
                <p className="text-[14px] text-text-secondary tracking-body">
                  API request processed
                </p>
                <p className="text-[12px] text-text-dim tracking-body">
                  Channel: 0penAI Production
                </p>
              </div>
              <span className="text-[12px] text-text-dim tracking-body">
                {i}m ago
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
