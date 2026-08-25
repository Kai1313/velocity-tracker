package service_test

import (
	"context"
	"errors"
	"testing"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type fakeProjectRepo struct {
	projects map[int64]*model.Project
	nextID   int64
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{projects: map[int64]*model.Project{}}
}

func (f *fakeProjectRepo) Create(ctx context.Context, p *model.Project) error {
	f.nextID++
	p.ID = f.nextID
	cp := *p
	f.projects[p.ID] = &cp
	return nil
}

func (f *fakeProjectRepo) Get(ctx context.Context, id int64) (*model.Project, error) {
	p, ok := f.projects[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (f *fakeProjectRepo) List(ctx context.Context) ([]model.Project, error) {
	var out []model.Project
	for _, p := range f.projects {
		out = append(out, *p)
	}
	return out, nil
}

func (f *fakeProjectRepo) Update(ctx context.Context, p *model.Project) error {
	if _, ok := f.projects[p.ID]; !ok {
		return apperr.ErrNotFound
	}
	cp := *p
	f.projects[p.ID] = &cp
	return nil
}

func (f *fakeProjectRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := f.projects[id]; !ok {
		return apperr.ErrNotFound
	}
	delete(f.projects, id)
	return nil
}

func TestProjectService_Create_RejectsBlankName(t *testing.T) {
	svc := service.NewProjectService(newFakeProjectRepo())
	_, err := svc.Create(context.Background(), "   ")
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
	}
}

func TestProjectService_Create_DefaultsToActive(t *testing.T) {
	svc := service.NewProjectService(newFakeProjectRepo())
	p, err := svc.Create(context.Background(), "Website Redesign")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if p.Status != model.ProjectActive {
		t.Fatalf("Create() status = %v, want Active", p.Status)
	}
}

func TestProjectService_Update_RejectsInvalidStatus(t *testing.T) {
	repo := newFakeProjectRepo()
	svc := service.NewProjectService(repo)
	p, _ := svc.Create(context.Background(), "Website Redesign")

	_, err := svc.Update(context.Background(), p.ID, "Website Redesign", model.ProjectStatus("Bogus"))
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("Update() = %v, want apperr.ErrValidation", err)
	}
}
