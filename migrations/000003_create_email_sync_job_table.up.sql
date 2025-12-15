-- Create email sync job table
CREATE TABLE IF NOT EXISTS email_sync_job (
    id VARCHAR(255) PRIMARY KEY,
    account_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    sync_type VARCHAR(20) NOT NULL DEFAULT 'initial', -- 'initial' or 'incremental' or 'webhook'
    
    -- Progress tracking
    emails_fetched INTEGER NOT NULL DEFAULT 0,
    page_token VARCHAR(255), -- Gmail pagination token for resuming
    last_synced_at TIMESTAMP, -- NULL = never synced (new jobs), used for round-robin
    
    -- Error handling
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    
    CONSTRAINT fk_account
        FOREIGN KEY (account_id)
        REFERENCES account(id)
        ON DELETE CASCADE,
    CONSTRAINT chk_status
        CHECK (status IN ('pending', 'processing', 'synced', 'completed', 'failed')),
    CONSTRAINT chk_sync_type
        CHECK (sync_type IN ('initial', 'incremental', 'webhook'))
);

-- Create index for efficient round-robin polling
-- NULL last_synced_at (new jobs) come first, then oldest synced jobs
CREATE INDEX idx_email_sync_job_round_robin ON email_sync_job(status, last_synced_at ASC NULLS FIRST, created_at ASC);

-- Create index for account lookup
CREATE INDEX idx_email_sync_job_account ON email_sync_job(account_id);
