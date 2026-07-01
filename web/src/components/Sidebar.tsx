interface SidebarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  onSelect: (id: string) => void
}

export function Sidebar({ apps, onSelect }: SidebarProps) {
  return (
    <div className="w-48 bg-gray-800 border-r border-gray-700 p-4">
      <h2 className="text-white font-bold mb-4">UniOps</h2>
      <nav className="space-y-1">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className="w-full text-left px-3 py-2 rounded text-gray-300 hover:bg-gray-700"
          >
            {app.icon} {app.title}
          </button>
        ))}
      </nav>
    </div>
  )
}
