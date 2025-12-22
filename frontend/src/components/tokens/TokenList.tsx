import { useEffect, useMemo, useState } from 'react'
import { authAPI } from '../../services/api'

interface Token {
  id: number
  token_name: string
  expires_at: string
  created_at: string
  last_used_at?: string
}

interface TokenListProps {
  autoOpenGenerator?: boolean
}

export default function TokenList({ autoOpenGenerator = false }: TokenListProps) {
  const [tokens, setTokens] = useState<Token[]>([])
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<'list' | 'create'>(autoOpenGenerator ? 'create' : 'list')
  const [search, setSearch] = useState('')
  const [formState, setFormState] = useState({ tokenName: '', expiresAt: '' })
  const [generatedToken, setGeneratedToken] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(
    null
  )

  useEffect(() => {
    loadTokens()
  }, [])

  const loadTokens = async () => {
    setLoading(true)
    try {
      const response = await authAPI.listTokens()
      setTokens(response.data.tokens || [])
    } catch (error) {
      console.error('Failed to load tokens:', error)
      setFeedback({ type: 'error', message: 'Failed to load tokens' })
    } finally {
      setLoading(false)
    }
  }

  const filteredTokens = useMemo(() => {
    return tokens.filter((token) =>
      token.token_name.toLowerCase().includes(search.trim().toLowerCase())
    )
  }, [tokens, search])

  const summary = useMemo(() => {
    const now = Date.now()
    const active = tokens.filter((token) => new Date(token.expires_at).getTime() > now).length
    return {
      total: tokens.length,
      active,
      expired: tokens.length - active,
    }
  }, [tokens])

  const handleDelete = async (id: number) => {
    if (!window.confirm('Revoke this token? Clients using it will lose access.')) return
    try {
      await authAPI.deleteToken(id)
      setFeedback({ type: 'success', message: 'Token revoked' })
      loadTokens()
    } catch (error) {
      console.error('Failed to delete token:', error)
      setFeedback({ type: 'error', message: 'Failed to revoke token' })
    }
  }

  const handleGenerate = async (event: React.FormEvent) => {
    event.preventDefault()
    setSubmitting(true)
    setGeneratedToken(null)
    try {
      const response = await authAPI.generateToken(formState.tokenName, formState.expiresAt)
      setGeneratedToken(response.data.token)
      setFormState({ tokenName: '', expiresAt: '' })
      loadTokens()
    } catch (error) {
      console.error('Failed to generate token', error)
      setFeedback({ type: 'error', message: 'Failed to generate token' })
    } finally {
      setSubmitting(false)
    }
  }

  const isExpired = (expiresAt: string) => new Date(expiresAt) < new Date()

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-gray-900">API Tokens</h1>
          <p className="text-sm text-gray-500">
            Issue scoped credentials for automation and service integrations.
          </p>
        </div>
        <div className="flex rounded-full border border-gray-200 bg-white p-1 shadow-sm">
          <button
            className={`rounded-full px-4 py-1.5 text-sm font-semibold ${
              tab === 'list' ? 'bg-indigo-600 text-white' : 'text-gray-600'
            }`}
            onClick={() => setTab('list')}
          >
            Manage tokens
          </button>
          <button
            className={`rounded-full px-4 py-1.5 text-sm font-semibold ${
              tab === 'create' ? 'bg-indigo-600 text-white' : 'text-gray-600'
            }`}
            onClick={() => setTab('create')}
          >
            Generate token
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

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <TokenSummary label="Total tokens" value={summary.total} />
        <TokenSummary label="Active" value={summary.active} tone="green" />
        <TokenSummary label="Expired" value={summary.expired} tone="red" />
      </div>

      {tab === 'list' ? (
        <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <p className="text-sm font-semibold text-gray-900">Active credentials</p>
              <p className="text-xs text-gray-500">
                Rotate tokens regularly to keep integrations secure.
              </p>
            </div>
            <input
              type="search"
              placeholder="Filter by token name..."
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200 md:w-64"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          <div className="mt-4 overflow-hidden rounded-xl border border-gray-100">
            {loading ? (
              <div className="py-12 text-center text-gray-500">Loading tokens…</div>
            ) : filteredTokens.length === 0 ? (
              <div className="py-12 text-center text-gray-500">
                No tokens found. Try a different search.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-100 text-sm">
                  <thead className="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    <tr>
                      <th className="px-4 py-3">Name</th>
                      <th className="px-4 py-3">Expires</th>
                      <th className="px-4 py-3">Last used</th>
                      <th className="px-4 py-3">Status</th>
                      <th className="px-4 py-3 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100 text-gray-700">
                    {filteredTokens.map((token) => (
                      <tr key={token.id}>
                        <td className="px-4 py-3 font-medium text-gray-900">{token.token_name}</td>
                        <td className="px-4 py-3">
                          {new Date(token.expires_at).toLocaleString()}
                        </td>
                        <td className="px-4 py-3">
                          {token.last_used_at
                            ? new Date(token.last_used_at).toLocaleString()
                            : 'Never'}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                              isExpired(token.expires_at)
                                ? 'bg-red-100 text-red-700'
                                : 'bg-green-100 text-green-700'
                            }`}
                          >
                            {isExpired(token.expires_at) ? 'Expired' : 'Active'}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <button
                            onClick={() => handleDelete(token.id)}
                            className="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50"
                          >
                            Revoke
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Generate one-time token</h2>
            <p className="text-sm text-gray-500">
              Tokens are shown only once. Store securely in your secret manager.
            </p>
          </div>
          {generatedToken && (
            <div className="mt-4 rounded-xl border border-yellow-200 bg-yellow-50 p-4 text-sm text-yellow-800">
              <p className="font-semibold">Copy now — this value will disappear if you leave.</p>
              <div className="mt-2 rounded-lg bg-white p-3 font-mono text-xs text-gray-900">
                {generatedToken}
              </div>
              <div className="mt-3 flex gap-2">
                <button
                  onClick={() => navigator.clipboard.writeText(generatedToken)}
                  className="rounded-lg border border-yellow-300 px-3 py-1 text-xs font-semibold text-yellow-900 hover:bg-white"
                >
                  Copy token
                </button>
                <button
                  onClick={() => setGeneratedToken(null)}
                  className="rounded-lg border border-yellow-300 px-3 py-1 text-xs font-semibold text-yellow-900 hover:bg-white"
                >
                  Hide token
                </button>
              </div>
            </div>
          )}
          <form onSubmit={handleGenerate} className="mt-5 space-y-4">
            <div>
              <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Token name
              </label>
              <input
                required
                value={formState.tokenName}
                onChange={(e) => setFormState((prev) => ({ ...prev, tokenName: e.target.value }))}
                className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
                placeholder="e.g. production-webhook"
              />
            </div>
            <div>
              <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Expiration
              </label>
              <input
                required
                type="datetime-local"
                value={formState.expiresAt}
                onChange={(e) => setFormState((prev) => ({ ...prev, expiresAt: e.target.value }))}
                className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
              />
            </div>
            <button
              type="submit"
              disabled={submitting}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {submitting ? 'Generating…' : 'Generate access token'}
            </button>
          </form>
        </div>
      )}
    </div>
  )
}

interface TokenSummaryProps {
  label: string
  value: number
  tone?: 'default' | 'green' | 'red'
}

function TokenSummary({ label, value, tone = 'default' }: TokenSummaryProps) {
  const toneStyles: Record<string, string> = {
    default: 'bg-white border-gray-100 text-gray-900',
    green: 'bg-green-50 border-green-100 text-green-800',
    red: 'bg-red-50 border-red-100 text-red-800',
  }

  return (
    <div className={`rounded-2xl border p-4 shadow-sm ${toneStyles[tone]}`}>
      <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">{label}</p>
      <p className="mt-2 text-2xl font-semibold">{value}</p>
    </div>
  )
}
