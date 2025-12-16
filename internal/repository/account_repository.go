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
		SELECT id, "accountId", "providerId", "userId", "accessToken", "refreshToken", 
		       "idToken", "accessTokenExpiresAt", "refreshTokenExpiresAt", scope, password,
		       "createdAt", "updatedAt"
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

// UpdateTokens updates access token, refresh token, and their expiry times
func (r *AccountRepository) UpdateTokens(ctx context.Context, accountID string, accessToken string, refreshToken string, accessTokenExpiresAt time.Time) error {
	query := `
		UPDATE account
		SET "accessToken" = $1,
		    "refreshToken" = $2,
		    "accessTokenExpiresAt" = $3,
		    "updatedAt" = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, accessToken, refreshToken, accessTokenExpiresAt, time.Now(), accountID)
	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	return nil
}
