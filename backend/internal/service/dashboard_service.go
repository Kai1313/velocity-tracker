package service

import (
	"context"
	"time"

	"velocity-tracker/backend/internal/model"
)

type DashboardRepository interface {
	SprintSummaries(ctx context.Context) ([]model.SprintSummary, error)
	DeveloperBreakdown(ctx context.Context, sprintID int64) ([]model.DeveloperSummary, error)
	TicketEntries(ctx context.Context, sprintID int64) ([]model.SprintEntryDetail, error)
	ProjectSprintPoints(ctx context.Context) ([]model.ProjectSprintPoints, error)
}

type DashboardService struct {
	repo    DashboardRepository
	sprints SprintLookup
}

func NewDashboardService(repo DashboardRepository, sprints SprintLookup) *DashboardService {
	return &DashboardService{repo: repo, sprints: sprints}
}

func (s *DashboardService) SprintSummaries(ctx context.Context) ([]model.SprintSummary, error) {
	return s.repo.SprintSummaries(ctx)
}

// SprintDeveloperBreakdown 404s via SprintLookup if the sprint itself
// doesn't exist, rather than silently returning an empty breakdown.
func (s *DashboardService) SprintDeveloperBreakdown(ctx context.Context, sprintID int64) ([]model.DeveloperSummary, error) {
	if _, err := s.sprints.Get(ctx, sprintID); err != nil {
		return nil, err
	}
	return s.repo.DeveloperBreakdown(ctx, sprintID)
}

// SprintTicketBreakdown splits a sprint's ticket entries into freshly-planned
// "current" tickets and tickets continued from a prior sprint ("carried
// over"), by whether each entry names a carried-from sprint. 404s via
// SprintLookup if the sprint itself doesn't exist.
func (s *DashboardService) SprintTicketBreakdown(ctx context.Context, sprintID int64) (*model.SprintTicketBreakdown, error) {
	if _, err := s.sprints.Get(ctx, sprintID); err != nil {
		return nil, err
	}
	entries, err := s.repo.TicketEntries(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	breakdown := &model.SprintTicketBreakdown{Current: []model.SprintEntryDetail{}, CarriedOver: []model.SprintEntryDetail{}}
	for _, e := range entries {
		if e.CarriedFromSprintName != nil {
			breakdown.CarriedOver = append(breakdown.CarriedOver, e)
		} else {
			breakdown.Current = append(breakdown.Current, e)
		}
	}
	return breakdown, nil
}

// ProjectSprintHealth returns the early-warning velocity comparison
// ("Chart 42C") for every Active project's Open-sprint work, one row per
// (project, sprint) pair.
func (s *DashboardService) ProjectSprintHealth(ctx context.Context) ([]model.ProjectSprintHealth, error) {
	points, err := s.repo.ProjectSprintPoints(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	health := make([]model.ProjectSprintHealth, len(points))
	for i, p := range points {
		health[i] = ComputeProjectSprintHealth(p, now)
	}
	return health, nil
}

// ComputeProjectSprintHealth is the pure day-count and velocity calculation
// behind ProjectSprintHealth, exported so it can be unit-tested with a fixed
// `now` instead of depending on the wall clock.
//
// Both velocity terms use committed-only points (CommittedDonePoints, not
// CommittedDonePoints+LateAdd): a team finishing late-added tickets while
// original commitments stall must not read as "on pace" just because more
// points got marked Done. Late-add work is surfaced separately on the model
// instead, as a visible signal rather than one baked into the ratio.
//
// Days are calendar days, computed from each row's own sprint dates — a
// project split across multiple concurrently-open sprints (legal, since the
// one-open-sprint-per-*ticket* trigger in the DB doesn't limit how many
// sprints a project's tickets collectively touch) gets one independent row
// per sprint rather than a merged countdown.
func ComputeProjectSprintHealth(p model.ProjectSprintPoints, now time.Time) model.ProjectSprintHealth {
	today := now.UTC().Truncate(24 * time.Hour)
	start := p.SprintStartDate.UTC().Truncate(24 * time.Hour)
	end := p.SprintEndDate.UTC().Truncate(24 * time.Hour)

	daysElapsed := int(today.Sub(start).Hours() / 24)
	if daysElapsed < 0 {
		daysElapsed = 0
	}
	daysRemaining := int(end.Sub(today).Hours() / 24)
	overdue := daysRemaining <= 0

	h := model.ProjectSprintHealth{
		ProjectID:       p.ProjectID,
		ProjectName:     p.ProjectName,
		SprintID:        p.SprintID,
		SprintName:      p.SprintName,
		CommittedPoints: p.CommittedPoints,
		DonePoints:      p.CommittedDonePoints,
		LateAddPoints:   p.LateAddPoints,
		DaysElapsed:     daysElapsed,
		DaysRemaining:   daysRemaining,
		Overdue:         overdue,
	}

	if !overdue {
		required := float64(p.CommittedPoints-p.CommittedDonePoints) / float64(daysRemaining)
		h.RequiredVelocity = &required
	}
	if daysElapsed > 0 {
		achieved := float64(p.CommittedDonePoints) / float64(daysElapsed)
		h.AchievedVelocity = &achieved
	}
	h.Status = sprintHealthStatus(h)
	return h
}

func sprintHealthStatus(h model.ProjectSprintHealth) model.SprintHealthStatus {
	if h.Overdue {
		return model.HealthCritical
	}
	if h.RequiredVelocity != nil && *h.RequiredVelocity <= 0 {
		return model.HealthOnTrack // nothing committed left to do
	}
	if h.AchievedVelocity == nil {
		return model.HealthOnTrack // sprint's first day, pace not yet measurable
	}
	switch ratio := *h.AchievedVelocity / *h.RequiredVelocity; {
	case ratio >= 1:
		return model.HealthOnTrack
	case ratio >= 0.6:
		return model.HealthAtRisk
	default:
		return model.HealthCritical
	}
}
