-- Function to create account sync job on account insert
-- Uses gen_random_uuid() for UUID generation
CREATE OR REPLACE FUNCTION create_account_sync_job()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO account_sync_job (id, account_id, status, created_at, updated_at)
    VALUES (
        gen_random_uuid()::text,
        NEW.id,
        'pending',
        NOW(),
        NOW()
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger on account table insert
CREATE TRIGGER account_insert_trigger
    AFTER INSERT ON account
    FOR EACH ROW
    EXECUTE FUNCTION create_account_sync_job();
