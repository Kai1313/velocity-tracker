package model

import "time"

type Role string

const (
	RoleLead      Role = "Lead"
	RoleDeveloper Role = "Developer"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role Role   `json:"role"`
}

type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "Active"
	ProjectArchived ProjectStatus = "Archived"
)

type Project struct {
	ID     int64         `json:"id"`
	Name   string        `json:"name"`
	Status ProjectStatus `json:"status"`
}

type Ticket struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"projectId"`
	Title       string `json:"title"`
	StoryPoints int    `json:"storyPoints"`
	AssigneeID  *int64 `json:"assigneeId"`
}

// TicketDetail is a Ticket plus its current status, computed from the
// ticket's most recent SprintEntry rather than stored on Ticket itself.
type TicketDetail struct {
	Ticket
	CurrentStatus *EntryStatus `json:"currentStatus"`
}

type SprintStatus string

const (
	SprintOpen   SprintStatus = "Open"
	SprintClosed SprintStatus = "Closed"
)

type Sprint struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	StartDate time.Time    `json:"startDate"`
	EndDate   time.Time    `json:"endDate"`
	Status    SprintStatus `json:"status"`
}

type EntryStatus string

const (
	EntryDone      EntryStatus = "Done"
	EntryNotDone   EntryStatus = "NotDone"
	EntryCancelled EntryStatus = "Cancelled"
)

// SprintSummary is one sprint's aggregate workload/done totals, computed
// from its SprintEntry rows (Cancelled entries excluded). "Workload" and
// "Done" are a deliberately simpler v1 metric than the CONTEXT.md Sprint
// Velocity definition (no Committed/Late-Add split yet) — see ADR-0004.
type SprintSummary struct {
	SprintID        int64  `json:"sprintId"`
	SprintName      string `json:"sprintName"`
	WorkloadPoints  int    `json:"workloadPoints"`
	DonePoints      int    `json:"donePoints"`
	WorkloadTickets int    `json:"workloadTickets"`
	DoneTickets     int    `json:"doneTickets"`
}

// DeveloperSummary is one developer's workload/done totals within a single
// sprint. UserID is nil for tickets with no assignee, grouped under "Unassigned"
// so its points aren't silently dropped from the per-developer view.
type DeveloperSummary struct {
	UserID          *int64 `json:"userId"`
	Name            string `json:"name"`
	WorkloadPoints  int    `json:"workloadPoints"`
	DonePoints      int    `json:"donePoints"`
	WorkloadTickets int    `json:"workloadTickets"`
	DoneTickets     int    `json:"doneTickets"`
}

type SprintEntry struct {
	ID                    int64       `json:"id"`
	TicketID              int64       `json:"ticketId"`
	SprintID              int64       `json:"sprintId"`
	Status                EntryStatus `json:"status"`
	AddedAfterSprintStart bool        `json:"addedAfterSprintStart"`
	CarriedFrom           *int64      `json:"carriedFrom"`
	PointsAtEntry         int         `json:"pointsAtEntry"`
	CreatedAt             time.Time   `json:"createdAt"`
}

// SprintEntryDetail is one ticket's SprintEntry within a single sprint,
// joined with the ticket/project/assignee names the sprint detail page's
// ticket tables need. CarriedFromSprintName is set only when the entry
// continues from a prior sprint's unfinished entry (CarriedFrom is not nil
// on the underlying SprintEntry), naming that one prior sprint rather than
// walking the full carry-over chain back to the ticket's original sprint.
type SprintEntryDetail struct {
	EntryID               int64       `json:"entryId"`
	TicketID              int64       `json:"ticketId"`
	TicketTitle           string      `json:"ticketTitle"`
	ProjectName           string      `json:"projectName"`
	AssigneeName          string      `json:"assigneeName"`
	Status                EntryStatus `json:"status"`
	AddedAfterSprintStart bool        `json:"addedAfterSprintStart"`
	PointsAtEntry         int         `json:"pointsAtEntry"`
	CarriedFromSprintName *string     `json:"carriedFromSprintName"`
}

// SprintTicketBreakdown splits a sprint's entries into freshly-planned
// "current" tickets and tickets continued from a prior sprint's unfinished
// entry ("carriedOver"), for the sprint detail page's two ticket tables.
type SprintTicketBreakdown struct {
	Current     []SprintEntryDetail `json:"current"`
	CarriedOver []SprintEntryDetail `json:"carriedOver"`
}

// ProjectSprintPoints is one project's raw committed/done/late-add totals
// within a single currently-open sprint, plus that sprint's own date range.
// "Committed" means points_at_entry for entries that were NOT added after
// the sprint started; late-add points are tracked separately rather than
// folded in, so a project's original commitment isn't diluted by mid-sprint
// scope changes (see ADR-0004's Committed/Late-Add distinction, not yet
// computed anywhere else in the app). Cancelled entries are excluded, per
// the same domain rule SprintSummaries/DeveloperBreakdown already follow.
//
// A project with tickets split across more than one concurrently-open
// sprint produces one row per (project, sprint) pair here — Sprint doesn't
// belong to a Project (ADR-0003), so there is no single "active sprint per
// project" to collapse rows into.
type ProjectSprintPoints struct {
	ProjectID           int64     `json:"projectId"`
	ProjectName         string    `json:"projectName"`
	SprintID            int64     `json:"sprintId"`
	SprintName          string    `json:"sprintName"`
	SprintStartDate     time.Time `json:"sprintStartDate"`
	SprintEndDate       time.Time `json:"sprintEndDate"`
	CommittedPoints     int       `json:"committedPoints"`
	CommittedDonePoints int       `json:"committedDonePoints"`
	LateAddPoints       int       `json:"lateAddPoints"`
}

// SprintHealthStatus is the early-warning signal for one ProjectSprintHealth
// row, comparing the velocity still required to finish committed work
// against the velocity actually being achieved.
type SprintHealthStatus string

const (
	HealthOnTrack  SprintHealthStatus = "OnTrack"
	HealthAtRisk   SprintHealthStatus = "AtRisk"
	HealthCritical SprintHealthStatus = "Critical"
)

// ProjectSprintHealth is a ProjectSprintPoints row plus the derived
// day-count and velocity comparison ("Chart 42C"). RequiredVelocity is nil
// when the sprint is Overdue (days remaining <= 0) rather than a huge or
// infinite number — the point is triggering the alarm, not precision.
// AchievedVelocity is nil on the sprint's first day (days elapsed == 0),
// since 0 SP/day on day one reads as "the team achieved nothing," which is
// misleading rather than informative.
type ProjectSprintHealth struct {
	ProjectID        int64              `json:"projectId"`
	ProjectName      string             `json:"projectName"`
	SprintID         int64              `json:"sprintId"`
	SprintName       string             `json:"sprintName"`
	CommittedPoints  int                `json:"committedPoints"`
	DonePoints       int                `json:"donePoints"`
	LateAddPoints    int                `json:"lateAddPoints"`
	DaysElapsed      int                `json:"daysElapsed"`
	DaysRemaining    int                `json:"daysRemaining"`
	Overdue          bool               `json:"overdue"`
	RequiredVelocity *float64           `json:"requiredVelocity"`
	AchievedVelocity *float64           `json:"achievedVelocity"`
	Status           SprintHealthStatus `json:"status"`
}
