package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ruslan/job-hunter/internal/domain"
)

func (s *Server) listApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := s.apps.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if apps == nil {
		apps = []domain.Application{}
	}
	writeJSON(w, http.StatusOK, apps)
}

func (s *Server) createApplication(w http.ResponseWriter, r *http.Request) {
	var body struct {
		JobID string `json:"job_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	jobID, err := uuid.Parse(body.JobID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job_id")
		return
	}

	cv, _ := s.cvs.Latest(r.Context())
	var cvID *uuid.UUID
	if cv != nil {
		cvID = &cv.ID
	}

	app := domain.Application{
		ID:        uuid.New(),
		JobID:     jobID,
		CVID:      cvID,
		Status:    domain.StatusNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.apps.Create(r.Context(), app); err != nil {
		s.log.Error("create application", "err", err)
		writeError(w, http.StatusInternalServerError, "create failed")
		return
	}
	writeJSON(w, http.StatusCreated, app)
}

func (s *Server) getApplication(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	app, err := s.apps.GetByID(r.Context(), id)
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var body struct {
		Status domain.ApplicationStatus `json:"status"`
		Notes  string                   `json:"notes"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := s.apps.UpdateStatus(r.Context(), id, body.Status, body.Notes); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(body.Status)})
}

// adaptCV calls Claude to tailor the CV for the job, then saves it.
func (s *Server) adaptCV(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	app, err := s.apps.GetByID(r.Context(), id)
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}

	cv, _ := s.cvs.Latest(r.Context())
	if cv == nil {
		writeError(w, http.StatusBadRequest, "no CV uploaded")
		return
	}

	adapted, err := s.adapter.AdaptCV(r.Context(), *cv, *app.Job)
	if err != nil {
		s.log.Error("adapt cv", "err", err)
		writeError(w, http.StatusInternalServerError, "adaptation failed")
		return
	}

	if err := s.apps.SetAdaptedCV(r.Context(), id, adapted); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"adapted_cv": adapted})
}

// ── util ──────────────────────────────────────────────────────────────────────

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
