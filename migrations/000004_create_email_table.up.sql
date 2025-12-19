-- Create email table for storing raw email data (for LLM fine-tuning)
-- This table is temporary and will be removed once LLM is sufficiently trained
-- No foreign keys - standalone table for training data collection
CREATE TABLE email (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    gmail_message_id TEXT NOT NULL,
    gmail_thread_id TEXT,
    
    -- Email headers
    "from" TEXT,
    "to" TEXT,
    cc TEXT,
    bcc TEXT,
    subject TEXT,
    
    -- Email content
    body_text TEXT,
    body_html TEXT,
    snippet TEXT,
    
    -- Metadata
    received_at TIMESTAMP,
    internal_date TIMESTAMP,
    labels TEXT[], -- Gmail labels as array
    
    -- Raw data for fine-tuning
    raw_headers JSONB, -- All headers as JSON
    raw_payload JSONB, -- Full message payload
    
    -- Attachments info
    has_attachments BOOLEAN DEFAULT FALSE,
    attachments JSONB, -- Array of attachment metadata
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for querying by account
CREATE INDEX idx_email_account_id ON email(account_id);

-- Index for querying by Gmail message ID (prevent duplicates)
CREATE UNIQUE INDEX idx_email_gmail_message_id ON email(gmail_message_id);

-- Index for querying by received date
CREATE INDEX idx_email_received_at ON email(received_at DESC);

-- Index for full-text search on subject and body (for analysis)
CREATE INDEX idx_email_subject_text ON email USING gin(to_tsvector('english', subject));
CREATE INDEX idx_email_body_text ON email USING gin(to_tsvector('english', body_text));
