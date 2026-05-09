# go-http-server

Production-ready HTTP server template written in Go, using only the standard library plus one small dependency (`golang.org/x/time/rate` for rate limiting).

## Features

| Feature | Detail |
|---|---|
| **Routing** | Go 1.22+ `net/http` enhanced ServeMux (`METHOD /path/{param}`) |
| **Graceful shutdown** | Drains in-flight requests on SIGINT/SIGTERM |
| **Structured logging** | `log/slog` — JSON in production, text in development |
| **Rate limiting** | Per-IP token bucket, configurable RPS + burst |
| **CORS** | Configurable origins, methods, and headers |
| **Security headers** | OWASP-recommended headers (CSP, HSTS, X-Frame-Options…) |
| **Request ID** | Propagates or generates `X-Request-ID` on every request |
| **Panic recovery** | Logs stack trace, returns 500 (never crashes the process) |
| **TLS** | Optional — TLS 1.2+ with strong cipher suites |
| **Configuration** | Environment variables with validation and defaults |
| **Health probes** | `GET /health` (liveness) and `GET /ready` (readiness) |
| **Docker** | Multi-stage build → `scratch` image (~6 MB) |
| **CI** | GitHub Actions (vet + race detector + build) |

## Project layout

```
.
├── cmd/server/         # Entry point (main.go)
├── internal/
│   ├── config/         # Environment-based configuration + validation
│   ├── handler/        # HTTP handlers and router
│   ├── middleware/     # Chain, CORS, logging, rate limiter, recovery, request ID, security
│   └── server/         # net/http.Server wrapper with graceful shutdown
├── Dockerfile
├── Makefile
└── .env.example
```

## Quick start

```bash
# Run locally (human-readable logs)
make run

# Run tests
make test

# Run tests with race detector
make test-race

# Build binary
make build
./bin/server
```

## Configuration

Copy `.env.example` to `.env.local` and source it, or export variables directly.

| Variable | Default | Description |
|---|---|---|
| `HOST` | `0.0.0.0` | Bind address |
| `PORT` | `8080` | Listen port |
| `READ_TIMEOUT` | `5s` | Max time to read request |
| `WRITE_TIMEOUT` | `10s` | Max time to write response |
| `IDLE_TIMEOUT` | `120s` | Keep-alive idle timeout |
| `SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown drain window |
| `LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` |
| `LOG_FORMAT` | `json` | `json` \| `text` |
| `TLS_CERT_FILE` | _(empty)_ | Path to TLS certificate (PEM) |
| `TLS_KEY_FILE` | _(empty)_ | Path to TLS private key (PEM) |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated origins |
| `CORS_ALLOWED_METHODS` | `GET,POST,…` | Comma-separated methods |
| `CORS_ALLOWED_HEADERS` | see `.env.example` | Comma-separated headers |
| `RATE_LIMIT_RPS` | `100` | Requests per second per IP |
| `RATE_LIMIT_BURST` | `200` | Burst size per IP |
| `ENV` | `development` | `development` \| `staging` \| `production` |

## API endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe |
| `GET` | `/ready` | Readiness probe |
| `GET` | `/api/v1/items` | List items |
| `POST` | `/api/v1/items` | Create item |
| `GET` | `/api/v1/items/{id}` | Get item by ID |
| `DELETE` | `/api/v1/items/{id}` | Delete item |

## Docker

```bash
make docker-build
make docker-run
```

## Adding middleware

Middleware follows the standard `func(http.Handler) http.Handler` signature.
Register it with `middleware.Chain` in `cmd/server/main.go`:

```go
chain := middleware.Chain(
    middleware.RequestID,
    middleware.Logger(logger),
    middleware.Recovery(logger),
    middleware.SecureHeaders,
    middleware.CORS(cfg),
    rateLimiter.Middleware,
    yourMiddleware,   // ← add here
)
```

## Adding routes

Register handlers in `internal/handler/routes.go`:

```go
mux.HandleFunc("GET /api/v1/users/{id}", GetUser)
mux.HandleFunc("POST /api/v1/users", CreateUser)
```
