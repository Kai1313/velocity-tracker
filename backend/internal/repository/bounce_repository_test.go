package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/repository"
)

func bounceTestSprint(t *testing.T, tx pgx.Tx, name string, status model.SprintStatus) *model.Sprint {
	t.Helper()
	// The dates are irrelevant to bounce behaviour; sprints here are told apart by creation order.
	s := &model.Sprint{Name: name, StartDate: fixedDate(2026, 8, 1), EndDate: fixedDate(2026, 8, 14), Status: status}
	if err := repository.NewSprintRepository(tx).Create(context.Background(), s); err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	return s
}

// bounceTestEntry creates an entry and sets its bounce count: nil leaves the
// column NULL ("not tracked"), mirroring entries that predate the feature. It
// writes the column directly because UpdateBounceCount refuses closed sprints,
// and several tests need counts on entries in closed sprints.
func bounceTestEntry(t *testing.T, tx pgx.Tx, ticketID, sprintID int64, status model.EntryStatus, points int, bounces *int) *model.SprintEntry {
	t.Helper()
	e := &model.SprintEntry{TicketID: ticketID, SprintID: sprintID, Status: status, PointsAtEntry: points}
	if err := repository.NewSprintEntryRepository(tx).Create(context.Background(), e); err != nil {
		t.Fatalf("create sprint entry: %v", err)
	}
	if _, err := tx.Exec(context.Background(), `UPDATE sprint_entry SET bounce_count = $1 WHERE id = $2`, bounces, e.ID); err != nil {
		t.Fatalf("set bounce_count: %v", err)
	}
	e.BounceCount = bounces
	return e
}

func bounceTestCount(n int) *int { return &n }

// noLimit asks ClosedSprintPoints for every closed sprint in the shared test DB.
const noLimit = 1 << 30

func TestSprintEntryRepository_BounceCount_DefaultsToZeroAndRoundTrips(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()
	project := mustCreateProject(t, tx, "Bounce Project")
	ticket := mustCreateTicket(t, tx, project.ID)
	sprint := bounceTestSprint(t, tx, "Sprint 1", model.SprintOpen)

	entry := &model.SprintEntry{TicketID: ticket.ID, SprintID: sprint.ID, Status: model.EntryNotDone, PointsAtEntry: 3}
	repo := repository.NewSprintEntryRepository(tx)
	if err := repo.Create(ctx, entry); err != nil {
		t.Fatalf("create: %v", err)
	}
	if entry.BounceCount == nil || *entry.BounceCount != 0 {
		t.Fatalf("new entry BounceCount = %v, want 0 (tracked from creation)", entry.BounceCount)
	}

	if err := repo.UpdateBounceCount(ctx, entry.ID, 4); err != nil {
		t.Fatalf("UpdateBounceCount() unexpected error: %v", err)
	}
	got, err := repo.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if got.BounceCount == nil || *got.BounceCount != 4 {
		t.Fatalf("Get() BounceCount = %v, want 4", got.BounceCount)
	}

	// A full-entry Update must leave the bounce count alone.
	got.Status = model.EntryDone
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	after, err := repo.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("Get() after Update() unexpected error: %v", err)
	}
	if after.BounceCount == nil || *after.BounceCount != 4 {
		t.Fatalf("BounceCount after Update() = %v, want 4", after.BounceCount)
	}
}

func TestSprintEntryRepository_UpdateBounceCount_NotFound(t *testing.T) {
	tx := withTx(t)
	err := repository.NewSprintEntryRepository(tx).UpdateBounceCount(context.Background(), 999999999, 1)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("UpdateBounceCount() for missing entry = %v, want apperr.ErrNotFound", err)
	}
}

func TestSprintEntryRepository_UpdateBounceCount_RefusesClosedSprint(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()
	project := mustCreateProject(t, tx, "Bounce Project")
	ticket := mustCreateTicket(t, tx, project.ID)
	closed := bounceTestSprint(t, tx, "Closed", model.SprintClosed)
	entry := bounceTestEntry(t, tx, ticket.ID, closed.ID, model.EntryNotDone, 3, bounceTestCount(2))

	repo := repository.NewSprintEntryRepository(tx)
	err := repo.UpdateBounceCount(ctx, entry.ID, 5)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("UpdateBounceCount() on a closed sprint's entry = %v, want apperr.ErrConflict", err)
	}
	got, err := repo.Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if got.BounceCount == nil || *got.BounceCount != 2 {
		t.Fatalf("BounceCount after refused update = %v, want it unchanged at 2", got.BounceCount)
	}
}

