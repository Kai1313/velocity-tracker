package handler

import (
	"context"
	"net/http"
	"time"

	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type SprintHandler struct {
	svc *service.SprintService
}

func NewSprintHandler(svc *service.SprintService) *SprintHandler {
	return &SprintHandler{svc: svc}
}

type createSprintRequest struct {
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type updateSprintRequest struct {
	Name      string             `json:"name"`
	StartDate time.Time          `json:"startDate"`
	EndDate   time.Time          `json:"endDate"`
	Status    model.SprintStatus `json:"status"`
}

func (h *SprintHandler) Create(w http.ResponseWriter, r *http.Request) {
	handleCreate(w, r, func(ctx context.Context, req createSprintRequest) (*model.Sprint, error) {
		return h.svc.Create(ctx, req.Name, req.StartDate, req.EndDate)
	})
}

func (h *SprintHandler) List(w http.ResponseWriter, r *http.Request) {
	handleList(w, r, h.svc.List)
}

func (h *SprintHandler) Get(w http.ResponseWriter, r *http.Request) {
	handleGet(w, r, h.svc.Get)
}

func (h *SprintHandler) Update(w http.ResponseWriter, r *http.Request) {
	handleUpdate(w, r, func(ctx context.Context, id int64, req updateSprintRequest) (*model.Sprint, error) {
		return h.svc.Update(ctx, id, req.Name, req.StartDate, req.EndDate, req.Status)
	})
}

func (h *SprintHandler) Delete(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, h.svc.Delete)
}
