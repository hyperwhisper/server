-- =====================
-- TRIAL API KEY QUERIES
-- =====================

-- name: CreateTrialAPIKey :one
INSERT INTO trial_api_keys (key_hash, key_prefix, device_fingerprint, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTrialAPIKeyByHash :one
SELECT * FROM trial_api_keys WHERE key_hash = $1 AND revoked_at IS NULL;

-- name: GetTrialAPIKeyByFingerprint :one
SELECT * FROM trial_api_keys WHERE device_fingerprint = $1;

-- name: GetTrialAPIKeyByID :one
SELECT * FROM trial_api_keys WHERE id = $1;

-- name: UpdateTrialAPIKeyLastUsed :exec
UPDATE trial_api_keys SET last_used_at = NOW() WHERE id = $1;

-- name: RevokeTrialAPIKey :exec
UPDATE trial_api_keys SET revoked_at = NOW() WHERE id = $1;

-- name: RegenerateTrialAPIKey :one
UPDATE trial_api_keys
SET key_hash = $2, key_prefix = $3
WHERE id = $1
RETURNING *;

-- name: CountTrialAPIKeys :one
SELECT COUNT(*) FROM trial_api_keys;

-- name: CountActiveTrialAPIKeys :one
SELECT COUNT(*) FROM trial_api_keys WHERE revoked_at IS NULL AND expires_at > NOW();

-- name: ListTrialAPIKeys :many
SELECT * FROM trial_api_keys ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- =====================
-- TRIAL USAGE QUERIES
-- =====================

-- name: CreateTrialUsageLog :one
INSERT INTO trial_usage (trial_key_id, deepgram_params, client_ip)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTrialUsageComplete :exec
UPDATE trial_usage
SET ended_at = NOW(),
    duration_seconds = $2,
    status = 'completed',
    bytes_sent = $3
WHERE id = $1;

-- name: UpdateTrialUsageError :exec
UPDATE trial_usage
SET ended_at = NOW(),
    status = 'error',
    error_message = $2,
    bytes_sent = $3
WHERE id = $1;

-- name: UpdateTrialUsageTimeout :exec
UPDATE trial_usage
SET ended_at = NOW(),
    status = 'timeout',
    bytes_sent = $2
WHERE id = $1;

-- name: GetTrialUsageLog :one
SELECT * FROM trial_usage WHERE id = $1;

-- name: ListTrialUsageLogs :many
SELECT * FROM trial_usage WHERE trial_key_id = $1 ORDER BY started_at DESC LIMIT $2 OFFSET $3;

-- name: CountTrialUsageLogs :one
SELECT COUNT(*) FROM trial_usage WHERE trial_key_id = $1;

-- name: CountTrialSessions :one
SELECT COUNT(*) FROM trial_usage WHERE trial_key_id = $1;

-- name: GetTrialUsageSummary :one
SELECT
    COUNT(*) as total_sessions,
    COALESCE(SUM(duration_seconds), 0)::DECIMAL(12,3) as total_duration_seconds,
    COALESCE(SUM(bytes_sent), 0) as total_bytes_sent
FROM trial_usage
WHERE trial_key_id = $1;

-- name: GetTrialUsageSummaryActive :one
SELECT
    COUNT(*) as total_sessions,
    COALESCE(SUM(duration_seconds), 0)::DECIMAL(12,3) as total_duration_seconds,
    COALESCE(SUM(bytes_sent), 0) as total_bytes_sent
FROM trial_usage
WHERE trial_key_id = $1 AND (status = 'completed' OR status = 'timeout');

-- =====================
-- TRIAL LIMITS QUERIES
-- =====================

-- name: GetTrialLimits :one
SELECT * FROM trial_limits WHERE id = 1;

-- name: UpdateTrialLimits :one
UPDATE trial_limits
SET max_duration_seconds = $1,
    max_sessions = $2,
    max_session_duration_seconds = $3,
    expiry_days = $4,
    updated_at = NOW()
WHERE id = 1
RETURNING *;

-- =====================
-- ADMIN TRIAL QUERIES
-- =====================

-- name: ListAllTrialAPIKeys :many
SELECT
    tak.*,
    COALESCE(usage_stats.total_sessions, 0)::bigint as total_sessions,
    COALESCE(usage_stats.total_duration_seconds, 0)::DECIMAL(12,3) as total_duration_seconds
FROM trial_api_keys tak
LEFT JOIN (
    SELECT
        trial_key_id,
        COUNT(*) as total_sessions,
        SUM(duration_seconds) as total_duration_seconds
    FROM trial_usage
    GROUP BY trial_key_id
) usage_stats ON tak.id = usage_stats.trial_key_id
ORDER BY tak.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAllTrialUsageSummary :one
SELECT
    COUNT(DISTINCT tak.id) as total_trial_keys,
    COUNT(DISTINCT CASE WHEN tak.revoked_at IS NULL AND tak.expires_at > NOW() THEN tak.id END) as active_trial_keys,
    COALESCE(SUM(tu.duration_seconds), 0)::DECIMAL(12,3) as total_duration_seconds,
    COUNT(tu.id) as total_sessions,
    COALESCE(SUM(tu.bytes_sent), 0) as total_bytes_sent
FROM trial_api_keys tak
LEFT JOIN trial_usage tu ON tak.id = tu.trial_key_id
WHERE tu.started_at >= sqlc.arg(start_date) AND tu.started_at < sqlc.arg(end_date);

-- name: ListAllTrialUsageLogs :many
SELECT
    tu.*,
    tak.key_prefix,
    tak.device_fingerprint
FROM trial_usage tu
JOIN trial_api_keys tak ON tu.trial_key_id = tak.id
ORDER BY tu.started_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllTrialUsageLogs :one
SELECT COUNT(*) FROM trial_usage;

-- name: CleanupExpiredTrialKeys :exec
UPDATE trial_api_keys SET revoked_at = NOW() WHERE expires_at <= NOW() AND revoked_at IS NULL;

-- name: UnrevokeTrialAPIKey :exec
UPDATE trial_api_keys SET revoked_at = NULL WHERE id = $1;

-- name: DeleteTrialAPIKey :exec
DELETE FROM trial_api_keys WHERE id = $1;
