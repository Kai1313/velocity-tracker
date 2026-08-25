package handler

import (
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
	var req createSprintRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	sp, err := h.svc.Create(r.Context(), req.Name, req.StartDate, req.EndDate)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sp)
}

func (h *SprintHandler) List(w http.ResponseWriter, r *http.Request) {
	sprints, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sprints)
}

func (h *SprintHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	sp, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}

func (h *SprintHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req updateSprintRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	sp, err := h.svc.Update(r.Context(), id, req.Name, req.StartDate, req.EndDate, req.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}

func (h *SprintHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
