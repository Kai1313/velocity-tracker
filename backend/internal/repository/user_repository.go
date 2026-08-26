package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/model"
)

type UserRepository struct {
	db dbtx
}

func NewUserRepository(db dbtx) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, role) VALUES ($1, $2) RETURNING id`,
		u.Name, u.Role,
	).Scan(&u.ID)
	return wrapMutationErr(err, false)
}

func (r *UserRepository) Get(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, role FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Role)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	return u, nil
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, role FROM users ORDER BY LOWER(name)`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Role); err != nil {
			return nil, wrapReadErr(err)
		}
		users = append(users, u)
	}
	return users, wrapReadErr(rows.Err())
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET name = $1, role = $2 WHERE id = $3`,
		u.Name, u.Role, u.ID,
	)
	if err != nil {
		return wrapMutationErr(err, false)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return wrapMutationErr(err, true)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}
