CREATE TABLE IF NOT EXISTS blacklist_tokens (
    id SERIAL PRIMARY KEY,
    token_hash VARCHAR(256) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token_hash (token_hash),
    INDEX idx_expires_at (expires_at)
);
