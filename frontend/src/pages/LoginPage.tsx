import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Shield, AlertCircle } from 'lucide-react'
import { api } from '../services/api'

export default function LoginPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const response = await api.login(email, password)
      if (response?.access_token) {
        api.setToken(response.access_token)
        navigate('/dashboard')
      } else {
        setError(response?.error || 'Invalid credentials')
      }
    } catch {
      setError('Network error. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen px-6 bg-bg-deep">
      <div className="w-full max-w-[400px]">
        <div className="card p-10">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-raycast-red/15 mb-4">
              <Shield className="w-7 h-7 text-raycast-red" />
            </div>
            <h1 className="text-[28px] font-semibold text-text-primary tracking-tight">
              KeyRaccoon
            </h1>
            <p className="text-[14px] text-text-tertiary mt-2 tracking-body">
              Admin Dashboard
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@keyraccoon.com"
                required
                autoComplete="email"
                className="input-dark"
              />
            </div>

            <div>
              <label className="block text-[14px] font-medium text-text-tertiary mb-2 tracking-body">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter your password"
                required
                autoComplete="current-password"
                className="input-dark"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full mt-2 disabled:opacity-50"
            >
              {loading ? 'Signing in...' : 'Sign In'}
            </button>

            {error && (
              <div className="mt-4 p-3 bg-red-transparent border border-raycast-red/30 rounded-lg flex items-start gap-2.5">
                <AlertCircle className="w-5 h-5 text-raycast-red flex-shrink-0 mt-0.5" />
                <p className="text-[14px] font-medium text-raycast-red tracking-body">
                  {error}
                </p>
              </div>
            )}
          </form>
        </div>
      </div>
    </div>
  )
}
