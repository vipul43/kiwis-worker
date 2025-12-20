# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Account watcher service with polling mechanism (10-second interval)
- PostgreSQL trigger to automatically create sync jobs on account insert
- Account sync job table with status tracking (pending/processing/completed/failed)
- Email sync job table with fair round-robin processing
- Email table for storing raw email data (for LLM fine-tuning)
- Fair round-robin: new accounts (last_synced_at=NULL) get picked first, then oldest synced jobs
- No priority field: uses last_synced_at for natural round-robin ordering
- Round-robin email fetching: processes one account at a time
- Job lifecycle: pending→processing→synced (email) or completed (account)
- Sync types: initial (historical), incremental (manual re-sync), webhook (real-time)
- Status types: pending, processing, synced, completed, failed
- Synced status: marks completion of historical sync, ready for webhook
- Completed status: account setup complete, email sync job created
- Fair round-robin: last_synced_at updated after each batch to push job to back of queue
- Reverse chronological fetching: newest emails first for recent payment dues
- Pagination support: fetches 50 emails per batch, resumes from last page token
- Email sync limits: max 10,000 emails or 1 year of history per account
- Token refresh logic with automatic expiry checking (5-minute buffer)
- UUID-based IDs: all job IDs use UUIDs for flexibility
- Gmail API integration: OAuth2 token refresh, email fetching, comprehensive data extraction
- Token management: automatic refresh and database updates
- Email parsing: extracts from, to, cc, bcc, subject, body (text/html), attachments, headers
- Email storage: stores complete email data including raw headers, payload, and attachments metadata
- Pagination: Gmail API pagination with nextPageToken support
- GetProcessingJobs: fetches stuck jobs from crashes/restarts
- Infinite retry with failed status: jobs retry forever, failures go to failed status
- Graceful shutdown handling with context cancellation
- Database migrations using golang-migrate
- Makefile commands for build, run, migration management, testing, formatting, and linting
- CI/CD pipeline with GitHub Actions (format check, lint, test, build)
- CASCADE delete on account removal (automatically removes sync jobs)
- Clean architecture with separation of concerns (config, database, models, repository, service, watcher)
- Type-safe enums in Go code with VARCHAR storage in database
- Connection pooling configuration
- Environment-based configuration via .env file
- Comprehensive unit tests for all layers (config, models, repository, service)
- Test coverage reporting with HTML output
- Mock-based testing using go-sqlmock for database operations

### Changed

- Database column naming: uses camelCase to match Prisma/frontend schema
- Status field from ENUM type to VARCHAR(50) with CHECK constraint for easier schema evolution
- AccountProcessor now uses interface for better testability
- Watcher now handles both account sync and email sync jobs
- Account setup creates email sync job after completion
- All SQL queries use quoted camelCase column names (e.g., "accountId", "userId")
- Removed max retry limit: jobs now retry infinitely
- Error handlers: failures go to failed status (not pending) for clear state tracking
- Watcher fetches pending, failed, AND processing jobs for crash recovery
- Email sync partial success: stays in processing status (not pending)
- ProcessEmailSyncJob: updates job object in-place to avoid extra DB queries
- Failed jobs: update last_synced_at on failure to prevent queue blocking
- Round-robin: queries pre-sorted by last_synced_at, no additional sorting needed
- Improved error messages: specific validation for missing access/refresh tokens

### Removed

- Email table and all email storage functionality (simplified architecture for payment extraction focus)
- Environment-based configuration (no longer needed without email storage)
- Email table migration files (000004_create_email_table.up.sql and .down.sql)
- Email model (internal/models/email.go)
- Email repository (internal/repository/email_repository.go)
- Email storage logic from EmailProcessor
- EmailProcessor dependencies: emailRepo and environment parameters
- Configuration struct Environment field
- ENVIRONMENT variable from .env and .env.example

### Fixed

- Email job handler now passes job by pointer to preserve in-place updates from ProcessEmailSyncJob

### Technical

- Foreign key constraint with ON DELETE CASCADE
- Composite index on (status, created_at) for efficient polling
- Composite index on (status, last_synced_at ASC) for round-robin
- Unique index on gmail_message_id to prevent duplicate emails
- Full-text search indexes on email subject and body for analysis
- Unique constraint on account_id (one job per account)
- SSL mode configurable via DATABASE_URL query parameter
- Dependency injection pattern for testability
- Gmail API client interface for testability
- Separate GetPendingJobs(), GetFailedJobs(), and GetProcessingJobs() repository methods
- Account sync: processes pending + failed + processing jobs (up to 5 each)
- Email sync: fetches 1 from each status, sorts combined list by last_synced_at for fairness
- Email table: no foreign keys (standalone for LLM training data)
- JSONB storage for raw headers, payload, and attachments
- BulkCreate for efficient email insertion with ON CONFLICT DO NOTHING
