import { useEffect, useState } from 'react'
import type { Plan, ResourceChange } from './types/plan'
import Summary from './components/Summary'
import ResourceTree from './components/ResourceTree'
import DiffView from './components/DiffView'
import RiskPanel from './components/RiskPanel'
import AIAnalysis from './components/AIAnalysis'

const globalStyles = `
  *, *::before, *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  :root {
    --bg-primary: #0a0e14;
    --bg-secondary: #0f1419;
    --bg-tertiary: #151b23;
    --bg-elevated: #1a2029;
    --border: #1e2832;
    --border-active: #2d3b4a;
    --text-primary: #c8d3de;
    --text-secondary: #6b7d8f;
    --text-muted: #3f4e5c;
    --accent-green: #00e59b;
    --accent-green-dim: #00b87a;
    --accent-red: #ff4a6e;
    --accent-red-dim: #cc3a57;
    --accent-yellow: #ffb84d;
    --accent-yellow-dim: #cc9440;
    --accent-blue: #4dc4ff;
    --accent-purple: #b07dff;
    --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
    --font-sans: 'DM Sans', -apple-system, sans-serif;
    --glow-green: 0 0 20px rgba(0, 229, 155, 0.15);
    --glow-red: 0 0 20px rgba(255, 74, 110, 0.15);
    --glow-yellow: 0 0 20px rgba(255, 184, 77, 0.15);
  }

  html, body, #root {
    height: 100%;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: var(--font-sans);
    -webkit-font-smoothing: antialiased;
    overflow: hidden;
  }

  ::-webkit-scrollbar {
    width: 6px;
  }
  ::-webkit-scrollbar-track {
    background: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background: var(--border-active);
    border-radius: 3px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: var(--text-muted);
  }

  @keyframes fadeInUp {
    from { opacity: 0; transform: translateY(8px); }
    to { opacity: 1; transform: translateY(0); }
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
`

const s: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100vh',
    position: 'relative',
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '12px 24px',
    borderBottom: '1px solid var(--border)',
    background: 'var(--bg-secondary)',
    flexShrink: 0,
  },
  headerLeft: {
    display: 'flex',
    alignItems: 'center',
    gap: '16px',
  },
  logo: {
    fontFamily: 'var(--font-mono)',
    fontSize: '16px',
    fontWeight: 700,
    color: 'var(--accent-green)',
    letterSpacing: '-0.5px',
  },
  logoSub: {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    color: 'var(--text-muted)',
    letterSpacing: '0.5px',
    textTransform: 'uppercase' as const,
  },
  headerMeta: {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    color: 'var(--text-secondary)',
    display: 'flex',
    gap: '16px',
  },
  body: {
    display: 'flex',
    flex: 1,
    overflow: 'hidden',
  },
  leftPanel: {
    width: '380px',
    minWidth: '320px',
    borderRight: '1px solid var(--border)',
    display: 'flex',
    flexDirection: 'column',
    background: 'var(--bg-secondary)',
    overflow: 'hidden',
  },
  rightPanel: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    overflow: 'hidden',
  },
  loading: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100vh',
    flexDirection: 'column',
    gap: '16px',
  },
  loadingText: {
    fontFamily: 'var(--font-mono)',
    fontSize: '14px',
    color: 'var(--accent-green)',
    animation: 'pulse 1.5s ease-in-out infinite',
  },
  noChanges: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100vh',
    flexDirection: 'column',
    gap: '12px',
  },
  noChangesIcon: {
    fontSize: '48px',
    color: 'var(--accent-green)',
    fontFamily: 'var(--font-mono)',
  },
  noChangesText: {
    fontFamily: 'var(--font-sans)',
    fontSize: '18px',
    fontWeight: 600,
    color: 'var(--text-primary)',
  },
  noChangesSub: {
    fontFamily: 'var(--font-mono)',
    fontSize: '12px',
    color: 'var(--text-secondary)',
  },
  errorBox: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100vh',
    flexDirection: 'column',
    gap: '12px',
  },
  errorText: {
    fontFamily: 'var(--font-mono)',
    fontSize: '14px',
    color: 'var(--accent-red)',
    maxWidth: '500px',
    textAlign: 'center',
  },
  scrollArea: {
    flex: 1,
    overflow: 'auto',
  },
  tabBar: {
    display: 'flex',
    borderBottom: '1px solid var(--border)',
    background: 'var(--bg-secondary)',
    flexShrink: 0,
  },
  tab: {
    padding: '10px 20px',
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 500,
    color: 'var(--text-secondary)',
    cursor: 'pointer',
    border: 'none',
    background: 'none',
    borderBottom: '2px solid transparent',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.5px',
    transition: 'all 0.15s ease',
  },
  tabActive: {
    color: 'var(--accent-green)',
    borderBottomColor: 'var(--accent-green)',
  },
}

