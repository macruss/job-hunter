const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

async function req(method, path, body) {
  const opts = { method, headers: { 'Content-Type': 'application/json' } }
  if (body) opts.body = JSON.stringify(body)
  const res = await fetch(BASE + path, opts)
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}

// CV
export const uploadCV = (text) =>
  fetch(BASE + '/cv/upload', { method: 'POST', body: text }).then(r => r.json())

export const getCV = () => req('GET', '/cv')

// Jobs
export const listJobs = (source = '', limit = 50, offset = 0) =>
  req('GET', `/jobs?source=${source}&limit=${limit}&offset=${offset}`)

export const scrapeJobs = (keywords, remote = true) =>
  req('POST', '/jobs/scrape', { keywords, remote })

// Applications
export const listApplications = () => req('GET', '/applications')

export const createApplication = (jobId) =>
  req('POST', '/applications', { job_id: jobId })

export const updateStatus = (id, status, notes = '') =>
  req('PATCH', `/applications/${id}/status`, { status, notes })

export const adaptCV = (id) => req('POST', `/applications/${id}/adapt-cv`)
