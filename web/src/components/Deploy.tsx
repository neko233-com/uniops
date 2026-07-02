import { useState, useEffect, useRef } from 'react'

interface Deployment {
  id: number
  server_id: number
  type: string
  status: string
  config: string
  logs: string
  triggered_by: string
  created_at: string
  completed_at?: string
  server?: { name: string; host: string }
}

interface DeployProps {
  serverId: number
}

export function Deploy({ serverId }: DeployProps) {
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [loading, setLoading] = useState(false)
  const [deployType, setDeployType] = useState('full')
  const [serviceName, setServiceName] = useState('uniops')
  const [binaryUrl, setBinaryUrl] = useState('')
  const [appPort, setAppPort] = useState(6020)
  const [domain, setDomain] = useState('_')
  const [nginxPort, setNginxPort] = useState(80)
  const [viewingLogs, setViewingLogs] = useState<Deployment | null>(null)
  const logRef = useRef<HTMLPreElement>(null)

  const token = localStorage.getItem('token')
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  }

  const fetchDeployments = async () => {
    if (!serverId) return
    try {
      const res = await fetch(`/api/deploy/server/${serverId}`, { headers })
      if (res.ok) {
        setDeployments(await res.json())
      }
    } catch (err) {
      console.error('Fetch deployments error:', err)
    }
  }

  useEffect(() => {
    fetchDeployments()
  }, [serverId])

  const handleDeploy = async () => {
    if (!serverId || loading) return
    setLoading(true)

    try {
      const res = await fetch('/api/deploy', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          server_id: serverId,
          type: deployType,
          service_name: serviceName,
          binary_url: binaryUrl,
          app_port: appPort,
          domain,
          nginx_port: nginxPort,
        }),
      })

      if (res.ok) {
        await fetchDeployments()
      }
    } catch (err) {
      console.error('Deploy error:', err)
    }
    setLoading(false)
  }

  const watchLogs = (deployment: Deployment) => {
    setViewingLogs(deployment)
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${proto}//${location.host}/api/deploy/${deployment.id}/ws`)

    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data)
      if (logRef.current) {
        logRef.current.textContent += msg.data + '\n'
        logRef.current.scrollTop = logRef.current.scrollHeight
      }
    }

    ws.onclose = () => {
      fetchDeployments()
    }
  }

  const statusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-400'
      case 'failed': return 'text-red-400'
      case 'running': return 'text-amber-400'
      default: return 'text-gray-400'
    }
  }

  if (!serverId) {
    return (
      <div className="text-center text-gray-500 py-8">Select a server to manage deployments</div>
    )
  }

  return (
    <div className="flex flex-col h-[calc(100vh-150px)]">
      {/* Deploy form */}
      <div className="rog-panel p-4 mb-4">
          <h3 className="text-sm font-bold mb-3">New Deployment</h3>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-gray-400">Type</label>
              <select
                value={deployType}
                onChange={(e) => setDeployType(e.target.value)}
                className="rog-input w-full"
              >
                <option value="full">Full (Backend + Nginx)</option>
                <option value="nginx">Nginx Only</option>
                <option value="backend">Backend Only</option>
              </select>
            </div>
            <div>
              <label className="text-xs text-gray-400">Service Name</label>
              <input
                value={serviceName}
                onChange={(e) => setServiceName(e.target.value)}
                className="rog-input w-full"
              />
            </div>
            {(deployType === 'backend' || deployType === 'full') && (
              <>
                <div>
                  <label className="text-xs text-gray-400">Binary URL</label>
                  <input
                    value={binaryUrl}
                    onChange={(e) => setBinaryUrl(e.target.value)}
                    placeholder="https://github.com/.../releases/download/..."
                    className="rog-input w-full"
                  />
                </div>
                <div>
                  <label className="text-xs text-gray-400">App Port</label>
                  <input
                    type="number"
                    value={appPort}
                    onChange={(e) => setAppPort(parseInt(e.target.value))}
                    className="rog-input w-full"
                  />
                </div>
              </>
            )}
            {(deployType === 'nginx' || deployType === 'full') && (
              <>
                <div>
                  <label className="text-xs text-gray-400">Domain</label>
                  <input
                    value={domain}
                    onChange={(e) => setDomain(e.target.value)}
                    placeholder="_ for default"
                    className="rog-input w-full"
                  />
                </div>
                <div>
                  <label className="text-xs text-gray-400">Nginx Port</label>
                  <input
                    type="number"
                    value={nginxPort}
                    onChange={(e) => setNginxPort(parseInt(e.target.value))}
                    className="rog-input w-full"
                  />
                </div>
              </>
            )}
          </div>
          <button
            onClick={handleDeploy}
            disabled={loading}
            className="rog-btn rog-btn-primary mt-3"
          >
            {loading ? 'Deploying...' : 'Deploy'}
          </button>
        </div>

        {/* Deployment history */}
        <div className="flex-1 overflow-auto">
          <h3 className="text-sm font-bold mb-2">History</h3>
          {deployments.length === 0 && (
            <div className="text-gray-500 text-sm">No deployments yet</div>
          )}
          {deployments.map((d) => (
            <div key={d.id} className="rog-panel p-3 mb-2 flex items-center justify-between">
              <div>
                <span className="font-mono text-sm">#{d.id}</span>
                <span className="ml-2 text-xs text-gray-400">{d.type}</span>
                <span className={`ml-2 text-xs ${statusColor(d.status)}`}>{d.status}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-gray-500">
                  {new Date(d.created_at).toLocaleString()}
                </span>
                <button
                  onClick={() => watchLogs(d)}
                  className="rog-btn text-xs"
                >
                  Logs
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* Log viewer modal */}
        {viewingLogs && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50" onClick={() => setViewingLogs(null)}>
            <div className="rog-panel w-3/4 h-3/4 flex flex-col p-4" onClick={(e) => e.stopPropagation()}>
              <div className="flex justify-between items-center mb-2">
                <h3 className="text-sm font-bold">
                  Deployment #{viewingLogs.id} ({viewingLogs.type}) - {viewingLogs.status}
                </h3>
                <button onClick={() => setViewingLogs(null)} className="rog-btn text-xs">Close</button>
              </div>
              <pre
                ref={logRef}
                className="flex-1 overflow-auto text-xs font-mono rog-panel p-3 whitespace-pre-wrap"
              >
                {viewingLogs.logs}
              </pre>
            </div>
          </div>
        )}
    </div>
  )
}
