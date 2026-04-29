import { useState } from 'react'
import Jobs from './pages/Jobs'
import Applications from './pages/Applications'
import CVPage from './pages/CV'
import './index.css'

const TABS = [
  { id: 'cv',           label: '📄 CV' },
  { id: 'jobs',         label: '🔍 Jobs' },
  { id: 'applications', label: '📋 Applications' },
]

export default function App() {
  const [tab, setTab] = useState('cv')

  return (
    <div className="app">
      <header className="header">
        <h1>🎯 Job Hunter</h1>
        <nav className="tabs">
          {TABS.map(t => (
            <button
              key={t.id}
              className={`tab ${tab === t.id ? 'active' : ''}`}
              onClick={() => setTab(t.id)}
            >
              {t.label}
            </button>
          ))}
        </nav>
      </header>

      <main className="main">
        {tab === 'cv'           && <CVPage />}
        {tab === 'jobs'         && <Jobs />}
        {tab === 'applications' && <Applications />}
      </main>
    </div>
  )
}
