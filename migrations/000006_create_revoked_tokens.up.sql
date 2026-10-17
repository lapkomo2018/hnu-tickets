CREATE TABLE revoked_tokens (
    id UUID PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_revoked_tokens_expiry ON revoked_tokens(expires_at);
