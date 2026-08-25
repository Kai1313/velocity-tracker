package service

import (
	"context"
	"fmt"
	"strings"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
)

type ProjectRepository interface {
	Create(ctx context.Context, p *model.Project) error
	Get(ctx context.Context, id int64) (*model.Project, error)
	List(ctx context.Context) ([]model.Project, error)
	Update(ctx context.Context, p *model.Project) error
	Delete(ctx context.Context, id int64) error
}

type ProjectService struct {
	repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, name string) (*model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", apperr.ErrValidation)
	}
	p := &model.Project{Name: name, Status: model.ProjectActive}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Get(ctx context.Context, id int64) (*model.Project, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProjectService) List(ctx context.Context) ([]model.Project, error) {
	return s.repo.List(ctx)
}

func (s *ProjectService) Update(ctx context.Context, id int64, name string, status model.ProjectStatus) (*model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", apperr.ErrValidation)
	}
	switch status {
	case model.ProjectActive, model.ProjectArchived:
	default:
		return nil, fmt.Errorf("%w: status must be Active or Archived", apperr.ErrValidation)
	}
	p := &model.Project{ID: id, Name: name, Status: status}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
