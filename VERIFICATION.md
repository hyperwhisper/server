# Authentication System Verification Guide

## Prerequisites

1. Ensure PostgreSQL is running
2. Set up environment variables in `.env`:
   ```bash
   DATABASE_URL=postgres://user:password@localhost:5432/hyperwhisper?sslmode=disable
   JWT_SECRET=your-secure-random-secret-at-least-32-characters
   ACCESS_TOKEN_EXPIRY=5       # minutes (default: 5)
   REFRESH_TOKEN_EXPIRY=7      # days (default: 7)
   APP_ENV=dev                 # 'dev' disables password validation
   ```

---

## Step 1: Run Database Migrations

```bash
./tmp/main migrate up
```

Expected: Migration should apply successfully, creating the `users` table.

---

## Step 2: Test API Endpoints

### 2.1 Test Signup

```bash
curl -X POST http://localhost:1323/api/v1/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123!@#",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**Expected Response (201 Created):**
```json
{
  "user": {
    "id": "uuid-here",
    "username": "testuser",
    "email": "test@example.com",
    "first_name": "Test",
    "last_name": "User",
    "user_type": "user",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "access_token": "jwt-token-here",
  "expires_in": 300
}
```

### 2.2 Test Signin with Email

```bash
curl -X POST http://localhost:1323/api/v1/signin \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "test@example.com",
    "password": "Test123!@#"
  }'
```

**Expected Response (200 OK):** Same structure as signup response.

### 2.3 Test Signin with Username

```bash
curl -X POST http://localhost:1323/api/v1/signin \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "testuser",
    "password": "Test123!@#"
  }'
```

**Expected Response (200 OK):** Same structure as signup response.

### 2.4 Test Token Refresh

Use the cookies from signin response, or pass refresh token in body:

```bash
curl -X POST http://localhost:1323/api/v1/token_refresh \
  -H "Content-Type: application/json" \
  -H "Cookie: refresh_token=<refresh_token_from_signin>" \
  -c - -b -
```

Or with body:
```bash
curl -X POST http://localhost:1323/api/v1/token_refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token_here>"}'
```

**Expected Response (200 OK):**
```json
{
  "access_token": "new-jwt-token-here",
  "expires_in": 300
}
```

### 2.5 Test Protected Route (/me)

```bash
curl http://localhost:1323/api/v1/me \
  -H "Authorization: Bearer <access_token_here>"
```

**Expected Response (200 OK):**
```json
{
  "id": "uuid-here",
  "username": "testuser",
  "email": "test@example.com",
  "first_name": "Test",
  "last_name": "User",
  "user_type": "user",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### 2.6 Test Signout

```bash
curl -X POST http://localhost:1323/api/v1/signout
```

**Expected Response (200 OK):**
```json
{
  "message": "signed out successfully"
}
```

---

## Step 3: Test Error Cases

### 3.1 Duplicate Email

```bash
curl -X POST http://localhost:1323/api/v1/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "email": "test@example.com",
    "password": "Test123!@#"
  }'
```

**Expected Response (409 Conflict):**
```json
{
  "error": "email already taken",
  "details": {"email": "this email is already registered"}
}
```

### 3.2 Invalid Credentials

```bash
curl -X POST http://localhost:1323/api/v1/signin \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "test@example.com",
    "password": "wrongpassword"
  }'
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "invalid credentials"
}
```

### 3.3 Missing Token on Protected Route

```bash
curl http://localhost:1323/api/v1/me
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "missing authentication token"
}
```

---

## Step 4: Test Nuxt Frontend

Start the dev server:
```bash
./tmp/main serve --dev
```

### 4.1 Signup Flow
1. Visit `http://localhost:1323/signup`
2. Fill in the form and submit
3. **Expected:** Redirect to `/dashboard` on success

### 4.2 Signin Flow
1. Sign out from dashboard
2. Visit `http://localhost:1323/signin`
3. Enter email/username and password
4. **Expected:** Redirect to `/dashboard` on success

### 4.3 Guest Middleware (Redirect Authenticated Users)
1. While logged in, visit `http://localhost:1323/signin`
2. **Expected:** Redirect to `/dashboard`

3. While logged in, visit `http://localhost:1323/signup`
4. **Expected:** Redirect to `/dashboard`

### 4.4 Auth Middleware (Protect Dashboard)
1. Sign out
2. Visit `http://localhost:1323/dashboard` directly
3. **Expected:** Redirect to `/signin`

### 4.5 Session Persistence
1. Sign in successfully
2. Refresh the page
3. **Expected:** Still logged in (token refreshed from cookie)

---

## Files Created

### Go Server
- `migrations/000001_create_users_table.up.sql`
- `migrations/000001_create_users_table.down.sql`
- `sqlc.yaml`
- `internal/db/queries/users.sql`
- `internal/db/sqlc/` (generated)
- `internal/auth/password.go`
- `internal/auth/jwt.go`
- `internal/auth/middleware.go`
- `internal/handlers/auth.go`

### Nuxt Frontend
- `web/app/types/auth.ts`
- `web/app/composables/useAuth.ts`
- `web/app/middleware/auth.ts`
- `web/app/middleware/guest.ts`
- `web/app/pages/signup.vue`
- `web/app/pages/signin.vue`
- `web/app/pages/dashboard.vue`

### Modified
- `cmd/serve.go`

---

## Environment Variables Reference

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://localhost:5432/hyperwhisper?sslmode=disable` |
| `JWT_SECRET` | Secret for signing JWT tokens | `hyperwhisper-dev-secret-change-in-production` |
| `ACCESS_TOKEN_EXPIRY` | Access token validity in minutes | `5` |
| `REFRESH_TOKEN_EXPIRY` | Refresh token validity in days | `7` |
| `APP_ENV` | Environment mode (`dev` disables password validation) | - |
