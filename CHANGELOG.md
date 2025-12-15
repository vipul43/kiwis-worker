# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Account watcher service with polling mechanism (10-second interval)
- PostgreSQL trigger to automatically create sync jobs on account insert
- Account sync job table with status tracking (pending/processing/completed/failed)
- **Email sync job table with fair round-robin processing**
- **Fair round-robin: new accounts (last_synced_at=NULL) get picked first, then oldest synced jobs**
- **No priority field: uses last_synced_at for natural round-robin ordering**
- **Round-robin email fetching: processes one account at a time**
- **Job lifecycle: pending→processing→pending→...→synced→completed**
- **Sync types: initial (historical), incremental (manual re-sync), webhook (real-time)**
- **Status types: pending, processing, synced, completed, failed**
- **Synced status: marks completion of historical sync, ready for webhook**
- **Completed status: webhook setup complete, job fully finished**
- **Fair round-robin: last_synced_at updated after each batch to push job to back of queue**
- **Reverse chronological fetching: newest emails first for recent payment dues**
- **Pagination support: fetches 50 emails per batch, resumes from last page token**
- **Email sync limits: max 10,000 emails or 1 year of history per account**
- **Token refresh logic with automatic expiry checking**
- **Failed job handling: skipped in round-robin until manually reset**
- **UUID-based IDs: all job IDs use UUIDs for flexibility**
- Retry logic with configurable max attempts (default: 3)
- Graceful shutdown handling with context cancellation
- Database migrations using golang-migrate
- Makefile commands for build, run, migration management, and testing
- Test setup SQL file for database initialization with snake_case columns
- CASCADE delete on account removal (automatically removes sync jobs)
- Clean architecture with separation of concerns (config, database, models, repository, service, watcher)
- Type-safe enums in Go code with VARCHAR storage in database
- Connection pooling configuration
- Environment-based configuration via .env file
- Comprehensive unit tests for all layers (config, models, repository, service)
- Test coverage reporting with HTML output
- Mock-based testing using go-sqlmock for database operations

### Changed

- Database column naming convention to snake_case for PostgreSQL standards
- Status field from ENUM type to VARCHAR(50) with CHECK constraint for easier schema evolution
- AccountProcessor now uses interface for better testability
- **Watcher now handles both account sync and email sync jobs**
- **Account setup creates email sync job after completion**

### Technical

- Foreign key constraint with ON DELETE CASCADE
- Composite index on (status, created_at) for efficient polling
- **Composite index on (status, priority ASC, last_synced_at ASC) for round-robin**
- Unique constraint on account_id (one job per account)
- SSL mode configurable via DATABASE_URL query parameter
- Dependency injection pattern for testability
- **Gmail API client interface for testability (implementation pending)**
