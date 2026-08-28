package handler

import (
	"net/http"
	"strconv"

	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/repository"
	"velocity-tracker/backend/internal/service"
)

type SprintEntryHandler struct {
	svc *service.SprintEntryService
}

func NewSprintEntryHandler(svc *service.SprintEntryService) *SprintEntryHandler {
	return &SprintEntryHandler{svc: svc}
}

type createSprintEntryRequest struct {
	TicketID              int64             `json:"ticketId"`
	SprintID              int64             `json:"sprintId"`
	Status                model.EntryStatus `json:"status"`
	AddedAfterSprintStart bool              `json:"addedAfterSprintStart"`
	CarriedFrom           *int64            `json:"carriedFrom"`
	PointsAtEntry         int               `json:"pointsAtEntry"`
}

type updateSprintEntryRequest struct {
	Status                model.EntryStatus `json:"status"`
	AddedAfterSprintStart bool              `json:"addedAfterSprintStart"`
	CarriedFrom           *int64            `json:"carriedFrom"`
	PointsAtEntry         int               `json:"pointsAtEntry"`
}

func (h *SprintEntryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSprintEntryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	e, err := h.svc.Create(r.Context(), req.TicketID, req.SprintID, req.Status, req.AddedAfterSprintStart, req.CarriedFrom, req.PointsAtEntry)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// parseSprintEntryFilter reads sprintId/projectId/status/carriedOver/search
// query params. Malformed sprintId/projectId are treated as absent rather
// than a 400 — an unfiltered result is a safer fallback than rejecting the
// request over a stray query param.
func parseSprintEntryFilter(r *http.Request) repository.SprintEntryFilter {
	q := r.URL.Query()
	var f repository.SprintEntryFilter
	if v := q.Get("sprintId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.SprintID = &id
		}
	}
	if v := q.Get("projectId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.ProjectID = &id
		}
	}
	if v := q.Get("status"); v != "" {
		status := model.EntryStatus(v)
		f.Status = &status
	}
	if q.Get("carriedOver") == "true" {
		f.CarriedOverOnly = true
	}
	if v := q.Get("search"); v != "" {
		f.Search = &v
	}
	return f
}

func (h *SprintEntryHandler) List(w http.ResponseWriter, r *http.Request) {
	entries, err := h.svc.List(r.Context(), parseSprintEntryFilter(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *SprintEntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	e, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *SprintEntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req updateSprintEntryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	e, err := h.svc.Update(r.Context(), id, req.Status, req.AddedAfterSprintStart, req.CarriedFrom, req.PointsAtEntry)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *SprintEntryHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
