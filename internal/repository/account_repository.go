package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Account struct {
	ID                    string
	AccountID             string
	ProviderID            string
	UserID                string
	AccessToken           *string
	RefreshToken          *string
	IDToken               *string
	AccessTokenExpiresAt  *time.Time
	RefreshTokenExpiresAt *time.Time
	Scope                 *string
	Password              *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

var ErrAccountNotFound = fmt.Errorf("account not found")

// GetByID retrieves account by ID
func (r *AccountRepository) GetByID(ctx context.Context, accountID string) (*Account, error) {
	query := `
		SELECT id, account_id, provider_id, user_id, access_token, refresh_token, 
		       id_token, access_token_expires_at, refresh_token_expires_at, scope, password,
		       created_at, updated_at
		FROM account
		WHERE id = $1
	`

	var acc Account
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&acc.ID,
		&acc.AccountID,
		&acc.ProviderID,
		&acc.UserID,
		&acc.AccessToken,
		&acc.RefreshToken,
		&acc.IDToken,
		&acc.AccessTokenExpiresAt,
		&acc.RefreshTokenExpiresAt,
		&acc.Scope,
		&acc.Password,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &acc, nil
}
