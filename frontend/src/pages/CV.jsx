import { useState, useEffect } from 'react'
import { getCV, uploadCV } from '../api/client'

export default function CVPage() {
  const [cv, setCV] = useState(null)
  const [text, setText] = useState('')
  const [loading, setLoading] = useState(false)
  const [msg, setMsg] = useState('')

  useEffect(() => {
    getCV().then(setCV).catch(() => {})
  }, [])

  const handleUpload = async () => {
    if (!text.trim()) return
    setLoading(true)
    setMsg('')
    try {
      const saved = await uploadCV(text)
      setCV(saved)
      setMsg('✅ CV uploaded and skills parsed!')
    } catch (e) {
      setMsg('❌ ' + e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <h2>Your CV</h2>

      {cv ? (
        <div className="card">
          <div className="card-header">
            <span>{cv.file_name}</span>
            <span className="badge">
              {cv.parsed_skills?.length ?? 0} skills detected
            </span>
          </div>

          {cv.parsed_skills?.length > 0 && (
            <div className="skills">
              {cv.parsed_skills.map(s => (
                <span key={s} className="skill-tag">{s}</span>
              ))}
            </div>
          )}

          <details className="cv-preview">
            <summary>View raw text</summary>
            <pre>{cv.content}</pre>
          </details>
        </div>
      ) : (
        <p className="hint">No CV uploaded yet.</p>
      )}

      <div className="section">
        <h3>{cv ? 'Replace CV' : 'Upload CV'}</h3>
        <p className="hint">Paste your CV text below (PDF text extraction coming soon)</p>
        <textarea
          className="cv-textarea"
          value={text}
          onChange={e => setText(e.target.value)}
          placeholder="Paste your CV here..."
          rows={15}
        />
        <button
          className="btn btn-primary"
          onClick={handleUpload}
          disabled={loading || !text.trim()}
        >
          {loading ? 'Analyzing...' : '⬆ Upload & Analyze'}
        </button>
        {msg && <p className="msg">{msg}</p>}
      </div>
    </div>
  )
}
