package handler

import (
	"context"
	"net/http"

	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type TicketHandler struct {
	svc *service.TicketService
}

func NewTicketHandler(svc *service.TicketService) *TicketHandler {
	return &TicketHandler{svc: svc}
}

type ticketRequest struct {
	ProjectID   int64  `json:"projectId"`
	Title       string `json:"title"`
	StoryPoints int    `json:"storyPoints"`
	AssigneeID  *int64 `json:"assigneeId"`
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	handleCreate(w, r, func(ctx context.Context, req ticketRequest) (*model.Ticket, error) {
		return h.svc.Create(ctx, req.ProjectID, req.Title, req.StoryPoints, req.AssigneeID)
	})
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	handleList(w, r, h.svc.List)
}

func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	handleGet(w, r, h.svc.Get)
}

func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	handleUpdate(w, r, func(ctx context.Context, id int64, req ticketRequest) (*model.Ticket, error) {
		return h.svc.Update(ctx, id, req.ProjectID, req.Title, req.StoryPoints, req.AssigneeID)
	})
}

func (h *TicketHandler) Delete(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, h.svc.Delete)
}
