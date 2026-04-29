package storage

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/ruslan/job-hunter/internal/domain"
)

type CVRepo struct{ db *DB }

func NewCVRepo(db *DB) *CVRepo { return &CVRepo{db} }

func (r *CVRepo) Save(ctx context.Context, cv domain.CV) error {
	const q = `
		INSERT INTO cvs (id, file_name, content, parsed_skills, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := r.db.ExecContext(ctx, q,
		cv.ID, cv.FileName, cv.Content,
		pq.Array(cv.ParsedSkills), cv.CreatedAt, cv.UpdatedAt,
	)
	return err
}

// Latest returns the most recently uploaded CV.
func (r *CVRepo) Latest(ctx context.Context) (*domain.CV, error) {
	const q = `
		SELECT id, file_name, content, parsed_skills, created_at, updated_at
		FROM cvs ORDER BY created_at DESC LIMIT 1`
	var cv domain.CV
	err := r.db.QueryRowContext(ctx, q).Scan(
		&cv.ID, &cv.FileName, &cv.Content,
		pq.Array(&cv.ParsedSkills), &cv.CreatedAt, &cv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &cv, err
}
