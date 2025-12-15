package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/payment-tracker/internal/repository"
)

type mockAccountRepository struct {
	getByIDFunc func(ctx context.Context, accountID string) (*repository.Account, error)
}

func (m *mockAccountRepository) GetByID(ctx context.Context, accountID string) (*repository.Account, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, accountID)
	}
	return nil, nil
}

func TestAccountProcessor_ProcessAccount_Success(t *testing.T) {
	accessToken := "token123"
	mockRepo := &mockAccountRepository{
		getByIDFunc: func(ctx context.Context, accountID string) (*repository.Account, error) {
			return &repository.Account{
				ID:          accountID,
				UserID:      "user-123",
				AccessToken: &accessToken,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	processor := NewAccountProcessor(mockRepo)

	err := processor.ProcessAccount(context.Background(), "acc-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAccountProcessor_ProcessAccount_MissingToken(t *testing.T) {
	mockRepo := &mockAccountRepository{
		getByIDFunc: func(ctx context.Context, accountID string) (*repository.Account, error) {
			return &repository.Account{
				ID:          accountID,
				UserID:      "user-123",
				AccessToken: nil, // Missing token
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	processor := NewAccountProcessor(mockRepo)

	err := processor.ProcessAccount(context.Background(), "acc-123")
	if err == nil {
		t.Fatal("expected error for missing access token, got nil")
	}

	expectedMsg := "account missing access token"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestAccountProcessor_ProcessAccount_AccountNotFound(t *testing.T) {
	mockRepo := &mockAccountRepository{
		getByIDFunc: func(ctx context.Context, accountID string) (*repository.Account, error) {
			return nil, errors.New("account not found")
		},
	}

	processor := NewAccountProcessor(mockRepo)

	err := processor.ProcessAccount(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent account, got nil")
	}
}
