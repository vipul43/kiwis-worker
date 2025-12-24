package models

import "time"

// Account represents a user's OAuth account
// Note: Column names use camelCase to match Prisma/frontend schema
type Account struct {
	ID                    string     `gorm:"column:id;primaryKey"`
	AccountID             string     `gorm:"column:accountId"`
	ProviderID            string     `gorm:"column:providerId"`
	UserID                string     `gorm:"column:userId"`
	AccessToken           *string    `gorm:"column:accessToken"`
	RefreshToken          *string    `gorm:"column:refreshToken"`
	IDToken               *string    `gorm:"column:idToken"`
	AccessTokenExpiresAt  *time.Time `gorm:"column:accessTokenExpiresAt"`
	RefreshTokenExpiresAt *time.Time `gorm:"column:refreshTokenExpiresAt"`
	Scope                 *string    `gorm:"column:scope"`
	Password              *string    `gorm:"column:password"`
	CreatedAt             time.Time  `gorm:"column:createdAt"`
	UpdatedAt             time.Time  `gorm:"column:updatedAt"`
}

// TableName specifies the table name for GORM
func (Account) TableName() string {
	return "account"
}
