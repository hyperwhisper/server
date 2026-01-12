-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, first_name, last_name, user_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByEmailOrUsername :one
SELECT * FROM users WHERE email = $1 OR username = $1;

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: CheckUsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at ASC LIMIT $1 OFFSET $2;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users SET
    username = COALESCE(NULLIF($2, ''), username),
    email = COALESCE(NULLIF($3, ''), email),
    first_name = COALESCE(NULLIF($4, ''), first_name),
    last_name = COALESCE(NULLIF($5, ''), last_name),
    user_type = COALESCE(NULLIF($6, ''), user_type),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- Refresh token queries (only refresh tokens are tracked, access tokens are stateless)

-- name: CreateRefreshToken :one
INSERT INTO tokens (token_jti, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: IsRefreshTokenRevoked :one
SELECT EXISTS(SELECT 1 FROM tokens WHERE token_jti = $1 AND revoked_at IS NOT NULL);

-- name: RevokeRefreshToken :exec
UPDATE tokens SET revoked_at = NOW(), revoked_reason = $2 WHERE token_jti = $1;

-- name: RevokeUserRefreshTokens :exec
UPDATE tokens SET revoked_at = NOW(), revoked_reason = $2 WHERE user_id = $1 AND revoked_at IS NULL;

-- name: GetRefreshTokenByJTI :one
SELECT * FROM tokens WHERE token_jti = $1;

-- name: ListRefreshTokens :many
SELECT * FROM tokens ORDER BY issued_at DESC LIMIT $1 OFFSET $2;

-- name: ListUserRefreshTokens :many
SELECT * FROM tokens WHERE user_id = $1 ORDER BY issued_at DESC LIMIT $2 OFFSET $3;

-- name: CountRefreshTokens :one
SELECT COUNT(*) FROM tokens;

-- name: CountUserRefreshTokens :one
SELECT COUNT(*) FROM tokens WHERE user_id = $1;

-- name: CountActiveRefreshTokens :one
SELECT COUNT(*) FROM tokens WHERE revoked_at IS NULL AND expires_at > NOW();

-- name: ListActiveRefreshTokens :many
SELECT * FROM tokens WHERE revoked_at IS NULL AND expires_at > NOW() ORDER BY issued_at DESC LIMIT $1 OFFSET $2;

-- name: CleanupExpiredRefreshTokens :exec
DELETE FROM tokens WHERE expires_at <= NOW();
