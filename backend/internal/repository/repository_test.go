package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/repository"
)

func fixedDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// withTx opens a transaction against TEST_DATABASE_URL and rolls it back
// when the test finishes, so tests never leave data behind or interfere
// with each other despite sharing one real Postgres instance.
func withTx(t *testing.T) pgx.Tx {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := repository.RunMigrations(url); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	return tx
}

func mustCreateProject(t *testing.T, tx pgx.Tx, name string) *model.Project {
	t.Helper()
	repo := repository.NewProjectRepository(tx)
	p := &model.Project{Name: name, Status: model.ProjectActive}
	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("create project: %v", err)
	}
	return p
}

func mustCreateTicket(t *testing.T, tx pgx.Tx, projectID int64) *model.Ticket {
	t.Helper()
	repo := repository.NewTicketRepository(tx)
	ticket := &model.Ticket{ProjectID: projectID, Title: "Ticket", StoryPoints: 3}
	if err := repo.Create(context.Background(), ticket); err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	return ticket
}

func TestProjectRepository_Delete_BlockedWhileReferenced(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Website Redesign")
	mustCreateTicket(t, tx, project.ID)

	err := repository.NewProjectRepository(tx).Delete(ctx, project.ID)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("Delete() referenced project = %v, want apperr.ErrConflict", err)
	}
}

func TestProjectRepository_Delete_AllowedWhenUnreferenced(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	project := mustCreateProject(t, tx, "Unused Project")

	if err := repository.NewProjectRepository(tx).Delete(ctx, project.ID); err != nil {
		t.Fatalf("Delete() unreferenced project unexpected error: %v", err)
	}
}

func TestTicketRepository_Create_RejectsUnknownProject(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()

	ticket := &model.Ticket{ProjectID: 999999, Title: "Orphan", StoryPoints: 1}
	err := repository.NewTicketRepository(tx).Create(ctx, ticket)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("Create() with unknown project = %v, want apperr.ErrValidation", err)
	}
}

func TestTicketRepository_Get_ComputesCurrentStatusFromLatestSprintEntry(t *testing.T) {
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
	entry := &model.SprintEntry{TicketID: ticket.ID, SprintID: sprint.ID, Status: model.EntryNotDone, PointsAtEntry: ticket.StoryPoints}
	if err := entryRepo.Create(ctx, entry); err != nil {
		t.Fatalf("create sprint entry: %v", err)
	}

	detail, err := repository.NewTicketRepository(tx).Get(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("get ticket: %v", err)
	}
	if detail.CurrentStatus == nil || *detail.CurrentStatus != model.EntryNotDone {
		t.Fatalf("Get() currentStatus = %v, want NotDone", detail.CurrentStatus)
	}
}

func TestSprintEntryRepository_UniquePerTicketAndSprint(t *testing.T) {
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
	first := &model.SprintEntry{TicketID: ticket.ID, SprintID: sprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3}
	if err := entryRepo.Create(ctx, first); err != nil {
		t.Fatalf("create first entry: %v", err)
	}

	second := &model.SprintEntry{TicketID: ticket.ID, SprintID: sprint.ID, Status: model.EntryDone, PointsAtEntry: 3}
	err := entryRepo.Create(ctx, second)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("Create() duplicate ticket/sprint entry = %v, want apperr.ErrConflict", err)
	}
}
