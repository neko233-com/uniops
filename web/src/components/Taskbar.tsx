interface TaskbarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  activeApp: string | null
  onSelect: (id: string) => void
}

export function Taskbar({ apps, activeApp, onSelect }: TaskbarProps) {
  return (
    <div className="h-12 bg-gray-800 border-t border-gray-700 flex items-center px-4">
      <div className="flex gap-2">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className={`px-3 py-1 rounded text-sm ${
              activeApp === app.id
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
          >
            {app.icon} {app.title}
          </button>
        ))}
      </div>
      <div className="ml-auto text-gray-400 text-sm">
        UniOps
      </div>
    </div>
  )
}
