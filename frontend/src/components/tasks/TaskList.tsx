import { useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { tasksAPI } from '../../services/api'
import TaskFormModal, { TaskFormValues } from './TaskFormModal'

interface Task {
  id: number
  name: string
  url: string
  type: string
  status: string
  config: string
  created_at: string
  updated_at: string
}

type ModalState =
  | null
  | {
      mode: 'create' | 'edit'
      task?: Task
    }

export default function TaskList() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [processing, setProcessing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(
    null
  )
  const [filters, setFilters] = useState({ search: '', status: 'all', type: 'all' })
  const [selectedTaskIds, setSelectedTaskIds] = useState<number[]>([])
  const [modalState, setModalState] = useState<ModalState>(null)
  const [pollingEnabled, setPollingEnabled] = useState(true)
  const [searchParams, setSearchParams] = useSearchParams()

  useEffect(() => {
    loadTasks()
  }, [])

  useEffect(() => {
    const shouldOpenModal = searchParams.get('new') === '1'
    if (shouldOpenModal) {
      setModalState({ mode: 'create' })
      searchParams.delete('new')
      setSearchParams(searchParams, { replace: true })
    }
  }, [searchParams, setSearchParams])

  useEffect(() => {
    if (!pollingEnabled) return
    const interval = setInterval(() => loadTasks(false), 15000)
    return () => clearInterval(interval)
  }, [pollingEnabled])

  const loadTasks = async (showSpinner = true) => {
    if (showSpinner) {
      setLoading(true)
    }
    try {
      const response = await tasksAPI.list(100, 0)
      setTasks(response.data.tasks || [])
    } catch (error) {
      console.error('Failed to load tasks:', error)
      setFeedback({ type: 'error', message: 'Failed to load tasks' })
    } finally {
      if (showSpinner) {
        setLoading(false)
      }
    }
  }

  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      const matchesStatus = filters.status === 'all' || task.status === filters.status
      const matchesType = filters.type === 'all' || task.type === filters.type
      const matchesSearch =
        filters.search === '' ||
        task.name.toLowerCase().includes(filters.search.toLowerCase()) ||
        task.url.toLowerCase().includes(filters.search.toLowerCase()) ||
        task.id.toString() === filters.search

      return matchesStatus && matchesType && matchesSearch
    })
  }, [tasks, filters])

  const statusSummary = useMemo(() => {
    return tasks.reduce<Record<string, number>>((acc, task) => {
      acc[task.status] = (acc[task.status] || 0) + 1
      return acc
    }, {})
  }, [tasks])

  const toggleTaskSelection = (taskId: number) => {
    setSelectedTaskIds((prev) =>
      prev.includes(taskId) ? prev.filter((id) => id !== taskId) : [...prev, taskId]
    )
  }

  const toggleSelectAll = () => {
    if (selectedTaskIds.length === filteredTasks.length) {
      setSelectedTaskIds([])
    } else {
      setSelectedTaskIds(filteredTasks.map((task) => task.id))
    }
  }

  const handleBulkAction = async (action: 'start' | 'stop' | 'pause') => {
    if (selectedTaskIds.length === 0) return
    setProcessing(true)
    try {
      await Promise.all(
        selectedTaskIds.map((id) => {
          if (action === 'start') return tasksAPI.start(id)
          if (action === 'stop') return tasksAPI.stop(id)
          return tasksAPI.pause(id)
        })
      )
      setFeedback({ type: 'success', message: `Bulk ${action} command dispatched` })
      setSelectedTaskIds([])
      loadTasks(false)
    } catch (error) {
      console.error(`Failed to ${action} tasks`, error)
      setFeedback({ type: 'error', message: `Failed to ${action} selected tasks` })
    } finally {
      setProcessing(false)
    }
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      running: 'bg-green-100 text-green-800',
      completed: 'bg-blue-100 text-blue-800',
      failed: 'bg-red-100 text-red-800',
      pending: 'bg-yellow-100 text-yellow-800',
      paused: 'bg-gray-100 text-gray-800',
      stopped: 'bg-gray-100 text-gray-800',
    }
    return `${colors[status] || 'bg-gray-100 text-gray-800'}`
  }

  const handleSingleAction = async (id: number, action: 'start' | 'stop' | 'pause') => {
    setProcessing(true)
    try {
      if (action === 'start') await tasksAPI.start(id)
      if (action === 'stop') await tasksAPI.stop(id)
      if (action === 'pause') await tasksAPI.pause(id)
      setFeedback({ type: 'success', message: `Task ${action}ed successfully` })
      loadTasks(false)
    } catch (error) {
      console.error(`Failed to ${action} task`, error)
      setFeedback({ type: 'error', message: `Failed to ${action} task` })
    } finally {
      setProcessing(false)
    }
  }

  const handleModalSubmit = async (values: TaskFormValues) => {
    setSaving(true)
    try {
      if (modalState?.mode === 'edit' && modalState.task) {
        await tasksAPI.update(modalState.task.id, values)
        setFeedback({ type: 'success', message: 'Task updated successfully' })
      } else {
        await tasksAPI.create(values)
        setFeedback({ type: 'success', message: 'Task created successfully' })
      }
      setModalState(null)
      loadTasks(false)
    } catch (error) {
      console.error('Failed to save task', error)
      setFeedback({ type: 'error', message: 'Failed to save task' })
    } finally {
      setSaving(false)
    }
  }

  const modalInitialValues: TaskFormValues | undefined = modalState?.task
    ? {
        name: modalState.task.name,
        url: modalState.task.url,
        type: modalState.task.type,
        config: modalState.task.config || '{}',
      }
    : undefined

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-gray-900">Tasks</h1>
          <p className="text-sm text-gray-500">
            Monitor, orchestrate, and configure crawler executions.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => setPollingEnabled((prev) => !prev)}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            {pollingEnabled ? 'Pause auto-refresh' : 'Resume auto-refresh'}
          </button>
          <button
            onClick={() => loadTasks()}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Refresh
          </button>
          <button
            onClick={() => setModalState({ mode: 'create' })}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
          >
            + New Task
          </button>
        </div>
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

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <SummaryCard label="Total" value={tasks.length} />
        <SummaryCard label="Running" value={statusSummary['running'] || 0} tone="green" />
        <SummaryCard label="Pending" value={statusSummary['pending'] || 0} tone="amber" />
        <SummaryCard label="Failed" value={statusSummary['failed'] || 0} tone="red" />
      </div>

      <div className="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <label className="text-xs font-semibold text-gray-500 uppercase">Search</label>
            <input
              type="text"
              placeholder="Filter by id, name, url..."
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              value={filters.search}
              onChange={(e) => setFilters((prev) => ({ ...prev, search: e.target.value }))}
            />
          </div>
          <div>
            <label className="text-xs font-semibold text-gray-500 uppercase">Status</label>
            <select
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              value={filters.status}
              onChange={(e) => setFilters((prev) => ({ ...prev, status: e.target.value }))}
            >
              <option value="all">All statuses</option>
              <option value="pending">Pending</option>
              <option value="running">Running</option>
              <option value="paused">Paused</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="stopped">Stopped</option>
            </select>
          </div>
          <div>
            <label className="text-xs font-semibold text-gray-500 uppercase">Type</label>
            <select
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              value={filters.type}
              onChange={(e) => setFilters((prev) => ({ ...prev, type: e.target.value }))}
            >
              <option value="all">All types</option>
              <option value="api">API crawler</option>
              <option value="browser">Browser crawler</option>
            </select>
          </div>
        </div>

        {selectedTaskIds.length > 0 && (
          <div className="mt-4 flex flex-wrap items-center gap-2 rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-600">
            <span className="font-medium">{selectedTaskIds.length} selected</span>
            <button
              disabled={processing}
              onClick={() => handleBulkAction('start')}
              className="rounded border border-gray-200 px-2 py-1 text-xs font-semibold text-gray-700 hover:bg-white disabled:opacity-50"
            >
              Start
            </button>
            <button
              disabled={processing}
              onClick={() => handleBulkAction('pause')}
              className="rounded border border-gray-200 px-2 py-1 text-xs font-semibold text-gray-700 hover:bg-white disabled:opacity-50"
            >
              Pause
            </button>
            <button
              disabled={processing}
              onClick={() => handleBulkAction('stop')}
              className="rounded border border-gray-200 px-2 py-1 text-xs font-semibold text-gray-700 hover:bg-white disabled:opacity-50"
            >
              Stop
            </button>
            <button
              onClick={() => setSelectedTaskIds([])}
              className="ml-auto text-xs text-gray-500 hover:text-gray-700"
            >
              Clear selection
            </button>
          </div>
        )}
      </div>

      <div className="overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-sm">
        {loading ? (
          <div className="py-16 text-center text-gray-500">Loading tasks…</div>
        ) : filteredTasks.length === 0 ? (
          <div className="py-16 text-center text-gray-500">No tasks match the current filters.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-100 text-sm">
              <thead className="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                <tr>
                  <th className="px-4 py-3">
                    <input
                      type="checkbox"
                      checked={selectedTaskIds.length > 0 && selectedTaskIds.length === filteredTasks.length}
                      onChange={toggleSelectAll}
                      className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    />
                  </th>
                  <th className="px-4 py-3">Task</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Type</th>
                  <th className="px-4 py-3">URL</th>
                  <th className="px-4 py-3">Updated</th>
                  <th className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filteredTasks.map((task) => {
                  const isSelected = selectedTaskIds.includes(task.id)
                  return (
                    <tr key={task.id} className={isSelected ? 'bg-indigo-50/30' : undefined}>
                      <td className="px-4 py-3">
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => toggleTaskSelection(task.id)}
                          className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                        />
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex flex-col">
                          <Link
                            to={`/tasks/${task.id}`}
                            className="font-semibold text-gray-900 hover:text-indigo-600"
                          >
                            {task.name}
                          </Link>
                          <span className="text-xs text-gray-500">#{task.id}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold capitalize ${getStatusColor(
                            task.status
                          )}`}
                        >
                          {task.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 capitalize text-gray-700">{task.type}</td>
                      <td className="px-4 py-3 max-w-xs">
                        <span className="truncate text-xs text-gray-500">{task.url}</span>
                      </td>
                      <td className="px-4 py-3 text-gray-600">
                        {new Date(task.updated_at || task.created_at).toLocaleString()}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-2">
                          {(task.status === 'pending' ||
                            task.status === 'paused' ||
                            task.status === 'failed' ||
                            task.status === 'stopped') && (
                            <button
                              disabled={processing}
                              onClick={() => handleSingleAction(task.id, 'start')}
                              className="rounded-lg border border-gray-200 px-2 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                            >
                              Start
                            </button>
                          )}
                          {task.status === 'running' && (
                            <>
                              <button
                                disabled={processing}
                                onClick={() => handleSingleAction(task.id, 'pause')}
                                className="rounded-lg border border-gray-200 px-2 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                              >
                                Pause
                              </button>
                              <button
                                disabled={processing}
                                onClick={() => handleSingleAction(task.id, 'stop')}
                                className="rounded-lg border border-gray-200 px-2 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                              >
                                Stop
                              </button>
                            </>
                          )}
                          <button
                            onClick={() => setModalState({ mode: 'edit', task })}
                            className="rounded-lg border border-gray-200 px-2 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50"
                          >
                            Edit
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <TaskFormModal
        open={modalState !== null}
        mode={modalState?.mode ?? 'create'}
        onClose={() => setModalState(null)}
        initialValues={modalInitialValues}
        submitting={saving}
        onSubmit={handleModalSubmit}
      />
    </div>
  )
}

interface SummaryCardProps {
  label: string
  value: number
  tone?: 'default' | 'green' | 'amber' | 'red'
}

const toneStyles: Record<string, { ring: string; text: string }> = {
  default: { ring: 'ring-gray-100', text: 'text-gray-900' },
  green: { ring: 'ring-green-100', text: 'text-green-700' },
  amber: { ring: 'ring-yellow-100', text: 'text-yellow-700' },
  red: { ring: 'ring-red-100', text: 'text-red-700' },
}

function SummaryCard({ label, value, tone = 'default' }: SummaryCardProps) {
  const styles = toneStyles[tone] ?? toneStyles.default
  return (
    <div className={`rounded-2xl border border-gray-100 bg-white p-4 shadow-sm ring-1 ${styles.ring}`}>
      <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">{label}</p>
      <p className={`mt-2 text-2xl font-semibold ${styles.text}`}>{value}</p>
    </div>
  )
}

