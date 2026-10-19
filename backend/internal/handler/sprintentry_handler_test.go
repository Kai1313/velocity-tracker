package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
)

func TestParseSprintEntryFilter_RejectsInvalidStatus(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/sprint-entries?status=Bogus", nil)
	_, err := parseSprintEntryFilter(r)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("parseSprintEntryFilter() error = %v, want apperr.ErrValidation", err)
	}
}

func TestParseSprintEntryFilter_AcceptsValidStatus(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/sprint-entries?status=Done", nil)
	f, err := parseSprintEntryFilter(r)
	if err != nil {
		t.Fatalf("parseSprintEntryFilter() unexpected error: %v", err)
	}
	if f.Status == nil || *f.Status != model.EntryDone {
		t.Fatalf("parseSprintEntryFilter() status = %v, want Done", f.Status)
	}
}

func TestParseSprintEntryFilter_MalformedSprintIDTreatedAsAbsent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/sprint-entries?sprintId=notanumber", nil)
	f, err := parseSprintEntryFilter(r)
	if err != nil {
		t.Fatalf("parseSprintEntryFilter() unexpected error: %v", err)
	}
	if f.SprintID != nil {
		t.Fatalf("parseSprintEntryFilter() sprintId = %v, want nil (absent)", f.SprintID)
	}
}

// The bounceCount field is required: a body without it must be rejected
// (422) rather than treated as 0, which would silently reset the count.
func TestSprintEntryHandler_SetBounceCount_RequiresBounceCount(t *testing.T) {
	h := &SprintEntryHandler{} // validation fails before the service is touched
	r := httptest.NewRequest(http.MethodPatch, "/sprint-entries/1/bounce-count", strings.NewReader(`{}`))
	r.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.SetBounceCount(rec, r)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
}
