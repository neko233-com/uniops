import { useState, useEffect } from 'react'

interface Session {
  id: number
  user_id: number
  server_id: number
  start_time: string
  end_time: string
  status: string
}

interface ReplayEntry {
  timestamp: string
  type: string
  data: string
}

export function Audit() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [selectedSession, setSelectedSession] = useState<Session | null>(null)
  const [replay, setReplay] = useState<ReplayEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [replaying, setReplaying] = useState(false)
  const [replayIndex, setReplayIndex] = useState(0)

  useEffect(() => {
    fetchSessions()
  }, [])

  const fetchSessions = async () => {
    try {
      const res = await fetch('/api/audit/sessions')
      if (res.ok) {
        const data = await res.json()
        setSessions(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch sessions:', err)
    }
    setLoading(false)
  }

  const fetchReplay = async (sessionId: number) => {
    try {
      const res = await fetch(`/api/audit/sessions/${sessionId}/replay`)
      if (res.ok) {
        const data = await res.json()
        setReplay(data || [])
        setSelectedSession(sessions.find(s => s.id === sessionId) || null)
        setReplayIndex(0)
      }
    } catch (err) {
      console.error('Failed to fetch replay:', err)
    }
  }

  const startReplay = () => {
    if (replay.length === 0) return
    setReplaying(true)
    setReplayIndex(0)
  }

  const stopReplay = () => {
    setReplaying(false)
  }

  useEffect(() => {
    if (!replaying || replayIndex >= replay.length) {
      if (replayIndex >= replay.length && replaying) {
        setReplaying(false)
      }
      return
    }

    const timer = setTimeout(() => {
      setReplayIndex(prev => prev + 1)
    }, 100)

    return () => clearTimeout(timer)
  }, [replaying, replayIndex, replay.length])

  if (loading) {
    return <div className="p-4 text-center">Loading...</div>
  }

  return (
    <div className="flex h-[calc(100vh-200px)]">
      {/* Session list */}
      <div className="w-1/3 border-r border-gray-700 overflow-auto">
        {sessions.length === 0 ? (
          <div className="p-4 text-center text-gray-500">No sessions recorded</div>
        ) : (
          <div className="space-y-2 p-2">
            {sessions.map(session => (
              <div
                key={session.id}
                className={`rog-panel p-3 cursor-pointer hover:bg-gray-800 ${
                  selectedSession?.id === session.id ? 'border-l-2 border-l-red-500' : ''
                }`}
                onClick={() => fetchReplay(session.id)}
              >
                <div className="flex justify-between items-center">
                  <span className="font-mono text-sm">Session #{session.id}</span>
                  <span className={`text-xs px-2 py-1 rounded ${
                    session.status === 'active' ? 'bg-green-900 text-green-300' : 'bg-gray-700 text-gray-400'
                  }`}>
                    {session.status}
                  </span>
                </div>
                <div className="text-xs text-gray-500 mt-1">
                  {session.start_time}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Replay viewer */}
      <div className="flex-1 flex flex-col">
        {selectedSession ? (
          <>
            {/* Controls */}
            <div className="rog-panel-header flex items-center gap-4">
              <span className="text-sm">Session #{selectedSession.id}</span>
              <button
                onClick={replaying ? stopReplay : startReplay}
                className="rog-btn rog-btn-primary text-sm"
              >
                {replaying ? 'Stop' : 'Play'}
              </button>
              <span className="text-sm text-gray-400">
                {replayIndex} / {replay.length}
              </span>
            </div>

            {/* Terminal output */}
            <div className="flex-1 rog-terminal p-4 font-mono text-sm overflow-auto">
              {replay.slice(0, replayIndex).map((entry, i) => (
                <div key={i} className={entry.type === 'input' ? 'text-green-400' : 'text-gray-300'}>
                  {entry.data}
                </div>
              ))}
              {replaying && replayIndex < replay.length && (
                <span className="animate-pulse">█</span>
              )}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            Select a session to view replay
          </div>
        )}
      </div>
    </div>
  )
}
