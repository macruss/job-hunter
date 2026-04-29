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
│  │  · Djinni   │  │ (Anthropic)│  │  /cv        │  │
│  │  · DOU      │  │            │  │  /jobs      │  │
│  │  · Work.ua  │  │ · AdaptCV  │  │  /apps      │  │
│  │  · LinkedIn │  │ · ScoreMatch│  │             │  │
│  └──────────────┘  │ · ParseSkills│ └─────────────┘  │
│                    └────────────┘                    │
└──────────────────────────┬──────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────┐
│                  PostgreSQL                          │
│  cvs │ jobs (source, external_id, match_score)       │
│  applications (status, adapted_cv, notes)            │
└─────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Set up environment

```bash
cp .env.example .env
# fill in ANTHROPIC_API_KEY
# optionally: LINKEDIN_SESSION_COOKIE (your li_at cookie)
```

### 2. Start database

```bash
make db
```

### 3. Run backend

```bash
make dev
```

### 4. Run frontend

```bash
make ui
# opens at http://localhost:3000
```

## Workflow

1. **Upload CV** → paste text in the CV tab. AI extracts skills automatically.
2. **Scrape jobs** → enter keywords (e.g. "Go backend gRPC"), click Scrape.
   - Searches Djinni, DOU, Work.ua (+ LinkedIn if cookie set)
   - Each job gets an AI match score 0–100 against your CV
3. **Browse jobs** → sorted by match score, filter by source
4. **Track** → click "+ Track Application" on any job
5. **Adapt CV** → in Applications tracker, click "🤖 Adapt CV" — Claude tailors your CV for that specific role
6. **Track progress** → move cards through kanban: New → Applied → Screening → Interview → Offer/Rejected

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/cv/upload` | Upload CV text |
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
| `ANTHROPIC_API_KEY` | yes | For CV scoring and adaptation |
| `LINKEDIN_SESSION_COOKIE` | no | Your `li_at` cookie for LinkedIn |
| `PORT` | no | HTTP port (default 8080) |

## Next Steps

- [ ] PDF upload support (use `pdfcpu` or `unipdf` to extract text)
- [ ] Scheduled auto-scraping (cron via `robfig/cron`)
- [ ] Email/Telegram notifications for new high-score jobs
- [ ] Cover letter generation
- [ ] Export adapted CV to `.docx`
