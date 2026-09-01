package handler

import (
	"net/http"

	"velocity-tracker/backend/internal/service"
)

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
