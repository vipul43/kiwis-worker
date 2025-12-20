-- Revert to the original INSERT-only trigger
DROP TRIGGER IF EXISTS account_sync_trigger ON account;

-- Restore original function (INSERT only)
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

-- Restore original INSERT-only trigger
CREATE TRIGGER account_insert_trigger
    AFTER INSERT ON account
    FOR EACH ROW
    EXECUTE FUNCTION create_account_sync_job();
