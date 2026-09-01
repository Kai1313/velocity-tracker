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
