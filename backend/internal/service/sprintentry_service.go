package service

import (
	"context"
	"fmt"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
)

type SprintEntryRepository interface {
	Create(ctx context.Context, e *model.SprintEntry) error
	Get(ctx context.Context, id int64) (*model.SprintEntry, error)
	List(ctx context.Context) ([]model.SprintEntry, error)
	Update(ctx context.Context, e *model.SprintEntry) error
	Delete(ctx context.Context, id int64) error
}

// SprintLookup is the narrow slice of SprintRepository the SprintEntry
// service needs, to check whether an entry's parent sprint is closed.
type SprintLookup interface {
	Get(ctx context.Context, id int64) (*model.Sprint, error)
}

type SprintEntryService struct {
	repo    SprintEntryRepository
	sprints SprintLookup
}

func NewSprintEntryService(repo SprintEntryRepository, sprints SprintLookup) *SprintEntryService {
	return &SprintEntryService{repo: repo, sprints: sprints}
}

func validateEntryFields(ticketID, sprintID int64, status model.EntryStatus, points int) error {
	if ticketID <= 0 {
		return fmt.Errorf("%w: ticketId is required", apperr.ErrValidation)
	}
	if sprintID <= 0 {
		return fmt.Errorf("%w: sprintId is required", apperr.ErrValidation)
	}
	if points <= 0 {
		return fmt.Errorf("%w: pointsAtEntry must be positive", apperr.ErrValidation)
	}
	switch status {
	case model.EntryDone, model.EntryNotDone, model.EntryCancelled:
	default:
		return fmt.Errorf("%w: status must be Done, NotDone, or Cancelled", apperr.ErrValidation)
	}
	return nil
}

// validateCarriedFrom enforces that a carriedFrom reference actually denotes
// a carry-over: the same ticket's own NotDone entry from a sprint that has
// already closed. Without this, carriedFrom is just an unchecked foreign key
// and can point at an unrelated ticket, a Done/Cancelled entry, or an entry
// in a still-open sprint — silently corrupting carry-over metrics.
func (s *SprintEntryService) validateCarriedFrom(ctx context.Context, ticketID int64, carriedFrom *int64) error {
	if carriedFrom == nil {
		return nil
	}
	source, err := s.repo.Get(ctx, *carriedFrom)
	if err != nil {
		return fmt.Errorf("%w: carriedFrom entry %d not found", apperr.ErrValidation, *carriedFrom)
	}
	if source.TicketID != ticketID {
		return fmt.Errorf("%w: carriedFrom entry %d belongs to a different ticket", apperr.ErrValidation, *carriedFrom)
	}
	if source.Status != model.EntryNotDone {
		return fmt.Errorf("%w: carriedFrom entry %d must have status NotDone", apperr.ErrValidation, *carriedFrom)
	}
	sourceSprint, err := s.sprints.Get(ctx, source.SprintID)
	if err != nil {
		return fmt.Errorf("%w: carriedFrom entry %d's sprint not found", apperr.ErrValidation, *carriedFrom)
	}
	if sourceSprint.Status != model.SprintClosed {
		return fmt.Errorf("%w: carriedFrom entry %d's sprint must be Closed", apperr.ErrValidation, *carriedFrom)
	}
	return nil
}

func (s *SprintEntryService) Create(ctx context.Context, ticketID, sprintID int64, status model.EntryStatus, addedAfterStart bool, carriedFrom *int64, points int) (*model.SprintEntry, error) {
	if err := validateEntryFields(ticketID, sprintID, status, points); err != nil {
		return nil, err
	}
	if err := s.validateCarriedFrom(ctx, ticketID, carriedFrom); err != nil {
		return nil, err
	}
	e := &model.SprintEntry{
		TicketID:              ticketID,
		SprintID:              sprintID,
		Status:                status,
		AddedAfterSprintStart: addedAfterStart,
		CarriedFrom:           carriedFrom,
		PointsAtEntry:         points,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *SprintEntryService) Get(ctx context.Context, id int64) (*model.SprintEntry, error) {
	return s.repo.Get(ctx, id)
}

func (s *SprintEntryService) List(ctx context.Context) ([]model.SprintEntry, error) {
	return s.repo.List(ctx)
}

// Update rejects edits once the entry's parent sprint is Closed — a closed
// sprint's SprintEntry rows are locked history, per CONTEXT.md's "Close
// Sprint" definition.
func (s *SprintEntryService) Update(ctx context.Context, id int64, status model.EntryStatus, addedAfterStart bool, carriedFrom *int64, points int) (*model.SprintEntry, error) {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := validateEntryFields(existing.TicketID, existing.SprintID, status, points); err != nil {
		return nil, err
	}
	if err := s.validateCarriedFrom(ctx, existing.TicketID, carriedFrom); err != nil {
		return nil, err
	}

	sprint, err := s.sprints.Get(ctx, existing.SprintID)
	if err != nil {
		return nil, err
	}
	if sprint.Status == model.SprintClosed {
		return nil, fmt.Errorf("%w: sprint %d is closed, its sprint entries are locked", apperr.ErrConflict, sprint.ID)
	}

	existing.Status = status
	existing.AddedAfterSprintStart = addedAfterStart
	existing.CarriedFrom = carriedFrom
	existing.PointsAtEntry = points
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *SprintEntryService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
