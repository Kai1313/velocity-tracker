package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/model"
)

type SprintEntryRepository struct {
	db dbtx
}

func NewSprintEntryRepository(db dbtx) *SprintEntryRepository {
	return &SprintEntryRepository{db: db}
}

func (r *SprintEntryRepository) Create(ctx context.Context, e *model.SprintEntry) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO sprint_entry (ticket_id, sprint_id, status, added_after_sprint_start, carried_from, points_at_entry)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`,
		e.TicketID, e.SprintID, e.Status, e.AddedAfterSprintStart, e.CarriedFrom, e.PointsAtEntry,
	).Scan(&e.ID, &e.CreatedAt)
	return wrapMutationErr(err, false)
}

func (r *SprintEntryRepository) Get(ctx context.Context, id int64) (*model.SprintEntry, error) {
	e := &model.SprintEntry{}
	err := r.db.QueryRow(ctx,
		`SELECT id, ticket_id, sprint_id, status, added_after_sprint_start, carried_from, points_at_entry, created_at
		 FROM sprint_entry WHERE id = $1`, id,
	).Scan(&e.ID, &e.TicketID, &e.SprintID, &e.Status, &e.AddedAfterSprintStart, &e.CarriedFrom, &e.PointsAtEntry, &e.CreatedAt)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	return e, nil
}

func (r *SprintEntryRepository) List(ctx context.Context) ([]model.SprintEntry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, ticket_id, sprint_id, status, added_after_sprint_start, carried_from, points_at_entry, created_at
		 FROM sprint_entry ORDER BY id`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	entries := []model.SprintEntry{}
	for rows.Next() {
		var e model.SprintEntry
		if err := rows.Scan(&e.ID, &e.TicketID, &e.SprintID, &e.Status, &e.AddedAfterSprintStart, &e.CarriedFrom, &e.PointsAtEntry, &e.CreatedAt); err != nil {
			return nil, wrapReadErr(err)
		}
		entries = append(entries, e)
	}
	return entries, wrapReadErr(rows.Err())
}

func (r *SprintEntryRepository) Update(ctx context.Context, e *model.SprintEntry) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE sprint_entry SET status = $1, added_after_sprint_start = $2, carried_from = $3, points_at_entry = $4 WHERE id = $5`,
		e.Status, e.AddedAfterSprintStart, e.CarriedFrom, e.PointsAtEntry, e.ID,
	)
	if err != nil {
		return wrapMutationErr(err, false)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}

func (r *SprintEntryRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM sprint_entry WHERE id = $1`, id)
	if err != nil {
		return wrapMutationErr(err, true)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}
