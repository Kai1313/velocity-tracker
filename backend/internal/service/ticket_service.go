package service

import (
	"context"
	"fmt"
	"strings"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
)

type TicketRepository interface {
	Create(ctx context.Context, t *model.Ticket) error
	Get(ctx context.Context, id int64) (*model.TicketDetail, error)
	List(ctx context.Context) ([]model.TicketDetail, error)
	Update(ctx context.Context, t *model.Ticket) error
	Delete(ctx context.Context, id int64) error
}

type TicketService struct {
	repo TicketRepository
}

func NewTicketService(repo TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func validateTicketFields(projectID int64, title string, storyPoints int) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("%w: title is required", apperr.ErrValidation)
	}
	if projectID <= 0 {
		return "", fmt.Errorf("%w: projectId is required", apperr.ErrValidation)
	}
	if storyPoints <= 0 {
		return "", fmt.Errorf("%w: storyPoints must be positive", apperr.ErrValidation)
	}
	return title, nil
}

func (s *TicketService) Create(ctx context.Context, projectID int64, title string, storyPoints int, assigneeID *int64) (*model.Ticket, error) {
	title, err := validateTicketFields(projectID, title, storyPoints)
	if err != nil {
		return nil, err
	}
	t := &model.Ticket{ProjectID: projectID, Title: title, StoryPoints: storyPoints, AssigneeID: assigneeID}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TicketService) Get(ctx context.Context, id int64) (*model.TicketDetail, error) {
	return s.repo.Get(ctx, id)
}

func (s *TicketService) List(ctx context.Context) ([]model.TicketDetail, error) {
	return s.repo.List(ctx)
}

func (s *TicketService) Update(ctx context.Context, id, projectID int64, title string, storyPoints int, assigneeID *int64) (*model.Ticket, error) {
	title, err := validateTicketFields(projectID, title, storyPoints)
	if err != nil {
		return nil, err
	}
	t := &model.Ticket{ID: id, ProjectID: projectID, Title: title, StoryPoints: storyPoints, AssigneeID: assigneeID}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TicketService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
