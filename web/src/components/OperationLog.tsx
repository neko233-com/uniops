import { useState, useEffect } from 'react'

interface LogEntry {
  timestamp: string
  user: string
  user_id: number
  action: string
  method: string
  path: string
  status: number
  ip: string
  detail?: string
}

export function OperationLog() {
  const [dates, setDates] = useState<string[]>([])
  const [selectedDate, setSelectedDate] = useState('')
  const [entries, setEntries] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [searchUser, setSearchUser] = useState('')
  const [searchAction, setSearchAction] = useState('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [cleanStart, setCleanStart] = useState('')
  const [cleanEnd, setCleanEnd] = useState('')
  const [message, setMessage] = useState('')

  const token = localStorage.getItem('token')
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  }

  useEffect(() => {
    fetchDates()
  }, [])

  const fetchDates = async () => {
    try {
      const res = await fetch('/api/oplog/dates', { headers })
      if (res.ok) {
        const data = await res.json()
        setDates(data || [])
        if (data.length > 0 && !selectedDate) {
          setSelectedDate(data[0])
          fetchEntries(data[0])
        }
      }
    } catch (err) {
      console.error('Fetch dates error:', err)
    }
  }

  const fetchEntries = async (date: string) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/oplog/read?date=${date}`, { headers })
      if (res.ok) {
        const data = await res.json()
        setEntries(data || [])
      }
    } catch (err) {
      console.error('Fetch entries error:', err)
    }
    setLoading(false)
  }

  const handleSearch = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      if (searchUser) params.set('user', searchUser)
      if (searchAction) params.set('action', searchAction)
      if (searchKeyword) params.set('keyword', searchKeyword)
      const res = await fetch(`/api/oplog/search?${params}`, { headers })
      if (res.ok) {
        const data = await res.json()
        setEntries(data || [])
      }
    } catch (err) {
      console.error('Search error:', err)
    }
    setLoading(false)
  }

  const handleClean = async () => {
    if (!cleanStart || !cleanEnd) {
      setMessage('Please select start and end dates')
      return
    }
    if (!confirm(`Delete operation logs from ${cleanStart} to ${cleanEnd}?`)) return

    try {
      const res = await fetch('/api/oplog', {
        method: 'DELETE',
        headers,
        body: JSON.stringify({ start_date: cleanStart, end_date: cleanEnd }),
      })
      if (res.ok) {
        const data = await res.json()
        setMessage(`Deleted ${data.deleted} log file(s)`)
        fetchDates()
      }
    } catch (err) {
      setMessage('Cleanup failed')
    }
  }

  const statusBadge = (status: number) => {
    if (status >= 200 && status < 300) return 'bg-green-900/50 text-green-400'
    if (status >= 400 && status < 500) return 'bg-amber-900/50 text-amber-400'
    if (status >= 500) return 'bg-red-900/50 text-red-400'
    return 'bg-gray-700 text-gray-400'
  }

  return (
    <div className="space-y-6">
      {/* Search panel */}
      <div className="rog-panel p-4">
        <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">Search Logs</h3>
        <div className="flex gap-3 flex-wrap">
          <input
            placeholder="User"
            value={searchUser}
            onChange={e => setSearchUser(e.target.value)}
            className="rog-input text-sm flex-1 min-w-[120px]"
          />
          <input
            placeholder="Action (e.g. POST)"
            value={searchAction}
            onChange={e => setSearchAction(e.target.value)}
            className="rog-input text-sm flex-1 min-w-[120px]"
          />
          <input
            placeholder="Keyword (path/detail)"
            value={searchKeyword}
            onChange={e => setSearchKeyword(e.target.value)}
            className="rog-input text-sm flex-1 min-w-[150px]"
          />
          <button onClick={handleSearch} className="rog-btn rog-btn-primary text-sm px-4">
            Search
          </button>
        </div>
      </div>

      {/* Date selector + cleanup */}
      <div className="rog-panel p-4">
        <div className="flex items-center justify-between flex-wrap gap-3">
          <div className="flex items-center gap-3">
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">By Date</h3>
            <div className="flex gap-1 flex-wrap">
              {dates.map(date => (
                <button
                  key={date}
                  onClick={() => { setSelectedDate(date); fetchEntries(date) }}
                  className={`text-xs px-3 py-1 rounded transition-colors ${
                    selectedDate === date
                      ? 'bg-red-900/50 text-red-300 border border-red-700'
                      : 'bg-gray-800 text-gray-400 hover:bg-gray-700 border border-gray-700'
                  }`}
                >
                  {date}
                </button>
              ))}
              {dates.length === 0 && <span className="text-xs text-gray-600">No logs yet</span>}
            </div>
          </div>

          {/* Cleanup */}
          <div className="flex items-center gap-2">
            <input
              type="date"
              value={cleanStart}
              onChange={e => setCleanStart(e.target.value)}
              className="rog-input text-xs py-1 px-2"
            />
            <span className="text-gray-600">~</span>
            <input
              type="date"
              value={cleanEnd}
              onChange={e => setCleanEnd(e.target.value)}
              className="rog-input text-xs py-1 px-2"
            />
            <button onClick={handleClean} className="rog-btn text-xs text-red-400 px-3 py-1 border-red-800 hover:border-red-500">
              Clean
            </button>
            {message && <span className="text-xs text-gray-500">{message}</span>}
          </div>
        </div>
      </div>

      {/* Log table */}
      <div className="rog-panel overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-800">
                <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Time</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">User</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Method</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Path</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Status</th>
                <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">IP</th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-500">Loading...</td></tr>
              )}
              {!loading && entries.length === 0 && (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-600">No log entries</td></tr>
              )}
              {entries.map((entry, i) => (
                <tr key={i} className="border-b border-gray-800/50 hover:bg-gray-800/30 transition-colors">
                  <td className="px-4 py-2.5 font-mono text-xs text-gray-400">
                    {entry.timestamp ? new Date(entry.timestamp).toLocaleTimeString() : '-'}
                  </td>
                  <td className="px-4 py-2.5 text-gray-300">{entry.user}</td>
                  <td className="px-4 py-2.5">
                    <span className={`text-xs px-2 py-0.5 rounded font-mono ${
                      entry.method === 'GET' ? 'bg-blue-900/40 text-blue-400' :
                      entry.method === 'POST' ? 'bg-green-900/40 text-green-400' :
                      entry.method === 'DELETE' ? 'bg-red-900/40 text-red-400' :
                      'bg-gray-700 text-gray-400'
                    }`}>
                      {entry.method}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 font-mono text-xs text-gray-400 max-w-[300px] truncate">
                    {entry.path}
                  </td>
                  <td className="px-4 py-2.5">
                    <span className={`text-xs px-2 py-0.5 rounded ${statusBadge(entry.status)}`}>
                      {entry.status}
                    </span>
                  </td>
                  <td className="px-4 py-2.5 text-xs text-gray-500">{entry.ip}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="px-4 py-2 border-t border-gray-800 text-xs text-gray-600">
          {entries.length} entries
        </div>
      </div>
    </div>
  )
}
