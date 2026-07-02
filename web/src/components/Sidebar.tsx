interface NavItem {
  id: string
  label: string
  icon: string
  group: string
}

interface SidebarProps {
  items: NavItem[]
  activeId: string
  onSelect: (id: string) => void
}

export function Sidebar({ items, activeId, onSelect }: SidebarProps) {
  // Group items by group name
  const groups = items.reduce<Record<string, NavItem[]>>((acc, item) => {
    if (!acc[item.group]) acc[item.group] = []
    acc[item.group].push(item)
    return acc
  }, {})

  return (
    <aside className="w-56 rog-sidebar flex flex-col shrink-0 border-r border-gray-800">
      {/* Logo */}
      <div className="px-5 py-5 border-b border-gray-800">
        <h1 className="text-xl font-bold rog-brand tracking-wider">UNIOPS</h1>
        <p className="text-[10px] text-gray-600 mt-1 uppercase tracking-widest">Agent-First Bastion</p>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-3 overflow-y-auto">
        {Object.entries(groups).map(([group, navItems]) => (
          <div key={group} className="mb-2">
            <div className="px-5 py-2 text-[10px] font-semibold text-gray-600 uppercase tracking-wider">
              {group}
            </div>
            {navItems.map(item => (
              <button
                key={item.id}
                onClick={() => onSelect(item.id)}
                className={`w-full text-left px-5 py-2.5 flex items-center gap-3 text-sm transition-all
                  ${activeId === item.id
                    ? 'bg-gray-800/60 text-white border-l-2 border-red-500'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/30 border-l-2 border-transparent'
                  }`}
              >
                <span className="text-base w-5 text-center">{item.icon}</span>
                <span>{item.label}</span>
              </button>
            ))}
          </div>
        ))}
      </nav>

      {/* Footer */}
      <div className="px-5 py-3 border-t border-gray-800">
        <div className="text-[10px] text-gray-600">UniOps v1.0.1</div>
      </div>
    </aside>
  )
}
