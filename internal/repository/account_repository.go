package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vipul43/kiwis-worker/internal/models"
	"gorm.io/gorm"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// GetByID retrieves account by ID
func (r *AccountRepository) GetByID(ctx context.Context, accountID string) (*models.Account, error) {
	var account models.Account
	result := r.db.WithContext(ctx).First(&account, "id = ?", accountID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", result.Error)
	}
	return &account, nil
}

// UpdateTokens updates access token, refresh token, and their expiry times
func (r *AccountRepository) UpdateTokens(ctx context.Context, accountID string, accessToken string, refreshToken string, accessTokenExpiresAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.Account{}).
		Where("id = ?", accountID).
		Updates(map[string]interface{}{
			"accessToken":          accessToken,
			"refreshToken":         refreshToken,
			"accessTokenExpiresAt": accessTokenExpiresAt,
			"updatedAt":            time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update tokens: %w", result.Error)
	}
	return nil
}
