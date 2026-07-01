import { useEffect, useRef } from 'react'
import { Terminal as XTerminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

interface TerminalProps {
  serverId: number
}

export function Terminal({ serverId }: TerminalProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<XTerminal | null>(null)

  useEffect(() => {
    if (!containerRef.current) return

    const term = new XTerminal({
      theme: {
        background: '#1a1a2e',
        foreground: '#eaeaea',
      },
    })

    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(containerRef.current)
    fitAddon.fit()

    termRef.current = term

    const ws = new WebSocket(`ws://localhost:8080/ws/terminal/${serverId}`)

    ws.onopen = () => {
      ws.send(JSON.stringify({
        type: 'resize',
        data: JSON.stringify({ width: term.cols, height: term.rows }),
      }))
    }

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      if (msg.type === 'output') {
        term.write(msg.data)
      }
    }

    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', data }))
      }
    })

    const resizeHandler = () => {
      fitAddon.fit()
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          data: JSON.stringify({ width: term.cols, height: term.rows }),
        }))
      }
    }

    window.addEventListener('resize', resizeHandler)

    return () => {
      window.removeEventListener('resize', resizeHandler)
      ws.close()
      term.dispose()
    }
  }, [serverId])

  return <div ref={containerRef} className="h-full w-full" />
}
