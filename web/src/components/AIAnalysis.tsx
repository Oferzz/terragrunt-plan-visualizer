import { useState } from 'react'
import type { AIAnalysisData } from '../types/plan'

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: '24px',
    animation: 'fadeInUp 0.2s ease-out',
  },
  empty: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    padding: '80px 24px',
    gap: '16px',
  },
  emptyIcon: {
    fontFamily: 'var(--font-mono)',
    fontSize: '32px',
    color: 'var(--accent-purple)',
    opacity: 0.6,
  },
  emptyTitle: {
    fontFamily: 'var(--font-sans)',
    fontSize: '16px',
    fontWeight: 600,
    color: 'var(--text-primary)',
  },
  emptyDesc: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-secondary)',
    textAlign: 'center',
    maxWidth: '400px',
    lineHeight: 1.6,
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
  finding: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-primary)',
    padding: '8px 12px',
    background: 'var(--bg-tertiary)',
    borderRadius: '4px',
    marginBottom: '6px',
    borderLeft: '2px solid var(--accent-yellow)',
    lineHeight: 1.5,
  },
  recommendation: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-primary)',
    padding: '8px 12px',
    background: 'var(--bg-tertiary)',
    borderRadius: '4px',
    marginBottom: '6px',
    borderLeft: '2px solid var(--accent-green)',
    lineHeight: 1.5,
  },
  riskSummary: {
    fontFamily: 'var(--font-sans)',
    fontSize: '14px',
    color: 'var(--text-primary)',
    lineHeight: 1.6,
    padding: '16px',
    background: 'var(--bg-tertiary)',
    borderRadius: '6px',
    border: '1px solid var(--border)',
  },
  fetchButton: {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 600,
    padding: '10px 24px',
    background: 'rgba(176, 125, 255, 0.12)',
    color: 'var(--accent-purple)',
    border: '1px solid rgba(176, 125, 255, 0.2)',
    borderRadius: '4px',
    cursor: 'pointer',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.5px',
    transition: 'all 0.15s ease',
    marginTop: '8px',
  },
}

export default function AIAnalysis() {
  const [analysis, setAnalysis] = useState<AIAnalysisData | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchAnalysis = async () => {
    setLoading(true)
    try {
      const res = await fetch('/api/ai-analysis')
      if (res.ok) {
        const data = await res.json()
        setAnalysis(data)
      }
    } catch {
      // AI analysis is optional
    } finally {
      setLoading(false)
    }
  }

  if (!analysis) {
    return (
      <div style={styles.empty}>
        <div style={styles.emptyIcon}>AI</div>
        <div style={styles.emptyTitle}>AI Security Analysis</div>
        <div style={styles.emptyDesc}>
          When running via Claude Code, AI analysis is automatically populated with security
          findings, risk assessment, and recommendations.
        </div>
        <button
          style={{
            ...styles.fetchButton,
            opacity: loading ? 0.5 : 1,
            cursor: loading ? 'default' : 'pointer',
          }}
          onClick={fetchAnalysis}
          disabled={loading}
          onMouseEnter={e => {
            if (!loading) e.currentTarget.style.background = 'rgba(176, 125, 255, 0.2)'
          }}
          onMouseLeave={e => {
            e.currentTarget.style.background = 'rgba(176, 125, 255, 0.12)'
          }}
        >
          {loading ? 'Fetching...' : 'Fetch Analysis'}
        </button>
      </div>
    )
  }

  return (
    <div style={styles.container}>
      {analysis.risk_summary && (
        <div style={styles.section}>
          <div style={styles.sectionTitle}>Risk Summary</div>
          <div style={styles.riskSummary}>{analysis.risk_summary}</div>
        </div>
      )}

      {analysis.findings && analysis.findings.length > 0 && (
        <div style={styles.section}>
          <div style={styles.sectionTitle}>Findings</div>
          {analysis.findings.map((f, i) => (
            <div key={i} style={styles.finding}>{f}</div>
          ))}
        </div>
      )}

      {analysis.recommendations && analysis.recommendations.length > 0 && (
        <div style={styles.section}>
          <div style={styles.sectionTitle}>Recommendations</div>
          {analysis.recommendations.map((r, i) => (
            <div key={i} style={styles.recommendation}>{r}</div>
          ))}
        </div>
      )}
    </div>
  )
}
