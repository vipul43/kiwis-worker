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
	ID          string            `gorm:"column:id;primaryKey"`
	AccountID   string            `gorm:"column:account_id;uniqueIndex"`
	Status      AccountSyncStatus `gorm:"column:status"`
	Attempts    int               `gorm:"column:attempts"`
	LastError   *string           `gorm:"column:last_error"`
	CreatedAt   time.Time         `gorm:"column:created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at"`
	ProcessedAt *time.Time        `gorm:"column:processed_at"`
}

// TableName specifies the table name for GORM
func (AccountSyncJob) TableName() string {
	return "account_sync_job"
}
