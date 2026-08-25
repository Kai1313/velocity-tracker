package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/model"
)

type TicketRepository struct {
	db dbtx
}

func NewTicketRepository(db dbtx) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ctx context.Context, t *model.Ticket) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO ticket (project_id, title, story_points, assignee_id)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		t.ProjectID, t.Title, t.StoryPoints, t.AssigneeID,
	).Scan(&t.ID)
	return wrapMutationErr(err, false)
}

// currentStatusQuery joins each ticket to the status of its most recent
// SprintEntry (by created_at), per the domain rule that a ticket's status
// is never stored directly. LEFT JOIN so a ticket with no SprintEntry yet
// still returns, with CurrentStatus left nil.
const ticketDetailSelect = `
	SELECT t.id, t.project_id, t.title, t.story_points, t.assignee_id, se.status
	FROM ticket t
	LEFT JOIN LATERAL (
		SELECT status FROM sprint_entry
		WHERE ticket_id = t.id
		ORDER BY created_at DESC
		LIMIT 1
	) se ON true
`

func (r *TicketRepository) Get(ctx context.Context, id int64) (*model.TicketDetail, error) {
	d := &model.TicketDetail{}
	err := r.db.QueryRow(ctx, ticketDetailSelect+` WHERE t.id = $1`, id).Scan(
		&d.ID, &d.ProjectID, &d.Title, &d.StoryPoints, &d.AssigneeID, &d.CurrentStatus,
	)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	return d, nil
}

func (r *TicketRepository) List(ctx context.Context) ([]model.TicketDetail, error) {
	rows, err := r.db.Query(ctx, ticketDetailSelect+` ORDER BY t.id`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	tickets := []model.TicketDetail{}
	for rows.Next() {
		var d model.TicketDetail
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.StoryPoints, &d.AssigneeID, &d.CurrentStatus); err != nil {
			return nil, wrapReadErr(err)
		}
		tickets = append(tickets, d)
	}
	return tickets, wrapReadErr(rows.Err())
}

func (r *TicketRepository) Update(ctx context.Context, t *model.Ticket) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE ticket SET project_id = $1, title = $2, story_points = $3, assignee_id = $4 WHERE id = $5`,
		t.ProjectID, t.Title, t.StoryPoints, t.AssigneeID, t.ID,
	)
	if err != nil {
		return wrapMutationErr(err, false)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}

func (r *TicketRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM ticket WHERE id = $1`, id)
	if err != nil {
		return wrapMutationErr(err, true)
	}
	if tag.RowsAffected() == 0 {
		return wrapReadErr(pgx.ErrNoRows)
	}
	return nil
}
