package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"github.com/vipul43/kiwis-worker/internal/models"
)

type EmailRepository struct {
	db *sql.DB
}

func NewEmailRepository(db *sql.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// Create inserts a new email record
func (r *EmailRepository) Create(ctx context.Context, email models.Email) error {
	query := `
		INSERT INTO email (
			id, account_id, gmail_message_id, gmail_thread_id,
			"from", "to", cc, bcc, subject,
			body_text, body_html, snippet,
			received_at, internal_date, labels,
			raw_headers, raw_payload,
			has_attachments, attachments,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21
		)
		ON CONFLICT (gmail_message_id) DO NOTHING
	`

	rawHeadersJSON, err := json.Marshal(email.RawHeaders)
	if err != nil {
		return fmt.Errorf("failed to marshal raw headers: %w", err)
	}

	rawPayloadJSON, err := json.Marshal(email.RawPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal raw payload: %w", err)
	}

	attachmentsJSON, err := json.Marshal(email.Attachments)
	if err != nil {
		return fmt.Errorf("failed to marshal attachments: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		email.ID,
		email.AccountID,
		email.GmailMessageID,
		email.GmailThreadID,
		email.From,
		email.To,
		email.CC,
		email.BCC,
		email.Subject,
		email.BodyText,
		email.BodyHTML,
		email.Snippet,
		email.ReceivedAt,
		email.InternalDate,
		pq.Array(email.Labels),
		rawHeadersJSON,
		rawPayloadJSON,
		email.HasAttachments,
		attachmentsJSON,
		email.CreatedAt,
		email.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert email: %w", err)
	}

	return nil
}

// BulkCreate inserts multiple emails in a single transaction
func (r *EmailRepository) BulkCreate(ctx context.Context, emails []models.Email) error {
	if len(emails) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback is safe to call even after commit
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO email (
			id, account_id, gmail_message_id, gmail_thread_id,
			"from", "to", cc, bcc, subject,
			body_text, body_html, snippet,
			received_at, internal_date, labels,
			raw_headers, raw_payload,
			has_attachments, attachments,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21
		)
		ON CONFLICT (gmail_message_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, email := range emails {
		rawHeadersJSON, err := json.Marshal(email.RawHeaders)
		if err != nil {
			return fmt.Errorf("failed to marshal raw headers: %w", err)
		}

		rawPayloadJSON, err := json.Marshal(email.RawPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal raw payload: %w", err)
		}

		attachmentsJSON, err := json.Marshal(email.Attachments)
		if err != nil {
			return fmt.Errorf("failed to marshal attachments: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			email.ID,
			email.AccountID,
			email.GmailMessageID,
			email.GmailThreadID,
			email.From,
			email.To,
			email.CC,
			email.BCC,
			email.Subject,
			email.BodyText,
			email.BodyHTML,
			email.Snippet,
			email.ReceivedAt,
			email.InternalDate,
			pq.Array(email.Labels),
			rawHeadersJSON,
			rawPayloadJSON,
			email.HasAttachments,
			attachmentsJSON,
			email.CreatedAt,
			email.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert email %s: %w", email.GmailMessageID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Exists checks if an email with the given Gmail message ID already exists
func (r *EmailRepository) Exists(ctx context.Context, gmailMessageID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM email WHERE gmail_message_id = $1)`
	err := r.db.QueryRowContext(ctx, query, gmailMessageID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}
