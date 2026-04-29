package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/ruslan/job-hunter/internal/domain"
)

type ApplicationRepo struct{ db *DB }

func NewApplicationRepo(db *DB) *ApplicationRepo { return &ApplicationRepo{db} }

func (r *ApplicationRepo) Create(ctx context.Context, a domain.Application) error {
	const q = `
		INSERT INTO applications (id, job_id, cv_id, adapted_cv, status, notes, applied_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.ExecContext(ctx, q,
		a.ID, a.JobID, a.CVID, a.AdaptedCV, a.Status,
		a.Notes, a.AppliedAt, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *ApplicationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, notes string) error {
	const q = `UPDATE applications SET status=$2, notes=$3, updated_at=$4 WHERE id=$1`
	_, err := r.db.ExecContext(ctx, q, id, status, notes, time.Now())
	return err
}

func (r *ApplicationRepo) SetAdaptedCV(ctx context.Context, id uuid.UUID, adaptedCV string) error {
	const q = `UPDATE applications SET adapted_cv=$2, updated_at=$3 WHERE id=$1`
	_, err := r.db.ExecContext(ctx, q, id, adaptedCV, time.Now())
	return err
}

func (r *ApplicationRepo) List(ctx context.Context) ([]domain.Application, error) {
	const q = `
		SELECT a.id, a.job_id, a.cv_id, a.adapted_cv, a.status, a.notes,
		       a.applied_at, a.created_at, a.updated_at,
		       j.title, j.company, j.url, j.source, j.match_score
		FROM applications a
		JOIN jobs j ON j.id = a.job_id
		ORDER BY a.updated_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []domain.Application
	for rows.Next() {
		var a domain.Application
		a.Job = &domain.Job{}
		err := rows.Scan(
			&a.ID, &a.JobID, &a.CVID, &a.AdaptedCV, &a.Status, &a.Notes,
			&a.AppliedAt, &a.CreatedAt, &a.UpdatedAt,
			&a.Job.Title, &a.Job.Company, &a.Job.URL, &a.Job.Source, &a.Job.MatchScore,
		)
		if err != nil {
			return nil, err
		}
		a.Job.ID = a.JobID
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func (r *ApplicationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	const q = `
		SELECT a.id, a.job_id, a.cv_id, a.adapted_cv, a.status, a.notes,
		       a.applied_at, a.created_at, a.updated_at,
		       j.id, j.title, j.company, j.url, j.source, j.description, j.match_score
		FROM applications a
		JOIN jobs j ON j.id = a.job_id
		WHERE a.id = $1`

	var a domain.Application
	a.Job = &domain.Job{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&a.ID, &a.JobID, &a.CVID, &a.AdaptedCV, &a.Status, &a.Notes,
		&a.AppliedAt, &a.CreatedAt, &a.UpdatedAt,
		&a.Job.ID, &a.Job.Title, &a.Job.Company, &a.Job.URL,
		&a.Job.Source, &a.Job.Description, &a.Job.MatchScore,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}
