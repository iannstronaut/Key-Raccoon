import { BarChart3, TrendingUp, Clock } from 'lucide-react'

export default function AnalyticsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-[24px] font-medium text-white tracking-body">
          Analytics
        </h2>
        <p className="text-[16px] text-text-muted mt-1 tracking-body">
          Usage statistics and insights
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card-elevated p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-lg bg-raycast-blue/15 flex items-center justify-center">
              <TrendingUp className="w-5 h-5 text-raycast-blue" />
            </div>
            <div>
              <h3 className="text-[20px] font-medium text-white tracking-body">
                Token Usage Over Time
              </h3>
              <p className="text-[14px] text-text-muted tracking-body">
                Daily token consumption trends
              </p>
            </div>
          </div>
          <div className="flex items-center justify-center h-64 bg-white/[0.02] rounded-lg">
            <div className="text-center">
              <BarChart3 className="w-12 h-12 text-text-dark mx-auto mb-3" />
              <p className="text-[16px] text-text-muted tracking-body">
                Chart data will appear here
              </p>
              <p className="text-[14px] text-text-dim mt-1 tracking-body">
                Connect analytics backend to visualize
              </p>
            </div>
          </div>
        </div>

        <div className="card-elevated p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-lg bg-raycast-green/15 flex items-center justify-center">
              <Clock className="w-5 h-5 text-raycast-green" />
            </div>
            <div>
              <h3 className="text-[20px] font-medium text-white tracking-body">
                API Key Usage
              </h3>
              <p className="text-[14px] text-text-muted tracking-body">
                Usage breakdown by API key
              </p>
            </div>
          </div>
          <div className="flex items-center justify-center h-64 bg-white/[0.02] rounded-lg">
            <div className="text-center">
              <BarChart3 className="w-12 h-12 text-text-dark mx-auto mb-3" />
              <p className="text-[16px] text-text-muted tracking-body">
                Chart data will appear here
              </p>
              <p className="text-[14px] text-text-dim mt-1 tracking-body">
                Connect analytics backend to visualize
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="card-elevated p-6">
        <h3 className="text-[20px] font-medium text-white tracking-body mb-4">
          Recent Activity
        </h3>
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="flex items-center gap-4 p-4 bg-white/[0.02] rounded-lg"
            >
              <div className="w-2 h-2 rounded-full bg-raycast-blue" />
              <div className="flex-1">
                <p className="text-[16px] text-text-secondary tracking-body">
                  API request processed
                </p>
                <p className="text-[14px] text-text-dim tracking-body">
                  Channel: OpenAI Production
                </p>
              </div>
              <span className="text-[14px] text-text-dim tracking-body">
                {i}m ago
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
