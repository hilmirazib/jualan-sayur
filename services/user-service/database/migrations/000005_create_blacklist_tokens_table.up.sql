CREATE TABLE IF NOT EXISTS blacklist_tokens (
    id SERIAL PRIMARY KEY,
    token_hash VARCHAR(256) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_token_hash ON blacklist_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_expires_at ON blacklist_tokens (expires_at);
