package watcher

import (
	"context"
	"log"
	"time"

	"github.com/vipul43/kiwis-worker/internal/config"
	"github.com/vipul43/kiwis-worker/internal/repository"
	"github.com/vipul43/kiwis-worker/internal/service"
)

type Watcher struct {
	cfg              *config.Config
	accountJobRepo   *repository.AccountSyncJobRepository
	emailJobRepo     *repository.EmailSyncJobRepository
	accountProcessor *service.AccountProcessor
	emailProcessor   *service.EmailProcessor
}

func New(
	cfg *config.Config,
	accountJobRepo *repository.AccountSyncJobRepository,
	emailJobRepo *repository.EmailSyncJobRepository,
	accountProcessor *service.AccountProcessor,
	emailProcessor *service.EmailProcessor,
) *Watcher {
	return &Watcher{
		cfg:              cfg,
		accountJobRepo:   accountJobRepo,
		emailJobRepo:     emailJobRepo,
		accountProcessor: accountProcessor,
		emailProcessor:   emailProcessor,
	}
}

// Start begins watching for pending jobs (both account and email sync)
func (w *Watcher) Start(ctx context.Context) error {
	log.Println("Starting watcher for account and email sync jobs...")

	// Process any pending jobs from previous runs
	if err := w.processAllPendingJobs(ctx); err != nil {
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
			if err := w.processAllPendingJobs(ctx); err != nil {
				log.Printf("Error processing jobs: %v", err)
			}
		}
	}
}

// processAllPendingJobs processes both account sync and email sync jobs
func (w *Watcher) processAllPendingJobs(ctx context.Context) error {
	// Process account sync jobs first (new accounts)
	if err := w.processAccountSyncJobs(ctx); err != nil {
		log.Printf("Error processing account sync jobs: %v", err)
	}

	// Process email sync jobs (round-robin by priority)
	if err := w.processEmailSyncJobs(ctx); err != nil {
		log.Printf("Error processing email sync jobs: %v", err)
	}

	return nil
}

// processAccountSyncJobs processes pending and failed account sync jobs
func (w *Watcher) processAccountSyncJobs(ctx context.Context) error {
	// Get pending jobs
	pendingJobs, err := w.accountJobRepo.GetPendingJobs(ctx, 5)
	if err != nil {
		return err
	}

	// Get failed jobs for retry
	failedJobs, err := w.accountJobRepo.GetFailedJobs(ctx, 5)
	if err != nil {
		return err
	}

	// Combine both lists
	jobs := append(pendingJobs, failedJobs...)

	if len(jobs) == 0 {
		return nil
	}

	log.Printf("Found %d account sync job(s) to process", len(jobs))

	for _, job := range jobs {
		if err := w.processAccountJob(ctx, job); err != nil {
			log.Printf("Failed to process account job %s: %v", job.ID, err)
		}
	}

	return nil
}

// processEmailSyncJobs processes pending and failed email sync jobs (round-robin)
func (w *Watcher) processEmailSyncJobs(ctx context.Context) error {
	// Fetch 1 pending job for round-robin behavior
	pendingJobs, err := w.emailJobRepo.GetPendingJobs(ctx, 1)
	if err != nil {
		return err
	}

	// If no pending, try failed jobs
	if len(pendingJobs) == 0 {
		failedJobs, err := w.emailJobRepo.GetFailedJobs(ctx, 1)
		if err != nil {
			return err
		}

		if len(failedJobs) == 0 {
			return nil
		}

		job := failedJobs[0]
		log.Printf("Retrying failed email sync job: %s (account: %s, attempts: %d)", job.ID, job.AccountID, job.Attempts)

		if err := w.processEmailJob(ctx, job); err != nil {
			log.Printf("Failed to process email job %s: %v", job.ID, err)
		}

		return nil
	}

	job := pendingJobs[0]
	log.Printf("Found email sync job: %s (account: %s, last_synced: %v)", job.ID, job.AccountID, job.LastSyncedAt)

	if err := w.processEmailJob(ctx, job); err != nil {
		log.Printf("Failed to process email job %s: %v", job.ID, err)
	}

	return nil
}
