-- Drop old token_blacklist table
DROP TABLE IF EXISTS token_blacklist;

-- Create new tokens table (only tracks refresh tokens)
CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_jti VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE NULL,
    revoked_reason VARCHAR(255) NULL
);

CREATE INDEX idx_tokens_jti ON tokens(token_jti);
CREATE INDEX idx_tokens_user ON tokens(user_id);
CREATE INDEX idx_tokens_expires ON tokens(expires_at);
CREATE INDEX idx_tokens_revoked ON tokens(revoked_at) WHERE revoked_at IS NOT NULL;
