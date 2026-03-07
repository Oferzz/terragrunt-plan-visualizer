import type { FeatureContext } from '../types/plan'

interface Props {
  featureContext: FeatureContext
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: '24px',
    animation: 'fadeInUp 0.2s ease-out',
  },
  section: {
    marginBottom: '24px',
  },
  sectionTitle: {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 600,
    color: 'var(--text-secondary)',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.5px',
    marginBottom: '12px',
    paddingBottom: '8px',
    borderBottom: '1px solid var(--border)',
  },
  summaryBar: {
    display: 'flex',
    gap: '12px',
    marginBottom: '16px',
  },
  badge: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    fontWeight: 600,
    padding: '6px 12px',
    borderRadius: '4px',
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
  },
  dot: {
    width: '8px',
    height: '8px',
    borderRadius: '50%',
  },
  meta: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-secondary)',
    padding: '12px 16px',
    background: 'var(--bg-tertiary)',
    borderRadius: '6px',
    border: '1px solid var(--border)',
    lineHeight: 1.8,
  },
  metaLabel: {
    color: 'var(--text-muted)',
    marginRight: '8px',
  },
  fileList: {
    listStyle: 'none',
    padding: 0,
  },
  fileItem: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-primary)',
    padding: '6px 12px',
    background: 'var(--bg-tertiary)',
    borderRadius: '4px',
    marginBottom: '4px',
    borderLeft: '2px solid var(--accent-blue)',
  },
  resourceItem: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--accent-green)',
    padding: '6px 12px',
    background: 'var(--bg-tertiary)',
    borderRadius: '4px',
    marginBottom: '4px',
    borderLeft: '2px solid var(--accent-green)',
  },
  moduleItem: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--accent-purple)',
    padding: '6px 12px',
    background: 'var(--bg-tertiary)',
    borderRadius: '4px',
    marginBottom: '4px',
    borderLeft: '2px solid var(--accent-purple)',
  },
  errorBox: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--accent-red)',
    padding: '12px 16px',
    background: 'rgba(255, 74, 110, 0.08)',
    borderRadius: '6px',
    border: '1px solid rgba(255, 74, 110, 0.2)',
  },
}

export default function FeaturePanel({ featureContext: fc }: Props) {
  if (fc.error) {
    return (
      <div style={styles.container}>
        <div style={styles.section}>
          <div style={styles.sectionTitle}>Feature Analysis</div>
          <div style={styles.errorBox}>{fc.error}</div>
        </div>
      </div>
    )
  }

  return (
    <div style={styles.container}>
      <div style={styles.section}>
        <div style={styles.sectionTitle}>Summary</div>
        <div style={styles.summaryBar}>
          <span style={{
            ...styles.badge,
            background: 'rgba(0, 229, 155, 0.12)',
            color: 'var(--accent-green)',
          }}>
            <span style={{ ...styles.dot, background: 'var(--accent-green)' }} />
            {fc.expected_count} expected
          </span>
          <span style={{
            ...styles.badge,
            background: 'rgba(255, 184, 77, 0.12)',
            color: 'var(--accent-yellow)',
          }}>
            <span style={{ ...styles.dot, background: 'var(--accent-yellow)' }} />
            {fc.indirect_count} indirect
          </span>
          <span style={{
            ...styles.badge,
            background: 'rgba(255, 74, 110, 0.12)',
            color: 'var(--accent-red)',
          }}>
            <span style={{ ...styles.dot, background: 'var(--accent-red)' }} />
            {fc.unrelated_count} unrelated
          </span>
        </div>
      </div>

      <div style={styles.section}>
        <div style={styles.sectionTitle}>Git Diff Info</div>
        <div style={styles.meta}>
          <div><span style={styles.metaLabel}>base branch:</span>{fc.base_branch}</div>
          <div><span style={styles.metaLabel}>files changed:</span>{fc.files_changed?.length ?? 0}</div>
          <div><span style={styles.metaLabel}>resources in diff:</span>{fc.resources_in_diff?.length ?? 0}</div>
          <div><span style={styles.metaLabel}>modules in diff:</span>{fc.modules_in_diff?.length ?? 0}</div>
        </div>
      </div>

      {fc.files_changed && fc.files_changed.length > 0 && (
        <div style={styles.section}>
          <div style={styles.sectionTitle}>Changed Files</div>
          <ul style={styles.fileList}>
            {fc.files_changed.map((f, i) => (
              <li key={i} style={styles.fileItem}>{f}</li>
            ))}
          </ul>
        </div>
      )}

      {fc.resources_in_diff && fc.resources_in_diff.length > 0 && (
        <div style={styles.section}>
          <div style={styles.sectionTitle}>Resources in Diff</div>
          <ul style={styles.fileList}>
            {fc.resources_in_diff.map((r, i) => (
              <li key={i} style={styles.resourceItem}>{r}</li>
            ))}
          </ul>
        </div>
      )}

      {fc.modules_in_diff && fc.modules_in_diff.length > 0 && (
        <div style={styles.section}>
          <div style={styles.sectionTitle}>Modules in Diff</div>
          <ul style={styles.fileList}>
            {fc.modules_in_diff.map((m, i) => (
              <li key={i} style={styles.moduleItem}>{m}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
