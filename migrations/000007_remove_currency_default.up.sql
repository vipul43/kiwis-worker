-- Remove default value from currency column
-- We don't want to assume INR, LLM should always provide the currency
ALTER TABLE payment ALTER COLUMN currency DROP DEFAULT;
