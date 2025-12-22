import { useEffect, useState } from 'react'

export interface TaskFormValues {
  name: string
  url: string
  type: string
  config: string
}

interface TaskFormModalProps {
  open: boolean
  mode: 'create' | 'edit'
  initialValues?: TaskFormValues
  submitting: boolean
  onSubmit: (values: TaskFormValues) => Promise<void> | void
  onClose: () => void
}

const typeOptions = [
  { value: 'api', label: 'API crawler' },
  { value: 'browser', label: 'Browser crawler' },
]

export default function TaskFormModal({
  open,
  mode,
  initialValues,
  submitting,
  onSubmit,
  onClose,
}: TaskFormModalProps) {
  const [values, setValues] = useState<TaskFormValues>(
    initialValues ?? {
      name: '',
      url: '',
      type: 'api',
      config: JSON.stringify({ crawler_type: 'embroidery_api' }, null, 2),
    }
  )
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (initialValues) {
      setValues(initialValues)
    } else {
      setValues({
        name: '',
        url: '',
        type: 'api',
        config: JSON.stringify({ crawler_type: 'embroidery_api' }, null, 2),
      })
    }
  }, [initialValues])

  if (!open) {
    return null
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    setError(null)
    try {
      JSON.parse(values.config || '{}')
    } catch {
      setError('Config must be valid JSON')
      return
    }

    await onSubmit(values)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/40 px-4">
      <div className="w-full max-w-2xl rounded-2xl bg-white shadow-2xl">
        <div className="flex items-center justify-between border-b border-gray-100 px-6 py-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">Tasks</p>
            <h2 className="text-xl font-semibold text-gray-900">
              {mode === 'create' ? 'Create new task' : 'Edit task'}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="rounded-full border border-gray-200 px-3 py-1 text-sm text-gray-600 hover:bg-gray-50"
          >
            Close
          </button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-6">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Name
              </label>
              <input
                required
                type="text"
                value={values.name}
                onChange={(e) => setValues((prev) => ({ ...prev, name: e.target.value }))}
                className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              />
            </div>
            <div>
              <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Type
              </label>
              <select
                value={values.type}
                onChange={(e) => setValues((prev) => ({ ...prev, type: e.target.value }))}
                className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              >
                {typeOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Target URL
            </label>
            <input
              required
              type="url"
              value={values.url}
              onChange={(e) => setValues((prev) => ({ ...prev, url: e.target.value }))}
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              placeholder="https://www.embroiderydesigns.com/es/prdsrch"
            />
          </div>

          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Config (JSON)
            </label>
            <textarea
              rows={8}
              value={values.config}
              onChange={(e) => setValues((prev) => ({ ...prev, config: e.target.value }))}
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 font-mono text-xs focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
            {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
          </div>

          <div className="flex flex-wrap items-center justify-between gap-2 pt-2">
            <button
              type="button"
              onClick={() =>
                setValues((prev) => ({
                  ...prev,
                  config: JSON.stringify({ crawler_type: 'embroidery_api' }, null, 2),
                }))
              }
              className="text-xs font-semibold text-gray-500 hover:text-gray-700"
            >
              Reset to embroidery preset
            </button>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={submitting}
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
              >
                {submitting ? 'Saving…' : mode === 'create' ? 'Create task' : 'Save changes'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  )
}

