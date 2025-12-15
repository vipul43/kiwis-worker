package models

import "time"

type AccountSyncStatus string

const (
	StatusPending    AccountSyncStatus = "pending"
	StatusProcessing AccountSyncStatus = "processing"
	StatusCompleted  AccountSyncStatus = "completed"
	StatusFailed     AccountSyncStatus = "failed"
)

type AccountSyncJob struct {
	ID          string
	AccountID   string
	Status      AccountSyncStatus
	Attempts    int
	LastError   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ProcessedAt *time.Time
}
