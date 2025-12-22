import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import TaskFormModal, { TaskFormValues } from './TaskFormModal'
import { tasksAPI } from '../../services/api'

interface Task {
  id: number
  name: string
  url: string
  type: string
  status: string
  config: string
  created_at: string
  updated_at: string
  started_at?: string
  completed_at?: string
}

interface CrawlResult {
  id: number
  url: string
  method: string
  status_code: number
  response_time: number
  created_at: string
}

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>()
  const [task, setTask] = useState<Task | null>(null)
  const [results, setResults] = useState<CrawlResult[]>([])
  const [loadingTask, setLoadingTask] = useState(true)
  const [loadingResults, setLoadingResults] = useState(true)
  const [processing, setProcessing] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(
    null
  )

  useEffect(() => {
    if (!id) return
    loadTask()
    loadResults()
  }, [id])

  useEffect(() => {
    if (!id) return
    const interval = setInterval(() => {
      loadTask(false)
      loadResults(false)
    }, 15000)
    return () => clearInterval(interval)
  }, [id])

  const loadTask = async (showSpinner = true) => {
    if (!id) return
    if (showSpinner) setLoadingTask(true)
    try {
      const response = await tasksAPI.get(parseInt(id, 10))
      setTask(response.data)
    } catch (error) {
      console.error('Failed to load task', error)
      setFeedback({ type: 'error', message: 'Failed to load task' })
    } finally {
      if (showSpinner) setLoadingTask(false)
    }
  }

  const loadResults = async (showSpinner = true) => {
    if (!id) return
    if (showSpinner) setLoadingResults(true)
    try {
      const response = await tasksAPI.getResults(parseInt(id, 10), 5, 0)
      setResults(response.data.results || [])
    } catch (error) {
      console.error('Failed to load results', error)
    } finally {
      if (showSpinner) setLoadingResults(false)
    }
  }

  const handleAction = async (action: 'start' | 'stop' | 'pause') => {
    if (!id) return
    setProcessing(true)
    try {
      if (action === 'start') await tasksAPI.start(parseInt(id, 10))
      if (action === 'stop') await tasksAPI.stop(parseInt(id, 10))
      if (action === 'pause') await tasksAPI.pause(parseInt(id, 10))
      setFeedback({ type: 'success', message: `Task ${action} command queued` })
      loadTask(false)
      loadResults(false)
    } catch (error) {
      console.error(`Failed to ${action} task`, error)
      setFeedback({ type: 'error', message: `Failed to ${action} task` })
    } finally {
      setProcessing(false)
    }
  }

  const handleModalSubmit = async (values: TaskFormValues) => {
    if (!task) return
    setSaving(true)
    try {
      await tasksAPI.update(task.id, values)
      setFeedback({ type: 'success', message: 'Task updated successfully' })
      setModalOpen(false)
      loadTask(false)
    } catch (error) {
      console.error('Failed to update task', error)
      setFeedback({ type: 'error', message: 'Failed to update task' })
    } finally {
      setSaving(false)
    }
  }

  const formattedConfig = useMemo(() => {
    try {
      return JSON.stringify(JSON.parse(task?.config || '{}'), null, 2)
    } catch {
      return task?.config || '{}'
    }
  }, [task])

  if (loadingTask) {
    return <div className="py-12 text-center text-gray-500">Loading task…</div>
  }

  if (!task) {
    return <div className="py-12 text-center text-gray-500">Task not found</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-xs uppercase tracking-wider text-gray-500">Task #{task.id}</p>
          <h1 className="text-3xl font-semibold text-gray-900">{task.name}</h1>
          <div className="mt-2 inline-flex items-center rounded-full bg-gray-100 px-3 py-1 text-xs font-semibold capitalize text-gray-700">
            {task.type} crawler
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          {(task.status === 'pending' ||
            task.status === 'paused' ||
            task.status === 'failed' ||
            task.status === 'stopped') && (
            <button
              onClick={() => handleAction('start')}
              disabled={processing}
              className="rounded-lg border border-gray-200 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              Start
            </button>
          )}
          {task.status === 'running' && (
            <>
              <button
                onClick={() => handleAction('pause')}
                disabled={processing}
                className="rounded-lg border border-gray-200 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
              >
                Pause
              </button>
              <button
                onClick={() => handleAction('stop')}
                disabled={processing}
                className="rounded-lg border border-gray-200 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
              >
                Stop
              </button>
            </>
          )}
          <button
            onClick={() => setModalOpen(true)}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
          >
            Edit Task
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

      <div className="grid gap-4 md:grid-cols-4">
        <Metric label="Status" value={task.status} />
        <Metric
          label="Created"
          value={new Date(task.created_at).toLocaleString()}
        />
        <Metric
          label="Started"
          value={task.started_at ? new Date(task.started_at).toLocaleString() : '—'}
        />
        <Metric
          label="Completed"
          value={task.completed_at ? new Date(task.completed_at).toLocaleString() : '—'}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <section className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Configuration</h2>
            <button
              onClick={() => navigator.clipboard.writeText(formattedConfig)}
              className="text-xs font-semibold text-gray-500 hover:text-gray-700"
            >
              Copy JSON
            </button>
          </div>
          <pre className="mt-4 max-h-64 overflow-auto rounded-xl bg-slate-900 p-4 text-xs text-slate-100">
            {formattedConfig}
          </pre>
        </section>

        <section className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Details</h2>
            <span className="text-xs text-gray-500">
              Last update {new Date(task.updated_at).toLocaleTimeString()}
            </span>
          </div>
          <dl className="mt-4 space-y-3 text-sm text-gray-700">
            <div>
              <dt className="font-medium text-gray-500">Target URL</dt>
              <dd className="mt-1 break-all text-gray-900">{task.url}</dd>
            </div>
            <div>
              <dt className="font-medium text-gray-500">Execution window</dt>
              <dd className="mt-1">
                {task.started_at
                  ? `${new Date(task.started_at).toLocaleString()}`
                  : 'Not started yet'}
              </dd>
            </div>
            <div>
              <dt className="font-medium text-gray-500">Duration</dt>
              <dd className="mt-1">
                {task.started_at && task.completed_at
                  ? `${Math.round(
                      (new Date(task.completed_at).getTime() -
                        new Date(task.started_at).getTime()) /
                        1000
                    )}s`
                  : task.started_at
                  ? `${Math.round(
                      (Date.now() - new Date(task.started_at).getTime()) / 1000
                    )}s elapsed`
                  : '—'}
              </dd>
            </div>
          </dl>
        </section>
      </div>

      <section className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Recent crawl results</h2>
            <p className="text-sm text-gray-500">Newest responses from this task</p>
          </div>
          <button
            onClick={() => loadResults()}
            className="rounded-lg border border-gray-200 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Refresh
          </button>
        </div>
        {loadingResults ? (
          <div className="py-8 text-center text-gray-500">Loading results…</div>
        ) : results.length === 0 ? (
          <div className="py-8 text-center text-gray-500">No results captured yet.</div>
        ) : (
          <div className="mt-4 overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-100 text-sm">
              <thead className="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                <tr>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Method</th>
                  <th className="px-4 py-3">URL</th>
                  <th className="px-4 py-3">Latency</th>
                  <th className="px-4 py-3">Timestamp</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 text-gray-700">
                {results.map((result) => (
                  <tr key={result.id}>
                    <td className="px-4 py-3 font-mono text-sm">{result.status_code}</td>
                    <td className="px-4 py-3 uppercase text-xs">{result.method}</td>
                    <td className="px-4 py-3 max-w-md">
                      <span className="truncate text-xs text-gray-500">{result.url}</span>
                    </td>
                    <td className="px-4 py-3">{result.response_time} ms</td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      {new Date(result.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <TaskFormModal
        open={modalOpen}
        mode="edit"
        initialValues={{
          name: task.name,
          url: task.url,
          type: task.type,
          config: formattedConfig,
        }}
        submitting={saving}
        onSubmit={handleModalSubmit}
        onClose={() => setModalOpen(false)}
      />
    </div>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
      <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">{label}</p>
      <p className="mt-2 text-sm text-gray-900">{value}</p>
    </div>
  )
}

