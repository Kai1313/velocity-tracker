package handler

import (
	"context"
	"net/http"

	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type userRequest struct {
	Name string     `json:"name"`
	Role model.Role `json:"role"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	handleCreate(w, r, func(ctx context.Context, req userRequest) (*model.User, error) {
		return h.svc.Create(ctx, req.Name, req.Role)
	})
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	handleList(w, r, h.svc.List)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	handleGet(w, r, h.svc.Get)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	handleUpdate(w, r, func(ctx context.Context, id int64, req userRequest) (*model.User, error) {
		return h.svc.Update(ctx, id, req.Name, req.Role)
	})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, h.svc.Delete)
}
