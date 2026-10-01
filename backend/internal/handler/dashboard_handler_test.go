package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"velocity-tracker/backend/internal/apperr"
)

func TestParseRetrospectiveLimit_DefaultsWhenAbsent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/dashboard/retrospective", nil)
	got, err := parseRetrospectiveLimit(r)
	if err != nil {
		t.Fatalf("parseRetrospectiveLimit() unexpected error: %v", err)
	}
	if got != defaultRetrospectiveLimit {
		t.Fatalf("parseRetrospectiveLimit() = %d, want default %d", got, defaultRetrospectiveLimit)
	}
}

func TestParseRetrospectiveLimit_AcceptsPositiveValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/dashboard/retrospective?limit=5", nil)
	got, err := parseRetrospectiveLimit(r)
	if err != nil {
		t.Fatalf("parseRetrospectiveLimit() unexpected error: %v", err)
	}
	if got != 5 {
		t.Fatalf("parseRetrospectiveLimit() = %d, want 5", got)
	}
}

func TestParseRetrospectiveLimit_RejectsNonPositiveOrMalformed(t *testing.T) {
	for _, v := range []string{"0", "-1", "abc"} {
		t.Run(v, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/dashboard/retrospective?limit="+v, nil)
			_, err := parseRetrospectiveLimit(r)
			if !errors.Is(err, apperr.ErrValidation) {
				t.Fatalf("parseRetrospectiveLimit(%q) error = %v, want apperr.ErrValidation", v, err)
			}
		})
	}
}
