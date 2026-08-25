package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/model"
)

type SprintRepository struct {
	db dbtx
}

func NewSprintRepository(db dbtx) *SprintRepository {
	return &SprintRepository{db: db}
}

func (r *SprintRepository) Create(ctx context.Context, s *model.Sprint) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO sprint (name, start_date, end_date, status) VALUES ($1, $2, $3, $4) RETURNING id`,
		s.Name, s.StartDate, s.EndDate, s.Status,
	).Scan(&s.ID)
	return wrapMutationErr(err, false)
}

func (r *SprintRepository) Get(ctx context.Context, id int64) (*model.Sprint, error) {
	s := &model.Sprint{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, start_date, end_date, status FROM sprint WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.Status)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	return s, nil
}

func (r *SprintRepository) List(ctx context.Context) ([]model.Sprint, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, start_date, end_date, status FROM sprint ORDER BY id`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	sprints := []model.Sprint{}
	for rows.Next() {
		var s model.Sprint
		if err := rows.Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.Status); err != nil {
			return nil, wrapReadErr(err)
		}
		sprints = append(sprints, s)
	}
	return sprints, wrapReadErr(rows.Err())
}

func (r *SprintRepository) Update(ctx context.Context, s *model.Sprint) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE sprint SET name = $1, start_date = $2, end_date = $3, status = $4 WHERE id = $5`,
		s.Name, s.StartDate, s.EndDate, s.Status, s.ID,
	)
	if err != nil {
		return wrapMutationErr(err, false)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}

func (r *SprintRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM sprint WHERE id = $1`, id)
	if err != nil {
		return wrapMutationErr(err, true)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}
