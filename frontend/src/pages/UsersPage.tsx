import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, X } from 'lucide-react'
import { api } from '../services/api'
import type { User } from '../types'

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [formData, setFormData] = useState({
    email: '',
    name: '',
    password: '',
    role: 'user',
  })

  useEffect(() => {
    loadUsers()
  }, [])

  async function loadUsers() {
    setLoading(true)
    try {
      const response = await api.getUsers()
      const data = response as { users?: User[] }
      setUsers(data?.users ?? [])
    } catch (err) {
      console.error('Failed to load users:', err)
    } finally {
      setLoading(false)
    }
  }

  async function handleCreateUser(e: React.FormEvent) {
    e.preventDefault()
    try {
      const response = await api.createUser(formData)
      if (response && typeof response === 'object' && !('error' in response)) {
        setModalOpen(false)
        setFormData({ email: '', name: '', password: '', role: 'user' })
        loadUsers()
      } else {
        alert('Error: ' + ((response as { error?: string })?.error || 'Failed to create user'))
      }
    } catch (err) {
      alert('Error creating user')
    }
  }

  async function handleDeleteUser(id: number) {
    if (!confirm('Are you sure you want to delete this user?')) return
    try {
      await api.deleteUser(id)
      loadUsers()
    } catch (err) {
      alert('Error deleting user')
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[20px] font-medium text-white tracking-body">
            Users
          </h2>
          <p className="text-[14px] text-text-muted mt-0.5 tracking-body">
            Manage user accounts
          </p>
        </div>
        <button onClick={() => setModalOpen(true)} className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Create User
        </button>
      </div>

      <div className="card-elevated overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border-medium">
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Email
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Name
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Role
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Status
                </th>
                <th className="text-left px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Last Login
                </th>
                <th className="text-right px-4 py-3 text-[12px] font-medium text-text-muted tracking-body">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-medium">
              {loading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    Loading...
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-6 text-center text-text-muted tracking-body text-[14px]">
                    No users found
                  </td>
                </tr>
              ) : (
                users.map((user) => (
                  <tr key={user.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-4 py-3 text-[14px] text-text-secondary tracking-body">
                      {user.email}
                    </td>
                    <td className="px-4 py-3 text-[14px] text-text-secondary tracking-body">
                      {user.name || '—'}
                    </td>
                    <td className="px-4 py-3">
                      <span className="badge">{user.role}</span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`badge ${user.is_active ? 'badge-success' : 'badge-danger'}`}>
                        {user.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-[14px] text-text-muted tracking-body">
                      {user.last_login
                        ? new Date(user.last_login).toLocaleString()
                        : 'Never'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button className="p-1.5 text-text-muted hover:text-white transition-colors rounded-lg hover:bg-white/5">
                          <Pencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteUser(user.id)}
                          className="p-1.5 text-text-muted hover:text-raycast-red transition-colors rounded-lg hover:bg-white/5"
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

      {/* Create User Modal */}
      {modalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 glass-overlay">
          <div className="glass-strong w-full max-w-md p-5 relative rounded-xl shadow-2xl border border-white/[0.1]">
            <button
              onClick={() => setModalOpen(false)}
              className="absolute top-3 right-3 p-1.5 text-text-muted hover:text-white transition-all rounded-lg hover:bg-white/[0.1]"
            >
              <X className="w-4 h-4" />
            </button>
            <h3 className="text-[18px] font-medium text-white tracking-body mb-4">
              Create User
            </h3>
            <form onSubmit={handleCreateUser} className="space-y-3">
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Email
                </label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  placeholder="user@example.com"
                  required
                  className="input-dark"
                />
              </div>
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Full Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="John Doe"
                  required
                  className="input-dark"
                />
              </div>
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Password
                </label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Secure password"
                  required
                  className="input-dark"
                />
              </div>
              <div>
                <label className="block text-[12px] font-medium text-text-tertiary mb-1.5 tracking-body">
                  Role
                </label>
                <select
                  value={formData.role}
                  onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                  className="input-dark"
                >
                  <option value="user">User</option>
                  <option value="admin">Admin</option>
                </select>
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
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
