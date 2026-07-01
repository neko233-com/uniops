import { useState } from 'react'

interface LoginProps {
  onLogin: (token: string, user: any) => void
}

export function Login({ onLogin }: LoginProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      if (res.ok) {
        const data = await res.json()
        localStorage.setItem('token', data.access_token)
        onLogin(data.access_token, data.user)
      } else {
        setError('Invalid credentials')
      }
    } catch (err) {
      setError('Connection failed')
    }
    setLoading(false)
  }

  return (
    <div className="h-screen flex items-center justify-center rog-bg">
      <div className="rog-panel w-96 p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold rog-brand">UNIOPS</h1>
          <p className="text-gray-500 mt-2">Agent-First Bastion Host</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full rog-input"
              required
            />
          </div>

          <div>
            <label className="block text-sm text-gray-400 mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rog-input"
              required
            />
          </div>

          {error && (
            <div className="text-red-500 text-sm">{error}</div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full rog-btn rog-btn-primary py-2"
          >
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="mt-6 text-center text-xs text-gray-600">
          Default: root / root
        </div>
      </div>
    </div>
  )
}