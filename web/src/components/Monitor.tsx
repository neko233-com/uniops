import { useState, useEffect } from 'react'

interface SystemMetrics {
  cpu: { usage: number }
  memory: { total: number; used: number; usage: number }
  disk: Array<{ mount: string; size: number; used: number; usage: number }>
  network: Array<{ interface: string; rx_bytes: number; tx_bytes: number }>
  hostname: string
  uptime: string
  load_avg: number[]
}

interface MonitorProps {
  serverId: number
}

export function Monitor({ serverId }: MonitorProps) {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchMetrics = async () => {
    try {
      const res = await fetch(`/api/monitor/${serverId}`)
      if (res.ok) {
        const data = await res.json()
        setMetrics(data)
      }
    } catch (err) {
      console.error('Failed to fetch metrics:', err)
    }
    setLoading(false)
  }

  useEffect(() => {
    fetchMetrics()
    const interval = setInterval(fetchMetrics, 5000)
    return () => clearInterval(interval)
  }, [serverId])

  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }

  const getUsageColor = (usage: number) => {
    if (usage < 50) return 'bg-green-500'
    if (usage < 80) return 'bg-yellow-500'
    return 'bg-red-500'
  }

  if (loading) {
    return <div className="p-4 text-center text-gray-400">Loading...</div>
  }

  if (!metrics) {
    return <div className="p-4 text-center text-gray-400">Failed to load metrics</div>
  }

  return (
    <div className="space-y-4 h-full overflow-auto">
      {/* Header */}
      <div className="rog-panel p-4">
        <div className="flex justify-between items-center">
          <div>
            <h3 className="text-lg font-bold">{metrics.hostname}</h3>
            <p className="text-gray-400 text-sm">{metrics.uptime}</p>
          </div>
          <div className="text-right">
            <p className="text-gray-400 text-sm">Load Average</p>
            <p className="font-mono">{metrics.load_avg?.map((l: number) => l.toFixed(2)).join(' ')}</p>
          </div>
        </div>
      </div>

      {/* CPU */}
      <div className="rog-panel p-4">
        <h4 className="mb-2 font-medium">CPU</h4>
        <div className="flex items-center gap-4">
          <div className="flex-1 rog-input h-6 rounded overflow-hidden">
            <div
              className={`h-full ${getUsageColor(metrics.cpu.usage)}`}
              style={{ width: `${metrics.cpu.usage}%` }}
            />
          </div>
          <span className="w-16 text-right">{metrics.cpu.usage.toFixed(1)}%</span>
        </div>
      </div>

      {/* Memory */}
      <div className="rog-panel p-4">
        <h4 className="mb-2 font-medium">Memory</h4>
        <div className="flex items-center gap-4">
          <div className="flex-1 rog-input h-6 rounded overflow-hidden">
            <div
              className={`h-full ${getUsageColor(metrics.memory.usage)}`}
              style={{ width: `${metrics.memory.usage}%` }}
            />
          </div>
          <span className="w-36 text-right text-sm">
            {formatBytes(metrics.memory.used)} / {formatBytes(metrics.memory.total)}
          </span>
        </div>
      </div>

      {/* Disk */}
      <div className="rog-panel p-4">
        <h4 className="mb-2 font-medium">Disk</h4>
        <div className="space-y-2">
          {metrics.disk?.map((d) => (
            <div key={d.mount} className="flex items-center gap-4">
              <span className="w-24 text-sm truncate">{d.mount}</span>
              <div className="flex-1 rog-input h-4 rounded overflow-hidden">
                <div
                  className={`h-full ${getUsageColor(d.usage)}`}
                  style={{ width: `${d.usage}%` }}
                />
              </div>
              <span className="w-36 text-right text-sm">
                {formatBytes(d.used)} / {formatBytes(d.size)}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Network */}
      <div className="rog-panel p-4">
        <h4 className="mb-2 font-medium">Network</h4>
        <div className="space-y-2">
          {metrics.network?.map((n) => (
            <div key={n.interface} className="flex items-center gap-4">
              <span className="w-24 text-sm">{n.interface}</span>
              <span className="text-sm text-green-400">↓ {formatBytes(n.rx_bytes)}</span>
              <span className="text-sm text-blue-400">↑ {formatBytes(n.tx_bytes)}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
