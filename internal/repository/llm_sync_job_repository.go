package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/vipul43/kiwis-worker/internal/models"
)

type LLMSyncJobRepository struct {
	db *sql.DB
}

func NewLLMSyncJobRepository(db *sql.DB) *LLMSyncJobRepository {
	return &LLMSyncJobRepository{db: db}
}

// Create creates a new LLM sync job
func (r *LLMSyncJobRepository) Create(ctx context.Context, job models.LLMSyncJob) error {
	query := `
		INSERT INTO llm_sync_job (
			id, account_id, message_id, status, last_synced_at, attempts,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		job.ID, job.AccountID, job.MessageID, job.Status, job.LastSyncedAt,
		job.Attempts, job.CreatedAt, job.UpdatedAt,
	)
	return err
}

// BulkCreate creates multiple LLM sync jobs in a single transaction
func (r *LLMSyncJobRepository) BulkCreate(ctx context.Context, jobs []models.LLMSyncJob) error {
	if len(jobs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		INSERT INTO llm_sync_job (
			id, account_id, message_id, status, last_synced_at, attempts,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (message_id) DO NOTHING
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, job := range jobs {
		_, err := stmt.ExecContext(ctx,
			job.ID, job.AccountID, job.MessageID, job.Status, job.LastSyncedAt,
			job.Attempts, job.CreatedAt, job.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetPendingJobs retrieves pending LLM sync jobs (round-robin by last_synced_at)
func (r *LLMSyncJobRepository) GetPendingJobs(ctx context.Context, limit int) ([]models.LLMSyncJob, error) {
	query := `
		SELECT id, account_id, message_id, status, last_synced_at, attempts,
		       last_error, created_at, updated_at, processed_at
		FROM llm_sync_job
		WHERE status = $1
		ORDER BY last_synced_at ASC NULLS FIRST, created_at ASC
		LIMIT $2
	`
	return r.queryJobs(ctx, query, models.LLMStatusPending, limit)
}

// GetFailedJobs retrieves failed LLM sync jobs for retry
func (r *LLMSyncJobRepository) GetFailedJobs(ctx context.Context, limit int) ([]models.LLMSyncJob, error) {
	query := `
		SELECT id, account_id, message_id, status, last_synced_at, attempts,
		       last_error, created_at, updated_at, processed_at
		FROM llm_sync_job
		WHERE status = $1
		ORDER BY last_synced_at ASC NULLS FIRST, created_at ASC
		LIMIT $2
	`
	return r.queryJobs(ctx, query, models.LLMStatusFailed, limit)
}

// GetProcessingJobs retrieves stuck processing jobs (crash recovery)
func (r *LLMSyncJobRepository) GetProcessingJobs(ctx context.Context, limit int) ([]models.LLMSyncJob, error) {
	query := `
		SELECT id, account_id, message_id, status, last_synced_at, attempts,
		       last_error, created_at, updated_at, processed_at
		FROM llm_sync_job
		WHERE status = $1
		ORDER BY last_synced_at ASC NULLS FIRST, created_at ASC
		LIMIT $2
	`
	return r.queryJobs(ctx, query, models.LLMStatusProcessing, limit)
}

// queryJobs is a helper function to query jobs
func (r *LLMSyncJobRepository) queryJobs(ctx context.Context, query string, args ...interface{}) ([]models.LLMSyncJob, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.LLMSyncJob
	for rows.Next() {
		var job models.LLMSyncJob
		err := rows.Scan(
			&job.ID, &job.AccountID, &job.MessageID, &job.Status, &job.LastSyncedAt,
			&job.Attempts, &job.LastError, &job.CreatedAt, &job.UpdatedAt, &job.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// UpdateStatus updates the status of an LLM sync job
func (r *LLMSyncJobRepository) UpdateStatus(ctx context.Context, id string, status string, lastError *string) error {
	now := time.Now()
	query := `
		UPDATE llm_sync_job
		SET status = $1, last_error = $2, updated_at = $3, last_synced_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, status, lastError, now, now, id)
	return err
}

// IncrementAttempts increments the attempts counter
func (r *LLMSyncJobRepository) IncrementAttempts(ctx context.Context, id string) error {
	query := `
		UPDATE llm_sync_job
		SET attempts = attempts + 1, updated_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
