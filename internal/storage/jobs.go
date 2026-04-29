package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/ruslan/job-hunter/internal/domain"
)

type JobRepo struct{ db *DB }

func NewJobRepo(db *DB) *JobRepo { return &JobRepo{db} }

// Upsert inserts a job or updates it if (source, external_id) already exists.
func (r *JobRepo) Upsert(ctx context.Context, j domain.Job) error {
	const q = `
		INSERT INTO jobs
			(id, external_id, source, title, company, description, url,
			 location, salary_min, salary_max, currency, tags, match_score, scraped_at, created_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (source, external_id)
		DO UPDATE SET
			title        = EXCLUDED.title,
			company      = EXCLUDED.company,
			description  = EXCLUDED.description,
			match_score  = EXCLUDED.match_score,
			scraped_at   = EXCLUDED.scraped_at`

	_, err := r.db.ExecContext(ctx, q,
		j.ID, j.ExternalID, j.Source,
		j.Title, j.Company, j.Description, j.URL,
		j.Location, j.SalaryMin, j.SalaryMax, j.Currency,
		pq.Array(j.Tags), j.MatchScore, j.ScrapedAt, j.CreatedAt,
	)
	return err
}

// List returns jobs ordered by match_score desc with optional source filter.
func (r *JobRepo) List(ctx context.Context, source string, limit, offset int) ([]domain.Job, error) {
	q := `SELECT id, external_id, source, title, company, description, url,
	             location, salary_min, salary_max, currency, tags, match_score,
	             scraped_at, created_at
	      FROM jobs`
	args := []any{}
	if source != "" {
		q += ` WHERE source = $1`
		args = append(args, source)
		q += fmt.Sprintf(` ORDER BY match_score DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	} else {
		q += fmt.Sprintf(` ORDER BY match_score DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	}
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func (r *JobRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	const q = `SELECT id, external_id, source, title, company, description, url,
	                  location, salary_min, salary_max, currency, tags, match_score,
	                  scraped_at, created_at
	           FROM jobs WHERE id = $1`
	rows, err := r.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs, err := scanJobs(rows)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, sql.ErrNoRows
	}
	return &jobs[0], nil
}

func scanJobs(rows *sql.Rows) ([]domain.Job, error) {
	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		err := rows.Scan(
			&j.ID, &j.ExternalID, &j.Source,
			&j.Title, &j.Company, &j.Description, &j.URL,
			&j.Location, &j.SalaryMin, &j.SalaryMax, &j.Currency,
			pq.Array(&j.Tags), &j.MatchScore,
			&j.ScrapedAt, &j.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}