func TestSprintEntryRepository_UpdateBounceCount_RejectsNegative(t *testing.T) {
	tx := withTx(t)
	project := mustCreateProject(t, tx, "Bounce Project")
	ticket := mustCreateTicket(t, tx, project.ID)
	sprint := bounceTestSprint(t, tx, "Sprint 1", model.SprintOpen)
	entry := bounceTestEntry(t, tx, ticket.ID, sprint.ID, model.EntryNotDone, 3, bounceTestCount(0))

	if err := repository.NewSprintEntryRepository(tx).UpdateBounceCount(context.Background(), entry.ID, -1); err == nil {
		t.Fatal("UpdateBounceCount(-1) succeeded, want the CHECK constraint to reject it")
	}
}

func TestDashboardRepository_DeveloperBreakdown_BounceFiguresCoverTrackedEntriesOnly(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()
	project := mustCreateProject(t, tx, "Bounce Project")
	sprint := bounceTestSprint(t, tx, "Sprint 1", model.SprintOpen)

	// Unassigned tickets: tracked 2 + 0, untracked 1, cancelled 1 (excluded).
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryNotDone, 3, bounceTestCount(2))
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryDone, 3, bounceTestCount(0))
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryDone, 3, nil)
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryCancelled, 3, bounceTestCount(9))

	breakdown, err := repository.NewDashboardRepository(tx).DeveloperBreakdown(ctx, sprint.ID)
	if err != nil {
		t.Fatalf("DeveloperBreakdown() unexpected error: %v", err)
	}
	if len(breakdown) != 1 || breakdown[0].UserID != nil {
		t.Fatalf("breakdown = %+v, want a single Unassigned row", breakdown)
	}
	got := breakdown[0]
	if got.Bounces != 2 || got.BounceTrackedTickets != 2 {
		t.Errorf("(Bounces, BounceTrackedTickets) = (%d, %d), want (2, 2)", got.Bounces, got.BounceTrackedTickets)
	}
}

func TestDashboardRepository_DeveloperBreakdown_ReportsZeroTrackedTicketsWhenNothingTracked(t *testing.T) {
	tx := withTx(t)
	project := mustCreateProject(t, tx, "Bounce Project")
	sprint := bounceTestSprint(t, tx, "Old Sprint", model.SprintClosed)
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryDone, 3, nil)

	breakdown, err := repository.NewDashboardRepository(tx).DeveloperBreakdown(context.Background(), sprint.ID)
	if err != nil {
		t.Fatalf("DeveloperBreakdown() unexpected error: %v", err)
	}
	if len(breakdown) != 1 || breakdown[0].BounceTrackedTickets != 0 || breakdown[0].Bounces != 0 {
		t.Fatalf("breakdown = %+v, want one row with 0 tracked tickets and 0 bounces", breakdown)
	}
}

func TestDashboardRepository_TicketEntries_ReturnsBounceCount(t *testing.T) {
	tx := withTx(t)
	project := mustCreateProject(t, tx, "Bounce Project")
	sprint := bounceTestSprint(t, tx, "Sprint 1", model.SprintOpen)
	tracked := mustCreateTicket(t, tx, project.ID)
	untracked := mustCreateTicket(t, tx, project.ID)
	bounceTestEntry(t, tx, tracked.ID, sprint.ID, model.EntryNotDone, 3, bounceTestCount(2))
	bounceTestEntry(t, tx, untracked.ID, sprint.ID, model.EntryNotDone, 3, nil)

	entries, err := repository.NewDashboardRepository(tx).TicketEntries(context.Background(), sprint.ID)
	if err != nil {
		t.Fatalf("TicketEntries() unexpected error: %v", err)
	}
	for _, e := range entries {
		switch e.TicketID {
		case tracked.ID:
			if e.BounceCount == nil || *e.BounceCount != 2 {
				t.Errorf("tracked BounceCount = %v, want 2", e.BounceCount)
			}
		case untracked.ID:
			if e.BounceCount != nil {
				t.Errorf("untracked BounceCount = %v, want nil", *e.BounceCount)
			}
		}
	}
}

