import { useState } from 'react'
import { Taskbar } from './Taskbar'
import { Sidebar } from './Sidebar'
import { Window } from './Window'
import { Terminal } from './Terminal'
import { FileManager } from './FileManager'

interface App {
  id: string
  title: string
  icon: string
  component: React.ComponentType
}

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '🖥️', component: () => <Terminal serverId={1} /> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <FileManager serverId={1} /> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <div>Monitor</div> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <div>Agent</div> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <div>Audit</div> },
]

export function Desktop() {
  const [activeApp, setActiveApp] = useState<string | null>(null)

  const activeAppData = apps.find(a => a.id === activeApp)

  return (
    <div className="h-screen flex flex-col rog-bg">
      <div className="flex-1 flex overflow-hidden">
        <Sidebar apps={apps} onSelect={setActiveApp} />
        <main className="flex-1 overflow-auto p-4">
          {activeAppData && (
            <Window title={activeAppData.title}>
              <activeAppData.component />
            </Window>
          )}
        </main>
      </div>
      <Taskbar apps={apps} activeApp={activeApp} onSelect={setActiveApp} />
    </div>
  )
}
