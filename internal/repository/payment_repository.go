package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vipul43/kiwis-worker/internal/models"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment
func (r *PaymentRepository) Create(ctx context.Context, payment models.Payment) error {
	metadataJSON, err := json.Marshal(payment.Metadata)
	if err != nil {
		return err
	}

	rawLlmResponseJSON, err := json.Marshal(payment.RawLlmResponse)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO payment (
			id, account_id, merchant, description, amount, currency, date,
			recurrence, status, category, external_reference, metadata,
			raw_llm_response, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err = r.db.ExecContext(ctx, query,
		payment.ID, payment.AccountID, payment.Merchant, payment.Description,
		payment.Amount, payment.Currency, payment.Date, payment.Recurrence,
		payment.Status, payment.Category, payment.ExternalReference,
		metadataJSON, rawLlmResponseJSON, payment.CreatedAt, payment.UpdatedAt,
	)
	return err
}

// BulkCreate creates multiple payments in a single transaction
func (r *PaymentRepository) BulkCreate(ctx context.Context, payments []models.Payment) error {
	if len(payments) == 0 {
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
		INSERT INTO payment (
			id, account_id, merchant, description, amount, currency, date,
			recurrence, status, category, external_reference, metadata,
			raw_llm_response, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, payment := range payments {
		metadataJSON, err := json.Marshal(payment.Metadata)
		if err != nil {
			return err
		}

		rawLlmResponseJSON, err := json.Marshal(payment.RawLlmResponse)
		if err != nil {
			return err
		}

		_, err = stmt.ExecContext(ctx,
			payment.ID, payment.AccountID, payment.Merchant, payment.Description,
			payment.Amount, payment.Currency, payment.Date, payment.Recurrence,
			payment.Status, payment.Category, payment.ExternalReference,
			metadataJSON, rawLlmResponseJSON, payment.CreatedAt, payment.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByAccountID retrieves all payments for an account
func (r *PaymentRepository) GetByAccountID(ctx context.Context, accountID string) ([]models.Payment, error) {
	query := `
		SELECT id, account_id, merchant, description, amount, currency, date,
		       recurrence, status, category, external_reference, metadata,
		       raw_llm_response, created_at, updated_at
		FROM payment
		WHERE account_id = $1
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var payment models.Payment
		var metadataJSON, rawLlmResponseJSON []byte

		err := rows.Scan(
			&payment.ID, &payment.AccountID, &payment.Merchant, &payment.Description,
			&payment.Amount, &payment.Currency, &payment.Date, &payment.Recurrence,
			&payment.Status, &payment.Category, &payment.ExternalReference,
			&metadataJSON, &rawLlmResponseJSON, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &payment.Metadata); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rawLlmResponseJSON, &payment.RawLlmResponse); err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	return payments, rows.Err()
}
