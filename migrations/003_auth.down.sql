-- Drop indexes
DROP INDEX IF EXISTS idx_api_tokens_expires_at;
DROP INDEX IF EXISTS idx_api_tokens_user_id;
DROP INDEX IF EXISTS idx_api_tokens_token_hash;
DROP INDEX IF EXISTS idx_users_username;

-- Drop tables
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS users;

