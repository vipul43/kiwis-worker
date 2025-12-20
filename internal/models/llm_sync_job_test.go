package models

import (
	"testing"
	"time"
)

func TestLLMSyncJobStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"pending", LLMStatusPending, "pending"},
		{"processing", LLMStatusProcessing, "processing"},
		{"completed", LLMStatusCompleted, "completed"},
		{"failed", LLMStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

func TestLLMSyncJob_Structure(t *testing.T) {
	now := time.Now()
	job := LLMSyncJob{
		ID:           "test-id",
		AccountID:    "account-123",
		MessageID:    "msg-456",
		Status:       LLMStatusPending,
		LastSyncedAt: &now,
		Attempts:     0,
		LastError:    nil,
		CreatedAt:    now,
		UpdatedAt:    now,
		ProcessedAt:  nil,
	}

	if job.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", job.ID)
	}
	if job.Status != LLMStatusPending {
		t.Errorf("Expected status 'pending', got %s", job.Status)
	}
	if job.MessageID != "msg-456" {
		t.Errorf("Expected MessageID 'msg-456', got %s", job.MessageID)
	}
}
