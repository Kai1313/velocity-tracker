package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
)

type SprintRepository interface {
	Create(ctx context.Context, s *model.Sprint) error
	Get(ctx context.Context, id int64) (*model.Sprint, error)
	List(ctx context.Context) ([]model.Sprint, error)
	Update(ctx context.Context, s *model.Sprint) error
	Delete(ctx context.Context, id int64) error
}

type SprintService struct {
	repo SprintRepository
}

func NewSprintService(repo SprintRepository) *SprintService {
	return &SprintService{repo: repo}
}

func validateSprintDates(name string, start, end time.Time) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("%w: name is required", apperr.ErrValidation)
	}
	if !start.Before(end) {
		return "", fmt.Errorf("%w: startDate must be before endDate", apperr.ErrValidation)
	}
	return name, nil
}

// Create always starts a sprint as Open — closing is an explicit later
// action (Update), never part of creation.
func (s *SprintService) Create(ctx context.Context, name string, start, end time.Time) (*model.Sprint, error) {
	name, err := validateSprintDates(name, start, end)
	if err != nil {
		return nil, err
	}
	sp := &model.Sprint{Name: name, StartDate: start, EndDate: end, Status: model.SprintOpen}
	if err := s.repo.Create(ctx, sp); err != nil {
		return nil, err
	}
	return sp, nil
}

func (s *SprintService) Get(ctx context.Context, id int64) (*model.Sprint, error) {
	return s.repo.Get(ctx, id)
}

func (s *SprintService) List(ctx context.Context) ([]model.Sprint, error) {
	return s.repo.List(ctx)
}

func (s *SprintService) Update(ctx context.Context, id int64, name string, start, end time.Time, status model.SprintStatus) (*model.Sprint, error) {
	name, err := validateSprintDates(name, start, end)
	if err != nil {
		return nil, err
	}
	switch status {
	case model.SprintOpen, model.SprintClosed:
	default:
		return nil, fmt.Errorf("%w: status must be Open or Closed", apperr.ErrValidation)
	}
	sp := &model.Sprint{ID: id, Name: name, StartDate: start, EndDate: end, Status: status}
	if err := s.repo.Update(ctx, sp); err != nil {
		return nil, err
	}
	return sp, nil
}

func (s *SprintService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
