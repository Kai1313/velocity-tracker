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
			COUNT(CASE WHEN se.status = 'Done' THEN 1 END)::int AS done_tickets,
			COALESCE(SUM(se.bounce_count), 0)::int AS bounces,
			COUNT(se.bounce_count)::int AS bounce_tracked_tickets
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
		if err := rows.Scan(&d.UserID, &d.Name, &d.WorkloadPoints, &d.DonePoints, &d.WorkloadTickets, &d.DoneTickets, &d.Bounces, &d.BounceTrackedTickets); err != nil {
			return nil, wrapReadErr(err)
		}
		breakdown = append(breakdown, d)
	}
	return breakdown, wrapReadErr(rows.Err())
}

// earlierBouncesSQL sums a ticket's bounces on its non-Cancelled entries in
// sprints created before `se`'s sprint (sprint ID order, the same recency
// convention as the rest of this file). Untracked (NULL) entries add nothing.
const earlierBouncesSQL = `(
	SELECT COALESCE(SUM(x.bounce_count), 0)::int
	FROM sprint_entry x
	WHERE x.ticket_id = se.ticket_id AND x.sprint_id < se.sprint_id AND x.status <> 'Cancelled'
)`

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
			origin_sprint.name AS carried_from_sprint_name,
			se.bounce_count
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
		if err := rows.Scan(&e.EntryID, &e.TicketID, &e.TicketTitle, &e.ProjectName, &e.AssigneeName, &e.Status, &e.AddedAfterSprintStart, &e.PointsAtEntry, &e.CarriedFromSprintName, &e.BounceCount); err != nil {
			return nil, wrapReadErr(err)
		}
		entries = append(entries, e)
	}
	return entries, wrapReadErr(rows.Err())
}

// SprintPoints returns committed/done/late-add point totals per currently
// Open sprint, summed across every Active project's tickets in it —
// developer capacity is shared across a team's sub-projects, so this
// deliberately does not break totals out per project (see ADR-0011).
// Cancelled entries are excluded, matching SprintSummaries/DeveloperBreakdown.
// Using an inner join (not the LEFT JOIN SprintSummaries uses) is
// deliberate: a sprint with no qualifying entries has nothing to report and
// should not appear as a zero row.
func (r *DashboardRepository) SprintPoints(ctx context.Context) ([]model.SprintPoints, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			s.id,
			s.name,
			s.start_date,
			s.end_date,
			COALESCE(SUM(CASE WHEN NOT se.added_after_sprint_start THEN se.points_at_entry ELSE 0 END), 0)::int AS committed_points,
			COALESCE(SUM(CASE WHEN NOT se.added_after_sprint_start AND se.status = 'Done' THEN se.points_at_entry ELSE 0 END), 0)::int AS committed_done_points,
			COALESCE(SUM(CASE WHEN se.added_after_sprint_start THEN se.points_at_entry ELSE 0 END), 0)::int AS late_add_points,
			COUNT(CASE WHEN NOT se.added_after_sprint_start THEN 1 END)::int AS committed_tickets,
			COUNT(CASE WHEN NOT se.added_after_sprint_start AND se.status = 'Done' THEN 1 END)::int AS committed_done_tickets,
			COUNT(CASE WHEN se.added_after_sprint_start THEN 1 END)::int AS late_add_tickets,
			COUNT(se.bounce_count)::int AS bounce_tracked_tickets,
			COALESCE(SUM(se.bounce_count), 0)::int AS total_bounces,
			COUNT(CASE WHEN se.bounce_count > 0 THEN 1 END)::int AS bounced_tickets
		FROM sprint_entry se
		JOIN ticket t ON t.id = se.ticket_id
		JOIN project p ON p.id = t.project_id
		JOIN sprint s ON s.id = se.sprint_id
		WHERE s.status = 'Open' AND p.status = 'Active' AND se.status <> 'Cancelled'
		GROUP BY s.id, s.name, s.start_date, s.end_date
		ORDER BY s.id
	`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	points := []model.SprintPoints{}
	for rows.Next() {
		var p model.SprintPoints
		if err := rows.Scan(
			&p.SprintID, &p.SprintName, &p.SprintStartDate, &p.SprintEndDate,
			&p.CommittedPoints, &p.CommittedDonePoints, &p.LateAddPoints,
			&p.CommittedTickets, &p.CommittedDoneTickets, &p.LateAddTickets,
			&p.BounceTrackedTickets, &p.TotalBounces, &p.BouncedTickets,
		); err != nil {
			return nil, wrapReadErr(err)
		}
		points = append(points, p)
	}
	return points, wrapReadErr(rows.Err())
}

// ClosedSprintPoints returns committed/done/late-add point totals for the
// most recently closed `limit` sprints, oldest first. Unlike SprintPoints
// (Open sprints, Active projects only), every project is counted regardless
// of its current status — see SprintRetrospective's doc comment for why.
func (r *DashboardRepository) ClosedSprintPoints(ctx context.Context, limit int) ([]model.SprintPoints, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM (
			SELECT
				s.id,
				s.name,
				s.start_date,
				s.end_date,
				COALESCE(SUM(CASE WHEN NOT se.added_after_sprint_start THEN se.points_at_entry ELSE 0 END), 0)::int AS committed_points,
				COALESCE(SUM(CASE WHEN NOT se.added_after_sprint_start AND se.status = 'Done' THEN se.points_at_entry ELSE 0 END), 0)::int AS committed_done_points,
				COALESCE(SUM(CASE WHEN se.added_after_sprint_start THEN se.points_at_entry ELSE 0 END), 0)::int AS late_add_points,
				COUNT(CASE WHEN NOT se.added_after_sprint_start THEN 1 END)::int AS committed_tickets,
				COUNT(CASE WHEN NOT se.added_after_sprint_start AND se.status = 'Done' THEN 1 END)::int AS committed_done_tickets,
				COUNT(CASE WHEN se.added_after_sprint_start THEN 1 END)::int AS late_add_tickets,
				COUNT(se.bounce_count)::int AS bounce_tracked_tickets,
				COALESCE(SUM(se.bounce_count), 0)::int AS total_bounces,
				COUNT(CASE WHEN se.bounce_count > 0 THEN 1 END)::int AS bounced_tickets
			FROM sprint_entry se
			JOIN ticket t ON t.id = se.ticket_id
			JOIN sprint s ON s.id = se.sprint_id
			WHERE s.status = 'Closed' AND se.status <> 'Cancelled'
			GROUP BY s.id, s.name, s.start_date, s.end_date
			ORDER BY s.id DESC
			LIMIT $1
		) recent
		ORDER BY id ASC
	`, limit)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	points := []model.SprintPoints{}
	for rows.Next() {
		var p model.SprintPoints
		if err := rows.Scan(
			&p.SprintID, &p.SprintName, &p.SprintStartDate, &p.SprintEndDate,
			&p.CommittedPoints, &p.CommittedDonePoints, &p.LateAddPoints,
			&p.CommittedTickets, &p.CommittedDoneTickets, &p.LateAddTickets,
			&p.BounceTrackedTickets, &p.TotalBounces, &p.BouncedTickets,
		); err != nil {
			return nil, wrapReadErr(err)
		}
		points = append(points, p)
	}
	return points, wrapReadErr(rows.Err())
}

