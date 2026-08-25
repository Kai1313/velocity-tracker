package service_test

import (
	"context"
	"errors"
	"testing"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type fakeDashboardRepo struct {
	breakdown map[int64][]model.DeveloperSummary
}

func (f *fakeDashboardRepo) SprintSummaries(ctx context.Context) ([]model.SprintSummary, error) {
	return nil, nil
}

func (f *fakeDashboardRepo) DeveloperBreakdown(ctx context.Context, sprintID int64) ([]model.DeveloperSummary, error) {
	return f.breakdown[sprintID], nil
}

func TestDashboardService_SprintDeveloperBreakdown_404sOnMissingSprint(t *testing.T) {
	repo := &fakeDashboardRepo{}
	sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{}}
	svc := service.NewDashboardService(repo, sprints)

	_, err := svc.SprintDeveloperBreakdown(context.Background(), 999)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("SprintDeveloperBreakdown() for missing sprint = %v, want apperr.ErrNotFound", err)
	}
}

func TestDashboardService_SprintDeveloperBreakdown_ReturnsDataForKnownSprint(t *testing.T) {
	repo := &fakeDashboardRepo{breakdown: map[int64][]model.DeveloperSummary{
		10: {{Name: "Alice", WorkloadPoints: 5, DonePoints: 5}},
	}}
	sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{
		10: {ID: 10, Status: model.SprintOpen},
	}}
	svc := service.NewDashboardService(repo, sprints)

	got, err := svc.SprintDeveloperBreakdown(context.Background(), 10)
	if err != nil {
		t.Fatalf("SprintDeveloperBreakdown() unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Alice" {
		t.Fatalf("SprintDeveloperBreakdown() = %+v, want [Alice]", got)
	}
}
