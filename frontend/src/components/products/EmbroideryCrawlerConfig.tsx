import { useEffect, useState } from 'react'
import { crawlerAPI } from '../../services/api'

interface CrawlConfigResponse {
  payload_overrides: Record<string, unknown>
  default_payload: Record<string, unknown>
  effective_payload: Record<string, unknown>
  updated_at?: string
}

type Status =
  | { type: 'success'; message: string }
  | { type: 'error'; message: string }
  | null

const stringify = (value: Record<string, unknown> | undefined) =>
  JSON.stringify(value ?? {}, null, 2)

const formatTimestamp = (value?: string) => {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

export default function EmbroideryCrawlerConfig() {
  const [config, setConfig] = useState<CrawlConfigResponse | null>(null)
  const [overridesDraft, setOverridesDraft] = useState('{}')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [status, setStatus] = useState<Status>(null)

  useEffect(() => {
    refreshConfig()
  }, [])

  const refreshConfig = async () => {
    setLoading(true)
    setStatus(null)
    try {
      const { data } = await crawlerAPI.getEmbroideryConfig()
      setConfig(data)
      setOverridesDraft(stringify(data.payload_overrides))
    } catch (error: any) {
      setStatus({
        type: 'error',
        message: error?.response?.data?.error || 'خطا در دریافت تنظیمات',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    let parsedOverrides: Record<string, unknown> = {}

    try {
      const trimmed = overridesDraft.trim()
      parsedOverrides = trimmed ? JSON.parse(trimmed) : {}
    } catch (error) {
      setStatus({ type: 'error', message: 'JSON نامعتبر است. لطفاً مقدار payload_overrides را بررسی کنید.' })
      return
    }

    setSaving(true)
    setStatus(null)
    try {
      const { data } = await crawlerAPI.updateEmbroideryConfig(parsedOverrides)
      setConfig(data)
      setOverridesDraft(stringify(data.payload_overrides))
      setStatus({ type: 'success', message: 'تنظیمات با موفقیت ذخیره شد.' })
    } catch (error: any) {
      setStatus({
        type: 'error',
        message: error?.response?.data?.error || 'ذخیره تنظیمات با خطا مواجه شد.',
      })
    } finally {
      setSaving(false)
    }
  }

  const handleReset = () => {
    setOverridesDraft('{}')
    setStatus(null)
  }

  return (
    <div className="space-y-6">
      <div className="bg-white shadow-sm rounded-2xl border border-gray-100">
        <div className="px-6 py-5 border-b border-gray-100 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-gray-900">Crawler Config</h1>
            <p className="text-sm text-gray-500 mt-1">
              JSON overrides برای بدنه درخواست Elasticsearch. این مقادیر روی payload پیش‌فرض اعمال می‌شوند.
            </p>
          </div>
          <div className="text-sm text-gray-500">
            آخرین بروزرسانی:{' '}
            <span className="font-medium text-gray-700">{formatTimestamp(config?.updated_at)}</span>
          </div>
        </div>

        <div className="p-6 space-y-4">
          {status && (
            <div
              className={`rounded-xl px-4 py-3 text-sm ${
                status.type === 'success'
                  ? 'bg-green-50 text-green-700 border border-green-100'
                  : 'bg-red-50 text-red-700 border border-red-100'
              }`}
            >
              {status.message}
            </div>
          )}

          {loading ? (
            <div className="text-center py-12 text-gray-500">در حال بارگذاری...</div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <label className="block text-sm font-medium text-gray-700">
                payload_overrides
                <textarea
                  className="mt-2 w-full rounded-xl border border-gray-200 font-mono text-sm p-4 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 min-h-[280px]"
                  value={overridesDraft}
                  onChange={(event) => setOverridesDraft(event.target.value)}
                  spellCheck={false}
                />
              </label>

              <div className="flex flex-wrap gap-3">
                <button
                  type="submit"
                  className="inline-flex items-center justify-center px-4 py-2 rounded-xl bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-500 disabled:opacity-60 disabled:cursor-not-allowed"
                  disabled={saving}
                >
                  {saving ? 'در حال ذخیره...' : 'ذخیره تغییرات'}
                </button>
                <button
                  type="button"
                  className="inline-flex items-center justify-center px-4 py-2 rounded-xl border border-gray-200 text-sm font-medium text-gray-700 hover:bg-gray-50"
                  onClick={handleReset}
                >
                  پاک کردن و بازگشت به پیش‌فرض
                </button>
                <button
                  type="button"
                  className="inline-flex items-center justify-center px-4 py-2 rounded-xl border border-gray-200 text-sm font-medium text-gray-700 hover:bg-gray-50"
                  onClick={refreshConfig}
                >
                  بارگذاری مجدد
                </button>
              </div>
            </form>
          )}
        </div>
      </div>

      {!loading && config && (
        <div className="grid gap-6 md:grid-cols-2">
          <PayloadCard title="Payload پایه (سیستمی)" data={config.default_payload} />
          <PayloadCard title="Payload نهایی (بعد از اعمال overrides)" data={config.effective_payload} />
        </div>
      )}
    </div>
  )
}

function PayloadCard({ title, data }: { title: string; data: Record<string, unknown> }) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden">
      <div className="px-5 py-4 border-b border-gray-100">
        <p className="text-sm font-semibold text-gray-900">{title}</p>
      </div>
      <pre className="p-5 text-xs bg-gray-50 text-gray-800 overflow-auto max-h-[450px]">
        {JSON.stringify(data ?? {}, null, 2)}
      </pre>
    </div>
  )
}

