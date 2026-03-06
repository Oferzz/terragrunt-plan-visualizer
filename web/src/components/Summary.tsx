import type { PlanSummary } from '../types/plan'

interface Props {
  summary: PlanSummary
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, 1fr)',
    gap: '1px',
    background: 'var(--border)',
    borderBottom: '1px solid var(--border)',
    flexShrink: 0,
  },
  card: {
    background: 'var(--bg-secondary)',
    padding: '14px 12px',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: '4px',
  },
  count: {
    fontFamily: 'var(--font-mono)',
    fontSize: '24px',
    fontWeight: 700,
    lineHeight: 1,
  },
  label: {
    fontFamily: 'var(--font-mono)',
    fontSize: '9px',
    fontWeight: 500,
    textTransform: 'uppercase' as const,
    letterSpacing: '1px',
    color: 'var(--text-muted)',
  },
}

const cardColors: Record<string, { count: string; shadow: string }> = {
  add: { count: 'var(--accent-green)', shadow: 'var(--glow-green)' },
  change: { count: 'var(--accent-yellow)', shadow: 'var(--glow-yellow)' },
  destroy: { count: 'var(--accent-red)', shadow: 'var(--glow-red)' },
  replace: { count: 'var(--accent-purple)', shadow: '0 0 20px rgba(176, 125, 255, 0.15)' },
}

export default function Summary({ summary }: Props) {
  const items = [
    { key: 'add', label: 'Add', value: summary.adds },
    { key: 'change', label: 'Change', value: summary.changes },
    { key: 'destroy', label: 'Destroy', value: summary.destroys },
    { key: 'replace', label: 'Replace', value: summary.replaces },
  ]

  return (
    <div style={styles.container}>
      {items.map(item => {
        const color = cardColors[item.key]
        return (
          <div key={item.key} style={styles.card}>
            <div
              style={{
                ...styles.count,
                color: item.value > 0 ? color.count : 'var(--text-muted)',
                textShadow: item.value > 0 ? color.shadow : 'none',
              }}
            >
              {item.value}
            </div>
            <div style={styles.label}>{item.label}</div>
          </div>
        )
      })}
    </div>
  )
}
