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

type fakeSprintRepo struct {
	sprints map[int64]*model.Sprint
	nextID  int64
}

func newFakeSprintRepo() *fakeSprintRepo {
	return &fakeSprintRepo{sprints: map[int64]*model.Sprint{}}
}

func (f *fakeSprintRepo) Create(ctx context.Context, s *model.Sprint) error {
	f.nextID++
	s.ID = f.nextID
	cp := *s
	f.sprints[s.ID] = &cp
	return nil
}

func (f *fakeSprintRepo) Get(ctx context.Context, id int64) (*model.Sprint, error) {
	s, ok := f.sprints[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *fakeSprintRepo) List(ctx context.Context) ([]model.Sprint, error) {
	var out []model.Sprint
	for _, s := range f.sprints {
		out = append(out, *s)
	}
	return out, nil
}

func (f *fakeSprintRepo) Update(ctx context.Context, s *model.Sprint) error {
	if _, ok := f.sprints[s.ID]; !ok {
		return apperr.ErrNotFound
	}
	cp := *s
	f.sprints[s.ID] = &cp
	return nil
}

func (f *fakeSprintRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := f.sprints[id]; !ok {
		return apperr.ErrNotFound
	}
	delete(f.sprints, id)
	return nil
}

func TestSprintService_Create_RejectsBadDateRange(t *testing.T) {
	svc := service.NewSprintService(newFakeSprintRepo())
	start := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	_, err := svc.Create(context.Background(), "Sprint 1", start, end)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("Create() with end before start = %v, want apperr.ErrValidation", err)
	}
}

func TestSprintService_Create_AlwaysStartsOpen(t *testing.T) {
	svc := service.NewSprintService(newFakeSprintRepo())
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)

	sp, err := svc.Create(context.Background(), "Sprint 1", start, end)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if sp.Status != model.SprintOpen {
		t.Fatalf("Create() status = %v, want Open", sp.Status)
	}
}

func TestSprintService_Update_RejectsInvalidStatus(t *testing.T) {
	repo := newFakeSprintRepo()
	svc := service.NewSprintService(repo)
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	sp, _ := svc.Create(context.Background(), "Sprint 1", start, end)

	_, err := svc.Update(context.Background(), sp.ID, "Sprint 1", start, end, model.SprintStatus("Bogus"))
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("Update() with bad status = %v, want apperr.ErrValidation", err)
	}
}

func TestSprintService_Update_CanClose(t *testing.T) {
	repo := newFakeSprintRepo()
	svc := service.NewSprintService(repo)
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	sp, _ := svc.Create(context.Background(), "Sprint 1", start, end)

	updated, err := svc.Update(context.Background(), sp.ID, "Sprint 1", start, end, model.SprintClosed)
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if updated.Status != model.SprintClosed {
		t.Fatalf("Update() status = %v, want Closed", updated.Status)
	}
}
