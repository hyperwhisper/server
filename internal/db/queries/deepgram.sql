-- =====================
-- API KEY QUERIES
-- =====================

-- name: CreateAPIKey :one
INSERT INTO api_keys (user_id, key_hash, key_prefix, name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL;

-- name: GetAPIKeyByID :one
SELECT * FROM api_keys WHERE id = $1;

-- name: ListUserAPIKeys :many
SELECT * FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountUserAPIKeys :one
SELECT COUNT(*) FROM api_keys WHERE user_id = $1;

-- name: CountActiveUserAPIKeys :one
SELECT COUNT(*) FROM api_keys WHERE user_id = $1 AND revoked_at IS NULL;

-- name: RevokeAPIKey :exec
UPDATE api_keys SET revoked_at = NOW() WHERE id = $1 AND user_id = $2;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = NOW() WHERE id = $1;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys WHERE id = $1 AND user_id = $2;

-- =====================
-- TRANSCRIPTION LOG QUERIES
-- =====================

-- name: CreateTranscriptionLog :one
INSERT INTO transcription_logs (user_id, api_key_id, deepgram_params, client_ip)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTranscriptionLogComplete :exec
UPDATE transcription_logs
SET ended_at = NOW(),
    duration_seconds = $2,
    status = 'completed',
    bytes_sent = $3
WHERE id = $1;

-- name: UpdateTranscriptionLogError :exec
UPDATE transcription_logs
SET ended_at = NOW(),
    status = 'error',
    error_message = $2,
    bytes_sent = $3
WHERE id = $1;

-- name: UpdateTranscriptionLogTimeout :exec
UPDATE transcription_logs
SET ended_at = NOW(),
    status = 'timeout',
    bytes_sent = $2
WHERE id = $1;

-- name: GetTranscriptionLog :one
SELECT * FROM transcription_logs WHERE id = $1;

-- name: ListUserTranscriptionLogs :many
SELECT * FROM transcription_logs WHERE user_id = $1 ORDER BY started_at DESC LIMIT $2 OFFSET $3;

-- name: CountUserTranscriptionLogs :one
SELECT COUNT(*) FROM transcription_logs WHERE user_id = $1;

-- name: GetUserUsageSummary :one
SELECT
    COUNT(*) as total_sessions,
    COALESCE(SUM(duration_seconds), 0)::DECIMAL(12,3) as total_duration_seconds,
    COALESCE(SUM(bytes_sent), 0) as total_bytes_sent
FROM transcription_logs
WHERE user_id = sqlc.arg(user_id) AND started_at >= sqlc.arg(start_date) AND started_at < sqlc.arg(end_date);

-- name: GetUserUsageSummaryByStatus :many
SELECT
    status,
    COUNT(*) as count,
    COALESCE(SUM(duration_seconds), 0)::DECIMAL(12,3) as total_duration_seconds
FROM transcription_logs
WHERE user_id = sqlc.arg(user_id) AND started_at >= sqlc.arg(start_date) AND started_at < sqlc.arg(end_date)
GROUP BY status;

-- =====================
-- ADMIN QUERIES
-- =====================

-- name: ListAllTranscriptionLogs :many
SELECT tl.*, u.username, u.email, ak.name as api_key_name
FROM transcription_logs tl
JOIN users u ON tl.user_id = u.id
JOIN api_keys ak ON tl.api_key_id = ak.id
ORDER BY tl.started_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllTranscriptionLogs :one
SELECT COUNT(*) FROM transcription_logs;

-- name: ListAllAPIKeys :many
SELECT ak.*, u.username, u.email
FROM api_keys ak
JOIN users u ON ak.user_id = u.id
ORDER BY ak.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllAPIKeys :one
SELECT COUNT(*) FROM api_keys;

-- name: GetSystemUsageSummary :one
SELECT
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(*) as total_sessions,
    COALESCE(SUM(duration_seconds), 0)::DECIMAL(12,3) as total_duration_seconds,
    COALESCE(SUM(bytes_sent), 0) as total_bytes_sent
FROM transcription_logs
WHERE started_at >= sqlc.arg(start_date) AND started_at < sqlc.arg(end_date);
