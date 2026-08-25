package service_test

import (
	"context"
	"errors"
	"testing"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type fakeTicketRepo struct {
	tickets map[int64]*model.Ticket
	nextID  int64
}

func newFakeTicketRepo() *fakeTicketRepo {
	return &fakeTicketRepo{tickets: map[int64]*model.Ticket{}}
}

func (f *fakeTicketRepo) Create(ctx context.Context, t *model.Ticket) error {
	f.nextID++
	t.ID = f.nextID
	cp := *t
	f.tickets[t.ID] = &cp
	return nil
}

func (f *fakeTicketRepo) Get(ctx context.Context, id int64) (*model.TicketDetail, error) {
	t, ok := f.tickets[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return &model.TicketDetail{Ticket: *t}, nil
}

func (f *fakeTicketRepo) List(ctx context.Context) ([]model.TicketDetail, error) {
	var out []model.TicketDetail
	for _, t := range f.tickets {
		out = append(out, model.TicketDetail{Ticket: *t})
	}
	return out, nil
}

func (f *fakeTicketRepo) Update(ctx context.Context, t *model.Ticket) error {
	if _, ok := f.tickets[t.ID]; !ok {
		return apperr.ErrNotFound
	}
	cp := *t
	f.tickets[t.ID] = &cp
	return nil
}

func (f *fakeTicketRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := f.tickets[id]; !ok {
		return apperr.ErrNotFound
	}
	delete(f.tickets, id)
	return nil
}

func TestTicketService_Create_Validation(t *testing.T) {
	svc := service.NewTicketService(newFakeTicketRepo())

	cases := []struct {
		name        string
		projectID   int64
		title       string
		storyPoints int
	}{
		{"missing project", 0, "Title", 3},
		{"blank title", 1, "   ", 3},
		{"zero points", 1, "Title", 0},
		{"negative points", 1, "Title", -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), c.projectID, c.title, c.storyPoints, nil)
			if !errors.Is(err, apperr.ErrValidation) {
				t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
			}
		})
	}
}

func TestTicketService_Create_Valid(t *testing.T) {
	svc := service.NewTicketService(newFakeTicketRepo())

	ticket, err := svc.Create(context.Background(), 1, "  Build homepage  ", 5, nil)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if ticket.Title != "Build homepage" {
		t.Fatalf("Create() title = %q, want trimmed %q", ticket.Title, "Build homepage")
	}
}
