-- API Keys table for user-generated API keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,  -- First 12 chars of key for identification (e.g., "hw_live_ab12")
    name VARCHAR(255) NOT NULL DEFAULT 'Default Key',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE NULL,
    revoked_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_active ON api_keys(user_id) WHERE revoked_at IS NULL;

-- Transcription logs for usage tracking
CREATE TABLE transcription_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE NULL,
    duration_seconds DECIMAL(12, 3) NULL,  -- Precise to milliseconds for billing
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'error', 'timeout')),
    error_message TEXT NULL,
    deepgram_params JSONB NOT NULL DEFAULT '{}',
    bytes_sent BIGINT NOT NULL DEFAULT 0,
    client_ip VARCHAR(45) NULL
);

CREATE INDEX idx_transcription_logs_user ON transcription_logs(user_id);
CREATE INDEX idx_transcription_logs_api_key ON transcription_logs(api_key_id);
CREATE INDEX idx_transcription_logs_started ON transcription_logs(started_at);
CREATE INDEX idx_transcription_logs_status ON transcription_logs(status);
CREATE INDEX idx_transcription_logs_user_date ON transcription_logs(user_id, started_at);
