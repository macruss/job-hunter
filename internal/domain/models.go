package domain

import (
	"time"

	"github.com/google/uuid"
)

// ── Source ────────────────────────────────────────────────────────────────────

type Source string

const (
	SourceDjinni   Source = "djinni"
	SourceLinkedIn Source = "linkedin"
	SourceDOU      Source = "dou"
	SourceWorkUA   Source = "workua"
)

// ── Job ───────────────────────────────────────────────────────────────────────

type Job struct {
	ID          uuid.UUID `json:"id"          db:"id"`
	ExternalID  string    `json:"external_id" db:"external_id"`
	Source      Source    `json:"source"      db:"source"`
	Title       string    `json:"title"       db:"title"`
	Company     string    `json:"company"     db:"company"`
	Description string    `json:"description" db:"description"`
	URL         string    `json:"url"         db:"url"`
	Location    string    `json:"location"    db:"location"`
	SalaryMin   *int      `json:"salary_min"  db:"salary_min"`
	SalaryMax   *int      `json:"salary_max"  db:"salary_max"`
	Currency    string    `json:"currency"    db:"currency"`
	Tags        []string  `json:"tags"        db:"tags"`
	MatchScore  float64   `json:"match_score" db:"match_score"`
	ScrapedAt   time.Time `json:"scraped_at"  db:"scraped_at"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
}

// ── Application ───────────────────────────────────────────────────────────────

type ApplicationStatus string

const (
	StatusNew       ApplicationStatus = "new"
	StatusApplied   ApplicationStatus = "applied"
	StatusScreening ApplicationStatus = "screening"
	StatusInterview ApplicationStatus = "interview"
	StatusOffer     ApplicationStatus = "offer"
	StatusRejected  ApplicationStatus = "rejected"
)

var AllStatuses = []ApplicationStatus{
	StatusNew, StatusApplied, StatusScreening,
	StatusInterview, StatusOffer, StatusRejected,
}

type Application struct {
	ID        uuid.UUID         `json:"id"         db:"id"`
	JobID     uuid.UUID         `json:"job_id"     db:"job_id"`
	CVID      *uuid.UUID        `json:"cv_id"      db:"cv_id"`
	AdaptedCV string            `json:"adapted_cv" db:"adapted_cv"`
	Status    ApplicationStatus `json:"status"     db:"status"`
	Notes     string            `json:"notes"      db:"notes"`
	AppliedAt *time.Time        `json:"applied_at" db:"applied_at"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"`

	Job *Job `json:"job,omitempty"` // eager join
}

// ── CV ────────────────────────────────────────────────────────────────────────

type CV struct {
	ID           uuid.UUID `json:"id"            db:"id"`
	FileName     string    `json:"file_name"     db:"file_name"`
	Content      string    `json:"content"       db:"content"`
	ParsedSkills []string  `json:"parsed_skills" db:"parsed_skills"`
	CreatedAt    time.Time `json:"created_at"    db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"    db:"updated_at"`
}

// ── Scrape filter ─────────────────────────────────────────────────────────────

type ScrapeFilter struct {
	Keywords []string
	Location string
	Remote   bool
}