// Earlier bounces sum the ticket's non-Cancelled entries in earlier sprints
// (sprint ID order); untracked entries add nothing.
func TestDashboardRepository_StuckTickets_EarlierBouncesSkipCancelledAndUntracked(t *testing.T) {
	tx := withTx(t)
	project := mustCreateProject(t, tx, "Bounce Project")
	sprint1 := bounceTestSprint(t, tx, "Sprint 1", model.SprintClosed)
	sprint2 := bounceTestSprint(t, tx, "Sprint 2", model.SprintClosed)
	sprint3 := bounceTestSprint(t, tx, "Sprint 3", model.SprintOpen)

	counted := mustCreateTicket(t, tx, project.ID)
	bounceTestEntry(t, tx, counted.ID, sprint1.ID, model.EntryNotDone, 3, bounceTestCount(2))
	bounceTestEntry(t, tx, counted.ID, sprint2.ID, model.EntryNotDone, 3, nil) // untracked: adds nothing
	bounceTestEntry(t, tx, counted.ID, sprint3.ID, model.EntryNotDone, 3, bounceTestCount(1))

	cancelledEarlier := mustCreateTicket(t, tx, project.ID)
	bounceTestEntry(t, tx, cancelledEarlier.ID, sprint1.ID, model.EntryCancelled, 3, bounceTestCount(7))
	bounceTestEntry(t, tx, cancelledEarlier.ID, sprint3.ID, model.EntryNotDone, 3, bounceTestCount(0)) // Open sprints allow one entry per ticket

	stuck, err := repository.NewDashboardRepository(tx).StuckTickets(context.Background(), 1)
	if err != nil {
		t.Fatalf("StuckTickets() unexpected error: %v", err)
	}
	var found bool
	for _, st := range stuck {
		switch st.TicketID {
		case counted.ID:
			found = true
			if st.BounceCount != 1 || st.EarlierBounces != 2 || st.TotalBounces != 3 {
				t.Errorf("counted = %+v, want (1 now, 2 earlier, 3 total)", st)
			}
		case cancelledEarlier.ID:
			t.Errorf("ticket whose only earlier bounces were on a Cancelled entry was reported stuck: %+v", st)
		}
	}
	if !found {
		t.Fatal("StuckTickets() missing the ticket with earlier bounces")
	}
}

func TestDashboardRepository_ClosedSprintPoints_BounceTotalsExcludeUntrackedAndCancelled(t *testing.T) {
	tx := withTx(t)
	project := mustCreateProject(t, tx, "Bounce Project")
	sprint := bounceTestSprint(t, tx, "Closed Sprint", model.SprintClosed)
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryDone, 3, bounceTestCount(2))
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryDone, 3, bounceTestCount(0))
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryNotDone, 3, bounceTestCount(1))
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryNotDone, 3, nil)
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprint.ID, model.EntryCancelled, 3, bounceTestCount(5))

	points, err := repository.NewDashboardRepository(tx).ClosedSprintPoints(context.Background(), noLimit)
	if err != nil {
		t.Fatalf("ClosedSprintPoints() unexpected error: %v", err)
	}
	for _, p := range points {
		if p.SprintID != sprint.ID {
			continue
		}
		if p.BounceTrackedTickets != 3 || p.TotalBounces != 3 || p.BouncedTickets != 2 {
			t.Errorf("(tracked, total, bounced) = (%d, %d, %d), want (3, 3, 2)", p.BounceTrackedTickets, p.TotalBounces, p.BouncedTickets)
		}
		return
	}
	t.Fatalf("ClosedSprintPoints() missing sprint %d", sprint.ID)
}

