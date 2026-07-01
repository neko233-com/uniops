interface SidebarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  onSelect: (id: string) => void
}

export function Sidebar({ apps, onSelect }: SidebarProps) {
  return (
    <div className="w-48 rog-sidebar flex flex-col">
      <div className="p-4 border-b border-gray-800">
        <h2 className="text-xl font-bold rog-brand">UNIOPS</h2>
        <p className="text-xs text-gray-500 mt-1">Agent-First Bastion</p>
      </div>
      <nav className="flex-1 p-2">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className="rog-sidebar-item w-full text-left"
          >
            <span className="mr-3">{app.icon}</span>
            {app.title}
          </button>
        ))}
      </nav>
      <div className="p-4 border-t border-gray-800">
        <div className="text-xs text-gray-500">v1.0.0</div>
      </div>
    </div>
  )
}
