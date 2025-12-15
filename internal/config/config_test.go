package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set required env var
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test" {
		t.Errorf("expected DATABASE_URL to be set, got %s", cfg.DatabaseURL)
	}

	// Check defaults
	if cfg.PollInterval != 10 {
		t.Errorf("expected PollInterval to be 10, got %d", cfg.PollInterval)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries to be 3, got %d", cfg.MaxRetries)
	}
	if cfg.ShutdownTimeout != 30 {
		t.Errorf("expected ShutdownTimeout to be 30, got %d", cfg.ShutdownTimeout)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	// Ensure DATABASE_URL is not set
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing, got nil")
	}

	expectedMsg := "DATABASE_URL is required"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}
