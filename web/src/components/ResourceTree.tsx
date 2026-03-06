import { useState, useMemo } from 'react'
import type { ResourceChange, Action } from '../types/plan'

interface Props {
  resources: ResourceChange[]
  selected: ResourceChange | null
  onSelect: (r: ResourceChange) => void
}

const actionConfig: Record<Action, { symbol: string; color: string; label: string }> = {
  'create': { symbol: '+', color: 'var(--accent-green)', label: 'create' },
  'update': { symbol: '~', color: 'var(--accent-yellow)', label: 'update' },
  'delete': { symbol: '-', color: 'var(--accent-red)', label: 'delete' },
  'replace': { symbol: '!', color: 'var(--accent-purple)', label: 'replace' },
  'create-before-delete': { symbol: '+/-', color: 'var(--accent-purple)', label: 'replace' },
  'delete-before-create': { symbol: '-/+', color: 'var(--accent-purple)', label: 'replace' },
}

const riskDotColor = {
  high: 'var(--accent-red)',
  medium: 'var(--accent-yellow)',
  low: 'var(--accent-green)',
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: '8px 0',
  },
  group: {
    marginBottom: '2px',
  },
  groupHeader: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '6px 16px',
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 600,
    color: 'var(--text-secondary)',
    cursor: 'pointer',
    userSelect: 'none',
    transition: 'background 0.1s ease',
  },
  chevron: {
    fontFamily: 'var(--font-mono)',
    fontSize: '10px',
    color: 'var(--text-muted)',
    width: '12px',
    transition: 'transform 0.15s ease',
  },
  groupCount: {
    fontFamily: 'var(--font-mono)',
    fontSize: '10px',
    color: 'var(--text-muted)',
    marginLeft: 'auto',
  },
  item: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '5px 16px 5px 36px',
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    color: 'var(--text-primary)',
    cursor: 'pointer',
    transition: 'background 0.1s ease',
    borderLeft: '2px solid transparent',
  },
  itemSelected: {
    background: 'var(--bg-elevated)',
    borderLeftColor: 'var(--accent-green)',
  },
  actionBadge: {
    fontFamily: 'var(--font-mono)',
    fontSize: '11px',
    fontWeight: 700,
    width: '24px',
    textAlign: 'center',
    flexShrink: 0,
  },
  riskDot: {
    width: '6px',
    height: '6px',
    borderRadius: '50%',
    flexShrink: 0,
    marginLeft: 'auto',
  },
  resourceName: {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    flex: 1,
    minWidth: 0,
  },
}

export default function ResourceTree({ resources, selected, onSelect }: Props) {
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())

  const grouped = useMemo(() => {
    const groups: Record<string, ResourceChange[]> = {}
    for (const r of resources) {
      if (!groups[r.type]) groups[r.type] = []
      groups[r.type].push(r)
    }
    return Object.entries(groups).sort(([a], [b]) => a.localeCompare(b))
  }, [resources])

  const toggleGroup = (type: string) => {
    setCollapsed(prev => {
      const next = new Set(prev)
      if (next.has(type)) next.delete(type)
      else next.add(type)
      return next
    })
  }

  return (
    <div style={styles.container}>
      {grouped.map(([type, items]) => {
        const isCollapsed = collapsed.has(type)
        return (
          <div key={type} style={styles.group}>
            <div
              style={styles.groupHeader}
              onClick={() => toggleGroup(type)}
              onMouseEnter={e => { e.currentTarget.style.background = 'var(--bg-tertiary)' }}
              onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}
            >
              <span style={{
                ...styles.chevron,
                transform: isCollapsed ? 'rotate(-90deg)' : 'rotate(0)',
              }}>
                v
              </span>
              <span>{type}</span>
              <span style={styles.groupCount}>{items.length}</span>
            </div>
            {!isCollapsed && items.map(r => {
              const cfg = actionConfig[r.action]
              const isSelected = selected?.address === r.address
              return (
                <div
                  key={r.address}
                  style={{
                    ...styles.item,
                    ...(isSelected ? styles.itemSelected : {}),
                  }}
                  onClick={() => onSelect(r)}
                  onMouseEnter={e => {
                    if (!isSelected) e.currentTarget.style.background = 'var(--bg-tertiary)'
                  }}
                  onMouseLeave={e => {
                    if (!isSelected) e.currentTarget.style.background = 'transparent'
                  }}
                >
                  <span style={{ ...styles.actionBadge, color: cfg.color }}>{cfg.symbol}</span>
                  <span style={styles.resourceName} title={r.address}>{r.name}</span>
                  <span style={{ ...styles.riskDot, background: riskDotColor[r.risk_level] }} />
                </div>
              )
            })}
          </div>
        )
      })}
    </div>
  )
}
