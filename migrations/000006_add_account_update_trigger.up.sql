-- Update the function to handle both INSERT and UPDATE
-- On INSERT: creates a new account sync job
-- On UPDATE: resets existing account sync job to pending status
CREATE OR REPLACE FUNCTION handle_account_sync_job()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Create new account sync job for new account
        INSERT INTO account_sync_job (id, account_id, status, created_at, updated_at)
        VALUES (
            gen_random_uuid()::text,
            NEW.id,
            'pending',
            NOW(),
            NOW()
        );
    ELSIF TG_OP = 'UPDATE' THEN
        -- Reset existing account sync job to pending status
        -- This will rerun the entire sync process from the beginning
        UPDATE account_sync_job
        SET status = 'pending',
            last_error = NULL,
            processed_at = NULL,
            updated_at = NOW()
        WHERE account_id = NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop old trigger and create new one for both INSERT and UPDATE
DROP TRIGGER IF EXISTS account_insert_trigger ON account;

CREATE TRIGGER account_sync_trigger
    AFTER INSERT OR UPDATE ON account
    FOR EACH ROW
    EXECUTE FUNCTION handle_account_sync_job();
