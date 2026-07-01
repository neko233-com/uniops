interface TaskbarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  activeApp: string | null
  onSelect: (id: string) => void
}

export function Taskbar({ apps, activeApp, onSelect }: TaskbarProps) {
  return (
    <div className="h-12 rog-taskbar flex items-center px-4">
      <div className="flex gap-2">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className={`rog-btn px-3 py-1 text-sm ${
              activeApp === app.id ? 'rog-btn-primary' : ''
            }`}
          >
            <span className="mr-2">{app.icon}</span>
            {app.title}
          </button>
        ))}
      </div>
      <div className="ml-auto rog-brand text-sm">
        UNIOPS
      </div>
    </div>
  )
}
