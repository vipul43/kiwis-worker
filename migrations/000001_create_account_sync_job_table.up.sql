-- Create account sync job table
CREATE TABLE IF NOT EXISTS account_sync_job (
    id VARCHAR(255) PRIMARY KEY,
    account_id VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    
    CONSTRAINT fk_account
        FOREIGN KEY (account_id)
        REFERENCES account(id)
        ON DELETE CASCADE,
    CONSTRAINT chk_status
        CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

-- Create index for efficient polling
CREATE INDEX idx_account_sync_job_status_created ON account_sync_job(status, created_at);
