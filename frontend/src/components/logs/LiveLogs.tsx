import { useEffect, useState, useRef } from 'react'
import { wsService } from '../../services/websocket'

interface LogMessage {
  type: string
  task_id?: number
  level?: string
  message?: string
  timestamp?: string
}

export default function LiveLogs() {
  const [logs, setLogs] = useState<LogMessage[]>([])
  const [taskID, setTaskID] = useState<string>('')
  const [connected, setConnected] = useState(false)
  const [levelFilter, setLevelFilter] = useState<'all' | 'info' | 'warn' | 'error'>('all')
  const [autoScroll, setAutoScroll] = useState(true)
  const [bufferSize, setBufferSize] = useState(500)
  const logsEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleConnected = () => setConnected(true)
    const handleDisconnected = () => setConnected(false)
    const handleIncoming = (message: LogMessage) => {
      setLogs((prev) => {
        const entry = { ...message, timestamp: new Date().toISOString() }
        const next = [...prev, entry]
        if (next.length > bufferSize) {
          return next.slice(next.length - bufferSize)
        }
        return next
      })
    }

    wsService.on('connected', handleConnected)
    wsService.on('disconnected', handleDisconnected)
    wsService.on('log', handleIncoming)
    wsService.on('task_status', handleIncoming)

    return () => {
      wsService.off('connected', handleConnected)
      wsService.off('disconnected', handleDisconnected)
      wsService.off('log', handleIncoming)
      wsService.off('task_status', handleIncoming)
      wsService.disconnect(true)
    }
  }, [bufferSize])

  useEffect(() => {
    if (autoScroll) {
      logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [logs, autoScroll])

  const filteredLogs = logs.filter((log) =>
    levelFilter === 'all' ? true : log.level === levelFilter
  )

  const handleConnect = () => {
    wsService.disconnect()
    const parsedID = taskID ? parseInt(taskID, 10) : NaN
    const id = Number.isNaN(parsedID) ? undefined : parsedID
    wsService.connect(id)
  }

  const handleDisconnect = () => {
    wsService.disconnect()
    setConnected(false)
  }

  const handleClear = () => setLogs([])

  const downloadLogs = () => {
    const payload = filteredLogs
      .map(
        (log) =>
          `${log.timestamp ?? ''} task=${log.task_id ?? '-'} level=${log.level ?? 'info'} ${
            log.message ?? ''
          }`
      )
      .join('\n')
    const blob = new Blob([payload], { type: 'text/plain' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = 'logs.txt'
    link.click()
    URL.revokeObjectURL(link.href)
  }

  const getLevelColor = (level?: string) => {
    switch (level) {
      case 'error':
        return 'text-red-600'
      case 'warn':
        return 'text-yellow-600'
      case 'info':
        return 'text-blue-600'
      default:
        return 'text-gray-600'
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-gray-900">Live Logs</h1>
          <p className="text-sm text-gray-500">Monitor crawler output in real-time.</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <input
            type="number"
            placeholder="Task filter"
            value={taskID}
            onChange={(e) => setTaskID(e.target.value)}
            className="w-32 rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
          />
          <select
            value={levelFilter}
            onChange={(e) => setLevelFilter(e.target.value as typeof levelFilter)}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
          >
            <option value="all">All levels</option>
            <option value="info">Info</option>
            <option value="warn">Warn</option>
            <option value="error">Error</option>
          </select>
          <select
            value={bufferSize}
            onChange={(e) => setBufferSize(Number(e.target.value))}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none"
          >
            {[200, 500, 1000].map((size) => (
              <option key={size} value={size}>
                {size} lines
              </option>
            ))}
          </select>
          {connected ? (
            <button
              onClick={handleDisconnect}
              className="rounded-lg border border-red-200 px-3 py-2 text-sm font-semibold text-red-700 hover:bg-red-50"
            >
              Disconnect
            </button>
          ) : (
            <button
              onClick={handleConnect}
              className="rounded-lg border border-green-200 px-3 py-2 text-sm font-semibold text-green-700 hover:bg-green-50"
            >
              Connect
            </button>
          )}
          <button
            onClick={handleClear}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
          >
            Clear
          </button>
          <button
            onClick={downloadLogs}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
          >
            Export
          </button>
          <div
            className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold ${
              connected ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
            }`}
          >
            {connected ? 'Connected' : 'Disconnected'}
          </div>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3 text-xs text-gray-500">
        <label className="flex items-center gap-1">
          <input
            type="checkbox"
            checked={autoScroll}
            onChange={(e) => setAutoScroll(e.target.checked)}
          />
          Auto-scroll
        </label>
        <span>{filteredLogs.length} entries</span>
      </div>

      <div className="h-[600px] overflow-y-auto rounded-2xl border border-gray-200 bg-black p-4 font-mono text-sm text-gray-100 shadow-inner">
        {filteredLogs.length === 0 ? (
          <div className="py-12 text-center text-gray-500">
            {connected ? 'Waiting for logs…' : 'Connect to start streaming logs.'}
          </div>
        ) : (
          filteredLogs.map((log, index) => (
            <div key={`${log.timestamp}-${index}`} className="mb-1 break-words">
              <span className="text-gray-500">
                {log.timestamp ? new Date(log.timestamp).toLocaleTimeString() : ''}
              </span>
              {log.task_id && (
                <span className="text-indigo-300 ml-2">Task {log.task_id}</span>
              )}
              {log.level && (
                <span className={`ml-2 font-semibold ${getLevelColor(log.level)}`}>
                  [{log.level.toUpperCase()}]
                </span>
              )}
              <span className="ml-2 text-gray-100">{log.message}</span>
            </div>
          ))
        )}
        <div ref={logsEndRef} />
      </div>
    </div>
  )
}

