package repository_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/repository"
)

func mustCreateUser(t *testing.T, tx pgx.Tx, name string) *model.User {
	t.Helper()
	repo := repository.NewUserRepository(tx)
	u := &model.User{Name: name, Role: model.RoleDeveloper}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func TestDashboardRepository_SprintSummaries_ExcludesCancelled(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Website Redesign")
	ticketA := mustCreateTicket(t, tx, project.ID)
	ticketB := mustCreateTicket(t, tx, project.ID)
	ticketC := mustCreateTicket(t, tx, project.ID)

	sprintRepo := repository.NewSprintRepository(tx)
	sprint := &model.Sprint{Name: "Sprint 1", StartDate: fixedDate(2026, 8, 1), EndDate: fixedDate(2026, 8, 14), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint); err != nil {
		t.Fatalf("create sprint: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	entries := []*model.SprintEntry{
		{TicketID: ticketA.ID, SprintID: sprint.ID, Status: model.EntryDone, PointsAtEntry: 5},
		{TicketID: ticketB.ID, SprintID: sprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3},
		{TicketID: ticketC.ID, SprintID: sprint.ID, Status: model.EntryCancelled, PointsAtEntry: 8},
	}
	for _, e := range entries {
		if err := entryRepo.Create(ctx, e); err != nil {
			t.Fatalf("create sprint entry: %v", err)
		}
	}

	summaries, err := repository.NewDashboardRepository(tx).SprintSummaries(ctx)
	if err != nil {
		t.Fatalf("SprintSummaries() unexpected error: %v", err)
	}

	var got *model.SprintSummary
	for i := range summaries {
		if summaries[i].SprintID == sprint.ID {
			got = &summaries[i]
		}
	}
	if got == nil {
		t.Fatalf("SprintSummaries() missing sprint %d in result", sprint.ID)
	}
	// Cancelled ticketC's 8 points must not appear in either total.
	if got.WorkloadPoints != 8 {
		t.Errorf("WorkloadPoints = %d, want 8 (5+3, cancelled excluded)", got.WorkloadPoints)
	}
	if got.DonePoints != 5 {
		t.Errorf("DonePoints = %d, want 5", got.DonePoints)
	}
	if got.WorkloadTickets != 2 {
		t.Errorf("WorkloadTickets = %d, want 2 (cancelled excluded)", got.WorkloadTickets)
	}
	if got.DoneTickets != 1 {
		t.Errorf("DoneTickets = %d, want 1", got.DoneTickets)
	}
}

func TestDashboardRepository_DeveloperBreakdown_GroupsUnassignedTickets(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Website Redesign")
	ticket := mustCreateTicket(t, tx, project.ID) // no assignee

	sprintRepo := repository.NewSprintRepository(tx)
	sprint := &model.Sprint{Name: "Sprint 1", StartDate: fixedDate(2026, 8, 1), EndDate: fixedDate(2026, 8, 14), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint); err != nil {
		t.Fatalf("create sprint: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	entry := &model.SprintEntry{TicketID: ticket.ID, SprintID: sprint.ID, Status: model.EntryDone, PointsAtEntry: ticket.StoryPoints}
	if err := entryRepo.Create(ctx, entry); err != nil {
		t.Fatalf("create sprint entry: %v", err)
	}

	breakdown, err := repository.NewDashboardRepository(tx).DeveloperBreakdown(ctx, sprint.ID)
	if err != nil {
		t.Fatalf("DeveloperBreakdown() unexpected error: %v", err)
	}
	if len(breakdown) != 1 {
		t.Fatalf("DeveloperBreakdown() = %d rows, want 1", len(breakdown))
	}
	if breakdown[0].UserID != nil {
		t.Errorf("UserID = %v, want nil (unassigned)", breakdown[0].UserID)
	}
	if breakdown[0].Name != "Unassigned" {
		t.Errorf("Name = %q, want %q", breakdown[0].Name, "Unassigned")
	}
	if breakdown[0].WorkloadPoints != ticket.StoryPoints {
		t.Errorf("WorkloadPoints = %d, want %d", breakdown[0].WorkloadPoints, ticket.StoryPoints)
	}
}

func TestDashboardRepository_DeveloperBreakdown_GroupsByAssignee(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Website Redesign")
	alice := mustCreateUser(t, tx, "Alice")
	bob := mustCreateUser(t, tx, "Bob")

	ticketRepo := repository.NewTicketRepository(tx)
	aliceTicket := &model.Ticket{ProjectID: project.ID, Title: "Alice's ticket", StoryPoints: 5, AssigneeID: &alice.ID}
	if err := ticketRepo.Create(ctx, aliceTicket); err != nil {
		t.Fatalf("create alice ticket: %v", err)
	}
	bobTicket := &model.Ticket{ProjectID: project.ID, Title: "Bob's ticket", StoryPoints: 3, AssigneeID: &bob.ID}
	if err := ticketRepo.Create(ctx, bobTicket); err != nil {
		t.Fatalf("create bob ticket: %v", err)
	}

	sprintRepo := repository.NewSprintRepository(tx)
	sprint := &model.Sprint{Name: "Sprint 1", StartDate: fixedDate(2026, 8, 1), EndDate: fixedDate(2026, 8, 14), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint); err != nil {
		t.Fatalf("create sprint: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: aliceTicket.ID, SprintID: sprint.ID, Status: model.EntryDone, PointsAtEntry: 5}); err != nil {
		t.Fatalf("create alice entry: %v", err)
	}
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: bobTicket.ID, SprintID: sprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3}); err != nil {
		t.Fatalf("create bob entry: %v", err)
	}

	breakdown, err := repository.NewDashboardRepository(tx).DeveloperBreakdown(ctx, sprint.ID)
	if err != nil {
		t.Fatalf("DeveloperBreakdown() unexpected error: %v", err)
	}
	if len(breakdown) != 2 {
		t.Fatalf("DeveloperBreakdown() = %d rows, want 2", len(breakdown))
	}

	byName := map[string]model.DeveloperSummary{}
	for _, d := range breakdown {
		byName[d.Name] = d
	}
	if got := byName["Alice"]; got.WorkloadPoints != 5 || got.DonePoints != 5 {
		t.Errorf("Alice = %+v, want workload=5 done=5", got)
	}
	if got := byName["Bob"]; got.WorkloadPoints != 3 || got.DonePoints != 0 {
		t.Errorf("Bob = %+v, want workload=3 done=0", got)
	}
}
