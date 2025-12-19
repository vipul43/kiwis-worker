package watcher

import (
	"context"
	"log"

	"github.com/vipul43/kiwis-worker/internal/models"
)

// processAccountJob processes a single account sync job
func (w *Watcher) processAccountJob(ctx context.Context, job models.AccountSyncJob) error {
	log.Printf("Processing account job %s for account %s", job.ID, job.AccountID)

	// Mark as processing
	if err := w.accountJobRepo.UpdateStatus(ctx, job.ID, models.StatusProcessing, nil); err != nil {
		return err
	}

	// Increment attempt counter
	if err := w.accountJobRepo.IncrementAttempts(ctx, job.ID); err != nil {
		log.Printf("Warning: failed to increment attempts: %v", err)
	}

	// Process the account
	err := w.accountProcessor.ProcessAccount(ctx, job.AccountID)
	if err != nil {
		return w.handleAccountJobError(ctx, job, err)
	}

	// Mark as completed
	if err := w.accountJobRepo.UpdateStatus(ctx, job.ID, models.StatusCompleted, nil); err != nil {
		return err
	}

	log.Printf("Successfully completed account job %s", job.ID)

	// Create initial email sync job for this account
	if err := w.emailProcessor.CreateInitialEmailSyncJob(ctx, job.AccountID); err != nil {
		log.Printf("Warning: failed to create email sync job for account %s: %v", job.AccountID, err)
		// Don't fail the account job if email job creation fails
	}

	return nil
}

// handleAccountJobError handles account job processing errors
// Sets status to failed for infinite retry
func (w *Watcher) handleAccountJobError(ctx context.Context, job models.AccountSyncJob, err error) error {
	errMsg := err.Error()
	newAttempts := job.Attempts + 1

	log.Printf("Account job %s failed (attempt %d): %v", job.ID, newAttempts, err)
	return w.accountJobRepo.UpdateStatus(ctx, job.ID, models.StatusFailed, &errMsg)
}
