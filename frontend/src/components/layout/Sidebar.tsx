import { Link, useLocation } from 'react-router-dom'

const sections = [
  {
    title: 'Overview',
    items: [
      { name: 'Dashboard', href: '/', description: 'Realtime stats', icon: '📊' },
      { name: 'Live Logs', href: '/logs', description: 'Crawler console', icon: '🛰️' },
    ],
  },
  {
    title: 'Operations',
    items: [
      { name: 'Tasks', href: '/tasks', description: 'Pipelines & runs', icon: '🧵' },
      { name: 'Products', href: '/products', description: 'Catalog data', icon: '🛍️' },
      { name: 'Crawler Config', href: '/crawler/config', description: 'Payload filters', icon: '⚙️' },
    ],
  },
  {
    title: 'Infrastructure',
    items: [
      { name: 'Proxies', href: '/proxies', description: 'Network pool', icon: '🌐' },
      { name: 'API Tokens', href: '/tokens', description: 'Programmatic access', icon: '🔑' },
    ],
  },
]

interface SidebarProps {
  collapsed: boolean
  isMobileOpen: boolean
  onCloseMobile: () => void
  onToggleCollapse: () => void
}

export default function Sidebar({
  collapsed,
  isMobileOpen,
  onCloseMobile,
  onToggleCollapse,
}: SidebarProps) {
  const location = useLocation()

  const isActivePath = (href: string) => {
    if (href === '/') {
      return location.pathname === '/'
    }
    return location.pathname.startsWith(href)
  }

  return (
    <>
      <div
        className={`fixed inset-0 bg-gray-900/40 z-30 lg:hidden ${
          isMobileOpen ? 'block' : 'hidden'
        }`}
        onClick={onCloseMobile}
      />
      <aside
        className={`fixed z-40 inset-y-0 left-0 transform bg-white border-r border-gray-100 transition-transform duration-200 ease-out lg:static lg:translate-x-0 flex flex-col ${
          isMobileOpen ? 'translate-x-0' : '-translate-x-full'
        } ${collapsed ? 'w-20' : 'w-64'} `}
      >
        <div className="flex items-center justify-between px-4 py-5 border-b border-gray-100">
          <div className="flex items-center space-x-3">
            <div className="h-9 w-9 rounded-lg bg-indigo-600 text-white flex items-center justify-center text-lg font-semibold">
              ED
            </div>
            {!collapsed && (
              <div>
                <p className="text-sm font-semibold text-gray-900">Embroidery Admin</p>
                <p className="text-xs text-gray-500">Crawler Control Center</p>
              </div>
            )}
          </div>
        </div>
        <nav className="flex-1 overflow-y-auto px-2 py-4">
          {sections.map((section) => (
            <div key={section.title} className="mb-6">
              {!collapsed && (
                <p className="px-2 text-xs font-semibold uppercase tracking-wider text-gray-500 mb-2">
                  {section.title}
                </p>
              )}
              <div className="space-y-1">
                {section.items.map((item) => {
                  const active = isActivePath(item.href)
                  return (
                    <Link
                      key={item.name}
                      to={item.href}
                      onClick={onCloseMobile}
                      className={`group flex items-center rounded-lg px-3 py-2 text-sm font-medium transition ${
                        active
                          ? 'bg-indigo-50 text-indigo-700 border border-indigo-100'
                          : 'text-gray-700 hover:bg-gray-50'
                      }`}
                    >
                      <span className="text-lg mr-3">{item.icon}</span>
                      {!collapsed && (
                        <span className="flex flex-col">
                          {item.name}
                          <span className="text-xs font-normal text-gray-500">
                            {item.description}
                          </span>
                        </span>
                      )}
                    </Link>
                  )
                })}
              </div>
            </div>
          ))}
        </nav>
        <div className="border-t border-gray-100 p-3">
          <button
            onClick={onToggleCollapse}
            className="w-full rounded-lg border border-gray-200 text-sm font-medium py-2 hover:bg-gray-50 transition"
          >
            {collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          </button>
        </div>
      </aside>
    </>
  )
}

