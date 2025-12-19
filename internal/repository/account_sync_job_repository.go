package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vipul43/kiwis-worker/internal/models"
)

type AccountSyncJobRepository struct {
	db *sql.DB
}

func NewAccountSyncJobRepository(db *sql.DB) *AccountSyncJobRepository {
	return &AccountSyncJobRepository{db: db}
}

// GetPendingJobs retrieves all pending account sync jobs
func (r *AccountSyncJobRepository) GetPendingJobs(ctx context.Context, limit int) ([]models.AccountSyncJob, error) {
	query := `
		SELECT id, account_id, status, attempts, last_error, created_at, updated_at, processed_at
		FROM account_sync_job
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// GetFailedJobs retrieves all failed account sync jobs for retry
func (r *AccountSyncJobRepository) GetFailedJobs(ctx context.Context, limit int) ([]models.AccountSyncJob, error) {
	query := `
		SELECT id, account_id, status, attempts, last_error, created_at, updated_at, processed_at
		FROM account_sync_job
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.StatusFailed, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// GetProcessingJobs retrieves account sync jobs stuck in processing state
func (r *AccountSyncJobRepository) GetProcessingJobs(ctx context.Context, limit int) ([]models.AccountSyncJob, error) {
	query := `
		SELECT id, account_id, status, attempts, last_error, created_at, updated_at, processed_at
		FROM account_sync_job
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.StatusProcessing, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query processing jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// UpdateStatus updates the job status
func (r *AccountSyncJobRepository) UpdateStatus(ctx context.Context, jobID string, status models.AccountSyncStatus, lastError *string) error {
	query := `
		UPDATE account_sync_job
		SET status = $1, last_error = $2, updated_at = $3, processed_at = $4
		WHERE id = $5
	`

	var processedAt *time.Time
	if status == models.StatusCompleted || status == models.StatusFailed {
		now := time.Now()
		processedAt = &now
	}

	_, err := r.db.ExecContext(ctx, query, status, lastError, time.Now(), processedAt, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// IncrementAttempts increments the retry attempt counter
func (r *AccountSyncJobRepository) IncrementAttempts(ctx context.Context, jobID string) error {
	query := `
		UPDATE account_sync_job
		SET attempts = attempts + 1, updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	return nil
}

// scanJobs scans database rows into AccountSyncJob slice
func (r *AccountSyncJobRepository) scanJobs(rows *sql.Rows) ([]models.AccountSyncJob, error) {
	var jobs []models.AccountSyncJob

	for rows.Next() {
		var job models.AccountSyncJob
		err := rows.Scan(
			&job.ID,
			&job.AccountID,
			&job.Status,
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
