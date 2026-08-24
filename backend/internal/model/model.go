package model

import "time"

type Role string

const (
	RoleLead      Role = "Lead"
	RoleDeveloper Role = "Developer"
)

type User struct {
	ID   int64
	Name string
	Role Role
}

type Ticket struct {
	ID          int64
	Title       string
	StoryPoints int
	AssigneeID  *int64
}

type SprintStatus string

const (
	SprintOpen   SprintStatus = "Open"
	SprintClosed SprintStatus = "Closed"
)

type Sprint struct {
	ID        int64
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    SprintStatus
}

type EntryStatus string

const (
	EntryDone      EntryStatus = "Done"
	EntryNotDone   EntryStatus = "NotDone"
	EntryCancelled EntryStatus = "Cancelled"
)

type SprintEntry struct {
	ID                    int64
	TicketID              int64
	SprintID              int64
	Status                EntryStatus
	AddedAfterSprintStart bool
	CarriedFrom           *int64
	PointsAtEntry         int
	CreatedAt             time.Time
}
