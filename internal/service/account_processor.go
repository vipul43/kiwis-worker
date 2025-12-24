package service

import (
	"context"
	"fmt"
	"log"

	"github.com/vipul43/kiwis-worker/internal/models"
)

// AccountRepository interface for dependency injection
type AccountRepository interface {
	GetByID(ctx context.Context, accountID string) (*models.Account, error)
}

type AccountProcessor struct {
	accountRepo AccountRepository
}

func NewAccountProcessor(accountRepo AccountRepository) *AccountProcessor {
	return &AccountProcessor{
		accountRepo: accountRepo,
	}
}

// EmailSyncJobCreator interface for creating email sync jobs
type EmailSyncJobCreator interface {
	CreateInitialEmailSyncJob(ctx context.Context, accountID string) error
}

// ProcessAccount processes the given account and creates initial email sync job
func (p *AccountProcessor) ProcessAccount(ctx context.Context, accountID string) error {
	// Fetch account details
	account, err := p.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Validate tokens exist
	if account.AccessToken == nil {
		return fmt.Errorf("account missing access token")
	}

	log.Printf("Processing account: %s (user: %s)", accountID, account.UserID)

	// Account setup is complete
	// Email sync job will be created by the watcher after this completes
	log.Printf("Account setup completed for account: %s", accountID)

	return nil
}
