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
