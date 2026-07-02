import { useState } from 'react'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

interface AgentChatProps {
  agentId: number
  serverId?: number
}

export function AgentChat({ agentId, serverId }: AgentChatProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)

  const sendMessage = async () => {
    if (!input.trim() || loading) return

    const userMessage: Message = { role: 'user', content: input }
    setMessages(prev => [...prev, userMessage])
    setInput('')
    setLoading(true)

    try {
      const res = await fetch('/api/agent/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          agent_id: agentId,
          server_id: serverId ?? 0,
          messages: [...messages, userMessage].map(m => ({
            role: m.role,
            content: m.content,
          })),
        }),
      })

      if (res.ok) {
        const data = await res.json()
        setMessages(prev => [...prev, { role: 'assistant', content: data.content }])
      }
    } catch (err) {
      console.error('Chat error:', err)
    }
    setLoading(false)
  }

  // Format message content: render tool call blocks specially
  const renderContent = (content: string) => {
    const parts = content.split(/(```tool\n[\s\S]*?\n```)/g)
    return parts.map((part, i) => {
      if (part.startsWith('```tool')) {
        const json = part.replace(/```tool\n|\n```/g, '')
        try {
          const parsed = JSON.parse(json)
          return (
            <div key={i} className="rog-panel p-2 my-1 border-l-2 border-amber-500 text-xs">
              <span className="text-amber-400 font-mono">⚡ {parsed.action}</span>
              <pre className="text-gray-400 mt-1 overflow-auto">{JSON.stringify(parsed.params, null, 2)}</pre>
            </div>
          )
        } catch {
          return <pre key={i} className="text-xs">{part}</pre>
        }
      }
      return <span key={i}>{part}</span>
    })
  }

  return (
    <div className="flex flex-col h-[calc(100vh-150px)]">
      <div className="flex-1 overflow-auto space-y-4 mb-4">
        {messages.length === 0 && (
          <div className="text-center text-gray-500 py-8">
            {serverId
              ? 'AI Agent ready. Ask it to deploy services or run commands on the target server.'
              : 'Select a server first, then ask the agent to deploy or run commands.'}
          </div>
        )}
        {messages.map((msg, i) => (
          <div
            key={i}
            className={`rog-panel p-3 ${
              msg.role === 'user' ? 'ml-12' : 'mr-12'
            }`}
          >
            <div className="text-xs text-gray-500 mb-1">
              {msg.role === 'user' ? 'You' : 'Agent'}
            </div>
            <div className="whitespace-pre-wrap">
              {msg.role === 'assistant' ? renderContent(msg.content) : msg.content}
            </div>
          </div>
        ))}
        {loading && (
          <div className="rog-panel p-3 mr-12">
            <div className="text-xs text-gray-500 mb-1">Agent</div>
            <div className="text-gray-400">Thinking...</div>
          </div>
        )}
      </div>

      <div className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
          placeholder={serverId ? 'e.g. "deploy nginx + backend on this server"' : 'Select a server first...'}
          className="flex-1 rog-input"
          disabled={loading || !serverId}
        />
        <button
          onClick={sendMessage}
          disabled={loading || !input.trim() || !serverId}
          className="rog-btn rog-btn-primary"
        >
          Send
        </button>
      </div>
    </div>
  )
}
