import type { PlanSummary } from '../types/plan'

interface Props {
  summary: PlanSummary
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    gap: '12px',
    padding: '12px 16px',
    borderBottom: '1px solid var(--border)',
    flexShrink: 0,
  },
  badge: {
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 500,
  },
  dot: {
    width: '8px',
    height: '8px',
    borderRadius: '50%',
    flexShrink: 0,
  },
}

export default function RiskPanel({ summary }: Props) {
  const risks = [
    { label: 'High', count: summary.high_risk, color: 'var(--accent-red)' },
    { label: 'Medium', count: summary.medium_risk, color: 'var(--accent-yellow)' },
    { label: 'Low', count: summary.low_risk, color: 'var(--accent-green)' },
  ]

  return (
    <div style={styles.container}>
      {risks.map(r => (
        <div key={r.label} style={styles.badge}>
          <div style={{ ...styles.dot, background: r.count > 0 ? r.color : 'var(--text-muted)' }} />
          <span style={{ color: r.count > 0 ? 'var(--text-primary)' : 'var(--text-muted)' }}>
            {r.count} {r.label}
          </span>
        </div>
      ))}
    </div>
  )
}
