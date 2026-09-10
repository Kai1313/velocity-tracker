package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func (f *fakeDashboardRepo) ProjectSprintPoints(ctx context.Context) ([]model.ProjectSprintPoints, error) {
	return nil, nil
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

func day(d int) time.Time {
	return time.Date(2026, 9, d, 0, 0, 0, 0, time.UTC)
}

func TestComputeProjectSprintHealth(t *testing.T) {
	base := model.ProjectSprintPoints{
		ProjectID: 1, ProjectName: "Core Platform",
		SprintID: 9, SprintName: "Sprint 9",
	}

	tests := []struct {
		name            string
		committed       int
		done            int
		start, end, now time.Time
		wantElapsed     int
		wantRemaining   int
		wantOverdue     bool
		wantRequiredNil bool
		wantRequired    float64
		wantAchievedNil bool
		wantAchieved    float64
		wantStatus      model.SprintHealthStatus
	}{
		{
			name:      "on track: achieved matches required",
			committed: 16, done: 8,
			start: day(1), end: day(9), now: day(5),
			wantElapsed: 4, wantRemaining: 4,
			wantRequired: 2, wantAchieved: 2,
			wantStatus: model.HealthOnTrack,
		},
		{
			name:      "at risk: achieved is 75% of required",
			committed: 14, done: 6,
			start: day(1), end: day(9), now: day(5),
			wantElapsed: 4, wantRemaining: 4,
			wantRequired: 2, wantAchieved: 1.5,
			wantStatus: model.HealthAtRisk,
		},
		{
			name:      "critical: achieved is 50% of required",
			committed: 12, done: 4,
			start: day(1), end: day(9), now: day(5),
			wantElapsed: 4, wantRemaining: 4,
			wantRequired: 2, wantAchieved: 1,
			wantStatus: model.HealthCritical,
		},
		{
			name:      "overdue forces critical with no required velocity number",
			committed: 10, done: 5,
			start: day(1), end: day(9), now: day(11),
			wantElapsed: 10, wantRemaining: -2, wantOverdue: true,
			wantRequiredNil: true,
			wantAchieved:    0.5,
			wantStatus:      model.HealthCritical,
		},
		{
			name:      "sprint's first day has no achieved velocity yet",
			committed: 10, done: 0,
			start: day(1), end: day(11), now: day(1),
			wantElapsed: 0, wantRemaining: 10,
			wantRequired:    1,
			wantAchievedNil: true,
			wantStatus:      model.HealthOnTrack,
		},
		{
			name:      "all committed work already done stays on track",
			committed: 5, done: 5,
			start: day(1), end: day(8), now: day(4),
			wantElapsed: 3, wantRemaining: 4,
			wantRequired: 0, wantAchieved: 5.0 / 3.0,
			wantStatus: model.HealthOnTrack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := base
			p.CommittedPoints = tt.committed
			p.CommittedDonePoints = tt.done
			p.SprintStartDate = tt.start
			p.SprintEndDate = tt.end

			got := service.ComputeProjectSprintHealth(p, tt.now)

			if got.DaysElapsed != tt.wantElapsed {
				t.Errorf("DaysElapsed = %d, want %d", got.DaysElapsed, tt.wantElapsed)
			}
			if got.DaysRemaining != tt.wantRemaining {
				t.Errorf("DaysRemaining = %d, want %d", got.DaysRemaining, tt.wantRemaining)
			}
			if got.Overdue != tt.wantOverdue {
				t.Errorf("Overdue = %v, want %v", got.Overdue, tt.wantOverdue)
			}
			if tt.wantRequiredNil {
				if got.RequiredVelocity != nil {
					t.Errorf("RequiredVelocity = %v, want nil", *got.RequiredVelocity)
				}
			} else if got.RequiredVelocity == nil || *got.RequiredVelocity != tt.wantRequired {
				t.Errorf("RequiredVelocity = %v, want %v", got.RequiredVelocity, tt.wantRequired)
			}
			if tt.wantAchievedNil {
				if got.AchievedVelocity != nil {
					t.Errorf("AchievedVelocity = %v, want nil", *got.AchievedVelocity)
				}
			} else if got.AchievedVelocity == nil || *got.AchievedVelocity != tt.wantAchieved {
				t.Errorf("AchievedVelocity = %v, want %v", got.AchievedVelocity, tt.wantAchieved)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.LateAddPoints != p.LateAddPoints {
				t.Errorf("LateAddPoints = %d, want %d (passthrough)", got.LateAddPoints, p.LateAddPoints)
			}
		})
	}
}
