package watcher

import (
	"context"
	"log"

	"github.com/vipul43/kiwis-worker/internal/models"
)

// processLLMSyncJobs processes pending, failed, and processing LLM sync jobs (round-robin batch)
func (w *Watcher) processLLMSyncJobs(ctx context.Context) error {
	// Fetch pending jobs (batch of 3)
	pendingJobs, err := w.llmJobRepo.GetPendingJobs(ctx, 3)
	if err != nil {
		return err
	}

	// Fetch failed jobs (batch of 3)
	failedJobs, err := w.llmJobRepo.GetFailedJobs(ctx, 3)
	if err != nil {
		return err
	}

	// Fetch processing jobs (stuck jobs, batch of 3)
	processingJobs, err := w.llmJobRepo.GetProcessingJobs(ctx, 3)
	if err != nil {
		return err
	}

	// Combine all jobs (already sorted by last_synced_at in individual queries)
	allJobs := append(pendingJobs, failedJobs...)
	allJobs = append(allJobs, processingJobs...)

	if len(allJobs) == 0 {
		return nil
	}

	log.Printf("Found %d LLM sync jobs to process (pending: %d, failed: %d, processing: %d)",
		len(allJobs), len(pendingJobs), len(failedJobs), len(processingJobs))

	// Mark all jobs as processing
	for _, job := range allJobs {
		if err := w.llmJobRepo.UpdateStatus(ctx, job.ID, models.LLMStatusProcessing, nil); err != nil {
			log.Printf("Warning: failed to update job %s to processing: %v", job.ID, err)
		}
		if err := w.llmJobRepo.IncrementAttempts(ctx, job.ID); err != nil {
			log.Printf("Warning: failed to increment attempts for job %s: %v", job.ID, err)
		}
	}

	// Process batch
	err = w.llmProcessor.ProcessLLMSyncJobs(ctx, allJobs)
	if err != nil {
		log.Printf("Error processing LLM sync jobs batch: %v", err)
		return err
	}

	log.Printf("Completed processing %d LLM sync jobs", len(allJobs))
	return nil
}