// StuckTickets returns NotDone entries in Open sprints (Active projects only,
// like SprintPoints — this is a live signal) whose ticket has bounced at
// least minBounces times across this entry and the ticket's entries in
// earlier sprints. Cancelled entries never count toward the total.
func (r *DashboardRepository) StuckTickets(ctx context.Context, minBounces int) ([]model.StuckTicket, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM (
			SELECT
				s.id,
				s.name,
				se.id AS entry_id,
				t.id AS ticket_id,
				t.title,
				p.name AS project_name,
				COALESCE(u.name, 'Unassigned') AS assignee_name,
				se.points_at_entry,
				COALESCE(se.bounce_count, 0) AS bounce_count,
				`+earlierBouncesSQL+` AS earlier_bounces
			FROM sprint_entry se
			JOIN sprint s ON s.id = se.sprint_id
			JOIN ticket t ON t.id = se.ticket_id
			JOIN project p ON p.id = t.project_id
			LEFT JOIN users u ON u.id = t.assignee_id
			WHERE s.status = 'Open' AND p.status = 'Active' AND se.status = 'NotDone'
		) candidate
		WHERE bounce_count + earlier_bounces >= $1
		ORDER BY id, bounce_count + earlier_bounces DESC, LOWER(title)
	`, minBounces)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	stuck := []model.StuckTicket{}
	for rows.Next() {
		var st model.StuckTicket
		if err := rows.Scan(
			&st.SprintID, &st.SprintName, &st.EntryID, &st.TicketID, &st.TicketTitle,
			&st.ProjectName, &st.AssigneeName, &st.PointsAtEntry, &st.BounceCount, &st.EarlierBounces,
		); err != nil {
			return nil, wrapReadErr(err)
		}
		st.TotalBounces = st.BounceCount + st.EarlierBounces
		stuck = append(stuck, st)
	}
	return stuck, wrapReadErr(rows.Err())
}

// BounceBySize groups the bounce data of every closed sprint's tracked,
// non-Cancelled entries into story-point bands (by the entry's own
// points_at_entry). Every band is always returned, with zero counts when it
// has no data. Counted per entry, so a ticket carried over twice counts
// three times in its band.
func (r *DashboardRepository) BounceBySize(ctx context.Context) ([]model.BounceBySize, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			b.band,
			COUNT(e.bounce_count)::int AS entries,
			COUNT(CASE WHEN e.bounce_count > 0 THEN 1 END)::int AS bounced_entries,
			COALESCE(SUM(e.bounce_count), 0)::int AS total_bounces
		FROM (VALUES (1, '1-2', 1, 2), (2, '3-5', 3, 5), (3, '6+', 6, 2147483647)) AS b(ord, band, lo, hi)
		LEFT JOIN (
			SELECT se.points_at_entry, se.bounce_count
			FROM sprint_entry se
			JOIN sprint s ON s.id = se.sprint_id
			WHERE s.status = 'Closed' AND se.status <> 'Cancelled' AND se.bounce_count IS NOT NULL
		) e ON e.points_at_entry BETWEEN b.lo AND b.hi
		GROUP BY b.ord, b.band
		ORDER BY b.ord
	`)
	if err != nil {
		return nil, wrapReadErr(err)
	}
	defer rows.Close()

	bands := []model.BounceBySize{}
	for rows.Next() {
		var b model.BounceBySize
		if err := rows.Scan(&b.Band, &b.Entries, &b.BouncedEntries, &b.TotalBounces); err != nil {
			return nil, wrapReadErr(err)
		}
		bands = append(bands, b)
	}
	return bands, wrapReadErr(rows.Err())
}
