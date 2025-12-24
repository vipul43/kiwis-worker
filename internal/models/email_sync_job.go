package models

import "time"

type EmailSyncStatus string

const (
	EmailStatusPending    EmailSyncStatus = "pending"    // Ready to fetch next batch
	EmailStatusProcessing EmailSyncStatus = "processing" // Currently fetching
	EmailStatusSynced     EmailSyncStatus = "synced"     // All historical emails fetched, waiting for webhook
	EmailStatusCompleted  EmailSyncStatus = "completed"  // Webhook setup complete, job finished
	EmailStatusFailed     EmailSyncStatus = "failed"     // Failed after max retries
)

type EmailSyncType string

const (
	SyncTypeInitial     EmailSyncType = "initial"     // Initial historical sync
	SyncTypeIncremental EmailSyncType = "incremental" // Incremental sync (manual re-sync)
	SyncTypeWebhook     EmailSyncType = "webhook"     // Real-time sync (webhook-triggered)
)

type EmailSyncJob struct {
	ID            string          `gorm:"column:id;primaryKey"`
	AccountID     string          `gorm:"column:account_id;index"`
	Status        EmailSyncStatus `gorm:"column:status;index"`
	SyncType      EmailSyncType   `gorm:"column:sync_type"`
	EmailsFetched int             `gorm:"column:emails_fetched"`
	PageToken     *string         `gorm:"column:page_token"`
	LastSyncedAt  *time.Time      `gorm:"column:last_synced_at"`
	Attempts      int             `gorm:"column:attempts"`
	LastError     *string         `gorm:"column:last_error"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
	ProcessedAt   *time.Time      `gorm:"column:processed_at"`
}

// TableName specifies the table name for GORM
func (EmailSyncJob) TableName() string {
	return "email_sync_job"
}
