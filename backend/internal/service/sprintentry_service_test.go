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

type fakeSprintEntryRepo struct {
	entries map[int64]*model.SprintEntry
	nextID  int64
}

func newFakeSprintEntryRepo() *fakeSprintEntryRepo {
	return &fakeSprintEntryRepo{entries: map[int64]*model.SprintEntry{}}
}

func (f *fakeSprintEntryRepo) Create(ctx context.Context, e *model.SprintEntry) error {
	f.nextID++
	e.ID = f.nextID
	e.CreatedAt = time.Now()
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeSprintEntryRepo) Get(ctx context.Context, id int64) (*model.SprintEntry, error) {
	e, ok := f.entries[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *fakeSprintEntryRepo) List(ctx context.Context) ([]model.SprintEntry, error) {
	var out []model.SprintEntry
	for _, e := range f.entries {
		out = append(out, *e)
	}
	return out, nil
}

func (f *fakeSprintEntryRepo) Update(ctx context.Context, e *model.SprintEntry) error {
	if _, ok := f.entries[e.ID]; !ok {
		return apperr.ErrNotFound
	}
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeSprintEntryRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := f.entries[id]; !ok {
		return apperr.ErrNotFound
	}
	delete(f.entries, id)
	return nil
}

type fakeSprintLookup struct {
	sprints map[int64]*model.Sprint
}

func (f *fakeSprintLookup) Get(ctx context.Context, id int64) (*model.Sprint, error) {
	s, ok := f.sprints[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return s, nil
}

func TestSprintEntryService_Update_RejectedWhenSprintClosed(t *testing.T) {
	repo := newFakeSprintEntryRepo()
	sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{
		10: {ID: 10, Status: model.SprintClosed},
	}}
	svc := service.NewSprintEntryService(repo, sprints)

	entry, err := svc.Create(context.Background(), 1, 10, model.EntryNotDone, false, nil, 5)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	_, err = svc.Update(context.Background(), entry.ID, model.EntryDone, false, nil, 5)
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("Update() on closed sprint = %v, want apperr.ErrConflict", err)
	}
}

func TestSprintEntryService_Update_AllowedWhenSprintOpen(t *testing.T) {
	repo := newFakeSprintEntryRepo()
	sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{
		10: {ID: 10, Status: model.SprintOpen},
	}}
	svc := service.NewSprintEntryService(repo, sprints)

	entry, err := svc.Create(context.Background(), 1, 10, model.EntryNotDone, false, nil, 5)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	updated, err := svc.Update(context.Background(), entry.ID, model.EntryDone, false, nil, 5)
	if err != nil {
		t.Fatalf("Update() on open sprint unexpected error: %v", err)
	}
	if updated.Status != model.EntryDone {
		t.Fatalf("Update() status = %v, want Done", updated.Status)
	}
}

func TestSprintEntryService_Create_ValidatesCarriedFrom(t *testing.T) {
	closedSprint := &model.Sprint{ID: 10, Status: model.SprintClosed}
	openSprint := &model.Sprint{ID: 11, Status: model.SprintOpen}

	newSvc := func() (*service.SprintEntryService, *fakeSprintEntryRepo) {
		repo := newFakeSprintEntryRepo()
		sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{10: closedSprint, 11: openSprint}}
		return service.NewSprintEntryService(repo, sprints), repo
	}

	t.Run("missing carriedFrom entry", func(t *testing.T) {
		svc, _ := newSvc()
		missing := int64(999)
		_, err := svc.Create(context.Background(), 1, 11, model.EntryDone, false, &missing, 5)
		if !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
		}
	})

	t.Run("different ticket", func(t *testing.T) {
		svc, _ := newSvc()
		source, err := svc.Create(context.Background(), 2, 10, model.EntryNotDone, false, nil, 5)
		if err != nil {
			t.Fatalf("setup Create() unexpected error: %v", err)
		}
		_, err = svc.Create(context.Background(), 1, 11, model.EntryDone, false, &source.ID, 5)
		if !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
		}
	})

	t.Run("source not NotDone", func(t *testing.T) {
		svc, _ := newSvc()
		source, err := svc.Create(context.Background(), 1, 10, model.EntryDone, false, nil, 5)
		if err != nil {
			t.Fatalf("setup Create() unexpected error: %v", err)
		}
		_, err = svc.Create(context.Background(), 1, 11, model.EntryDone, false, &source.ID, 5)
		if !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
		}
	})

	t.Run("source sprint still open", func(t *testing.T) {
		svc, _ := newSvc()
		source, err := svc.Create(context.Background(), 1, 11, model.EntryNotDone, false, nil, 5)
		if err != nil {
			t.Fatalf("setup Create() unexpected error: %v", err)
		}
		_, err = svc.Create(context.Background(), 1, 10, model.EntryNotDone, false, &source.ID, 5)
		if !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
		}
	})

	t.Run("valid carry-over", func(t *testing.T) {
		svc, _ := newSvc()
		source, err := svc.Create(context.Background(), 1, 10, model.EntryNotDone, false, nil, 5)
		if err != nil {
			t.Fatalf("setup Create() unexpected error: %v", err)
		}
		carried, err := svc.Create(context.Background(), 1, 11, model.EntryDone, false, &source.ID, 5)
		if err != nil {
			t.Fatalf("Create() unexpected error: %v", err)
		}
		if carried.CarriedFrom == nil || *carried.CarriedFrom != source.ID {
			t.Fatalf("Create() carriedFrom = %v, want %d", carried.CarriedFrom, source.ID)
		}
	})
}

func TestSprintEntryService_Create_ValidatesFields(t *testing.T) {
	repo := newFakeSprintEntryRepo()
	sprints := &fakeSprintLookup{sprints: map[int64]*model.Sprint{}}
	svc := service.NewSprintEntryService(repo, sprints)

	cases := []struct {
		name     string
		ticketID int64
		sprintID int64
		status   model.EntryStatus
		points   int
	}{
		{"missing ticket", 0, 1, model.EntryDone, 5},
		{"missing sprint", 1, 0, model.EntryDone, 5},
		{"zero points", 1, 1, model.EntryDone, 0},
		{"invalid status", 1, 1, model.EntryStatus("Bogus"), 5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), c.ticketID, c.sprintID, c.status, false, nil, c.points)
			if !errors.Is(err, apperr.ErrValidation) {
				t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
			}
		})
	}
}
