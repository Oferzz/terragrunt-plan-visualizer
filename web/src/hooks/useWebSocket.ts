import { useEffect, useRef, useState, useCallback } from 'react'
import type { ApplyLine } from '../types/plan'

interface UseWebSocketReturn {
  lines: ApplyLine[]
  isConnected: boolean
  isRunning: boolean
  connect: () => void
  disconnect: () => void
}

export function useWebSocket(url: string): UseWebSocketReturn {
  const [lines, setLines] = useState<ApplyLine[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [isRunning, setIsRunning] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)

  const connect = useCallback(() => {
    if (wsRef.current) return

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      setIsConnected(true)
      setIsRunning(true)
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const line: ApplyLine = {
          text: data.text || event.data,
          timestamp: data.timestamp || Date.now(),
          type: data.type || 'stdout',
        }
        setLines(prev => [...prev, line])

        if (data.type === 'status' && data.text === 'complete') {
          setIsRunning(false)
        }
      } catch {
        setLines(prev => [...prev, {
          text: event.data,
          timestamp: Date.now(),
          type: 'stdout',
        }])
      }
    }

    ws.onclose = () => {
      setIsConnected(false)
      setIsRunning(false)
      wsRef.current = null
    }

    ws.onerror = () => {
      setIsConnected(false)
      setIsRunning(false)
    }
  }, [url])

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
  }, [])

  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [])

  return { lines, isConnected, isRunning, connect, disconnect }
}
