import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { statsAPI, tasksAPI, productsAPI } from '../../services/api'

interface StatsResponse {
  tasks: {
    total: number
    by_status: Record<string, number>
  }
  proxies: {
    total: number
    active: number
  }
}

interface Task {
  id: number
  name: string
  status: string
  type: string
  updated_at: string
}

interface ProductStats {
  total_products?: number
  in_stock?: number
}

export default function Dashboard() {
  const [stats, setStats] = useState<StatsResponse | null>(null)
  const [productStats, setProductStats] = useState<ProductStats | null>(null)
  const [recentTasks, setRecentTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [polling, setPolling] = useState(true)

  useEffect(() => {
    refresh()
  }, [])

  useEffect(() => {
    if (!polling) return
    const interval = setInterval(() => refresh(false), 20000)
    return () => clearInterval(interval)
  }, [polling])

  const refresh = async (showSpinner = true) => {
    if (showSpinner) setLoading(true)
    try {
      const [statsRes, tasksRes, prodRes] = await Promise.all([
        statsAPI.get(),
        tasksAPI.list(5, 0),
        productsAPI.getStats(),
      ])
      setStats(statsRes.data)
      setRecentTasks(tasksRes.data.tasks || [])
      setProductStats(prodRes.data || {})
    } catch (error) {
      console.error('Failed to load dashboard data', error)
    } finally {
      if (showSpinner) setLoading(false)
    }
  }

  const taskStatusChart = useMemo(() => {
    if (!stats) return []
    const total = stats.tasks.total || 1
    return Object.entries(stats.tasks.by_status).map(([status, count]) => ({
      status,
      count,
      percent: Math.round((count / total) * 100),
    }))
  }, [stats])

  if (loading || !stats) {
    return <div className="py-12 text-center text-gray-500">Loading dashboard…</div>
  }

  const alerts: string[] = []
  if ((stats.tasks.by_status.failed || 0) > 0) {
    alerts.push(`${stats.tasks.by_status.failed} task(s) failed`)
  }
  if (stats.proxies.active < Math.max(3, Math.round(stats.proxies.total * 0.4))) {
    alerts.push('Proxy pool is running low')
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-gray-900">Control Center</h1>
          <p className="text-sm text-gray-500">Live snapshot of crawler health & throughput</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => refresh()}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
          >
            Refresh now
          </button>
          <button
            onClick={() => setPolling((prev) => !prev)}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
          >
            {polling ? 'Pause auto-refresh' : 'Resume auto-refresh'}
          </button>
        </div>
      </div>

      {alerts.length > 0 && (
        <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
          <p className="font-semibold">Attention required:</p>
          <ul className="mt-1 list-inside list-disc space-y-1">
            {alerts.map((alert) => (
              <li key={alert}>{alert}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total tasks" value={stats.tasks.total} />
        <StatCard
          label="Running tasks"
          value={stats.tasks.by_status.running || 0}
          tone="green"
        />
        <StatCard label="Active proxies" value={stats.proxies.active} tone="indigo" />
        <StatCard
          label="Products indexed"
          value={productStats?.total_products ?? 0}
          tone="purple"
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <section className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Task status mix</h2>
            <span className="text-xs text-gray-500">Updated {new Date().toLocaleTimeString()}</span>
          </div>
          <div className="mt-4 space-y-3">
            {taskStatusChart.map((entry) => (
              <div key={entry.status}>
                <div className="flex items-center justify-between text-sm text-gray-600">
                  <span className="capitalize">{entry.status}</span>
                  <span className="font-semibold text-gray-900">
                    {entry.count} / {entry.percent}%
                  </span>
                </div>
                <div className="mt-1 h-2 rounded-full bg-gray-100">
                  <div
                    className="h-2 rounded-full bg-indigo-500"
                    style={{ width: `${entry.percent}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Recent activity</h2>
            <Link
              to="/tasks"
              className="text-xs font-semibold text-indigo-600 hover:text-indigo-500"
            >
              View all
            </Link>
          </div>
          <div className="mt-4 divide-y divide-gray-100">
            {recentTasks.length === 0 ? (
              <p className="py-8 text-center text-sm text-gray-500">No tasks found.</p>
            ) : (
              recentTasks.map((task) => (
                <div key={task.id} className="flex items-center justify-between py-3">
                  <div>
                    <p className="font-medium text-gray-900">{task.name}</p>
                    <p className="text-xs uppercase tracking-wide text-gray-500">
                      #{task.id} • {task.type}
                    </p>
                  </div>
                  <div className="text-right">
                    <span className={`rounded-full px-2 py-0.5 text-xs font-semibold capitalize ${statusTone(task.status)}`}>
                      {task.status}
                    </span>
                    <p className="text-xs text-gray-500">
                      {new Date(task.updated_at).toLocaleTimeString()}
                    </p>
                  </div>
                </div>
              ))
            )}
          </div>
        </section>
      </div>
    </div>
  )
}

function StatCard({
  label,
  value,
  tone = 'gray',
}: {
  label: string
  value: number
  tone?: 'gray' | 'green' | 'indigo' | 'purple'
}) {
  const toneStyles: Record<string, string> = {
    gray: 'bg-white border-gray-100 text-gray-900',
    green: 'bg-green-50 border-green-100 text-green-800',
    indigo: 'bg-indigo-50 border-indigo-100 text-indigo-800',
    purple: 'bg-purple-50 border-purple-100 text-purple-800',
  }
  return (
    <div className={`rounded-2xl border p-4 shadow-sm ${toneStyles[tone]}`}>
      <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">{label}</p>
      <p className="mt-2 text-3xl font-semibold">{value}</p>
    </div>
  )
}

function statusTone(status: string) {
  switch (status) {
    case 'running':
      return 'bg-green-100 text-green-700'
    case 'failed':
      return 'bg-red-100 text-red-700'
    case 'pending':
      return 'bg-yellow-100 text-yellow-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

