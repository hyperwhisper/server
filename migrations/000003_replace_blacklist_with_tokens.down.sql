-- Drop new tokens table
DROP TABLE IF EXISTS tokens;

-- Recreate old token_blacklist table
CREATE TABLE token_blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_jti VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blacklisted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_token_blacklist_jti ON token_blacklist(token_jti);
CREATE INDEX idx_token_blacklist_user ON token_blacklist(user_id);
CREATE INDEX idx_token_blacklist_expires ON token_blacklist(expires_at);
