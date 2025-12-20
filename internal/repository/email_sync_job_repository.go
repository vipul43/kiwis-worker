package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vipul43/kiwis-worker/internal/models"
)

type EmailSyncJobRepository struct {
	db *sql.DB
}

func NewEmailSyncJobRepository(db *sql.DB) *EmailSyncJobRepository {
	return &EmailSyncJobRepository{db: db}
}

// GetPendingJobs retrieves pending email sync jobs in round-robin order
// New jobs (last_synced_at = NULL) get picked first, then oldest synced jobs
func (r *EmailSyncJobRepository) GetPendingJobs(ctx context.Context, limit int) ([]models.EmailSyncJob, error) {
	query := `
		SELECT id, account_id, status, sync_type, emails_fetched, 
		       page_token, last_synced_at, attempts, last_error, 
		       created_at, updated_at, processed_at
		FROM email_sync_job
		WHERE status = $1
		ORDER BY last_synced_at ASC NULLS FIRST, created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.EmailStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// GetFailedJobs retrieves failed email sync jobs for retry in round-robin order
func (r *EmailSyncJobRepository) GetFailedJobs(ctx context.Context, limit int) ([]models.EmailSyncJob, error) {
	query := `
		SELECT id, account_id, status, sync_type, emails_fetched, 
		       page_token, last_synced_at, attempts, last_error, 
		       created_at, updated_at, processed_at
		FROM email_sync_job
		WHERE status = $1
		ORDER BY last_synced_at ASC NULLS FIRST, created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.EmailStatusFailed, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// GetProcessingJobs retrieves email sync jobs stuck in processing state
func (r *EmailSyncJobRepository) GetProcessingJobs(ctx context.Context, limit int) ([]models.EmailSyncJob, error) {
	query := `
		SELECT id, account_id, status, sync_type, emails_fetched, 
		       page_token, last_synced_at, attempts, last_error, 
		       created_at, updated_at, processed_at
		FROM email_sync_job
		WHERE status = $1
		ORDER BY last_synced_at ASC NULLS FIRST, created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.EmailStatusProcessing, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query processing jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// Create creates a new email sync job
func (r *EmailSyncJobRepository) Create(ctx context.Context, job models.EmailSyncJob) error {
	query := `
		INSERT INTO email_sync_job (
			id, account_id, status, sync_type, 
			emails_fetched, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.AccountID,
		job.Status,
		job.SyncType,
		job.EmailsFetched,
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create email sync job: %w", err)
	}

	return nil
}

// UpdateProgress updates job progress (emails fetched, page token, last synced time)
// Updates last_synced_at to push job to back of round-robin queue
func (r *EmailSyncJobRepository) UpdateProgress(ctx context.Context, jobID string, emailsFetched int, pageToken *string) error {
	query := `
		UPDATE email_sync_job
		SET emails_fetched = $1, 
		    page_token = $2, 
		    last_synced_at = $3, 
		    updated_at = $4
		WHERE id = $5
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, emailsFetched, pageToken, now, now, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job progress: %w", err)
	}

	return nil
}

// UpdateStatus updates the job status
// For synced/completed/failed status, sets processed_at
// For processing status, clears processed_at
func (r *EmailSyncJobRepository) UpdateStatus(ctx context.Context, jobID string, status models.EmailSyncStatus, lastError *string) error {
	query := `
		UPDATE email_sync_job
		SET status = $1, last_error = $2, updated_at = $3, processed_at = $4
		WHERE id = $5
	`

	now := time.Now()
	var processedAt *time.Time

	// Set processed_at for terminal states (synced, completed, failed)
	// Clear it for processing state (job is being worked on)
	if status == models.EmailStatusSynced || status == models.EmailStatusCompleted || status == models.EmailStatusFailed {
		processedAt = &now
	}

	_, err := r.db.ExecContext(ctx, query, status, lastError, now, processedAt, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// IncrementAttempts increments the retry attempt counter
func (r *EmailSyncJobRepository) IncrementAttempts(ctx context.Context, jobID string) error {
	query := `
		UPDATE email_sync_job
		SET attempts = attempts + 1, updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	return nil
}

// scanJobs scans database rows into EmailSyncJob slice
func (r *EmailSyncJobRepository) scanJobs(rows *sql.Rows) ([]models.EmailSyncJob, error) {
	var jobs []models.EmailSyncJob

	for rows.Next() {
		var job models.EmailSyncJob
		err := rows.Scan(
			&job.ID,
			&job.AccountID,
			&job.Status,
			&job.SyncType,
			&job.EmailsFetched,
			&job.PageToken,
			&job.LastSyncedAt,
			&job.Attempts,
			&job.LastError,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return jobs, nil
}

// GetByID retrieves an email sync job by ID
func (r *EmailSyncJobRepository) GetByID(ctx context.Context, jobID string) (*models.EmailSyncJob, error) {
	query := `
		SELECT id, account_id, status, sync_type, emails_fetched, 
		       page_token, last_synced_at, attempts, last_error, 
		       created_at, updated_at, processed_at
		FROM email_sync_job
		WHERE id = $1
	`

	var job models.EmailSyncJob
	err := r.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID,
		&job.AccountID,
		&job.Status,
		&job.SyncType,
		&job.EmailsFetched,
		&job.PageToken,
		&job.LastSyncedAt,
		&job.Attempts,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.ProcessedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}
