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
	entries   map[int64][]model.SprintEntryDetail
}

func (f *fakeDashboardRepo) SprintSummaries(ctx context.Context) ([]model.SprintSummary, error) {
	return nil, nil
}

func (f *fakeDashboardRepo) DeveloperBreakdown(ctx context.Context, sprintID int64) ([]model.DeveloperSummary, error) {
	return f.breakdown[sprintID], nil
}

func (f *fakeDashboardRepo) TicketEntries(ctx context.Context, sprintID int64) ([]model.SprintEntryDetail, error) {
	return f.entries[sprintID], nil
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

func TestDashboardService_SprintTicketBreakdown_404sOnMissingSprint(t *testing.T) {
	repo := &fakeDashboardRepo{}
	sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{}}
	svc := service.NewDashboardService(repo, sprints)

	_, err := svc.SprintTicketBreakdown(context.Background(), 999)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("SprintTicketBreakdown() for missing sprint = %v, want apperr.ErrNotFound", err)
	}
}

func TestDashboardService_SprintTicketBreakdown_SplitsByCarriedFrom(t *testing.T) {
	origin := "Sprint 1"
	repo := &fakeDashboardRepo{entries: map[int64][]model.SprintEntryDetail{
		10: {
			{EntryID: 1, TicketTitle: "Fresh ticket", CarriedFromSprintName: nil},
			{EntryID: 2, TicketTitle: "Continued ticket", CarriedFromSprintName: &origin},
		},
	}}
	sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{
		10: {ID: 10, Status: model.SprintOpen},
	}}
	svc := service.NewDashboardService(repo, sprints)

	got, err := svc.SprintTicketBreakdown(context.Background(), 10)
	if err != nil {
		t.Fatalf("SprintTicketBreakdown() unexpected error: %v", err)
	}
	if len(got.Current) != 1 || got.Current[0].TicketTitle != "Fresh ticket" {
		t.Fatalf("Current = %+v, want [Fresh ticket]", got.Current)
	}
	if len(got.CarriedOver) != 1 || got.CarriedOver[0].TicketTitle != "Continued ticket" {
		t.Fatalf("CarriedOver = %+v, want [Continued ticket]", got.CarriedOver)
	}
}
