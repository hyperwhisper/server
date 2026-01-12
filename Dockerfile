# Build frontend
FROM oven/bun:1.3.5-slim AS frontend
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web .
RUN bun run generate

# Build backend
FROM golang:1.25.5-trixie AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN go build -o hweb .

# Runtime
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /app/hweb .
ENV APP_ENV=prod
EXPOSE 1323
CMD ["./hweb", "serve"]
