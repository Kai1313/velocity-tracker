package repository

import (
	"context"

	"velocity-tracker/backend/internal/model"
)

type DashboardRepository struct {
	db dbtx
}

func NewDashboardRepository(db dbtx) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// SprintSummaries returns workload/done totals per sprint, across all
// sprints, for the overview table. Cancelled entries are excluded from
// both totals per the domain rule in CONTEXT.md.
func (r *DashboardRepository) SprintSummaries(ctx context.Context) ([]model.SprintSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			s.id,
			s.name,
			COALESCE(SUM(se.points_at_entry), 0)::int AS workload_points,
			COALESCE(SUM(CASE WHEN se.status = 'Done' THEN se.points_at_entry ELSE 0 END), 0)::int AS done_points,
			COUNT(se.id)::int AS workload_tickets,
			COUNT(CASE WHEN se.status = 'Done' THEN 1 END)::int AS done_tickets
		FROM sprint s
		LEFT JOIN sprint_entry se ON se.sprint_id = s.id AND se.status <> 'Cancelled'
		GROUP BY s.id, s.name
		ORDER BY s.id
	`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	summaries := []model.SprintSummary{}
	for rows.Next() {
		var s model.SprintSummary
		if err := rows.Scan(&s.SprintID, &s.SprintName, &s.WorkloadPoints, &s.DonePoints, &s.WorkloadTickets, &s.DoneTickets); err != nil {
			return nil, wrapReadErr(err)
		}
		summaries = append(summaries, s)
	}
	return summaries, wrapReadErr(rows.Err())
}

// DeveloperBreakdown returns workload/done totals per developer within a
// single sprint. Tickets with no assignee are grouped under a nil UserID
// ("Unassigned") rather than dropped, so totals reconcile with SprintSummaries.
func (r *DashboardRepository) DeveloperBreakdown(ctx context.Context, sprintID int64) ([]model.DeveloperSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			u.id,
			COALESCE(u.name, 'Unassigned') AS name,
			COALESCE(SUM(se.points_at_entry), 0)::int AS workload_points,
			COALESCE(SUM(CASE WHEN se.status = 'Done' THEN se.points_at_entry ELSE 0 END), 0)::int AS done_points,
			COUNT(se.id)::int AS workload_tickets,
			COUNT(CASE WHEN se.status = 'Done' THEN 1 END)::int AS done_tickets
		FROM sprint_entry se
		JOIN ticket t ON t.id = se.ticket_id
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE se.sprint_id = $1 AND se.status <> 'Cancelled'
		GROUP BY u.id, u.name
		ORDER BY LOWER(name)
	`, sprintID)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	breakdown := []model.DeveloperSummary{}
	for rows.Next() {
		var d model.DeveloperSummary
		if err := rows.Scan(&d.UserID, &d.Name, &d.WorkloadPoints, &d.DonePoints, &d.WorkloadTickets, &d.DoneTickets); err != nil {
			return nil, wrapReadErr(err)
		}
		breakdown = append(breakdown, d)
	}
	return breakdown, wrapReadErr(rows.Err())
}

// TicketEntries returns every SprintEntry in a single sprint, joined with
// the ticket/project/assignee names and (when the entry was carried over)
// the name of the one prior sprint it was carried from. Cancelled entries
// are included — unlike the aggregate summaries above, this is a plain
// listing of what's in the sprint, not a metrics computation.
func (r *DashboardRepository) TicketEntries(ctx context.Context, sprintID int64) ([]model.SprintEntryDetail, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			se.id,
			se.ticket_id,
			t.title,
			p.name,
			COALESCE(u.name, 'Unassigned') AS assignee_name,
			se.status,
			se.added_after_sprint_start,
			se.points_at_entry,
			origin_sprint.name AS carried_from_sprint_name
		FROM sprint_entry se
		JOIN ticket t ON t.id = se.ticket_id
		JOIN project p ON p.id = t.project_id
		LEFT JOIN users u ON u.id = t.assignee_id
		LEFT JOIN sprint_entry origin_se ON origin_se.id = se.carried_from
		LEFT JOIN sprint origin_sprint ON origin_sprint.id = origin_se.sprint_id
		WHERE se.sprint_id = $1
		ORDER BY LOWER(t.title)
	`, sprintID)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	entries := []model.SprintEntryDetail{}
	for rows.Next() {
		var e model.SprintEntryDetail
		if err := rows.Scan(&e.EntryID, &e.TicketID, &e.TicketTitle, &e.ProjectName, &e.AssigneeName, &e.Status, &e.AddedAfterSprintStart, &e.PointsAtEntry, &e.CarriedFromSprintName); err != nil {
			return nil, wrapReadErr(err)
		}
		entries = append(entries, e)
	}
	return entries, wrapReadErr(rows.Err())
}
