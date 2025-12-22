import { useEffect, useMemo, useState } from 'react'
import { proxiesAPI } from '../../services/api'

interface Proxy {
  id: number
  host: string
  port: number
  type: string
  is_active: boolean
  failure_count: number
  last_checked?: string
  username?: string
}

export default function ProxyList() {
  const [proxies, setProxies] = useState<Proxy[]>([])
  const [loading, setLoading] = useState(true)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(
    null
  )
  const [formState, setFormState] = useState({
    host: '',
    port: '',
    type: 'http',
    username: '',
    password: '',
  })
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'healthy' | 'unhealthy'>('all')
  const [submitting, setSubmitting] = useState(false)
  const [testingId, setTestingId] = useState<number | null>(null)

  useEffect(() => {
    loadProxies()
  }, [])

  const loadProxies = async () => {
    setLoading(true)
    try {
      const response = await proxiesAPI.list()
      setProxies(response.data.proxies || [])
    } catch (error) {
      console.error('Failed to load proxies:', error)
      setFeedback({ type: 'error', message: 'Failed to load proxies' })
    } finally {
      setLoading(false)
    }
  }

  const filteredProxies = useMemo(() => {
    return proxies.filter((proxy) => {
      const matchesSearch =
        proxy.host.toLowerCase().includes(search.toLowerCase()) ||
        proxy.port.toString() === search.trim()
      const matchesStatus =
        statusFilter === 'all' ||
        (statusFilter === 'healthy'
          ? proxy.is_active && proxy.failure_count === 0
          : !proxy.is_active || proxy.failure_count > 0)
      return matchesSearch && matchesStatus
    })
  }, [proxies, search, statusFilter])

  const summary = useMemo(() => {
    const total = proxies.length
    const healthy = proxies.filter((proxy) => proxy.is_active && proxy.failure_count === 0).length
    const degraded = proxies.filter((proxy) => proxy.failure_count > 0).length
    return { total, healthy, degraded }
  }, [proxies])

  const handleDelete = async (id: number) => {
    if (!window.confirm('Remove this proxy from the pool?')) return
    try {
      await proxiesAPI.delete(id)
      setFeedback({ type: 'success', message: 'Proxy removed' })
      loadProxies()
    } catch (error) {
      console.error('Failed to delete proxy:', error)
      setFeedback({ type: 'error', message: 'Failed to delete proxy' })
    }
  }

  const handleCreate = async (event: React.FormEvent) => {
    event.preventDefault()
    setSubmitting(true)
    try {
      await proxiesAPI.create({
        host: formState.host,
        port: Number(formState.port),
        type: formState.type,
        username: formState.username || undefined,
        password: formState.password || undefined,
      })
      setFeedback({ type: 'success', message: 'Proxy added' })
      setFormState({ host: '', port: '', type: 'http', username: '', password: '' })
      loadProxies()
    } catch (error) {
      console.error('Failed to add proxy', error)
      setFeedback({ type: 'error', message: 'Failed to add proxy' })
    } finally {
      setSubmitting(false)
    }
  }

  const handleTest = async (proxy: Proxy) => {
    setTestingId(proxy.id)
    try {
      const response = await proxiesAPI.test({
        id: proxy.id,
        host: proxy.host,
        port: proxy.port,
        type: proxy.type,
        username: proxy.username,
      })
      setFeedback({
        type: response.data.healthy ? 'success' : 'error',
        message: response.data.healthy ? 'Proxy is healthy' : 'Proxy test failed',
      })
      loadProxies()
    } catch (error) {
      console.error('Failed to test proxy', error)
      setFeedback({ type: 'error', message: 'Failed to test proxy' })
    } finally {
      setTestingId(null)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-semibold text-gray-900">Proxy Pool</h1>
        <p className="text-sm text-gray-500">
          Manage your rotating proxies. Healthy entries are automatically prioritized.
        </p>
      </div>

      {feedback && (
        <div
          className={`rounded-lg p-4 text-sm ${
            feedback.type === 'success'
              ? 'bg-green-50 text-green-700 border border-green-100'
              : 'bg-red-50 text-red-700 border border-red-100'
          }`}
        >
          <div className="flex items-center justify-between">
            <span>{feedback.message}</span>
            <button className="text-xs" onClick={() => setFeedback(null)}>
              Dismiss
            </button>
          </div>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <ProxySummary label="Total proxies" value={summary.total} />
        <ProxySummary label="Healthy" value={summary.healthy} tone="green" />
        <ProxySummary label="Needs attention" value={summary.degraded} tone="amber" />
      </div>

      <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
        <h2 className="text-lg font-semibold text-gray-900">Add proxy</h2>
        <form onSubmit={handleCreate} className="mt-4 grid gap-4 md:grid-cols-4">
          <input
            required
            placeholder="Host"
            value={formState.host}
            onChange={(e) => setFormState((prev) => ({ ...prev, host: e.target.value }))}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
          />
          <input
            required
            type="number"
            placeholder="Port"
            value={formState.port}
            onChange={(e) => setFormState((prev) => ({ ...prev, port: e.target.value }))}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
          />
          <select
            value={formState.type}
            onChange={(e) => setFormState((prev) => ({ ...prev, type: e.target.value }))}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
          >
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
            <option value="socks5">SOCKS5</option>
          </select>
          <div className="flex gap-2">
            <input
              placeholder="Username (optional)"
              value={formState.username}
              onChange={(e) => setFormState((prev) => ({ ...prev, username: e.target.value }))}
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
            <input
              placeholder="Password"
              type="password"
              value={formState.password}
              onChange={(e) => setFormState((prev) => ({ ...prev, password: e.target.value }))}
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
          </div>
          <div className="md:col-span-4">
            <button
              type="submit"
              disabled={submitting}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {submitting ? 'Adding…' : 'Add proxy'}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Current pool</h2>
            <p className="text-sm text-gray-500">Auto-rotation pulls from healthy entries.</p>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <input
              type="search"
              placeholder="Search host or port"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as typeof statusFilter)}
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            >
              <option value="all">All statuses</option>
              <option value="healthy">Healthy</option>
              <option value="unhealthy">Needs attention</option>
            </select>
          </div>
        </div>

        <div className="mt-4 overflow-hidden rounded-xl border border-gray-100">
          {loading ? (
            <div className="py-12 text-center text-gray-500">Loading proxies…</div>
          ) : filteredProxies.length === 0 ? (
            <div className="py-12 text-center text-gray-500">No proxies match the filters.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-100 text-sm">
                <thead className="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <tr>
                    <th className="px-4 py-3">Host</th>
                    <th className="px-4 py-3">Port</th>
                    <th className="px-4 py-3">Type</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3">Failures</th>
                    <th className="px-4 py-3">Last check</th>
                    <th className="px-4 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100 text-gray-700">
                  {filteredProxies.map((proxy) => (
                    <tr key={proxy.id}>
                      <td className="px-4 py-3 font-medium text-gray-900">{proxy.host}</td>
                      <td className="px-4 py-3">{proxy.port}</td>
                      <td className="px-4 py-3 uppercase text-xs">{proxy.type}</td>
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                            proxy.is_active && proxy.failure_count === 0
                              ? 'bg-green-100 text-green-700'
                              : 'bg-red-100 text-red-700'
                          }`}
                        >
                          {proxy.is_active && proxy.failure_count === 0 ? 'Healthy' : 'Unhealthy'}
                        </span>
                      </td>
                      <td className="px-4 py-3">{proxy.failure_count}</td>
                      <td className="px-4 py-3 text-xs text-gray-500">
                        {proxy.last_checked
                          ? new Date(proxy.last_checked).toLocaleString()
                          : '—'}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => handleTest(proxy)}
                            className="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                            disabled={testingId === proxy.id}
                          >
                            {testingId === proxy.id ? 'Testing…' : 'Test'}
                          </button>
                          <button
                            onClick={() => handleDelete(proxy.id)}
                            className="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-red-600 hover:bg-red-50"
                          >
                            Remove
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function ProxySummary({
  label,
  value,
  tone = 'default',
}: {
  label: string
  value: number
  tone?: 'default' | 'green' | 'amber'
}) {
  const toneStyles: Record<string, string> = {
    default: 'bg-white border-gray-100 text-gray-900',
    green: 'bg-green-50 border-green-100 text-green-800',
    amber: 'bg-yellow-50 border-yellow-100 text-yellow-800',
  }

  return (
    <div className={`rounded-2xl border p-4 shadow-sm ${toneStyles[tone]}`}>
      <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">{label}</p>
      <p className="mt-2 text-2xl font-semibold">{value}</p>
    </div>
  )
}
