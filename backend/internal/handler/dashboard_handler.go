package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/service"
)

const defaultRetrospectiveLimit = 3

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) SprintSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.svc.SprintSummaries(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (h *DashboardHandler) SprintDeveloperBreakdown(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	breakdown, err := h.svc.SprintDeveloperBreakdown(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

func (h *DashboardHandler) SprintTicketBreakdown(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	breakdown, err := h.svc.SprintTicketBreakdown(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, breakdown)
}

func (h *DashboardHandler) SprintHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.svc.SprintHealth(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, health)
}

// parseRetrospectiveLimit reads the `limit` query param, defaulting to
// defaultRetrospectiveLimit when absent. Unlike parseSprintEntryFilter's
// malformed-sprintId handling, a bad limit here is rejected rather than
// silently defaulted — the caller asked for a specific window size, so
// silently ignoring a typo'd value would be more surprising than erroring.
func parseRetrospectiveLimit(r *http.Request) (int, error) {
	v := r.URL.Query().Get("limit")
	if v == "" {
		return defaultRetrospectiveLimit, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%w: limit must be a positive integer", apperr.ErrValidation)
	}
	return n, nil
}

// SprintRetrospective serves the last `limit` closed sprints' commitment
// metrics. limit defaults to 3 and must be a positive integer.
func (h *DashboardHandler) SprintRetrospective(w http.ResponseWriter, r *http.Request) {
	limit, err := parseRetrospectiveLimit(r)
	if err != nil {
		writeError(w, err)
		return
	}
	retro, err := h.svc.SprintRetrospective(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, retro)
}
