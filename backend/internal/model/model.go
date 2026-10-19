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
	// Bounce figures cover only entries with tracked bounce data
	// (BounceTrackedTickets of WorkloadTickets). The average is derived by
	// the caller from these raw counts. A sprint-scoped figure — see ADR-0013.
	Bounces              int `json:"bounces"`
	BounceTrackedTickets int `json:"bounceTrackedTickets"`
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
	// BounceCount is how many times QA sent this entry back to the developer
	// within this sprint. nil means "not tracked" (entries that predate the
	// feature), which is not the same as 0. See ADR-0013.
	BounceCount *int `json:"bounceCount"`
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
	BounceCount           *int        `json:"bounceCount"`
}

// SprintTicketBreakdown splits a sprint's entries into freshly-planned
// "current" tickets and tickets continued from a prior sprint's unfinished
// entry ("carriedOver"), for the sprint detail page's two ticket tables.
type SprintTicketBreakdown struct {
	Current     []SprintEntryDetail `json:"current"`
	CarriedOver []SprintEntryDetail `json:"carriedOver"`
}

// SprintPoints is one currently-open sprint's raw committed/done/late-add
// totals, summed across every Active project's tickets in it — developer
// capacity is shared across a team's sub-projects, so the totals are not
// broken out per project (see ADR-0011). "Committed" means points_at_entry
// for entries that were NOT added after the sprint started; late-add points
// are tracked separately rather than folded in, so the sprint's original
// commitment isn't diluted by mid-sprint scope changes (see ADR-0004's
// Committed/Late-Add distinction, not yet computed anywhere else in the
// app). Cancelled entries are excluded, per the same domain rule
// SprintSummaries/DeveloperBreakdown already follow.
type SprintPoints struct {
	SprintID            int64     `json:"sprintId"`
	SprintName          string    `json:"sprintName"`
	SprintStartDate     time.Time `json:"sprintStartDate"`
	SprintEndDate       time.Time `json:"sprintEndDate"`
	CommittedPoints     int       `json:"committedPoints"`
	CommittedDonePoints int       `json:"committedDonePoints"`
	LateAddPoints       int       `json:"lateAddPoints"`
	// Ticket counts cover the same sprint entries as the point totals above.
	CommittedTickets     int `json:"committedTickets"`
	CommittedDoneTickets int `json:"committedDoneTickets"`
	LateAddTickets       int `json:"lateAddTickets"`
	// Bounce totals cover only non-Cancelled entries with tracked bounce data.
	BounceTrackedTickets int `json:"bounceTrackedTickets"`
	TotalBounces         int `json:"totalBounces"`
	BouncedTickets       int `json:"bouncedTickets"`
}

// SprintRetrospective is one closed sprint's commitment-reliability
// metrics: Planning Accuracy and Late-Add Rate, per CONTEXT.md's
// definitions. Unlike SprintPoints/SprintHealth (Open sprints, Active
// projects only), this is computed across every project regardless of its
// current status — a retrospective on a past sprint should reflect what was
// actually committed and done at the time, not be reshaped by a project
// being archived afterward. Both ratios are nil when the sprint had no
// committed points, rather than a nonsensical 0% or a divide-by-zero.
type SprintRetrospective struct {
	SprintID            int64     `json:"sprintId"`
	SprintName          string    `json:"sprintName"`
	SprintStartDate     time.Time `json:"sprintStartDate"`
	SprintEndDate       time.Time `json:"sprintEndDate"`
	CommittedPoints     int       `json:"committedPoints"`
	CommittedDonePoints int       `json:"committedDonePoints"`
	LateAddPoints       int       `json:"lateAddPoints"`
	// Ticket counts cover the same sprint entries as the point totals above.
	CommittedTickets     int      `json:"committedTickets"`
	CommittedDoneTickets int      `json:"committedDoneTickets"`
	LateAddTickets       int      `json:"lateAddTickets"`
	PlanningAccuracy     *float64 `json:"planningAccuracy"`
	LateAddRate          *float64 `json:"lateAddRate"`
	// Bounce figures count only entries with tracked bounce data
	// (BounceTrackedTickets); a sprint that predates the feature has 0 tracked
	// tickets, shown as "not tracked" rather than 0. The per-ticket average is
	// derived by the caller from these raw counts.
	BounceTrackedTickets int `json:"bounceTrackedTickets"`
	TotalBounces         int `json:"totalBounces"`
	BouncedTickets       int `json:"bouncedTickets"`
}

// StuckTicket is a NotDone entry in an Open sprint whose ticket has bounced
// at least the stuck threshold's worth of times across all its entries.
type StuckTicket struct {
	SprintID       int64  `json:"sprintId"`
	SprintName     string `json:"sprintName"`
	EntryID        int64  `json:"entryId"`
	TicketID       int64  `json:"ticketId"`
	TicketTitle    string `json:"ticketTitle"`
	ProjectName    string `json:"projectName"`
	AssigneeName   string `json:"assigneeName"`
	PointsAtEntry  int    `json:"pointsAtEntry"`
	BounceCount    int    `json:"bounceCount"`
	EarlierBounces int    `json:"earlierBounces"`
	TotalBounces   int    `json:"totalBounces"`
}

// StuckTicketList is the stuck-tickets response: the tickets plus the
// threshold that selected them, so the UI never hard-codes the number.
type StuckTicketList struct {
	MinBounces int           `json:"minBounces"`
	Tickets    []StuckTicket `json:"tickets"`
}

// BounceBySize is one story-point band's bounce totals across every closed
// sprint's tracked, non-Cancelled entries. Counted per entry, so a ticket
// carried over twice appears three times.
type BounceBySize struct {
	Band           string `json:"band"`
	Entries        int    `json:"entries"`
	BouncedEntries int    `json:"bouncedEntries"`
	TotalBounces   int    `json:"totalBounces"`
}

// SprintHealthStatus is the early-warning signal for one SprintHealth
// row, comparing the velocity still required to finish committed work
// against the velocity actually being achieved.
type SprintHealthStatus string

const (
	HealthOnTrack  SprintHealthStatus = "OnTrack"
	HealthAtRisk   SprintHealthStatus = "AtRisk"
	HealthCritical SprintHealthStatus = "Critical"
)

// SprintHealth is a SprintPoints row plus the derived day-count and
// velocity comparison ("Chart 42C"). RequiredVelocity is nil when the
// sprint is Overdue (days remaining <= 0) rather than a huge or infinite
// number — the point is triggering the alarm, not precision.
// AchievedVelocity is nil on the sprint's first day (days elapsed == 0),
// since 0 SP/day on day one reads as "the team achieved nothing," which is
// misleading rather than informative.
type SprintHealth struct {
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
