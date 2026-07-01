import { useState } from 'react'
import { Taskbar } from './Taskbar'
import { Sidebar } from './Sidebar'
import { Window } from './Window'
import { Terminal } from './Terminal'
import { FileManager } from './FileManager'
import { Monitor } from './Monitor'
import { AgentChat } from './AgentChat'
import { Audit } from './Audit'
import { ServerSelector } from './ServerSelector'

interface App {
  id: string
  title: string
  icon: string
  component: React.ComponentType
}

interface DesktopProps {
  user: any
  onLogout: () => void
}

export function Desktop({ user, onLogout }: DesktopProps) {
  const [activeApp, setActiveApp] = useState<string | null>(null)
  const [selectedServerId, setSelectedServerId] = useState<number | null>(null)

  const apps: App[] = [
    { id: 'terminal', title: 'Terminal', icon: '🖥️', component: () => <Terminal serverId={selectedServerId ?? 0} /> },
    { id: 'files', title: 'Files', icon: '📁', component: () => <FileManager serverId={selectedServerId ?? 0} /> },
    { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <Monitor serverId={selectedServerId ?? 0} /> },
    { id: 'agent', title: 'Agent', icon: '🤖', component: () => <AgentChat agentId={1} /> },
    { id: 'audit', title: 'Audit', icon: '📋', component: () => <Audit /> },
  ]

  const activeAppData = apps.find(a => a.id === activeApp)

  return (
    <div className="h-screen flex flex-col rog-bg">
      <div className="flex-1 flex overflow-hidden">
        <Sidebar apps={apps} onSelect={setActiveApp} />
        <main className="flex-1 flex flex-col overflow-hidden">
          <div className="rog-panel-header flex items-center justify-between px-4 py-2 border-b border-gray-800">
            <ServerSelector onSelect={setSelectedServerId} selectedId={selectedServerId} />
            <div className="flex items-center gap-4">
              <span className="text-sm text-gray-400">{user?.username}</span>
              <button onClick={onLogout} className="rog-btn text-sm">Logout</button>
            </div>
          </div>
          <div className="flex-1 overflow-auto p-4">
            {activeAppData && (
              <Window title={activeAppData.title}>
                <activeAppData.component />
              </Window>
            )}
          </div>
        </main>
      </div>
      <Taskbar apps={apps} activeApp={activeApp} onSelect={setActiveApp} />
    </div>
  )
}
