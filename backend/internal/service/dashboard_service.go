package service

import (
	"context"

	"velocity-tracker/backend/internal/model"
)

type DashboardRepository interface {
	SprintSummaries(ctx context.Context) ([]model.SprintSummary, error)
	DeveloperBreakdown(ctx context.Context, sprintID int64) ([]model.DeveloperSummary, error)
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
