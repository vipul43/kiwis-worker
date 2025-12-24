-- Revert date column from TIMESTAMPTZ to TIMESTAMP (without timezone)
ALTER TABLE payment ALTER COLUMN date TYPE TIMESTAMP;
