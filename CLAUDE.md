# CLAUDE.md - Project Guidelines for hyperwhisper.dev

## Critical Rules

1. **No command execution**: Never execute `bun`, `go`, or `sqlc` commands directly. Always ask the user to run these commands manually.

2. **Static site only**: Nuxt is used exclusively for static site generation (SPA mode). No server-side rendering. All page data is hydrated via API calls to the Go Echo server.

3. **API prefix**: All Go API routes must start with `/api/v1/`

4. **Admin routes**:
   - Go backend: `/api/v1/admin/*`
   - Nuxt frontend: `/admin/*`

5. **UI components**: Always use shadcn-vue components with Tailwind CSS 4 for creating UI. Components are located in `web/app/components/ui/`.

---

## Project Overview

hyperwhisper.dev is a full-stack application with:
- **Backend**: Go + Echo framework (REST API)
- **Frontend**: Nuxt 4 + Vue 3 (Static SPA)
- **Database**: PostgreSQL with sqlc for type-safe queries
- **Auth**: JWT tokens (stateless access + stateful refresh tokens)

---

## Directory Structure

```
hyperwhisper.dev/
├── cmd/
│   └── serve.go              # Server startup & route configuration
├── internal/
│   ├── auth/
│   │   ├── jwt.go            # JWT generation & validation
│   │   ├── middleware.go     # Auth middleware (JWT + Admin)
│   │   └── password.go       # Password hashing (bcrypt)
│   ├── db/
│   │   ├── db.go             # PostgreSQL connection
│   │   ├── queries/          # SQL queries (sqlc input)
│   │   └── sqlc/             # Generated Go code
│   └── handlers/
│       ├── auth.go           # Auth endpoints
│       └── admin.go          # Admin endpoints
├── migrations/               # Database migrations (golang-migrate)
├── web/                      # Nuxt frontend
│   ├── app/
│   │   ├── components/
│   │   │   ├── ui/           # shadcn-vue components
│   │   │   ├── AppNavbar.vue
│   │   │   ├── AdminSidebar.vue
│   │   │   └── ThemeToggle.vue
│   │   ├── pages/            # File-based routing
│   │   ├── composables/
│   │   │   └── useAuth.ts    # Auth state management
│   │   ├── middleware/       # Route guards
│   │   └── types/            # TypeScript types
│   ├── dist/                 # Generated static files (embedded in prod)
│   ├── embed.go              # Go embed for static files
│   └── nuxt.config.ts
├── server.go                 # CLI entry point
├── sqlc.yaml                 # sqlc configuration
├── Dockerfile                # Production multi-stage build
└── docker-compose.yml        # Development environment
```

---

## Tech Stack

### Backend
- Go 1.25.5
- Echo v4 (web framework)
- sqlc (type-safe SQL)
- PostgreSQL
- JWT (golang-jwt/jwt/v5)
- bcrypt for passwords

### Frontend
- Nuxt 4 (SSR disabled, static mode)
- Vue 3
- TypeScript
- Tailwind CSS 4
- shadcn-vue (Reka UI based)
- Lucide icons

---

## Authentication System

### Token Types
- **Access Token**: 5 min expiry, stateless, stored in-memory + cookie backup
- **Refresh Token**: 7 days expiry, tracked in DB, single-use, HTTP-only cookie

### Key Endpoints
- `POST /api/v1/signup` - Register
- `POST /api/v1/signin` - Login
- `POST /api/v1/token_refresh` - Refresh tokens
- `POST /api/v1/signout` - Logout
- `GET /api/v1/me` - Current user (protected)

### Admin Endpoints
- `GET /api/v1/admin/users` - List users
- `POST /api/v1/admin/users` - Create user
- `DELETE /api/v1/admin/users/:id` - Delete user
- `GET /api/v1/admin/tokens` - List refresh tokens
- `POST /api/v1/admin/tokens/revoke` - Revoke token
- `POST /api/v1/admin/tokens/cleanup` - Cleanup expired

---

## Frontend Auth Flow

The `useAuth` composable manages authentication:
- On page load: Try existing access_token cookie via `/api/v1/me`
- If cookie auth succeeds: Set `accessToken.value = 'cookie'` (placeholder)
- If cookie fails: Try refresh token to get new access token
- Auto-refresh scheduled 30 seconds before expiry

### Middleware
- `auth.ts` - Protects routes, redirects to `/signin?redirect=<path>`
- `admin.ts` - Requires admin role
- `guest.ts` - Redirects authenticated users away from signin/signup

---

## Development

```bash
# Start dev environment (ask user to run)
docker-compose up

# The app runs on:
# - API: http://localhost:1323
# - Nuxt dev: http://localhost:3000 (proxied through API)
```

### Adding new SQL queries
1. Edit `internal/db/queries/users.sql`
2. Ask user to run: `sqlc generate`

### Adding new pages
- Create Vue file in `web/app/pages/`
- Use shadcn-vue components from `web/app/components/ui/`
- Apply appropriate middleware (`auth`, `admin`, or `guest`)

---

## Production Build

```bash
# Build frontend (ask user to run)
cd web && bun run generate

# Build Go binary (ask user to run)
go build -o hyperwhisper .

# Run production server
./hyperwhisper serve
```

The production binary embeds the static frontend files.

---

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing secret (change in production!)
- `ACCESS_TOKEN_EXPIRY` - Minutes (default: 5)
- `REFRESH_TOKEN_EXPIRY` - Days (default: 7)
- `APP_ENV` - 'dev' or 'prod'

---

## Code Patterns

### API Response Format
```go
// Success
c.JSON(http.StatusOK, data)

// Error
c.JSON(http.StatusBadRequest, ErrorResponse{
    Error:   "error message",
    Details: map[string]string{"field": "error"},
})
```

### Frontend API Calls
```typescript
const response = await $fetch<ResponseType>('/api/v1/endpoint', {
  method: 'POST',
  body: payload,
  credentials: 'include', // Always include for cookie auth
})
```

### Protected Page Setup
```vue
<script setup lang="ts">
definePageMeta({
  middleware: 'auth' // or 'admin' for admin pages
})
</script>
```
