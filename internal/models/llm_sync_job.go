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
	ID           string     `gorm:"column:id;primaryKey"`
	AccountID    string     `gorm:"column:account_id;index"`
	MessageID    string     `gorm:"column:message_id;uniqueIndex"`
	Status       string     `gorm:"column:status;index"`
	LastSyncedAt *time.Time `gorm:"column:last_synced_at"`
	Attempts     int        `gorm:"column:attempts"`
	LastError    *string    `gorm:"column:last_error"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	ProcessedAt  *time.Time `gorm:"column:processed_at"`
}

// TableName specifies the table name for GORM
func (LLMSyncJob) TableName() string {
	return "llm_sync_job"
}
