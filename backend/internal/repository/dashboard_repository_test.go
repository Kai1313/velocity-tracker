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

func TestDashboardRepository_TicketEntries_NamesCarriedFromSprintForCarriedOverEntriesOnly(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Website Redesign")
	freshTicket := mustCreateTicket(t, tx, project.ID)
	carriedTicket := mustCreateTicket(t, tx, project.ID)

	sprintRepo := repository.NewSprintRepository(tx)
	sprint1 := &model.Sprint{Name: "Sprint 1", StartDate: fixedDate(2026, 8, 1), EndDate: fixedDate(2026, 8, 14), Status: model.SprintClosed}
	if err := sprintRepo.Create(ctx, sprint1); err != nil {
		t.Fatalf("create sprint1: %v", err)
	}
	sprint2 := &model.Sprint{Name: "Sprint 2", StartDate: fixedDate(2026, 8, 15), EndDate: fixedDate(2026, 8, 28), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint2); err != nil {
		t.Fatalf("create sprint2: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	originEntry := &model.SprintEntry{TicketID: carriedTicket.ID, SprintID: sprint1.ID, Status: model.EntryNotDone, PointsAtEntry: 3}
	if err := entryRepo.Create(ctx, originEntry); err != nil {
		t.Fatalf("create origin entry: %v", err)
	}

	freshEntry := &model.SprintEntry{TicketID: freshTicket.ID, SprintID: sprint2.ID, Status: model.EntryNotDone, PointsAtEntry: 5}
	if err := entryRepo.Create(ctx, freshEntry); err != nil {
		t.Fatalf("create fresh entry: %v", err)
	}
	carriedEntry := &model.SprintEntry{TicketID: carriedTicket.ID, SprintID: sprint2.ID, Status: model.EntryNotDone, CarriedFrom: &originEntry.ID, PointsAtEntry: 3}
	if err := entryRepo.Create(ctx, carriedEntry); err != nil {
		t.Fatalf("create carried entry: %v", err)
	}

	entries, err := repository.NewDashboardRepository(tx).TicketEntries(ctx, sprint2.ID)
	if err != nil {
		t.Fatalf("TicketEntries() unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("TicketEntries() = %d rows, want 2", len(entries))
	}

	byTicketID := map[int64]model.SprintEntryDetail{}
	for _, e := range entries {
		byTicketID[e.TicketID] = e
	}

	fresh := byTicketID[freshTicket.ID]
	if fresh.CarriedFromSprintName != nil {
		t.Errorf("fresh entry CarriedFromSprintName = %v, want nil", *fresh.CarriedFromSprintName)
	}
	if fresh.ProjectName != "Website Redesign" {
		t.Errorf("fresh entry ProjectName = %q, want %q", fresh.ProjectName, "Website Redesign")
	}
	if fresh.AssigneeName != "Unassigned" {
		t.Errorf("fresh entry AssigneeName = %q, want %q", fresh.AssigneeName, "Unassigned")
	}

	carried := byTicketID[carriedTicket.ID]
	if carried.CarriedFromSprintName == nil || *carried.CarriedFromSprintName != "Sprint 1" {
		t.Errorf("carried entry CarriedFromSprintName = %v, want %q", carried.CarriedFromSprintName, "Sprint 1")
	}
}

func TestDashboardRepository_TicketEntries_IncludesCancelledEntries(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Website Redesign")
	ticket := mustCreateTicket(t, tx, project.ID)

	sprintRepo := repository.NewSprintRepository(tx)
	sprint := &model.Sprint{Name: "Sprint 1", StartDate: fixedDate(2026, 8, 1), EndDate: fixedDate(2026, 8, 14), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint); err != nil {
		t.Fatalf("create sprint: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: ticket.ID, SprintID: sprint.ID, Status: model.EntryCancelled, PointsAtEntry: 8}); err != nil {
		t.Fatalf("create cancelled entry: %v", err)
	}

	entries, err := repository.NewDashboardRepository(tx).TicketEntries(ctx, sprint.ID)
	if err != nil {
		t.Fatalf("TicketEntries() unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Status != model.EntryCancelled {
		t.Fatalf("TicketEntries() = %+v, want 1 Cancelled entry", entries)
	}
}

func TestDashboardRepository_SprintPoints_SplitsCommittedFromLateAdd(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Core Platform")
	committedTicket := mustCreateTicket(t, tx, project.ID)
	lateAddTicket := mustCreateTicket(t, tx, project.ID)
	cancelledTicket := mustCreateTicket(t, tx, project.ID)

	sprintRepo := repository.NewSprintRepository(tx)
	sprint := &model.Sprint{Name: "Sprint 9", StartDate: fixedDate(2026, 9, 1), EndDate: fixedDate(2026, 9, 14), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint); err != nil {
		t.Fatalf("create sprint: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	entries := []*model.SprintEntry{
		{TicketID: committedTicket.ID, SprintID: sprint.ID, Status: model.EntryDone, AddedAfterSprintStart: false, PointsAtEntry: 5},
		{TicketID: lateAddTicket.ID, SprintID: sprint.ID, Status: model.EntryNotDone, AddedAfterSprintStart: true, PointsAtEntry: 4},
		{TicketID: cancelledTicket.ID, SprintID: sprint.ID, Status: model.EntryCancelled, AddedAfterSprintStart: false, PointsAtEntry: 8},
	}
	for _, e := range entries {
		if err := entryRepo.Create(ctx, e); err != nil {
			t.Fatalf("create sprint entry: %v", err)
		}
	}

	points, err := repository.NewDashboardRepository(tx).SprintPoints(ctx)
	if err != nil {
		t.Fatalf("SprintPoints() unexpected error: %v", err)
	}

	var got *model.SprintPoints
	for i := range points {
		if points[i].SprintID == sprint.ID {
			got = &points[i]
		}
	}
	if got == nil {
		t.Fatalf("SprintPoints() missing sprint %d in result", sprint.ID)
	}
	if got.CommittedPoints != 5 {
		t.Errorf("CommittedPoints = %d, want 5 (cancelled and late-add excluded)", got.CommittedPoints)
	}
	if got.CommittedDonePoints != 5 {
		t.Errorf("CommittedDonePoints = %d, want 5", got.CommittedDonePoints)
	}
	if got.LateAddPoints != 4 {
		t.Errorf("LateAddPoints = %d, want 4", got.LateAddPoints)
	}
}

func TestDashboardRepository_SprintPoints_AggregatesAcrossProjectsInSameSprint(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	projectA := mustCreateProject(t, tx, "Website Redesign")
	projectB := mustCreateProject(t, tx, "Internal Tooling")
	ticketA := mustCreateTicket(t, tx, projectA.ID)
	ticketB := mustCreateTicket(t, tx, projectB.ID)

	sprintRepo := repository.NewSprintRepository(tx)
	sprint := &model.Sprint{Name: "Sprint 9", StartDate: fixedDate(2026, 9, 1), EndDate: fixedDate(2026, 9, 14), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint); err != nil {
		t.Fatalf("create sprint: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: ticketA.ID, SprintID: sprint.ID, Status: model.EntryDone, PointsAtEntry: 5}); err != nil {
		t.Fatalf("create project A entry: %v", err)
	}
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: ticketB.ID, SprintID: sprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3}); err != nil {
		t.Fatalf("create project B entry: %v", err)
	}

	points, err := repository.NewDashboardRepository(tx).SprintPoints(ctx)
	if err != nil {
		t.Fatalf("SprintPoints() unexpected error: %v", err)
	}

	var got *model.SprintPoints
	for i := range points {
		if points[i].SprintID == sprint.ID {
			got = &points[i]
		}
	}
	if got == nil {
		t.Fatalf("SprintPoints() missing sprint %d in result", sprint.ID)
	}
	// Two projects' points must land in one combined row, not split apart —
	// developer capacity is shared across a team's sub-projects (ADR-0011).
	if got.CommittedPoints != 8 {
		t.Errorf("CommittedPoints = %d, want 8 (5 from Website Redesign + 3 from Internal Tooling)", got.CommittedPoints)
	}
	if got.CommittedDonePoints != 5 {
		t.Errorf("CommittedDonePoints = %d, want 5", got.CommittedDonePoints)
	}
}

func TestDashboardRepository_SprintPoints_ExcludesClosedSprintsAndArchivedProjects(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	projectRepo := repository.NewProjectRepository(tx)
	sprintRepo := repository.NewSprintRepository(tx)
	entryRepo := repository.NewSprintEntryRepository(tx)

	openProject := mustCreateProject(t, tx, "Active Project")
	openTicket := mustCreateTicket(t, tx, openProject.ID)
	openSprint := &model.Sprint{Name: "Open Sprint", StartDate: fixedDate(2026, 9, 1), EndDate: fixedDate(2026, 9, 14), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, openSprint); err != nil {
		t.Fatalf("create open sprint: %v", err)
	}
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: openTicket.ID, SprintID: openSprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3}); err != nil {
		t.Fatalf("create open entry: %v", err)
	}

	closedProject := mustCreateProject(t, tx, "Project In Closed Sprint")
	closedTicket := mustCreateTicket(t, tx, closedProject.ID)
	closedSprint := &model.Sprint{Name: "Closed Sprint", StartDate: fixedDate(2026, 8, 1), EndDate: fixedDate(2026, 8, 14), Status: model.SprintClosed}
	if err := sprintRepo.Create(ctx, closedSprint); err != nil {
		t.Fatalf("create closed sprint: %v", err)
	}
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: closedTicket.ID, SprintID: closedSprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3}); err != nil {
		t.Fatalf("create closed-sprint entry: %v", err)
	}

	archivedProject := mustCreateProject(t, tx, "Archived Project")
	archivedTicket := mustCreateTicket(t, tx, archivedProject.ID)
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: archivedTicket.ID, SprintID: openSprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3}); err != nil {
		t.Fatalf("create archived-project entry: %v", err)
	}
	archivedProject.Status = model.ProjectArchived
	if err := projectRepo.Update(ctx, archivedProject); err != nil {
		t.Fatalf("archive project: %v", err)
	}

	points, err := repository.NewDashboardRepository(tx).SprintPoints(ctx)
	if err != nil {
		t.Fatalf("SprintPoints() unexpected error: %v", err)
	}

	bySprint := map[int64]model.SprintPoints{}
	for _, p := range points {
		bySprint[p.SprintID] = p
	}
	open, ok := bySprint[openSprint.ID]
	if !ok {
		t.Fatalf("SprintPoints() missing open sprint %d in result", openSprint.ID)
	}
	// Only the active project's 3 points, not the archived project's — its
	// entry in the same open sprint must not be summed in.
	if open.CommittedPoints != 3 {
		t.Errorf("open sprint CommittedPoints = %d, want 3 (archived project's points excluded)", open.CommittedPoints)
	}
	if _, ok := bySprint[closedSprint.ID]; ok {
		t.Errorf("SprintPoints() unexpectedly includes closed sprint %d", closedSprint.ID)
	}
}

