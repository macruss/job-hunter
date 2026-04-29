# Job Hunter

Personal job search automation: scrape → AI match scoring → CV adaptation → application tracking.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    React UI (port 3000)              │
│  CV Upload │ Jobs (scrape + filter) │ Kanban tracker │
└───────────────────────┬─────────────────────────────┘
                        │ REST /api/*
┌───────────────────────▼─────────────────────────────┐
│              Go HTTP Server (port 8080)              │
│                                                      │
│  ┌──────────────┐  ┌────────────┐  ┌─────────────┐  │
│  │ Scraper Mgr  │  │ CV Adapter │  │  API Routes │  │
│  │  · Djinni   │  │  (Ollama)  │  │  /cv        │  │
│  │  · DOU      │  │            │  │  /jobs      │  │
│  │  · Work.ua  │  │ · AdaptCV  │  │  /apps      │  │
│  │  · LinkedIn │  │ · ScoreMatch│  │             │  │
│  └──────────────┘  │ · ParseSkills│ └─────────────┘  │
│                    └────────────┘                    │
└──────────────────────────┬──────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────┐
│              Ollama (port 11434)                     │
│  local LLM · default model: llama3.2                 │
└──────────────────────────┬──────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────┐
│                  PostgreSQL                          │
│  cvs │ jobs (source, external_id, match_score)       │
│  applications (status, adapted_cv, notes)            │
└─────────────────────────────────────────────────────┘
```

## Quick Start

### Docker (recommended)

```bash
docker compose up --build
# backend → http://localhost:8080
# pull the LLM model on first run:
docker compose exec ollama ollama pull llama3.2
```

### Local dev

```bash
# 1. start postgres + ollama
make db
docker compose up -d ollama

# 2. pull the model
docker compose exec ollama ollama pull llama3.2

# 3. run backend
make dev

# 4. run frontend
make ui   # http://localhost:3000
```

## Workflow

1. **Upload CV** → upload a `.pdf` or plain text file in the CV tab. Skills extracted automatically.
2. **Scrape jobs** → enter keywords (e.g. "Go backend gRPC"), click Scrape.
   - Searches Djinni, DOU, Work.ua (+ LinkedIn if cookie set)
   - Each job gets an AI match score 0–100 against your CV
3. **Browse jobs** → sorted by match score, filter by source
4. **Track** → click "+ Track Application" on any job
5. **Adapt CV** → in Applications tracker, click "🤖 Adapt CV" — local LLM tailors your CV for that role
6. **Track progress** → move cards through kanban: New → Applied → Screening → Interview → Offer/Rejected

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/cv/upload` | Upload CV (`.pdf` or plain text) |
| GET | `/api/cv` | Get latest CV |
| GET | `/api/jobs?source=djinni&limit=50` | List jobs |
| POST | `/api/jobs/scrape` | Trigger scraping `{keywords, remote}` |
| GET | `/api/applications` | List all applications |
| POST | `/api/applications` | Create `{job_id}` |
| PATCH | `/api/applications/:id/status` | Update `{status, notes}` |
| POST | `/api/applications/:id/adapt-cv` | Generate adapted CV |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | yes | PostgreSQL DSN |
| `OLLAMA_URL` | no | Ollama base URL (default `http://localhost:11434`) |
| `OLLAMA_MODEL` | no | Model name (default `llama3.2`) |
| `LINKEDIN_SESSION_COOKIE` | no | Your `li_at` cookie for LinkedIn scraping |
| `PORT` | no | HTTP port (default `8080`) |

## Next Steps

- [ ] Scheduled auto-scraping (cron via `robfig/cron`)
- [ ] Email/Telegram notifications for new high-score jobs
- [ ] Cover letter generation
- [ ] Export adapted CV to `.docx`
