package watcher

import (
	"context"
	"log"

	"github.com/vipul43/kiwis-worker/internal/models"
)

// processEmailJob processes a single email sync job
func (w *Watcher) processEmailJob(ctx context.Context, job models.EmailSyncJob) error {
	log.Printf("Processing email job %s for account %s (type: %s, fetched: %d)",
		job.ID, job.AccountID, job.SyncType, job.EmailsFetched)

	// Mark as processing
	if err := w.emailJobRepo.UpdateStatus(ctx, job.ID, models.EmailStatusProcessing, nil); err != nil {
		return err
	}

	// Increment attempt counter
	if err := w.emailJobRepo.IncrementAttempts(ctx, job.ID); err != nil {
		log.Printf("Warning: failed to increment attempts: %v", err)
	}

	// Process the email sync job
	// ProcessEmailSyncJob updates the job object in-place with new values
	err := w.emailProcessor.ProcessEmailSyncJob(ctx, &job)
	if err != nil {
		return w.handleEmailJobError(ctx, job, err)
	}

	// Check if historical sync is complete
	if job.PageToken == nil && job.EmailsFetched >= 10000 {
		// Reached max emails and no more pages
		if err := w.emailJobRepo.UpdateStatus(ctx, job.ID, models.EmailStatusSynced, nil); err != nil {
			return err
		}
		log.Printf("Email sync job %s completed: reached max emails (%d)", job.ID, job.EmailsFetched)
		return nil
	}

	if job.PageToken == nil {
		// No more pages, but haven't reached max emails
		// This means we've fetched all available historical emails
		if err := w.emailJobRepo.UpdateStatus(ctx, job.ID, models.EmailStatusSynced, nil); err != nil {
			return err
		}
		log.Printf("Email sync job %s completed: no more emails to fetch (%d total)", job.ID, job.EmailsFetched)
		return nil
	}

	// Partial success: more pages to fetch
	// Stay in processing state (already set), last_synced_at updated by UpdateProgress
	// Job will be picked up again in next round (goes to back of queue due to last_synced_at)
	log.Printf("Email sync job %s has more pages, staying in processing (fetched: %d)", job.ID, job.EmailsFetched)
	return nil
}

// handleEmailJobError handles email job processing errors
// Sets status to failed and updates last_synced_at to push to back of queue
func (w *Watcher) handleEmailJobError(ctx context.Context, job models.EmailSyncJob, err error) error {
	errMsg := err.Error()
	newAttempts := job.Attempts + 1

	log.Printf("Email job %s failed (attempt %d): %v", job.ID, newAttempts, err)

	// Update last_synced_at to push failed job to back of queue
	// This prevents failed jobs from blocking the queue
	if err := w.emailJobRepo.UpdateProgress(ctx, job.ID, job.EmailsFetched, job.PageToken); err != nil {
		log.Printf("Warning: failed to update progress after error: %v", err)
	}

	// Set to failed status (will be picked up again in next round)
	return w.emailJobRepo.UpdateStatus(ctx, job.ID, models.EmailStatusFailed, &errMsg)
}
