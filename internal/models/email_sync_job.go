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
	ID            string
	AccountID     string
	Status        EmailSyncStatus
	SyncType      EmailSyncType
	EmailsFetched int
	PageToken     *string
	LastSyncedAt  *time.Time // NULL = never synced (new jobs get priority)
	Attempts      int
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ProcessedAt   *time.Time
}
