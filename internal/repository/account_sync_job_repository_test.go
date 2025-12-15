package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/yourusername/payment-tracker/internal/models"
)

func TestAccountSyncJobRepository_GetPendingJobs_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAccountSyncJobRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "account_id", "status", "attempts", "last_error",
		"created_at", "updated_at", "processed_at",
	}).AddRow(
		"job-1", "acc-1", "pending", 0, nil, now, now, nil,
	).AddRow(
		"job-2", "acc-2", "pending", 1, nil, now, now, nil,
	)

	mock.ExpectQuery("SELECT (.+) FROM account_sync_job WHERE status = \\$1").
		WithArgs(models.StatusPending, 10).
		WillReturnRows(rows)

	jobs, err := repo.GetPendingJobs(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	if jobs[0].ID != "job-1" {
		t.Errorf("expected first job ID to be job-1, got %s", jobs[0].ID)
	}
	if jobs[0].Status != models.StatusPending {
		t.Errorf("expected status pending, got %s", jobs[0].Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAccountSyncJobRepository_UpdateStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAccountSyncJobRepository(db)

	mock.ExpectExec("UPDATE account_sync_job SET status = \\$1").
		WithArgs(models.StatusCompleted, nil, sqlmock.AnyArg(), sqlmock.AnyArg(), "job-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateStatus(context.Background(), "job-1", models.StatusCompleted, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAccountSyncJobRepository_IncrementAttempts_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewAccountSyncJobRepository(db)

	mock.ExpectExec("UPDATE account_sync_job SET attempts = attempts \\+ 1").
		WithArgs(sqlmock.AnyArg(), "job-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.IncrementAttempts(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
