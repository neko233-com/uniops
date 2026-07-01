import { useState, useEffect } from 'react'

interface Server {
  id: number
  name: string
  host: string
  status: string
}

interface ServerSelectorProps {
  onSelect: (serverId: number) => void
  selectedId: number | null
}

export function ServerSelector({ onSelect, selectedId }: ServerSelectorProps) {
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchServers()
  }, [])

  const fetchServers = async () => {
    try {
      const res = await fetch('/api/servers', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
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

  if (loading) {
    return <div className="text-gray-500 text-sm">Loading servers...</div>
  }

  if (servers.length === 0) {
    return <div className="text-gray-500 text-sm">No servers configured</div>
  }

  return (
    <div className="flex items-center gap-2">
      <label className="text-sm text-gray-400">Server:</label>
      <select
        value={selectedId || ''}
        onChange={(e) => onSelect(Number(e.target.value))}
        className="rog-input text-sm px-2 py-1"
      >
        <option value="">Select server</option>
        {servers.map(server => (
          <option key={server.id} value={server.id}>
            {server.name} ({server.host})
          </option>
        ))}
      </select>
      {selectedId && (
        <span className={`text-xs px-2 py-1 rounded ${
          servers.find(s => s.id === selectedId)?.status === 'online'
            ? 'bg-green-900 text-green-300'
            : 'bg-gray-700 text-gray-400'
        }`}>
          {servers.find(s => s.id === selectedId)?.status || 'unknown'}
        </span>
      )}
    </div>
  )
}
