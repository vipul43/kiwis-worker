package models

import "time"

// LLM sync job status constants
const (
	LLMStatusPending    = "pending"
	LLMStatusProcessing = "processing"
	LLMStatusCompleted  = "completed"
	LLMStatusFailed     = "failed"
)

// LLMSyncJob represents a job for extracting payment information from an email using LLM
type LLMSyncJob struct {
	ID           string
	AccountID    string
	MessageID    string // Gmail message ID
	Status       string
	LastSyncedAt *time.Time
	Attempts     int
	LastError    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ProcessedAt  *time.Time
}