function App() {
  const [plan, setPlan] = useState<Plan | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedResource, setSelectedResource] = useState<ResourceChange | null>(null)
  const [activeTab, setActiveTab] = useState<'diff' | 'risk' | 'ai'>('diff')

  useEffect(() => {
    fetch('/api/plan')
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then((data: Plan) => {
        setPlan(data)
        if (data.resource_changes?.length > 0) {
          setSelectedResource(data.resource_changes[0])
        }
        setLoading(false)
      })
      .catch(err => {
        setError(err.message)
        setLoading(false)
      })
  }, [])

  if (loading) {
    return (
      <>
        <style>{globalStyles}</style>
        <div style={s.loading}>
          <div style={s.loadingText}>initializing tgviz...</div>
        </div>
      </>
    )
  }

  if (error) {
    return (
      <>
        <style>{globalStyles}</style>
        <div style={s.errorBox}>
          <div style={{ ...s.noChangesIcon, color: 'var(--accent-red)' }}>ERR</div>
          <div style={s.errorText}>{error}</div>
        </div>
      </>
    )
  }

  if (!plan || plan.resource_changes.length === 0) {
    return (
      <>
        <style>{globalStyles}</style>
        <div style={s.noChanges}>
          <div style={s.noChangesIcon}>OK</div>
          <div style={s.noChangesText}>Infrastructure is up to date</div>
          <div style={s.noChangesSub}>no changes detected in plan output</div>
        </div>
      </>
    )
  }

  const timestamp = plan.timestamp
    ? new Date(plan.timestamp).toLocaleString()
    : 'unknown'

  return (
    <>
      <style>{globalStyles}</style>
      <div style={s.container}>
        <header style={s.header}>
          <div style={s.headerLeft}>
            <div>
              <div style={s.logo}>tgviz</div>
              <div style={s.logoSub}>plan visualizer</div>
            </div>
          </div>
          <div style={s.headerMeta}>
            <span>tf {plan.terraform_version || '?'}</span>
            <span>{timestamp}</span>
            {plan.working_dir && <span>{plan.working_dir}</span>}
          </div>
        </header>

        <div style={s.body}>
          <div style={s.leftPanel}>
            <Summary summary={plan.summary} />
            <RiskPanel summary={plan.summary} />
            <div style={s.scrollArea}>
              <ResourceTree
                resources={plan.resource_changes}
                selected={selectedResource}
                onSelect={setSelectedResource}
              />
            </div>
          </div>

          <div style={s.rightPanel}>
            <div style={s.tabBar}>
              {(['diff', 'risk', 'ai'] as const).map(tab => (
                <button
                  key={tab}
                  style={{
                    ...s.tab,
                    ...(activeTab === tab ? s.tabActive : {}),
                  }}
                  onClick={() => setActiveTab(tab)}
                >
                  {tab === 'diff' ? 'Attribute Diff' : tab === 'risk' ? 'Risk Details' : 'AI Analysis'}
                </button>
              ))}
            </div>
            <div style={s.scrollArea}>
              {activeTab === 'diff' && selectedResource && (
                <DiffView resource={selectedResource} />
              )}
              {activeTab === 'risk' && selectedResource && (
                <DiffView resource={selectedResource} showRisk />
              )}
              {activeTab === 'ai' && <AIAnalysis />}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

export default App
