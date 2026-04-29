import { useState, useEffect } from 'react'
import { listJobs, scrapeJobs, createApplication } from '../api/client'

const SOURCES = ['', 'djinni', 'linkedin', 'dou', 'workua']

export default function Jobs() {
  const [jobs, setJobs] = useState([])
  const [source, setSource] = useState('')
  const [keywords, setKeywords] = useState('Golang Go backend')
  const [loading, setLoading] = useState(false)
  const [scraping, setScraping] = useState(false)
  const [msg, setMsg] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      const data = await listJobs(source)
      setJobs(data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [source])

  const handleScrape = async () => {
    const kws = keywords.split(/[\s,]+/).filter(Boolean)
    if (!kws.length) return
    setScraping(true)
    setMsg('')
    try {
      const res = await scrapeJobs(kws, true)
      setMsg(`✅ Scraped ${res.scraped}, saved ${res.saved} jobs (with AI match scores)`)
      load()
    } catch (e) {
      setMsg('❌ ' + e.message)
    } finally {
      setScraping(false)
    }
  }

  const handleApply = async (jobId) => {
    try {
      await createApplication(jobId)
      setMsg('✅ Added to applications tracker!')
    } catch (e) {
      setMsg('❌ ' + e.message)
    }
  }

  return (
    <div className="page">
      <h2>Job Search</h2>

      <div className="scrape-bar card">
        <input
          className="input"
          value={keywords}
          onChange={e => setKeywords(e.target.value)}
          placeholder="Keywords (e.g. Go backend gRPC)"
        />
        <button
          className="btn btn-primary"
          onClick={handleScrape}
          disabled={scraping}
        >
          {scraping ? '⏳ Scraping...' : '🔍 Scrape Jobs'}
        </button>
        {msg && <span className="msg">{msg}</span>}
      </div>

      <div className="filter-bar">
        <span>Filter by source:</span>
        {SOURCES.map(s => (
          <button
            key={s || 'all'}
            className={`tab-sm ${source === s ? 'active' : ''}`}
            onClick={() => setSource(s)}
          >
            {s || 'All'}
          </button>
        ))}
      </div>

      {loading ? (
        <p>Loading...</p>
      ) : jobs.length === 0 ? (
        <p className="hint">No jobs yet. Click "Scrape Jobs" to search.</p>
      ) : (
        <div className="job-list">
          {jobs.map(job => (
            <div key={job.id} className="job-card card">
              <div className="job-header">
                <div>
                  <h3>
                    <a href={job.url} target="_blank" rel="noreferrer">{job.title}</a>
                  </h3>
                  <span className="company">{job.company}</span>
                  {job.location && <span className="location"> · {job.location}</span>}
                </div>
                <div className="job-meta">
                  <MatchBadge score={job.match_score} />
                  <span className={`source-badge source-${job.source}`}>{job.source}</span>
                </div>
              </div>

              {job.tags?.length > 0 && (
                <div className="skills">
                  {job.tags.map(t => <span key={t} className="skill-tag">{t}</span>)}
                </div>
              )}

              {job.salary_min && (
                <p className="salary">
                  💰 {job.salary_min}–{job.salary_max} {job.currency}
                </p>
              )}

              <button
                className="btn btn-sm"
                onClick={() => handleApply(job.id)}
              >
                + Track Application
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function MatchBadge({ score }) {
  if (!score) return null
  const color = score >= 75 ? '#22c55e' : score >= 50 ? '#f59e0b' : '#ef4444'
  return (
    <span className="match-badge" style={{ background: color }}>
      {Math.round(score)}% match
    </span>
  )
}
