package watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vipul43/kiwis-worker/internal/config"
	"github.com/vipul43/kiwis-worker/internal/models"
	"github.com/vipul43/kiwis-worker/internal/repository"
	"github.com/vipul43/kiwis-worker/internal/service"
)

type Watcher struct {
	cfg              *config.Config
	accountJobRepo   *repository.AccountSyncJobRepository
	emailJobRepo     *repository.EmailSyncJobRepository
	llmJobRepo       *repository.LLMSyncJobRepository
	accountProcessor *service.AccountProcessor
	emailProcessor   *service.EmailProcessor
	llmProcessor     *service.LLMProcessor
}

func New(
	cfg *config.Config,
	accountJobRepo *repository.AccountSyncJobRepository,
	emailJobRepo *repository.EmailSyncJobRepository,
	llmJobRepo *repository.LLMSyncJobRepository,
	accountProcessor *service.AccountProcessor,
	emailProcessor *service.EmailProcessor,
	llmProcessor *service.LLMProcessor,
) *Watcher {
	return &Watcher{
		cfg:              cfg,
		accountJobRepo:   accountJobRepo,
		emailJobRepo:     emailJobRepo,
		llmJobRepo:       llmJobRepo,
		accountProcessor: accountProcessor,
		emailProcessor:   emailProcessor,
		llmProcessor:     llmProcessor,
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

	// Process LLM sync jobs (batch processing)
	if err := w.processLLMSyncJobs(ctx); err != nil {
		log.Printf("Error processing LLM sync jobs: %v", err)
	}

	return nil
}

// processAccountSyncJobs processes pending, failed, and processing account sync jobs
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

	// Get processing jobs (stuck jobs from crashes or errors)
	processingJobs, err := w.accountJobRepo.GetProcessingJobs(ctx, 5)
	if err != nil {
		return err
	}

	// Combine all lists
	jobs := append(pendingJobs, failedJobs...)
	jobs = append(jobs, processingJobs...)

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

// processEmailSyncJobs processes pending, failed, and processing email sync jobs (round-robin)
func (w *Watcher) processEmailSyncJobs(ctx context.Context) error {
	// Fetch pending jobs
	pendingJobs, err := w.emailJobRepo.GetPendingJobs(ctx, 1)
	if err != nil {
		return err
	}

	// Fetch failed jobs
	failedJobs, err := w.emailJobRepo.GetFailedJobs(ctx, 1)
	if err != nil {
		return err
	}

	// Fetch processing jobs (stuck jobs)
	processingJobs, err := w.emailJobRepo.GetProcessingJobs(ctx, 1)
	if err != nil {
		return err
	}

	// Combine all jobs (already sorted by last_synced_at in individual queries)
	allJobs := append(pendingJobs, failedJobs...)
	allJobs = append(allJobs, processingJobs...)

	if len(allJobs) == 0 {
		return nil
	}

	// Pick the first job (queries already sort by last_synced_at ASC NULLS FIRST)
	job := allJobs[0]

	statusMsg := ""
	if job.Status == models.EmailStatusProcessing {
		statusMsg = " (stuck in processing)"
	} else if job.Status == models.EmailStatusFailed {
		statusMsg = fmt.Sprintf(" (failed, attempt %d)", job.Attempts)
	}

	log.Printf("Found email sync job: %s (account: %s, status: %s%s)", job.ID, job.AccountID, job.Status, statusMsg)

	if err := w.processEmailJob(ctx, job); err != nil {
		log.Printf("Failed to process email job %s: %v", job.ID, err)
	}

	return nil
}
