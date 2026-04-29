package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ruslan/job-hunter/internal/adapter"
	"github.com/ruslan/job-hunter/internal/scraper"
	"github.com/ruslan/job-hunter/internal/storage"
)

type Server struct {
	jobs    *storage.JobRepo
	apps    *storage.ApplicationRepo
	cvs     *storage.CVRepo
	scraper *scraper.Manager
	adapter *adapter.CVAdapter
	log     *slog.Logger
}

func NewServer(
	jobs *storage.JobRepo,
	apps *storage.ApplicationRepo,
	cvs *storage.CVRepo,
	scraper *scraper.Manager,
	adapter *adapter.CVAdapter,
	log *slog.Logger,
) http.Handler {
	s := &Server{jobs: jobs, apps: apps, cvs: cvs,
		scraper: scraper, adapter: adapter, log: log}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors)

	r.Route("/api", func(r chi.Router) {
		// CV
		r.Post("/cv/upload", s.uploadCV)
		r.Get("/cv", s.getCV)

		// Jobs
		r.Get("/jobs", s.listJobs)
		r.Get("/jobs/{id}", s.getJob)
		r.Post("/jobs/scrape", s.scrapeJobs)

		// Applications
		r.Get("/applications", s.listApplications)
		r.Post("/applications", s.createApplication)
		r.Get("/applications/{id}", s.getApplication)
		r.Patch("/applications/{id}/status", s.updateStatus)
		r.Post("/applications/{id}/adapt-cv", s.adaptCV)
	})

	return r
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
