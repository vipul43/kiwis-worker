package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAccountRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAccountRepository(db)

	now := time.Now()
	expiresAt := now.Add(time.Hour)
	accessToken := "token123"
	refreshToken := "refresh123"

	rows := sqlmock.NewRows([]string{
		"id", "account_id", "provider_id", "user_id",
		"access_token", "refresh_token", "id_token",
		"access_token_expires_at", "refresh_token_expires_at",
		"scope", "password", "created_at", "updated_at",
	}).AddRow(
		"acc-123", "google-123", "google", "user-123",
		accessToken, refreshToken, nil,
		expiresAt, nil,
		"email", nil, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM account WHERE id = \\$1").
		WithArgs("acc-123").
		WillReturnRows(rows)

	account, err := repo.GetByID(context.Background(), "acc-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if account.ID != "acc-123" {
		t.Errorf("expected ID acc-123, got %s", account.ID)
	}
	if account.UserID != "user-123" {
		t.Errorf("expected UserID user-123, got %s", account.UserID)
	}
	if account.AccessToken == nil || *account.AccessToken != accessToken {
		t.Errorf("expected AccessToken %s, got %v", accessToken, account.AccessToken)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAccountRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAccountRepository(db)

	mock.ExpectQuery("SELECT (.+) FROM account WHERE id = \\$1").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent account, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
