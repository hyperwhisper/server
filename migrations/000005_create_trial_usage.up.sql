-- Trial API Keys table - stores trial keys linked to device fingerprints
CREATE TABLE trial_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(16) NOT NULL,  -- First 16 chars of key for identification (e.g., "hw_trial_ab12cd")
    device_fingerprint VARCHAR(255) NOT NULL UNIQUE,  -- Hash of machine ID + hardware info
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,  -- 90 days after creation
    last_used_at TIMESTAMP WITH TIME ZONE NULL,
    revoked_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_trial_api_keys_hash ON trial_api_keys(key_hash);
CREATE INDEX idx_trial_api_keys_fingerprint ON trial_api_keys(device_fingerprint);
CREATE INDEX idx_trial_api_keys_active ON trial_api_keys(expires_at) WHERE revoked_at IS NULL;

-- Trial Usage table - tracks per-session usage for trial keys
CREATE TABLE trial_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trial_key_id UUID NOT NULL REFERENCES trial_api_keys(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE NULL,
    duration_seconds DECIMAL(12, 3) NULL,  -- Precise to milliseconds
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'error', 'timeout')),
    error_message TEXT NULL,
    deepgram_params JSONB NOT NULL DEFAULT '{}',
    bytes_sent BIGINT NOT NULL DEFAULT 0,
    client_ip VARCHAR(45) NULL
);

CREATE INDEX idx_trial_usage_key ON trial_usage(trial_key_id);
CREATE INDEX idx_trial_usage_started ON trial_usage(started_at);
CREATE INDEX idx_trial_usage_status ON trial_usage(status);

-- Trial Limits table - configurable quota values (singleton row)
CREATE TABLE trial_limits (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),  -- Ensures only one row
    max_duration_seconds INTEGER NOT NULL DEFAULT 3600,  -- 60 minutes total
    max_sessions INTEGER NOT NULL DEFAULT 100,
    max_session_duration_seconds INTEGER NOT NULL DEFAULT 600,  -- 10 minutes per session
    expiry_days INTEGER NOT NULL DEFAULT 90,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default trial limits
INSERT INTO trial_limits (max_duration_seconds, max_sessions, max_session_duration_seconds, expiry_days)
VALUES (3600, 100, 600, 90);
