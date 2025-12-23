-- Change date column from TIMESTAMP to TIMESTAMPTZ (timestamp with timezone)
ALTER TABLE payment ALTER COLUMN date TYPE TIMESTAMPTZ;
