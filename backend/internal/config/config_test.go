package config_test

import (
	"testing"

	"velocity-tracker/backend/internal/config"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error when DATABASE_URL is unset")
	}
}

func TestLoad_UsesProvidedValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("Load() DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://example")
	}
	if cfg.Port != "9090" {
		t.Fatalf("Load() Port = %q, want %q", cfg.Port, "9090")
	}
}

func TestLoad_DefaultsPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("PORT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("Load() Port = %q, want default %q", cfg.Port, "8080")
	}
}
