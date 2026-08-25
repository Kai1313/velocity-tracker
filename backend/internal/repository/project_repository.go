package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/model"
)

type ProjectRepository struct {
	db dbtx
}

func NewProjectRepository(db dbtx) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *model.Project) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO project (name, status) VALUES ($1, $2) RETURNING id`,
		p.Name, p.Status,
	).Scan(&p.ID)
	return wrapMutationErr(err, false)
}

func (r *ProjectRepository) Get(ctx context.Context, id int64) (*model.Project, error) {
	p := &model.Project{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, status FROM project WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Status)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	return p, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]model.Project, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, status FROM project ORDER BY id`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	projects := []model.Project{}
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Status); err != nil {
			return nil, wrapReadErr(err)
		}
		projects = append(projects, p)
	}
	return projects, wrapReadErr(rows.Err())
}

func (r *ProjectRepository) Update(ctx context.Context, p *model.Project) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE project SET name = $1, status = $2 WHERE id = $3`,
		p.Name, p.Status, p.ID,
	)
	if err != nil {
		return wrapMutationErr(err, false)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM project WHERE id = $1`, id)
	if err != nil {
		return wrapMutationErr(err, true)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}
