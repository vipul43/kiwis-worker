-- Create llm_sync_job table for tracking LLM payment extraction
CREATE TABLE llm_sync_job (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    message_id TEXT NOT NULL, -- Gmail message ID
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    last_synced_at TIMESTAMP,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    
    CONSTRAINT fk_llm_sync_job_account
        FOREIGN KEY (account_id)
        REFERENCES account(id)
        ON DELETE CASCADE
);

-- Index for efficient polling by status and last synced time (round-robin)
CREATE INDEX idx_llm_sync_job_status_last_synced 
    ON llm_sync_job(status, last_synced_at ASC NULLS FIRST);

-- Index for querying by account
CREATE INDEX idx_llm_sync_job_account_id 
    ON llm_sync_job(account_id);

-- Index for preventing duplicate message processing
CREATE UNIQUE INDEX idx_llm_sync_job_message_id 
    ON llm_sync_job(message_id);

-- Index for created_at (useful for debugging and monitoring)
CREATE INDEX idx_llm_sync_job_created_at 
    ON llm_sync_job(created_at DESC);
