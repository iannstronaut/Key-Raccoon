import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { AlertCircle } from "lucide-react";
import { useAuth } from "../contexts/AuthContext";

export default function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const success = await login(email, password);
      if (success) {
        navigate("/dashboard");
      } else {
        setError("Invalid credentials");
      }
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen px-6 bg-bg-deep">
      <div className="w-full max-w-[400px]">
        <div className="glass-strong p-8 rounded-xl shadow-2xl border border-white/[0.1]">
          <div className="text-center mb-6">
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-xl backdrop-blur-sm mb-4 overflow-hidden">
              <img
                src="/keyraccoon_icon.png"
                alt="KeyRaccoon"
                className="w-20 h-20 object-contain"
              />
            </div>
            <h1 className="text-[24px] font-semibold text-text-primary tracking-tight">
              KeyRaccoon
            </h1>
            <p className="text-[12px] text-text-tertiary mt-1 tracking-body">
              Admin Dashboard
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
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
              <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
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
              className="btn-primary w-full mt-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Signing in..." : "Sign In"}
            </button>

            {error && (
              <div className="mt-3 p-3 glass-subtle border border-raycast-red/30 rounded-lg flex items-start gap-2.5">
                <AlertCircle className="w-4 h-4 text-raycast-red flex-shrink-0 mt-0.5" />
                <p className="text-[12px] font-medium text-raycast-red tracking-body">
                  {error}
                </p>
              </div>
            )}
          </form>
        </div>
      </div>
    </div>
  );
}
