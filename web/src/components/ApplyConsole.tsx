import { useEffect, useRef, useState } from 'react'
import type { ApplyLine } from '../types/plan'

interface Props {
  lines: ApplyLine[]
  isRunning: boolean
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
    background: 'var(--bg-primary)',
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '12px 16px',
    borderBottom: '1px solid var(--border)',
    flexShrink: 0,
  },
  statusDot: {
    width: '8px',
    height: '8px',
    borderRadius: '50%',
  },
  statusText: {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 600,
    textTransform: 'uppercase' as const,
    letterSpacing: '0.5px',
  },
  console: {
    flex: 1,
    overflow: 'auto',
    padding: '12px 16px',
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    lineHeight: 1.6,
  },
  line: {
    whiteSpace: 'pre-wrap' as const,
    wordBreak: 'break-all' as const,
  },
  lineStdout: {
    color: 'var(--text-primary)',
  },
  lineStderr: {
    color: 'var(--accent-red)',
  },
  lineStatus: {
    color: 'var(--accent-blue)',
    fontWeight: 500,
  },
  timestamp: {
    color: 'var(--text-muted)',
    marginRight: '8px',
    fontSize: '10px',
  },
}

export default function ApplyConsole({ lines, isRunning }: Props) {
  const consoleRef = useRef<HTMLDivElement>(null)
  const [autoScroll, setAutoScroll] = useState(true)

  useEffect(() => {
    if (autoScroll && consoleRef.current) {
      consoleRef.current.scrollTop = consoleRef.current.scrollHeight
    }
  }, [lines, autoScroll])

  const handleScroll = () => {
    if (!consoleRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = consoleRef.current
    setAutoScroll(scrollHeight - scrollTop - clientHeight < 50)
  }

  const formatTime = (ts: number) => {
    const d = new Date(ts)
    return d.toLocaleTimeString('en-US', { hour12: false })
  }

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <div style={{
          ...styles.statusDot,
          background: isRunning ? 'var(--accent-green)' : 'var(--text-muted)',
          animation: isRunning ? 'pulse 1s ease-in-out infinite' : 'none',
        }} />
        <span style={{
          ...styles.statusText,
          color: isRunning ? 'var(--accent-green)' : 'var(--text-secondary)',
        }}>
          {isRunning ? 'Applying...' : 'Complete'}
        </span>
      </div>
      <div style={styles.console} ref={consoleRef} onScroll={handleScroll}>
        {lines.map((line, i) => {
          const lineStyle = line.type === 'stderr'
            ? styles.lineStderr
            : line.type === 'status'
              ? styles.lineStatus
              : styles.lineStdout

          return (
            <div key={i} style={{ ...styles.line, ...lineStyle }}>
              <span style={styles.timestamp}>{formatTime(line.timestamp)}</span>
              {line.text}
            </div>
          )
        })}
      </div>
    </div>
  )
}
