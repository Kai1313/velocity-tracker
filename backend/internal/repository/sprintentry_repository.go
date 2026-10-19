package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
)

// SprintEntryFilter narrows List to matching rows; a nil field means that
// dimension is unfiltered. ProjectID and Search both require joining ticket,
// so the query only adds that join when at least one of them is set.
type SprintEntryFilter struct {
	SprintID        *int64
	ProjectID       *int64
	Status          *model.EntryStatus
	CarriedOverOnly bool
	Search          *string
}

type SprintEntryRepository struct {
	db dbtx
}

func NewSprintEntryRepository(db dbtx) *SprintEntryRepository {
	return &SprintEntryRepository{db: db}
}

func (r *SprintEntryRepository) Create(ctx context.Context, e *model.SprintEntry) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO sprint_entry (ticket_id, sprint_id, status, added_after_sprint_start, carried_from, points_at_entry)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, bounce_count`,
		e.TicketID, e.SprintID, e.Status, e.AddedAfterSprintStart, e.CarriedFrom, e.PointsAtEntry,
	).Scan(&e.ID, &e.CreatedAt, &e.BounceCount)
	return wrapMutationErr(err, false)
}

func (r *SprintEntryRepository) Get(ctx context.Context, id int64) (*model.SprintEntry, error) {
	e := &model.SprintEntry{}
	err := r.db.QueryRow(ctx,
		`SELECT id, ticket_id, sprint_id, status, added_after_sprint_start, carried_from, points_at_entry, created_at, bounce_count
		 FROM sprint_entry WHERE id = $1`, id,
	).Scan(&e.ID, &e.TicketID, &e.SprintID, &e.Status, &e.AddedAfterSprintStart, &e.CarriedFrom, &e.PointsAtEntry, &e.CreatedAt, &e.BounceCount)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	return e, nil
}

func (r *SprintEntryRepository) List(ctx context.Context, f SprintEntryFilter) ([]model.SprintEntry, error) {
	query := `SELECT se.id, se.ticket_id, se.sprint_id, se.status, se.added_after_sprint_start, se.carried_from, se.points_at_entry, se.created_at, se.bounce_count
		 FROM sprint_entry se`

	needsTicketJoin := f.ProjectID != nil || f.Search != nil
	if needsTicketJoin {
		query += ` JOIN ticket t ON t.id = se.ticket_id`
	}

	var conditions []string
	var args []any
	if f.SprintID != nil {
		args = append(args, *f.SprintID)
		conditions = append(conditions, fmt.Sprintf("se.sprint_id = $%d", len(args)))
	}
	if f.ProjectID != nil {
		args = append(args, *f.ProjectID)
		conditions = append(conditions, fmt.Sprintf("t.project_id = $%d", len(args)))
	}
	if f.Status != nil {
		args = append(args, *f.Status)
		conditions = append(conditions, fmt.Sprintf("se.status = $%d", len(args)))
	}
	if f.CarriedOverOnly {
		conditions = append(conditions, "se.carried_from IS NOT NULL")
	}
	if f.Search != nil && *f.Search != "" {
		args = append(args, "%"+*f.Search+"%")
		conditions = append(conditions, fmt.Sprintf("t.title ILIKE $%d", len(args)))
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY se.id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	entries := []model.SprintEntry{}
	for rows.Next() {
		var e model.SprintEntry
		if err := rows.Scan(&e.ID, &e.TicketID, &e.SprintID, &e.Status, &e.AddedAfterSprintStart, &e.CarriedFrom, &e.PointsAtEntry, &e.CreatedAt, &e.BounceCount); err != nil {
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

// UpdateBounceCount sets the entry's bounce count to an absolute value rather
// than incrementing, so a retried request is harmless. The write is itself
// conditional on the entry's sprint being Open, so a sprint closing between
// the service's check and this statement cannot let a change through. A
// refused write is a conflict if the entry exists, not found if it doesn't.
func (r *SprintEntryRepository) UpdateBounceCount(ctx context.Context, id int64, count int) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE sprint_entry SET bounce_count = $1
		 WHERE id = $2 AND sprint_id IN (SELECT id FROM sprint WHERE status = 'Open')`,
		count, id,
	)
	if err != nil {
		return wrapMutationErr(err, false)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM sprint_entry WHERE id = $1)`, id).Scan(&exists); err != nil {
		return wrapReadErr(err)
	}
	if !exists {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return fmt.Errorf("%w: entry %d's sprint is closed, its sprint entries are locked", apperr.ErrConflict, id)
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