func TestDashboardRepository_StuckTickets_NotDoneOpenSprintActiveProjectAtThreshold(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()
	project := mustCreateProject(t, tx, "Bounce Project")
	archived := mustCreateProject(t, tx, "Archived Bounce Project")
	archived.Status = model.ProjectArchived

	oldSprint := bounceTestSprint(t, tx, "Old", model.SprintClosed)
	open := bounceTestSprint(t, tx, "Open", model.SprintOpen)

	// carried: 2 earlier bounces, 0 now -> stuck from day one.
	carried := mustCreateTicket(t, tx, project.ID)
	bounceTestEntry(t, tx, carried.ID, oldSprint.ID, model.EntryNotDone, 5, bounceTestCount(2))
	bounceTestEntry(t, tx, carried.ID, open.ID, model.EntryNotDone, 5, bounceTestCount(0))
	// fresh: 2 bounces in this sprint alone.
	fresh := mustCreateTicket(t, tx, project.ID)
	bounceTestEntry(t, tx, fresh.ID, open.ID, model.EntryNotDone, 3, bounceTestCount(2))
	// not stuck: below threshold, already Done, or untracked.
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, open.ID, model.EntryNotDone, 3, bounceTestCount(1))
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, open.ID, model.EntryDone, 3, bounceTestCount(4))
	bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, open.ID, model.EntryNotDone, 3, nil)
	// archived project: excluded like Sprint Health.
	archivedTicket := mustCreateTicket(t, tx, archived.ID)
	bounceTestEntry(t, tx, archivedTicket.ID, open.ID, model.EntryNotDone, 3, bounceTestCount(5))
	if err := repository.NewProjectRepository(tx).Update(ctx, archived); err != nil {
		t.Fatalf("archive project: %v", err)
	}

	all, err := repository.NewDashboardRepository(tx).StuckTickets(ctx, 2)
	if err != nil {
		t.Fatalf("StuckTickets() unexpected error: %v", err)
	}
	var ours []model.StuckTicket
	for _, st := range all {
		if st.SprintID == open.ID {
			ours = append(ours, st)
		}
	}
	if len(ours) != 2 {
		t.Fatalf("stuck tickets in our sprint = %+v, want exactly the carried and fresh tickets", ours)
	}
	byTicket := map[int64]model.StuckTicket{}
	for _, st := range ours {
		byTicket[st.TicketID] = st
	}
	if st := byTicket[carried.ID]; st.BounceCount != 0 || st.EarlierBounces != 2 || st.TotalBounces != 2 || st.AssigneeName != "Unassigned" {
		t.Errorf("carried = %+v, want (0 now, 2 earlier, 2 total, Unassigned)", st)
	}
	if st := byTicket[fresh.ID]; st.BounceCount != 2 || st.EarlierBounces != 0 || st.TotalBounces != 2 {
		t.Errorf("fresh = %+v, want (2 now, 0 earlier, 2 total)", st)
	}
}

func TestDashboardRepository_BounceBySize_BandsClosedTrackedNonCancelledEntries(t *testing.T) {
	tx := withTx(t)
	ctx := context.Background()
	dash := repository.NewDashboardRepository(tx)

	// The query spans every closed sprint in the database, so assert on the
	// change this test's own rows cause rather than on absolute totals.
	before, err := dash.BounceBySize(ctx)
	if err != nil {
		t.Fatalf("BounceBySize() before: %v", err)
	}

	project := mustCreateProject(t, tx, "Bounce Project")
	closed := bounceTestSprint(t, tx, "Closed", model.SprintClosed)
	open := bounceTestSprint(t, tx, "Open", model.SprintOpen)
	add := func(status model.EntryStatus, points int, bounces *int, sprintID int64) {
		bounceTestEntry(t, tx, mustCreateTicket(t, tx, project.ID).ID, sprintID, status, points, bounces)
	}
	add(model.EntryDone, 2, bounceTestCount(1), closed.ID)
	add(model.EntryDone, 3, bounceTestCount(0), closed.ID)
	add(model.EntryNotDone, 5, bounceTestCount(2), closed.ID)
	add(model.EntryDone, 8, bounceTestCount(1), closed.ID)
	add(model.EntryDone, 6, bounceTestCount(0), closed.ID)
	add(model.EntryDone, 1, nil, closed.ID)                     // untracked: ignored
	add(model.EntryCancelled, 1, bounceTestCount(3), closed.ID) // cancelled: ignored
	add(model.EntryNotDone, 1, bounceTestCount(3), open.ID)     // open sprint: ignored

	after, err := dash.BounceBySize(ctx)
	if err != nil {
		t.Fatalf("BounceBySize() after: %v", err)
	}
	wantBands := []string{"1-2", "3-5", "6+"}
	if len(before) != 3 || len(after) != 3 {
		t.Fatalf("band counts = (%d, %d), want 3 bands always", len(before), len(after))
	}
	// delta = (entries, bouncedEntries, totalBounces) per band
	wantDelta := [][3]int{{1, 1, 1}, {2, 1, 2}, {2, 1, 1}}
	for i, band := range wantBands {
		if after[i].Band != band {
			t.Fatalf("band[%d] = %q, want %q (ordered by size)", i, after[i].Band, band)
		}
		got := [3]int{
			after[i].Entries - before[i].Entries,
			after[i].BouncedEntries - before[i].BouncedEntries,
			after[i].TotalBounces - before[i].TotalBounces,
		}
		if got != wantDelta[i] {
			t.Errorf("band %s delta (entries, bounced, total) = %v, want %v", band, got, wantDelta[i])
		}
	}
}
