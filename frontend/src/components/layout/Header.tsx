import { useMemo } from 'react'
import { useAuth } from '../../hooks/useAuth'
import { useLocation, useNavigate } from 'react-router-dom'

interface HeaderProps {
  onToggleSidebar?: () => void
  onToggleCollapse?: () => void
}

const titleMap: Array<{ path: string; label: string; description: string }> = [
  { path: '/', label: 'Dashboard', description: 'Snapshot of crawler health' },
  { path: '/tasks', label: 'Tasks', description: 'Manage crawls & schedules' },
  { path: '/products', label: 'Products', description: 'Catalog inventory' },
  { path: '/proxies', label: 'Proxies', description: 'Network pool oversight' },
  { path: '/tokens', label: 'API Tokens', description: 'Programmatic access' },
  { path: '/logs', label: 'Live Logs', description: 'Real-time telemetry' },
]

export default function Header({ onToggleSidebar, onToggleCollapse }: HeaderProps) {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const pageMeta = useMemo(() => {
    return (
      titleMap.find((entry) =>
        entry.path === '/' ? location.pathname === '/' : location.pathname.startsWith(entry.path)
      ) ?? titleMap[0]
    )
  }, [location.pathname])

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <header className="bg-white/90 backdrop-blur border-b border-gray-200 sticky top-0 z-20">
      <div className="px-4 lg:px-8 py-3 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <button
            type="button"
            onClick={() => onToggleSidebar?.()}
            className="lg:hidden inline-flex h-10 w-10 items-center justify-center rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50"
            aria-label="Toggle navigation"
          >
            ☰
          </button>
          <div>
            <p className="text-xs uppercase tracking-wider text-gray-500">Admin</p>
            <h1 className="text-xl font-semibold text-gray-900">{pageMeta.label}</h1>
            <p className="text-sm text-gray-500">{pageMeta.description}</p>
          </div>
        </div>
        <div className="flex items-center space-x-3">
          <button
            type="button"
            onClick={() => navigate('/tasks?new=1')}
            className="hidden sm:inline-flex items-center rounded-lg bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-500 transition"
          >
            + New Task
          </button>
          <button
            type="button"
            onClick={() => navigate('/tokens?tab=generate')}
            className="hidden lg:inline-flex items-center rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition"
          >
            Generate Token
          </button>
          <button
            type="button"
            onClick={() => onToggleCollapse?.()}
            className="hidden lg:inline-flex items-center rounded-lg border border-gray-200 px-2 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50 transition"
            aria-label="Collapse sidebar"
          >
            ⇆
          </button>
          <div className="flex items-center space-x-2 rounded-full bg-gray-50 px-3 py-1 border border-gray-200">
            <div className="h-8 w-8 rounded-full bg-indigo-100 text-indigo-700 flex items-center justify-center text-sm font-semibold">
              {user?.username?.slice(0, 2).toUpperCase()}
            </div>
            <div className="hidden md:block">
              <p className="text-sm font-medium text-gray-900">{user?.username}</p>
              <button
                onClick={handleLogout}
                className="text-xs text-gray-500 hover:text-gray-700 transition"
              >
                Logout
              </button>
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="md:hidden inline-flex items-center rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50 transition"
          >
            Logout
          </button>
        </div>
      </div>
    </header>
  )
}

