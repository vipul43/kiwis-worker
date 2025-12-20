package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/vipul43/kiwis-worker/internal/models"
	"github.com/vipul43/kiwis-worker/internal/repository"
)

const (
	MaxEmailsPerAccount = 10000 // Fetch max 10,000 emails per account
	EmailsPerPage       = 50    // Fetch 50 emails per batch
	InitialSyncDays     = 365   // Fetch last 1 year of emails for initial sync
)

type EmailProcessor struct {
	accountRepo      *repository.AccountRepository
	emailSyncJobRepo *repository.EmailSyncJobRepository
	llmSyncJobRepo   *repository.LLMSyncJobRepository
	gmailClient      GmailClient // Interface for Gmail API
}

// GmailClient interface for Gmail API operations
type GmailClient interface {
	FetchMessageIDs(ctx context.Context, accessToken string, query string, maxResults int, pageToken string) (*MessageIDFetchResult, error)
	FetchEmailByID(ctx context.Context, accessToken string, messageID string) (*EmailMessage, error)
	FetchEmails(ctx context.Context, accessToken string, query string, maxResults int, pageToken string) (*EmailFetchResult, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenRefreshResult, error)
}

type MessageIDFetchResult struct {
	MessageIDs    []string
	NextPageToken string
	TotalFetched  int
}

type EmailFetchResult struct {
	Messages      []EmailMessage
	NextPageToken string
	TotalFetched  int
}

type EmailMessage struct {
	ID             string
	ThreadID       string
	Subject        string
	From           string
	To             string
	CC             string
	BCC            string
	Date           time.Time
	InternalDate   time.Time
	BodyText       string
	BodyHTML       string
	Snippet        string
	Labels         []string
	RawHeaders     map[string]interface{}
	RawPayload     map[string]interface{}
	HasAttachments bool
	Attachments    []map[string]interface{}
}

type TokenRefreshResult struct {
	AccessToken  string
	ExpiresAt    time.Time
	RefreshToken string // May be same or new
}

func NewEmailProcessor(
	accountRepo *repository.AccountRepository,
	emailSyncJobRepo *repository.EmailSyncJobRepository,
	llmSyncJobRepo *repository.LLMSyncJobRepository,
	gmailClient GmailClient,
) *EmailProcessor {
	return &EmailProcessor{
		accountRepo:      accountRepo,
		emailSyncJobRepo: emailSyncJobRepo,
		llmSyncJobRepo:   llmSyncJobRepo,
		gmailClient:      gmailClient,
	}
}

