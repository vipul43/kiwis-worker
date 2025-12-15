package models

import (
	"testing"
	"time"
)

func TestAccountSyncStatus_Constants(t *testing.T) {
	tests := []struct {
		status   AccountSyncStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusProcessing, "processing"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestAccountSyncJob_Structure(t *testing.T) {
	now := time.Now()
	processedAt := now.Add(time.Minute)
	lastError := "test error"

	job := AccountSyncJob{
		ID:          "job-123",
		AccountID:   "acc-123",
		Status:      StatusPending,
		Attempts:    0,
		LastError:   &lastError,
		CreatedAt:   now,
		UpdatedAt:   now,
		ProcessedAt: &processedAt,
	}

	if job.ID != "job-123" {
		t.Errorf("expected ID to be job-123, got %s", job.ID)
	}
	if job.AccountID != "acc-123" {
		t.Errorf("expected AccountID to be acc-123, got %s", job.AccountID)
	}
	if job.Status != StatusPending {
		t.Errorf("expected Status to be pending, got %s", job.Status)
	}
	if job.Attempts != 0 {
		t.Errorf("expected Attempts to be 0, got %d", job.Attempts)
	}
	if job.LastError == nil || *job.LastError != "test error" {
		t.Errorf("expected LastError to be 'test error', got %v", job.LastError)
	}
	if job.ProcessedAt == nil {
		t.Error("expected ProcessedAt to be set")
	}
}
