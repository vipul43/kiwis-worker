-- Create user table (required for foreign key)
CREATE TABLE IF NOT EXISTS "user" (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified BOOLEAN DEFAULT false,
    image VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create account table with snake_case columns
CREATE TABLE IF NOT EXISTS account (
    id VARCHAR(255) PRIMARY KEY,
    account_id VARCHAR(255) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    id_token TEXT,
    access_token_expires_at TIMESTAMP,
    refresh_token_expires_at TIMESTAMP,
    scope TEXT,
    password TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES "user"(id)
        ON DELETE CASCADE
);

-- Insert test user
INSERT INTO "user" (id, name, email, created_at, updated_at)
VALUES ('user-123', 'Test User', 'test@example.com', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert test account (this will trigger the account_sync_job creation)
INSERT INTO account (
    id, 
    account_id, 
    provider_id, 
    user_id, 
    access_token, 
    refresh_token,
    access_token_expires_at,
    created_at, 
    updated_at
)
VALUES (
    'test-' || gen_random_uuid()::text,
    'acc-google-123',
    'google',
    'user-123',
    'ya29.test_access_token_here',
    'test_refresh_token_here',
    NOW() + INTERVAL '1 hour',
    NOW(),
    NOW()
);
