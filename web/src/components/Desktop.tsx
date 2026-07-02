import { useState, useEffect } from 'react'
import { Sidebar } from './Sidebar'
import { Terminal } from './Terminal'
import { FileManager } from './FileManager'
import { Monitor } from './Monitor'
import { AgentChat } from './AgentChat'
import { Audit } from './Audit'
import { ServerManager } from './ServerManager'
import { Deploy } from './Deploy'
import { OperationLog } from './OperationLog'

interface NavItem {
  id: string
  label: string
  icon: string
  group: string
}

interface DesktopProps {
  user: any
  onLogout: () => void
}

export function Desktop({ user, onLogout }: DesktopProps) {
  const [activePage, setActivePage] = useState('dashboard')
  const [selectedServerId, setSelectedServerId] = useState<number>(0)

  const navItems: NavItem[] = [
    { id: 'dashboard', label: 'Dashboard', icon: '📊', group: 'Overview' },
    { id: 'servers', label: 'Servers', icon: '🖥️', group: 'Infrastructure' },
    { id: 'terminal', label: 'Terminal', icon: '⌨️', group: 'Infrastructure' },
    { id: 'files', label: 'Files', icon: '📁', group: 'Infrastructure' },
    { id: 'monitor', label: 'Monitor', icon: '📈', group: 'Infrastructure' },
    { id: 'deploy', label: 'Deploy', icon: '🚀', group: 'Automation' },
    { id: 'agent', label: 'Agent', icon: '🤖', group: 'Automation' },
    { id: 'audit', label: 'Sessions', icon: '📋', group: 'Security' },
    { id: 'oplog', label: 'Op Logs', icon: '📝', group: 'Security' },
  ]

  const renderPage = () => {
    switch (activePage) {
      case 'dashboard':
        return <Dashboard user={user} />
      case 'servers':
        return <ServerManager />
      case 'terminal':
        return <Terminal serverId={selectedServerId} />
      case 'files':
        return <FileManager serverId={selectedServerId} />
      case 'monitor':
        return <Monitor serverId={selectedServerId} />
      case 'deploy':
        return <Deploy serverId={selectedServerId} />
      case 'agent':
        return <AgentChat agentId={1} serverId={selectedServerId} />
      case 'audit':
        return <Audit />
      case 'oplog':
        return <OperationLog />
      default:
        return <Dashboard user={user} />
    }
  }

  const needsServer = ['terminal', 'files', 'monitor', 'deploy', 'agent'].includes(activePage)
  const currentPage = navItems.find(n => n.id === activePage)

  return (
    <div className="h-screen flex rog-bg">
      {/* Left sidebar */}
      <Sidebar
        items={navItems}
        activeId={activePage}
        onSelect={setActivePage}
      />

      {/* Main area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-14 rog-panel-header flex items-center justify-between px-6 border-b border-gray-800 shrink-0">
          <div className="flex items-center gap-4">
            <h1 className="text-lg font-semibold text-white">
              {currentPage?.icon} {currentPage?.label}
            </h1>
            {needsServer && (
              <div className="flex items-center gap-2 ml-4">
                <label className="text-xs text-gray-500">Target:</label>
                <ServerQuickSelect
                  selectedId={selectedServerId}
                  onSelect={setSelectedServerId}
                />
              </div>
            )}
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-400">{user?.username}</span>
            <button onClick={onLogout} className="rog-btn text-xs py-1 px-3">
              Logout
            </button>
          </div>
        </header>

        {/* Content */}
        <main className="flex-1 overflow-auto p-6">
          {renderPage()}
        </main>
      </div>
    </div>
  )
}

// Inline server selector for header
function ServerQuickSelect({ selectedId, onSelect }: { selectedId: number; onSelect: (id: number) => void }) {
  const [servers, setServers] = useState<any[]>([])

  useEffect(() => {
    fetch('/api/servers', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
      .then(r => r.ok ? r.json() : [])
      .then(data => setServers(data || []))
      .catch(() => {})
  }, [])

  if (servers.length === 0) {
    return <span className="text-xs text-gray-500">No servers</span>
  }

  return (
    <select
      value={selectedId || ''}
      onChange={(e) => onSelect(Number(e.target.value))}
      className="rog-input text-xs py-1 px-2"
    >
      <option value="">Select server</option>
      {servers.map((s: any) => (
        <option key={s.id} value={s.id}>
          {s.name} ({s.host})
        </option>
      ))}
    </select>
  )
}

// Dashboard page
function Dashboard({ user }: { user: any }) {
  return (
    <div className="space-y-6">
      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Status" value="Online" color="text-green-400" />
        <StatCard label="User" value={user?.username || 'admin'} color="text-blue-400" />
        <StatCard label="Version" value="v1.0.1" color="text-purple-400" />
        <StatCard label="Platform" value="UniOps" color="text-amber-400" />
      </div>

      {/* Quick actions */}
      <div className="rog-panel p-6">
        <h2 className="text-sm font-semibold text-gray-400 mb-4">QUICK ACTIONS</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <QuickAction icon="🖥️" label="Add Server" desc="Register a new server" />
          <QuickAction icon="🚀" label="Deploy" desc="Deploy services to cluster" />
          <QuickAction icon="🤖" label="Agent Chat" desc="AI-powered operations" />
          <QuickAction icon="📝" label="Op Logs" desc="View operation history" />
        </div>
      </div>
    </div>
  )
}

function StatCard({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <div className="rog-panel p-4">
      <div className="text-xs text-gray-500 mb-1">{label}</div>
      <div className={`text-2xl font-bold ${color}`}>{value}</div>
    </div>
  )
}

function QuickAction({ icon, label, desc }: { icon: string; label: string; desc: string }) {
  return (
    <div className="rog-panel p-4 hover:border-gray-600 cursor-pointer transition-colors">
      <div className="text-2xl mb-2">{icon}</div>
      <div className="text-sm font-medium">{label}</div>
      <div className="text-xs text-gray-500 mt-1">{desc}</div>
    </div>
  )
}
