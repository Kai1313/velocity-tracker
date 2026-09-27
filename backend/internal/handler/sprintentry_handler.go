package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"velocity-tracker/backend/internal/apperr"
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
	handleCreate(w, r, func(ctx context.Context, req createSprintEntryRequest) (*model.SprintEntry, error) {
		return h.svc.Create(ctx, service.CreateSprintEntryInput{
			TicketID:              req.TicketID,
			SprintID:              req.SprintID,
			Status:                req.Status,
			AddedAfterSprintStart: req.AddedAfterSprintStart,
			CarriedFrom:           req.CarriedFrom,
			PointsAtEntry:         req.PointsAtEntry,
		})
	})
}

// parseSprintEntryFilter reads sprintId/projectId/status/carriedOver/search
// query params. Malformed sprintId/projectId are treated as absent rather
// than a 400 — an unfiltered result is a safer fallback than rejecting the
// request over a stray query param. status is different: it's an enum, so an
// invalid value is rejected rather than silently matching zero rows.
func parseSprintEntryFilter(r *http.Request) (repository.SprintEntryFilter, error) {
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
		switch status {
		case model.EntryDone, model.EntryNotDone, model.EntryCancelled:
			f.Status = &status
		default:
			return f, fmt.Errorf("%w: status must be Done, NotDone, or Cancelled", apperr.ErrValidation)
		}
	}
	if q.Get("carriedOver") == "true" {
		f.CarriedOverOnly = true
	}
	if v := q.Get("search"); v != "" {
		f.Search = &v
	}
	return f, nil
}

func (h *SprintEntryHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseSprintEntryFilter(r)
	if err != nil {
		writeError(w, err)
		return
	}
	entries, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *SprintEntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	handleGet(w, r, h.svc.Get)
}

func (h *SprintEntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	handleUpdate(w, r, func(ctx context.Context, id int64, req updateSprintEntryRequest) (*model.SprintEntry, error) {
		return h.svc.Update(ctx, id, service.UpdateSprintEntryInput{
			Status:                req.Status,
			AddedAfterSprintStart: req.AddedAfterSprintStart,
			CarriedFrom:           req.CarriedFrom,
			PointsAtEntry:         req.PointsAtEntry,
		})
	})
}

func (h *SprintEntryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, h.svc.Delete)
}
