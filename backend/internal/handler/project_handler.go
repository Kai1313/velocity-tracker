package handler

import (
	"context"
	"net/http"

	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type createProjectRequest struct {
	Name string `json:"name"`
}

type updateProjectRequest struct {
	Name   string              `json:"name"`
	Status model.ProjectStatus `json:"status"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	handleCreate(w, r, func(ctx context.Context, req createProjectRequest) (*model.Project, error) {
		return h.svc.Create(ctx, req.Name)
	})
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	handleList(w, r, h.svc.List)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	handleGet(w, r, h.svc.Get)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	handleUpdate(w, r, func(ctx context.Context, id int64, req updateProjectRequest) (*model.Project, error) {
		return h.svc.Update(ctx, id, req.Name, req.Status)
	})
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, h.svc.Delete)
}
