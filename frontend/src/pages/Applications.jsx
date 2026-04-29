import { useState, useEffect } from 'react'
import { listApplications, updateStatus, adaptCV } from '../api/client'

const COLUMNS = [
  { status: 'new',       label: '🆕 New',       color: '#6366f1' },
  { status: 'applied',   label: '📤 Applied',   color: '#3b82f6' },
  { status: 'screening', label: '🔎 Screening', color: '#f59e0b' },
  { status: 'interview', label: '🎙 Interview',  color: '#8b5cf6' },
  { status: 'offer',     label: '🎉 Offer',      color: '#22c55e' },
  { status: 'rejected',  label: '❌ Rejected',  color: '#ef4444' },
]

export default function Applications() {
  const [apps, setApps]   = useState([])
  const [selected, setSelected] = useState(null) // detail modal
  const [loading, setLoading]   = useState(false)
  const [msg, setMsg] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      setApps(await listApplications())
    } catch(e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const move = async (id, status) => {
    await updateStatus(id, status)
    setApps(prev => prev.map(a => a.id === id ? { ...a, status } : a))
  }

  const handleAdaptCV = async (id) => {
    setMsg('⏳ Adapting CV with AI...')
    try {
      const res = await adaptCV(id)
      setApps(prev => prev.map(a => a.id === id ? { ...a, adapted_cv: res.adapted_cv } : a))
      setMsg('✅ CV adapted!')
      if (selected?.id === id) setSelected(s => ({ ...s, adapted_cv: res.adapted_cv }))
    } catch (e) {
      setMsg('❌ ' + e.message)
    }
  }

  if (loading) return <p>Loading...</p>

  return (
    <div className="page">
      <h2>Application Tracker</h2>
      {msg && <p className="msg">{msg}</p>}
      {apps.length === 0 ? (
        <p className="hint">No applications yet. Go to Jobs and click "+ Track Application".</p>
      ) : (
        <div className="kanban">
          {COLUMNS.map(col => {
            const colApps = apps.filter(a => a.status === col.status)
            return (
              <div key={col.status} className="kanban-col">
                <div className="kanban-col-header" style={{ borderColor: col.color }}>
                  {col.label}
                  <span className="count">{colApps.length}</span>
                </div>
                {colApps.map(app => (
                  <AppCard
                    key={app.id}
                    app={app}
                    onMove={move}
                    onSelect={setSelected}
                    onAdapt={handleAdaptCV}
                  />
                ))}
              </div>
            )
          })}
        </div>
      )}

      {selected && (
        <Modal app={selected} onClose={() => setSelected(null)} onAdapt={handleAdaptCV} />
      )}
    </div>
  )
}

function AppCard({ app, onMove, onSelect, onAdapt }) {
  return (
    <div className="app-card card" onClick={() => onSelect(app)}>
      <h4>{app.job?.title}</h4>
      <p className="company">{app.job?.company}</p>
      <div className="card-actions" onClick={e => e.stopPropagation()}>
        <select
          className="status-select"
          value={app.status}
          onChange={e => onMove(app.id, e.target.value)}
        >
          {COLUMNS.map(c => (
            <option key={c.status} value={c.status}>{c.label}</option>
          ))}
        </select>
        {!app.adapted_cv && (
          <button
            className="btn btn-sm"
            onClick={() => onAdapt(app.id)}
          >🤖 Adapt CV</button>
        )}
        {app.adapted_cv && <span className="badge green">CV adapted ✓</span>}
      </div>
    </div>
  )
}

function Modal({ app, onClose, onAdapt }) {
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{app.job?.title} @ {app.job?.company}</h3>
          <button className="close-btn" onClick={onClose}>✕</button>
        </div>

        <p><a href={app.job?.url} target="_blank" rel="noreferrer">View job posting ↗</a></p>
        <p className="hint">Source: {app.job?.source} · Match: {Math.round(app.job?.match_score ?? 0)}%</p>

        {app.adapted_cv ? (
          <div className="section">
            <h4>🤖 Adapted CV</h4>
            <pre className="adapted-cv">{app.adapted_cv}</pre>
            <button className="btn" onClick={() => navigator.clipboard.writeText(app.adapted_cv)}>
              📋 Copy
            </button>
          </div>
        ) : (
          <button className="btn btn-primary" onClick={() => onAdapt(app.id)}>
            🤖 Generate Adapted CV
          </button>
        )}
      </div>
    </div>
  )
}
