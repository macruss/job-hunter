package api

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ruslan/job-hunter/internal/adapter"
	"github.com/ruslan/job-hunter/internal/domain"
)

// ── CV ────────────────────────────────────────────────────────────────────────

func (s *Server) uploadCV(w http.ResponseWriter, r *http.Request) {
	// Accept plain text or multipart form with a "cv" field
	r.ParseMultipartForm(5 << 20) // 5 MB

	var content, fileName string

	f, header, err := r.FormFile("cv")
	if err == nil {
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			writeError(w, http.StatusBadRequest, "cannot read file")
			return
		}
		fileName = header.Filename
		if strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
			content, err = adapter.ExtractPDFText(data)
			if err != nil {
				writeError(w, http.StatusBadRequest, "cannot parse PDF: "+err.Error())
				return
			}
		} else {
			content = string(data)
		}
	} else {
		// fallback: raw body as text
		data, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "cannot read body")
			return
		}
		content = string(data)
		fileName = "cv.txt"
	}

	if strings.TrimSpace(content) == "" {
		writeError(w, http.StatusBadRequest, "empty CV content")
		return
	}

	skills, err := s.adapter.ParseSkills(r.Context(), content)
	if err != nil {
		s.log.Warn("skill parse failed", "err", err)
		skills = []string{}
	}

	cv := domain.CV{
		ID:           uuid.New(),
		FileName:     fileName,
		Content:      content,
		ParsedSkills: skills,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.cvs.Save(r.Context(), cv); err != nil {
		s.log.Error("save cv", "err", err)
		writeError(w, http.StatusInternalServerError, "save failed")
		return
	}
	writeJSON(w, http.StatusCreated, cv)
}

func (s *Server) getCV(w http.ResponseWriter, r *http.Request) {
	cv, err := s.cvs.Latest(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cv == nil {
		writeError(w, http.StatusNotFound, "no CV uploaded yet")
		return
	}
	writeJSON(w, http.StatusOK, cv)
}

// ── Jobs ──────────────────────────────────────────────────────────────────────

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	source := q.Get("source")
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	jobs, err := s.jobs.List(r.Context(), source, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	job, err := s.jobs.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// scrapeJobs triggers on-demand scraping for given keywords.
func (s *Server) scrapeJobs(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Keywords []string `json:"keywords"`
		Remote   bool     `json:"remote"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if len(body.Keywords) == 0 {
		writeError(w, http.StatusBadRequest, "keywords required")
		return
	}

	filter := domain.ScrapeFilter{Keywords: body.Keywords, Remote: body.Remote}

	// Load latest CV for scoring
	cv, _ := s.cvs.Latest(r.Context())

	jobs := s.scraper.ScrapeAll(r.Context(), filter)
	saved := 0
	for i := range jobs {
		if cv != nil {
			score, _, err := s.adapter.ScoreMatch(r.Context(), *cv, jobs[i])
			if err == nil {
				jobs[i].MatchScore = score
			}
		}
		if err := s.jobs.Upsert(r.Context(), jobs[i]); err != nil {
			s.log.Warn("upsert job failed", "err", err)
			continue
		}
		saved++
	}

	writeJSON(w, http.StatusOK, map[string]int{
		"scraped": len(jobs),
		"saved":   saved,
	})
}
