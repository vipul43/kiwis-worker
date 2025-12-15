# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Account watcher service with polling mechanism (10-second interval)
- PostgreSQL trigger to automatically create sync jobs on account insert
- Account sync job table with status tracking (pending/processing/completed/failed)
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

### Technical

- Foreign key constraint with ON DELETE CASCADE
- Composite index on (status, created_at) for efficient polling
- Unique constraint on account_id (one job per account)
- SSL mode configurable via DATABASE_URL query parameter
- Dependency injection pattern for testability
