package watcher

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/payment-tracker/internal/config"
	"github.com/yourusername/payment-tracker/internal/models"
	"github.com/yourusername/payment-tracker/internal/repository"
	"github.com/yourusername/payment-tracker/internal/service"
)

type Watcher struct {
	cfg              *config.Config
	jobRepo          *repository.AccountSyncJobRepository
	accountProcessor *service.AccountProcessor
}

func New(
	cfg *config.Config,
	jobRepo *repository.AccountSyncJobRepository,
	accountProcessor *service.AccountProcessor,
) *Watcher {
	return &Watcher{
		cfg:              cfg,
		jobRepo:          jobRepo,
		accountProcessor: accountProcessor,
	}
}

// Start begins watching for pending account sync jobs
func (w *Watcher) Start(ctx context.Context) error {
	log.Println("Starting account sync watcher...")

	// Process any pending jobs from previous runs
	if err := w.processPendingJobs(ctx); err != nil {
		log.Printf("Warning: failed to process pending jobs on startup: %v", err)
	}

	// Start polling loop
	ticker := time.NewTicker(time.Duration(w.cfg.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Watcher shutting down...")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processPendingJobs(ctx); err != nil {
				log.Printf("Error processing jobs: %v", err)
			}
		}
	}
}

// processPendingJobs fetches and processes all pending jobs
func (w *Watcher) processPendingJobs(ctx context.Context) error {
	jobs, err := w.jobRepo.GetPendingJobs(ctx, 10)
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return nil
	}

	log.Printf("Found %d pending job(s)", len(jobs))

	for _, job := range jobs {
		if err := w.processJob(ctx, job); err != nil {
			log.Printf("Failed to process job %s: %v", job.ID, err)
		}
	}

	return nil
}

// processJob processes a single account sync job
func (w *Watcher) processJob(ctx context.Context, job models.AccountSyncJob) error {
	log.Printf("Processing job %s for account %s", job.ID, job.AccountID)

	// Mark as processing
	if err := w.jobRepo.UpdateStatus(ctx, job.ID, models.StatusProcessing, nil); err != nil {
		return err
	}

	// Increment attempt counter
	if err := w.jobRepo.IncrementAttempts(ctx, job.ID); err != nil {
		log.Printf("Warning: failed to increment attempts: %v", err)
	}

	// Process the account
	err := w.accountProcessor.ProcessAccount(ctx, job.AccountID)
	if err != nil {
		return w.handleJobError(ctx, job, err)
	}

	// Mark as completed
	if err := w.jobRepo.UpdateStatus(ctx, job.ID, models.StatusCompleted, nil); err != nil {
		return err
	}

	log.Printf("Successfully completed job %s", job.ID)
	return nil
}

// handleJobError handles job processing errors with retry logic
func (w *Watcher) handleJobError(ctx context.Context, job models.AccountSyncJob, err error) error {
	errMsg := err.Error()
	newAttempts := job.Attempts + 1

	// Check if max retries reached
	if newAttempts >= w.cfg.MaxRetries {
		log.Printf("Job %s failed after %d attempts: %v", job.ID, newAttempts, err)
		return w.jobRepo.UpdateStatus(ctx, job.ID, models.StatusFailed, &errMsg)
	}

	// Reset to pending for retry
	log.Printf("Job %s failed (attempt %d/%d), will retry: %v", job.ID, newAttempts, w.cfg.MaxRetries, err)
	return w.jobRepo.UpdateStatus(ctx, job.ID, models.StatusPending, &errMsg)
}
