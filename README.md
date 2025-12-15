# Kiwis Worker

An AI-powered app that securely extracts and displays your upcoming payment details from Gmail without storing your email data.

This repository contains the backend worker service that watches for new OAuth accounts and triggers email processing jobs.

## How It Works

1. Frontend inserts new Account (OAuth flow)
2. PostgreSQL trigger automatically creates AccountSyncJob (status: pending)
3. Go watcher polls for pending jobs every 10 seconds
4. Processes account (placeholder for Gmail API integration)
5. Updates job status to completed/failed

**Account Deletion**: When an account is deleted, the sync job is automatically deleted via CASCADE foreign key constraint.

## Quick Start

```bash
# Install dependencies
go mod download

# Install migration CLI
make migrate-install

# Setup database (creates tables with snake_case columns)
psql "$DATABASE_URL" -f test_setup.sql

# Run migrations
make migrate-up

# Start the service
make run
```

The service will:
- Connect to PostgreSQL
- Run pending migrations (creates `account_sync_job` table and trigger)
- Process any pending jobs from previous runs
- Start polling for new accounts

## Project Structure

```
.
├── cmd/watcher/              # Application entry point
├── internal/
│   ├── config/              # Configuration
│   ├── database/            # Connection & migrations
│   ├── models/              # Data structures (type-safe enums)
│   ├── repository/          # Data access layer
│   ├── service/             # Business logic
│   └── watcher/             # Polling & orchestration
├── migrations/              # SQL migrations
└── test_setup.sql          # Test database setup
```

## Configuration

Edit `.env`:
- `DATABASE_URL`: PostgreSQL connection string (add `?sslmode=disable` for local dev)

Example:
```
DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
```

Defaults (in code):
- Poll interval: 10 seconds
- Max retries: 3
- Shutdown timeout: 30 seconds

## Database Schema

### Account Table (snake_case columns)
- `id`, `account_id`, `provider_id`, `user_id`
- `access_token`, `refresh_token`, `id_token`
- `access_token_expires_at`, `refresh_token_expires_at`
- `scope`, `password`, `created_at`, `updated_at`

### Account Sync Job Table
- `id`, `account_id` (unique, FK to account)
- `status` (VARCHAR with CHECK constraint: pending/processing/completed/failed)
- `attempts`, `last_error`
- `created_at`, `updated_at`, `processed_at`

**Note**: Status is stored as VARCHAR (not enum) for easier schema evolution, with CHECK constraint for validation.

## Available Commands

```bash
# Development
make build              # Build the application
make run                # Run the application
make clean              # Clean build artifacts

# Dependencies
make deps               # Download Go dependencies

# Migrations
make migrate-install    # Install golang-migrate CLI
make migrate-up         # Apply all pending migrations
make migrate-down       # Rollback last migration
make migrate-status     # Show current migration version
make migrate-create name=your_migration  # Create new migration

# Testing
make test               # Run all tests
make test-coverage      # Run tests with coverage report
```

## Testing

### Unit Tests

Run all tests:
```bash
make test
```

Generate coverage report:
```bash
make test-coverage
# Opens coverage.html in browser
```

**Current Coverage:**
- Config: 100%
- Repository: 85%
- Service: 100%

### Integration Testing

Insert test account:
```bash
psql "$DATABASE_URL" -c "
INSERT INTO account (
    id, account_id, provider_id, user_id, 
    access_token, refresh_token, access_token_expires_at,
    created_at, updated_at
)
VALUES (
    'test-' || gen_random_uuid()::text,
    'acc-google-123',
    'google',
    'user-123',
    'ya29.test_token',
    'refresh_token',
    NOW() + INTERVAL '1 hour',
    NOW(),
    NOW()
);
"
```

Check job status:
```bash
psql "$DATABASE_URL" -c "SELECT * FROM account_sync_job ORDER BY created_at DESC LIMIT 5;"
```

View watcher logs:
```
Found 1 pending job(s)
Processing job <id> for account <account_id>
Processing account: <account_id> (user: <user_id>)
Successfully completed job <id>
```

## Architecture Decisions

- **Polling vs LISTEN/NOTIFY**: Chose polling for MVP simplicity and reliability
- **Trigger-based job creation**: Ensures no missed accounts even during downtime
- **VARCHAR status over ENUM**: Easier schema evolution without ALTER TYPE migrations
- **snake_case columns**: Standard PostgreSQL convention
- **Graceful shutdown**: Completes current job before exit
- **Retry logic**: Failed jobs retry up to 3 times before marking as failed

## Email Sync Strategy

### Fair Round-Robin with New Account Priority
- **New accounts** (`last_synced_at = NULL`): Get picked first
- **After first batch**: Join the fair pool with everyone else
- **Fair pool**: Oldest `last_synced_at` gets picked next
- **Query**: `ORDER BY last_synced_at ASC NULLS FIRST, created_at ASC`

### Fetching Behavior
- Fetches **50 emails per batch** to minimize memory usage
- Fetches in **reverse chronological order** (newest first) for recent payment dues
- Processes **one account at a time** (round-robin)
- **Resumes from last page** if interrupted
- Filters: **received emails only**, **excludes spam**
- Limits: **10,000 emails max** or **1 year of history** per account

### Job Lifecycle
```
pending → processing → pending → processing → ... → synced
                                                      ↓
                                            (webhook creates new job)
```

- **pending**: Ready to fetch next batch
- **processing**: Currently fetching
- **synced**: All historical emails fetched, waiting for webhook
- **failed**: Failed after max retries, skipped until manually reset

### Round-Robin Fairness
- After processing, `last_synced_at` is updated to NOW()
- Job goes to **back of queue** (oldest `last_synced_at` goes first)
- Prevents immediate re-processing of same account
- Ensures all accounts get equal turns

### Token Management
- Checks token expiry before each API call
- Auto-refreshes if expired or within 5 minutes of expiry
- Updates account table with new tokens and expiry times

## Next Steps

1. **Implement Gmail API client** (`internal/service/email_processor.go`):
   - OAuth2 token refresh
   - Gmail messages.list API
   - Gmail messages.get API (batch)
   
2. **Add emails table** for storing fetched emails

3. **Implement AI payment extraction** from email content

4. **Store extracted payments** in payments table

## Technologies

- **Go 1.21+**: Backend service
- **PostgreSQL**: Database with triggers
- **golang-migrate**: Database migrations
- **go-sqlmock**: Testing library for database mocks
- **Prisma**: Schema management (frontend)

## Conventions

- [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
- [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

## Collaborators

- **Vishnuprakash P**
  - [GitHub](https://github.com/vishkashpvp)
  - [Mail](mailto:vishkash.k@gmail.com)

- **Hassain Saheb S**
  - [GitHub](https://github.com/hafeezzshs)
  - [Mail](mailto:hafeezz.dev@gmail.com)
