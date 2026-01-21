<p align="center">
  <img src="web/public/logo.svg" alt="HyperWhisper Logo" width="128" height="128">
</p>

<h1 align="center">HyperWhisper Server</h1>

<p align="center">
  A full-stack speech-to-text transcription service built with Go/Echo and Nuxt 4, for real-time audio transcription.
</p>

---

## Features

- **JWT Authentication** - Stateless access tokens (5 min) + stateful refresh tokens (7 days)
- **API Key Management** - Generate and manage API keys for programmatic access
- **Trial System** - Device fingerprint-based trial keys with configurable limits
- **Usage Tracking** - Comprehensive transcription logging and analytics
- **Admin Dashboard** - User management, token management, and system analytics

## Tech Stack

**Backend:**
- Go 1.25+
- Echo v4 (web framework)
- PostgreSQL + sqlc (type-safe SQL)
- JWT authentication (golang-jwt/jwt/v5)

**Frontend:**
- Nuxt 4 (Vue 3, SPA mode)
- TypeScript
- Tailwind CSS 4
- shadcn-vue components

## Project Structure

```
server/
├── cmd/
│   ├── serve.go          # Server startup & routes
│   └── migrate.go        # Database migration CLI
├── internal/
│   ├── auth/             # JWT, middleware, password hashing
│   ├── db/
│   │   ├── queries/      # SQL query definitions
│   │   └── sqlc/         # Generated Go code
│   └── handlers/         # HTTP handlers
├── migrations/           # Database migrations
├── web/                  # Nuxt frontend
│   ├── app/
│   │   ├── components/   # Vue components
│   │   ├── pages/        # File-based routing
│   │   ├── composables/  # useAuth.ts
│   │   └── middleware/   # Route guards
│   └── dist/             # Built static files
├── server.go             # CLI entry point
├── sqlc.yaml             # sqlc configuration
├── Dockerfile            # Production build
└── docker-compose.yml    # Development environment
```

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for local development)
- Bun (for frontend development)

### Development

```bash
# Start the development environment
docker-compose up

# Services:
# - API: http://localhost:1323
# - Nuxt dev: http://localhost:3000 (proxied through API)
# - PostgreSQL: localhost:5432
```

### Database Migrations

```bash
# Run pending migrations
go run server.go migrate up

# Revert migrations
go run server.go migrate down

# Check current version
go run server.go migrate version

# Go to specific version
go run server.go migrate goto 3
```

### Production Build

```bash
# Build frontend
cd web && bun run generate

# Build Go binary
go build -o hweb .

# Run
./hweb serve
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://localhost:5432/hyperwhisper?sslmode=disable` |
| `JWT_SECRET` | JWT signing secret | `hyperwhisper-dev-secret-change-in-production` |
| `ACCESS_TOKEN_EXPIRY` | Access token expiry (minutes) | `5` |
| `REFRESH_TOKEN_EXPIRY` | Refresh token expiry (days) | `7` |
| `APP_ENV` | Environment (`dev` or `prod`) | `dev` |


## Authentication Flow

1. User signs in, receives access token (5 min) + refresh token (7 days)
2. Access token stored in memory, refresh token in HTTP-only cookie
3. On access token expiry, client calls `/token_refresh`
4. Refresh tokens are single-use and tracked in database

## License

[AGPLv3](LICENSE)
