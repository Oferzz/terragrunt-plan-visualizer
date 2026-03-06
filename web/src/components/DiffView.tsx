import type { ResourceChange, AttributeChange } from '../types/plan'

interface Props {
  resource: ResourceChange
  showRisk?: boolean
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: '24px',
    animation: 'fadeInUp 0.2s ease-out',
  },
  header: {
    marginBottom: '24px',
  },
  address: {
    fontFamily: 'var(--font-mono)',
    fontSize: '14px',
    fontWeight: 600,
    color: 'var(--text-primary)',
    marginBottom: '8px',
    wordBreak: 'break-all',
  },
  meta: {
    display: 'flex',
    gap: '12px',
    flexWrap: 'wrap',
  },
  metaTag: {
    fontFamily: 'var(--font-mono)',
    fontSize: '10px',
    fontWeight: 500,
    padding: '3px 8px',
    borderRadius: '3px',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.5px',
  },
  riskSection: {
    marginBottom: '24px',
    padding: '16px',
    background: 'var(--bg-tertiary)',
    borderRadius: '6px',
    border: '1px solid var(--border)',
  },
  riskTitle: {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 600,
    color: 'var(--text-secondary)',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.5px',
    marginBottom: '10px',
  },
  riskReason: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-primary)',
    padding: '4px 0',
    display: 'flex',
    alignItems: 'flex-start',
    gap: '8px',
  },
  riskBullet: {
    color: 'var(--accent-yellow)',
    flexShrink: 0,
    fontWeight: 700,
  },
  table: {
    width: '100%',
    borderCollapse: 'collapse' as const,
  },
  th: {
    fontFamily: 'var(--font-mono)',
    fontSize: '10px',
    fontWeight: 600,
    color: 'var(--text-muted)',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.5px',
    padding: '8px 12px',
    textAlign: 'left' as const,
    borderBottom: '1px solid var(--border)',
    position: 'sticky' as const,
    top: 0,
    background: 'var(--bg-primary)',
  },
  td: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    padding: '8px 12px',
    borderBottom: '1px solid var(--border)',
    verticalAlign: 'top' as const,
    lineHeight: 1.5,
  },
  attrName: {
    color: 'var(--accent-blue)',
    fontWeight: 500,
    whiteSpace: 'nowrap' as const,
  },
  oldValue: {
    color: 'var(--accent-red)',
    wordBreak: 'break-all' as const,
  },
  newValue: {
    color: 'var(--accent-green)',
    wordBreak: 'break-all' as const,
  },
  computed: {
    fontFamily: 'var(--font-mono)',
    fontSize: '10px',
    color: 'var(--text-muted)',
    fontStyle: 'italic',
  },
  noAttrs: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-muted)',
    padding: '24px',
    textAlign: 'center',
  },
  arrow: {
    color: 'var(--text-muted)',
    textAlign: 'center' as const,
    fontSize: '14px',
  },
}

const actionColors: Record<string, { bg: string; text: string }> = {
  create: { bg: 'rgba(0, 229, 155, 0.12)', text: 'var(--accent-green)' },
  update: { bg: 'rgba(255, 184, 77, 0.12)', text: 'var(--accent-yellow)' },
  delete: { bg: 'rgba(255, 74, 110, 0.12)', text: 'var(--accent-red)' },
  replace: { bg: 'rgba(176, 125, 255, 0.12)', text: 'var(--accent-purple)' },
  'create-before-delete': { bg: 'rgba(176, 125, 255, 0.12)', text: 'var(--accent-purple)' },
  'delete-before-create': { bg: 'rgba(176, 125, 255, 0.12)', text: 'var(--accent-purple)' },
}

const riskColors: Record<string, { bg: string; text: string }> = {
  high: { bg: 'rgba(255, 74, 110, 0.12)', text: 'var(--accent-red)' },
  medium: { bg: 'rgba(255, 184, 77, 0.12)', text: 'var(--accent-yellow)' },
  low: { bg: 'rgba(0, 229, 155, 0.12)', text: 'var(--accent-green)' },
}

function formatValue(val: unknown): string {
  if (val === null || val === undefined) return 'null'
  if (typeof val === 'string') return `"${val}"`
  if (typeof val === 'object') {
    try {
      return JSON.stringify(val, null, 2)
    } catch {
      return String(val)
    }
  }
  return String(val)
}

function AttrRow({ attr }: { attr: AttributeChange }) {
  const oldStr = formatValue(attr.old_value)
  const newStr = formatValue(attr.new_value)

  return (
    <tr>
      <td style={{ ...styles.td, ...styles.attrName }}>
        {attr.name}
        {attr.computed && <span style={styles.computed}> (computed)</span>}
      </td>
      <td style={{ ...styles.td, ...styles.oldValue }}>
        {attr.old_value !== null && attr.old_value !== undefined ? oldStr : (
          <span style={{ color: 'var(--text-muted)' }}>--</span>
        )}
      </td>
      <td style={{ ...styles.td, ...styles.arrow }}>&rarr;</td>
      <td style={{ ...styles.td, ...styles.newValue }}>
        {attr.new_value !== null && attr.new_value !== undefined ? newStr : (
          <span style={{ color: 'var(--text-muted)' }}>--</span>
        )}
      </td>
    </tr>
  )
}

export default function DiffView({ resource, showRisk }: Props) {
  const actionStyle = actionColors[resource.action] || actionColors.update
  const riskStyle = riskColors[resource.risk_level] || riskColors.low

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <div style={styles.address}>{resource.address}</div>
        <div style={styles.meta}>
          <span style={{
            ...styles.metaTag,
            background: actionStyle.bg,
            color: actionStyle.text,
          }}>
            {resource.action}
          </span>
          <span style={{
            ...styles.metaTag,
            background: riskStyle.bg,
            color: riskStyle.text,
          }}>
            {resource.risk_level} risk
          </span>
          <span style={{
            ...styles.metaTag,
            background: 'rgba(77, 196, 255, 0.08)',
            color: 'var(--accent-blue)',
          }}>
            {resource.provider_name}
          </span>
        </div>
      </div>

      {showRisk && resource.risk_reasons && resource.risk_reasons.length > 0 && (
        <div style={styles.riskSection}>
          <div style={styles.riskTitle}>Risk Analysis</div>
          {resource.risk_reasons.map((reason, i) => (
            <div key={i} style={styles.riskReason}>
              <span style={styles.riskBullet}>&gt;</span>
              <span>{reason}</span>
            </div>
          ))}
        </div>
      )}

      {resource.attributes && resource.attributes.length > 0 ? (
        <table style={styles.table}>
          <thead>
            <tr>
              <th style={styles.th}>Attribute</th>
              <th style={styles.th}>Before</th>
              <th style={{ ...styles.th, width: '30px' }}></th>
              <th style={styles.th}>After</th>
            </tr>
          </thead>
          <tbody>
            {resource.attributes.map(attr => (
              <AttrRow key={attr.name} attr={attr} />
            ))}
          </tbody>
        </table>
      ) : (
        <div style={styles.noAttrs}>No attribute-level changes available</div>
      )}
    </div>
  )
}
