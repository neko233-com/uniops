import { useState, useEffect } from 'react'

interface Server {
  id: number
  name: string
  host: string
  port: number
  username: string
  auth_type: string
  status: string
}

export function ServerManager() {
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    host: '',
    port: 22,
    username: 'root',
    auth_type: 'password',
    auth_data: '',
  })

  useEffect(() => {
    fetchServers()
  }, [])

  const fetchServers = async () => {
    try {
      const res = await fetch('/api/servers', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
      })
      if (res.ok) {
        const data = await res.json()
        setServers(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch servers:', err)
    }
    setLoading(false)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const res = await fetch('/api/servers', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(formData),
      })
      if (res.ok) {
        setShowForm(false)
        setFormData({ name: '', host: '', port: 22, username: 'root', auth_type: 'password', auth_data: '' })
        fetchServers()
      }
    } catch (err) {
      console.error('Failed to create server:', err)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this server?')) return
    try {
      await fetch(`/api/servers/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
      })
      fetchServers()
    } catch (err) {
      console.error('Failed to delete server:', err)
    }
  }

  if (loading) {
    return <div className="p-4 text-center text-gray-400">Loading...</div>
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <span className="text-sm text-gray-400">{servers.length} servers</span>
        <button onClick={() => setShowForm(!showForm)} className="rog-btn rog-btn-primary text-sm">
          {showForm ? 'Cancel' : 'Add Server'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="rog-panel p-4 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <input
              placeholder="Name"
              value={formData.name}
              onChange={e => setFormData({...formData, name: e.target.value})}
              className="rog-input"
              required
            />
            <input
              placeholder="Host"
              value={formData.host}
              onChange={e => setFormData({...formData, host: e.target.value})}
              className="rog-input"
              required
            />
            <input
              placeholder="Port"
              type="number"
              value={formData.port}
              onChange={e => setFormData({...formData, port: Number(e.target.value)})}
              className="rog-input"
            />
            <input
              placeholder="Username"
              value={formData.username}
              onChange={e => setFormData({...formData, username: e.target.value})}
              className="rog-input"
            />
            <select
              value={formData.auth_type}
              onChange={e => setFormData({...formData, auth_type: e.target.value})}
              className="rog-input"
            >
              <option value="password">Password</option>
              <option value="key">SSH Key</option>
            </select>
            <input
              placeholder={formData.auth_type === 'password' ? 'Password' : 'Private Key'}
              type={formData.auth_type === 'password' ? 'password' : 'text'}
              value={formData.auth_data}
              onChange={e => setFormData({...formData, auth_data: e.target.value})}
              className="rog-input"
            />
          </div>
          <button type="submit" className="rog-btn rog-btn-primary">Save</button>
        </form>
      )}

      <div className="space-y-2">
        {servers.map(server => (
          <div key={server.id} className="rog-panel p-3 flex items-center justify-between">
            <div>
              <span className="font-medium">{server.name}</span>
              <span className="text-gray-500 ml-2 text-sm">{server.host}:{server.port}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className={`text-xs px-2 py-1 rounded ${
                server.status === 'online' ? 'bg-green-900 text-green-300' : 'bg-gray-700 text-gray-400'
              }`}>
                {server.status || 'offline'}
              </span>
              <button onClick={() => handleDelete(server.id)} className="text-red-500 text-sm">Delete</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}