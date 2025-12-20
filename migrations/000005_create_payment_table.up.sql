-- Create payment table for storing extracted payment information
CREATE TABLE payment (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    merchant TEXT NOT NULL,
    description TEXT,
    amount NUMERIC(12, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'INR',
    date TIMESTAMP NOT NULL,
    recurrence TEXT CHECK (
        recurrence IN ('daily', 'weekly', 'biweekly', 'monthly', 'bimonthly', 'quarterly', 'semiannual', 'annual')
        OR recurrence IS NULL
    ),
    status TEXT NOT NULL CHECK (
        status IN ('draft', 'scheduled', 'upcoming', 'due', 'overdue', 'processing', 
                   'partially_paid', 'paid', 'failed', 'refunded', 'cancelled', 'written_off')
    ),
    category TEXT,
    external_reference TEXT,
    metadata JSONB,
    raw_llm_response JSONB, -- Store raw LLM response for debugging/audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_payment_account
        FOREIGN KEY (account_id)
        REFERENCES account(id)
        ON DELETE CASCADE
);

-- Index for querying by account
CREATE INDEX idx_payment_account_id 
    ON payment(account_id);

-- Index for querying by status
CREATE INDEX idx_payment_status 
    ON payment(status);

-- Index for querying by date (for upcoming payments)
CREATE INDEX idx_payment_date 
    ON payment(date DESC);

-- Index for querying by merchant
CREATE INDEX idx_payment_merchant 
    ON payment(merchant);

-- Composite index for account + date queries
CREATE INDEX idx_payment_account_date 
    ON payment(account_id, date DESC);