// ProcessEmailSyncJob processes a single email sync job
// Updates the job object in-place with new values after successful processing
func (p *EmailProcessor) ProcessEmailSyncJob(ctx context.Context, job *models.EmailSyncJob) error {
	log.Printf("Processing email sync job %s for account %s (type: %s, fetched: %d)",
		job.ID, job.AccountID, job.SyncType, job.EmailsFetched)

	// Fetch account details
	account, err := p.accountRepo.GetByID(ctx, job.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Validate tokens exist
	if account.AccessToken == nil || account.RefreshToken == nil {
		return fmt.Errorf("account missing tokens")
	}

	// Check if access token is expired and refresh if needed
	accessToken := *account.AccessToken
	if p.isTokenExpired(account.AccessTokenExpiresAt) {
		log.Printf("Access token expired for account %s, refreshing...", job.AccountID)
		newToken, err := p.refreshToken(ctx, account)
		if err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		accessToken = newToken
	}

	// Build Gmail query based on sync type
	query := p.buildGmailQuery(*job)

	// Determine how many emails to fetch in this batch
	remainingEmails := MaxEmailsPerAccount - job.EmailsFetched
	if remainingEmails <= 0 {
		log.Printf("Account %s has reached max emails limit (%d)", job.AccountID, MaxEmailsPerAccount)
		return nil // Job is complete
	}

	batchSize := EmailsPerPage
	if remainingEmails < batchSize {
		batchSize = remainingEmails
	}

	// Fetch emails from Gmail
	pageToken := ""
	if job.PageToken != nil {
		pageToken = *job.PageToken
	}

	log.Printf("Fetching %d message IDs for account %s (page_token: %s)", batchSize, job.AccountID, pageToken)

	result, err := p.gmailClient.FetchMessageIDs(ctx, accessToken, query, batchSize, pageToken)
	if err != nil {
		return fmt.Errorf("failed to fetch message IDs: %w", err)
	}

	log.Printf("Fetched %d message IDs for account %s", len(result.MessageIDs), job.AccountID)

	// Create LLM sync jobs for each message ID
	if len(result.MessageIDs) > 0 {
		llmJobs := make([]models.LLMSyncJob, 0, len(result.MessageIDs))
		now := time.Now()

		for _, messageID := range result.MessageIDs {
			llmJob := models.LLMSyncJob{
				ID:           uuid.New().String(),
				AccountID:    job.AccountID,
				MessageID:    messageID,
				Status:       models.LLMStatusPending,
				LastSyncedAt: nil, // NULL = new job, gets priority in round-robin
				Attempts:     0,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			llmJobs = append(llmJobs, llmJob)
		}

		// Bulk create LLM sync jobs
		if err := p.llmSyncJobRepo.BulkCreate(ctx, llmJobs); err != nil {
			return fmt.Errorf("failed to create LLM sync jobs: %w", err)
		}
		log.Printf("Created %d LLM sync jobs for account %s", len(llmJobs), job.AccountID)
	}

	// Update job progress
	newEmailsFetched := job.EmailsFetched + len(result.MessageIDs)
	var nextPageToken *string
	if result.NextPageToken != "" {
		nextPageToken = &result.NextPageToken
	}

	err = p.emailSyncJobRepo.UpdateProgress(ctx, job.ID, newEmailsFetched, nextPageToken)
	if err != nil {
		return fmt.Errorf("failed to update job progress: %w", err)
	}

	// Update the job object in-place with new values (since DB update succeeded)
	job.EmailsFetched = newEmailsFetched
	job.PageToken = nextPageToken

	log.Printf("Updated job %s: emails_fetched=%d, has_more=%v", job.ID, newEmailsFetched, nextPageToken != nil)

	return nil
}

// isTokenExpired checks if access token is expired or will expire within 5 minutes
func (p *EmailProcessor) isTokenExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return true // Assume expired if no expiry time
	}
	return time.Now().Add(5 * time.Minute).After(*expiresAt)
}

// refreshToken refreshes the access token and updates the account
func (p *EmailProcessor) refreshToken(ctx context.Context, account *repository.Account) (string, error) {
	if account.RefreshToken == nil {
		return "", fmt.Errorf("no refresh token available")
	}

	result, err := p.gmailClient.RefreshAccessToken(ctx, *account.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update account with new tokens
	err = p.accountRepo.UpdateTokens(ctx, account.ID, result.AccessToken, result.RefreshToken, result.ExpiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to update tokens in database: %w", err)
	}

	log.Printf("Token refreshed for account %s, expires at %s", account.ID, result.ExpiresAt)

	return result.AccessToken, nil
}

// buildGmailQuery builds the Gmail API query string based on sync type
// Fetches emails in REVERSE chronological order (newest first) for recent payment dues
func (p *EmailProcessor) buildGmailQuery(job models.EmailSyncJob) string {
	// Base query: only received emails (not sent), exclude spam
	baseQuery := "in:inbox -in:spam"

	// Add time filter based on sync type
	if job.SyncType == models.SyncTypeInitial {
		// Initial sync: fetch last 1 year (Gmail returns newest first by default)
		afterDate := time.Now().AddDate(-1, 0, 0).Format("2006/01/02")
		baseQuery += fmt.Sprintf(" after:%s", afterDate)
	} else if job.LastSyncedAt != nil {
		// Incremental sync: fetch emails after last sync
		afterDate := job.LastSyncedAt.Format("2006/01/02")
		baseQuery += fmt.Sprintf(" after:%s", afterDate)
	}

	log.Printf("Gmail query for job %s: %s (newest first)", job.ID, baseQuery)
	return baseQuery
}

// CreateInitialEmailSyncJob creates an initial email sync job for a new account
// New jobs have last_synced_at = NULL, giving them priority in round-robin
func (p *EmailProcessor) CreateInitialEmailSyncJob(ctx context.Context, accountID string) error {
	job := models.EmailSyncJob{
		ID:            uuid.New().String(),
		AccountID:     accountID,
		Status:        models.EmailStatusPending,
		SyncType:      models.SyncTypeInitial,
		EmailsFetched: 0,
		Attempts:      0,
		LastSyncedAt:  nil, // NULL = new job, gets priority in round-robin
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := p.emailSyncJobRepo.Create(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create email sync job: %w", err)
	}

	log.Printf("Created initial email sync job %s for account %s (will be picked first)", job.ID, accountID)
	return nil
}
