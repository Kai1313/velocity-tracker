package service

import (
	"context"

	"velocity-tracker/backend/internal/model"
)

type DashboardRepository interface {
	SprintSummaries(ctx context.Context) ([]model.SprintSummary, error)
	DeveloperBreakdown(ctx context.Context, sprintID int64) ([]model.DeveloperSummary, error)
	TicketEntries(ctx context.Context, sprintID int64) ([]model.SprintEntryDetail, error)
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