func TestDashboardRepository_SprintPoints_ProducesOneRowPerOpenSprint(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Core Platform")
	ticketInSprint8 := mustCreateTicket(t, tx, project.ID)
	ticketInSprint9 := mustCreateTicket(t, tx, project.ID)

	sprintRepo := repository.NewSprintRepository(tx)
	sprint8 := &model.Sprint{Name: "Sprint 8", StartDate: fixedDate(2026, 8, 15), EndDate: fixedDate(2026, 9, 4), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint8); err != nil {
		t.Fatalf("create sprint 8: %v", err)
	}
	sprint9 := &model.Sprint{Name: "Sprint 9", StartDate: fixedDate(2026, 9, 5), EndDate: fixedDate(2026, 9, 18), Status: model.SprintOpen}
	if err := sprintRepo.Create(ctx, sprint9); err != nil {
		t.Fatalf("create sprint 9: %v", err)
	}

	entryRepo := repository.NewSprintEntryRepository(tx)
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: ticketInSprint8.ID, SprintID: sprint8.ID, Status: model.EntryNotDone, PointsAtEntry: 5}); err != nil {
		t.Fatalf("create sprint8 entry: %v", err)
	}
	if err := entryRepo.Create(ctx, &model.SprintEntry{TicketID: ticketInSprint9.ID, SprintID: sprint9.ID, Status: model.EntryNotDone, PointsAtEntry: 7}); err != nil {
		t.Fatalf("create sprint9 entry: %v", err)
	}

	points, err := repository.NewDashboardRepository(tx).SprintPoints(ctx)
	if err != nil {
		t.Fatalf("SprintPoints() unexpected error: %v", err)
	}

	bySprint := map[int64]model.SprintPoints{}
	for _, p := range points {
		bySprint[p.SprintID] = p
	}
	if bySprint[sprint8.ID].CommittedPoints != 5 {
		t.Errorf("sprint8 row CommittedPoints = %d, want 5", bySprint[sprint8.ID].CommittedPoints)
	}
	if bySprint[sprint9.ID].CommittedPoints != 7 {
		t.Errorf("sprint9 row CommittedPoints = %d, want 7", bySprint[sprint9.ID].CommittedPoints)
	}
}
