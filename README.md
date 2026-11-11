# Bookstore API

A RESTful API for managing a book catalog, built with Go and PostgreSQL. It provides user registration and login with JWT-based authentication, and full CRUD operations on books for authenticated users.

The project structure follows [yonisaka/go-boilerplate](https://github.com/yonisaka/go-boilerplate): a CLI entry point, clean-architecture layers (handler, usecase, repository), lazy dependency injection, a shared JSON response envelope, and translated response messages.

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Deployment](#deployment)
- [Configuration](#configuration)
- [Database Migrations](#database-migrations)
- [API Reference](#api-reference)
- [Response Format](#response-format)
- [Testing](#testing)
- [Security Considerations](#security-considerations)
- [Known Limitations and Roadmap](#known-limitations-and-roadmap)

---

## Features

- **User authentication**: registration, login, and a profile endpoint secured with JWT Bearer tokens.
- **Book management**: create, list (with pagination), retrieve, update, and delete books.
- **Consistent responses**: every response, including errors and unknown routes, uses the same `{code, message, data, errors, meta}` envelope.
- **Translated messages**: response messages in English or Indonesian based on `Accept-Language`.
- **Input validation**: field-level validation errors in the `errors` object.
- **Strict request parsing**: rejects malformed JSON, unknown fields, wrong types, and oversized bodies.
- **Secure password storage**: passwords hashed with bcrypt (cost 12).
- **Structured logging**: JSON logs via logrus, with an access log entry for every request.
- **Panic recovery**: unexpected panics become clean `500` responses.
- **Graceful shutdown**: on `SIGINT`/`SIGTERM` the server drains in-flight requests, then closes the database pools.
- **Master/slave databases**: writes go to the master pool, reads to the slave pool. Both can point to the same database.
- **Versioned schema migrations**: SQL migrations managed with golang-migrate.

---

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.26 |
| CLI | [urfave/cli v2](https://github.com/urfave/cli) |
| HTTP routing | [gorilla/mux](https://github.com/gorilla/mux), wrapped with OpenTracing spans (`pkg/routerkit`) |
| Database | PostgreSQL 16 |
| Database driver | [pgx v5](https://github.com/jackc/pgx) with `pgxpool` |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Password hashing | [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) |
| Authentication | [golang-jwt v5](https://github.com/golang-jwt/jwt) (HS256) |
| Configuration | Environment variables, with [godotenv](https://github.com/joho/godotenv) for local development |
| Logging | [logrus](https://github.com/sirupsen/logrus) (`pkg/logger`) |
| Testing | [testify](https://github.com/stretchr/testify) and [uber-go/mock](https://github.com/uber-go/mock) |
| Local infrastructure | Docker Compose |

---

## Architecture

```
HTTP Request
     │
     ▼
┌─────────────────────────────┐
│ routerkit (gorilla/mux)     │  Matches method + path, starts a tracing span
├─────────────────────────────┤
│ router.handle               │  Panic recovery, access log, language, 1 MB body limit
├─────────────────────────────┤
│ middleware.Handle           │  Auth (protected routes only)
├─────────────────────────────┤
│ httpGateway → Handler       │  Decode request, call usecase, return dto.HTTPResponse
├─────────────────────────────┤
│ Usecase                     │  Business rules: normalization, validation, pagination, auth flow
├─────────────────────────────┤
│ Repository (datastore)      │  Parameterized SQL via pgx; master for writes, slave for reads
└─────────────────────────────┘
     │
     ▼
 PostgreSQL
```

### Design principles

- **Handlers return data, not writes.** Each endpoint is a `httphandler.Handler` whose `Handle(*http.Request)` returns a `dto.HTTPResponse`. `router.handle` is the only place that writes to the `http.ResponseWriter`, so headers, status codes and the JSON envelope are handled consistently.
- **Interfaces at the boundaries.** Repository interfaces live in `internal/entities/repository`, usecase interfaces in `internal/usecases`. Mocks are generated next to them with `mockgen`.
- **Lazy dependency injection.** `internal/di` builds every dependency through `Get*` functions. Shared resources (config, database pools, token manager) are created once with `sync.Once`. Database pools register a closer that runs on shutdown.
- **Domain errors.** Repositories translate driver errors into `entities` errors (for example, `pgx.ErrNoRows` becomes `entities.ErrNotFound`). Handlers map them to status codes in one place (`httphandler/response.go`).
- **Fail fast.** Configuration is validated and both database pools are pinged at startup.

---

## Project Structure

```
bookstore-api/
├── main.go                         # Entry point: loads config and runs the CLI
├── cmd/                            # CLI commands (http:start)
├── config/
│   ├── config.go                   # Environment-based configuration with validation
│   └── msg.yaml                    # Response messages per status code and language (embedded)
├── internal/
│   ├── adapters/httphandler/       # One Handler per endpoint, request decoding, error mapping
│   ├── consts/                     # Context keys, header names, log event names
│   ├── di/                         # Dependency wiring, routes, response writer
│   ├── dto/                        # Request/response objects and the HTTPResponse envelope
│   ├── entities/                   # Domain models and errors
│   │   └── repository/             # Repository interfaces and generated mocks
│   ├── infrastructure/datastore/   # PostgreSQL repositories and connection pools
│   ├── middleware/                 # Request middleware (Bearer auth)
│   ├── server/                     # HTTP server with graceful shutdown (gRPC stub)
│   └── usecases/                   # Business logic, interfaces and generated mocks
├── pkg/
│   ├── auth/                       # JWT and bcrypt helpers
│   ├── di/                         # Closer registry used during shutdown
│   ├── logger/                     # logrus wrapper
│   ├── msg/                        # Status code → translated message lookup
│   ├── routerkit/                  # gorilla/mux with OpenTracing
│   ├── locales/, util/, file/      # Helpers from the boilerplate
│   └── kvs/, ws/                   # Redis client and WebSocket hub (not wired into routes)
├── migrations/                     # Versioned SQL migrations (up and down)
├── development/docker-compose.yml  # PostgreSQL, migrations, optional API container
├── scripts/lint.sh, .golangci.yml  # Linting
├── .github/workflows/              # CI (tests; lint workflow available)
├── Makefile
└── .env.example
```

Test files (`*_test.go`) live alongside the code they test.

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.26 or later
- [Docker](https://www.docker.com/) with Docker Compose
- Optional: `make` (the commands below show both forms)

### 1. Configure environment variables

```bash
cp .env.example .env
```

Generate a secure random secret for `JWT_SECRET` and add it to `.env`:

```bash
openssl rand -base64 48
```

```powershell
# Windows PowerShell
$b = New-Object byte[] 48
[System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b)
[Convert]::ToBase64String($b)
```

### 2. Start PostgreSQL and run migrations

```bash
docker compose -f development/docker-compose.yml up -d db
docker compose -f development/docker-compose.yml --profile tools run --rm migrate up
```

Or `make db-up` and `make migrate-up`.

### 3. Start the server

```bash
go run . http:start
```

Or `make http`. Logs are JSON, one line per event:

```
{"level":"info","message":"starting bookstore-api on port 8080",...}
```

Press `Ctrl+C` to stop. The server finishes in-flight requests (up to 10 seconds) before exiting.

### 4. Verify

```bash
curl -i http://localhost:8080/liveness
```

> **Windows users:** in PowerShell, `curl` is an alias for `Invoke-WebRequest`. Use `curl.exe` to call the real curl binary.

### Build a binary

```bash
go build -o bin/bookstore-api .
./bin/bookstore-api http:start
```

The `.env` file is read from the current working directory. Variables already set in the environment take precedence.

---

## Deployment

`deploy/` contains a production Docker Compose stack: PostgreSQL (not exposed to the internet) and the API image built from `Dockerfile` (distroless, non-root). Neither publishes a port.

The public entry point is a separate reverse proxy project (Caddy on ports 80/443) that is shared by every app on the server. Apps join the external Docker network `web` and the proxy reaches them by network alias; this API is `bookstore-api:8080`. Stopping or redeploying one app never takes the proxy down for the others.

```
internet ──► proxy (Caddy :80/:443) ──web network──► bookstore-api:8080 ──► db
                                     └─────────────► other-app:3000
```

On the server, from the repository root:

```bash
docker network create web                # once per server
cp deploy/.env.example deploy/.env       # fill in passwords and JWT_SECRET
docker compose -f deploy/docker-compose.yml up -d --build
docker compose -f deploy/docker-compose.yml --profile tools run --rm migrate up
```

The API container reports its own health (`bookstore-api healthcheck`, which calls `/liveness`), visible in `docker compose -f deploy/docker-compose.yml ps`.

A minimal proxy project (for example in `/opt/proxy`) needs a `Caddyfile`:

```
:80 {
	reverse_proxy bookstore-api:8080
}
```

and a `docker-compose.yml` running `caddy:2.10-alpine` with ports `80:80` and `443:443`, the `Caddyfile` mounted at `/etc/caddy/Caddyfile`, volumes for `/data` and `/config`, and the external `web` network. To enable HTTPS, replace `:80` with a domain whose A record points to the server and restart Caddy; it obtains and renews certificates automatically.

### Continuous deployment

Every push to `master` runs `.github/workflows/deploy.yml`: the tests run first, then GitHub Actions connects to the server over SSH and runs `deploy/deploy.sh` (pull, build, migrate, restart, wait for the container health check). Pull requests to `master` only run the tests.

The SSH key used by GitHub Actions is pinned on the server to that script with a forced command, so a leaked key can only trigger a deploy:

```
restrict,command="/opt/bookstore-api/deploy/deploy.sh" ssh-ed25519 AAAA... github-actions-deploy
```

Required repository secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY` (private key), and `VPS_KNOWN_HOSTS` (output of `ssh-keyscan <host>`, or the server's `/etc/ssh/ssh_host_*_key.pub` prefixed with the host).

Change files under `deploy/` through commits, not by editing them on the server, or `git pull --ff-only` in the deploy script may fail.

---

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_NAME` | No | `bookstore-api` | Service name in logs and traces. |
| `APP_ENV` | No | `development` | Environment name in logs. |
| `APP_PORT` | No | `8080` | Port the HTTP server listens on. |
| `APP_DEFAULT_LANG` | No | `en` | Message language when `Accept-Language` is missing or unsupported (`en`, `id`). |
| `APP_LOG_LEVEL` | No | `info` | logrus level (`debug`, `info`, `warn`, `error`). |
| `APP_READ_TIMEOUT` | No | `10` | HTTP read timeout in seconds. |
| `APP_WRITE_TIMEOUT` | No | `15` | HTTP write timeout in seconds. |
| `POSTGRES_HOST_MASTER` | No | `localhost` | Master database host (writes). |
| `POSTGRES_PORT_MASTER` | No | `5432` | Master database port. |
| `POSTGRES_USER_MASTER` | No | `postgres` | Master database user. |
| `POSTGRES_PASSWORD_MASTER` | No | `postgres` | Master database password. |
| `POSTGRES_DB_MASTER` | No | `postgres` | Master database name. |
| `POSTGRES_SSLMODE_MASTER` | No | `disable` | `disable` only for local databases; use `require` or stricter over a network. |
| `POSTGRES_*_SLAVE` | No | same as master | Same settings for the read replica. Point them at the master if you have no replica. |
| `JWT_SECRET` | **Yes** | — | Secret used to sign tokens. At least 32 characters. |
| `JWT_TTL` | No | `24h` | Token lifetime in Go duration format (`15m`, `1h`, `24h`). |

`WS_*` variables configure the unused WebSocket package (`pkg/ws`); see `config/config.go`.

> Never commit `.env`. It is listed in `.gitignore`.

> On Windows, prefer `127.0.0.1` over `localhost` for database hosts. `localhost` may resolve to IPv6 (`::1`), which Docker Desktop port forwarding does not always handle reliably.

---

## Database Migrations

Migrations are plain SQL files in `migrations/`, named `{version}_{description}.up.sql` and `{version}_{description}.down.sql`.

```bash
make migrate-up      # apply all pending migrations
make migrate-down    # roll back the last migration
```

Without `make`, use `docker compose -f development/docker-compose.yml --profile tools run --rm migrate up` (or `down 1`, `version`).

> Never edit a migration that has already been applied to a shared environment. Create a new migration instead.

### Schema

**users**: `id BIGSERIAL PK`, `name VARCHAR(100)`, `email VARCHAR(255) UNIQUE`, `password_hash VARCHAR(255)`, `created_at`, `updated_at`.

**books**: `id BIGSERIAL PK`, `title VARCHAR(255)`, `author VARCHAR(255)`, `published_year INTEGER NULL`, `created_at`, `updated_at`.

---

## API Reference

**Base URL:** `http://localhost:8080`

### Endpoints overview

| Method | Path | Description | Auth |
|---|---|---|---|
| `GET` | `/liveness` | Health check (pings both database pools) | No |
| `POST` | `/auth/register` | Register a new user | No |
| `POST` | `/auth/login` | Log in and obtain an access token | No |
| `GET` | `/auth/me` | Get the authenticated user | Yes |
| `GET` | `/books` | List books (paginated) | Yes |
| `POST` | `/books` | Create a book | Yes |
| `GET` | `/books/{id}` | Get a book by ID | Yes |
| `PUT` | `/books/{id}` | Replace a book | Yes |
| `DELETE` | `/books/{id}` | Delete a book | Yes |

Protected endpoints require `Authorization: Bearer <access_token>`. The scheme name is case-insensitive. A missing, malformed, invalid, or expired token returns `401` with a `WWW-Authenticate` header.

### `GET /liveness`

`200` with `"message": "Service is healthy"` and `"data": {"version": "<commit>"}`, or `503` when a database is unreachable. The version is the short Git commit the image was built from (`dev` for local builds), so it shows which commit is deployed.

### `POST /auth/register`

```json
{ "name": "Budi Santoso", "email": "budi@example.com", "password": "secret123" }
```

| Field | Rules |
|---|---|
| `name` | Required. Max 100 characters. Trimmed. |
| `email` | Required. Valid address, max 255 characters. Lowercased. |
| `password` | Required. Min 8 characters, max 72 bytes (bcrypt limit). Not trimmed. |

`201` with the user in `data`. Errors: `400`, `409` email already registered, `422`.

### `POST /auth/login`

```json
{ "email": "budi@example.com", "password": "secret123" }
```

`200` with `data: {access_token, token_type, expires_at}` and `Cache-Control: no-store`. Errors: `400`, `401` invalid email or password.

### `GET /auth/me`

`200` with the user in `data`. `401` if the token is invalid or the account no longer exists.

### Books

Book object: `id`, `title`, `author`, `published_year` (integer or `null`), `created_at`, `updated_at`.

| Field | Rules |
|---|---|
| `title` | Required. Max 255 characters. Trimmed. |
| `author` | Required. Max 255 characters. Trimmed. |
| `published_year` | Optional. Between 1 and the current year. |

- `POST /books`: `201` with the book in `data` and a `Location: /books/{id}` header. Errors: `400`, `413` body over 1 MB, `422`.
- `GET /books?page=1&page_size=10`: `200`, books in `data` (always an array), `meta: {page, per_page}`. `page` below 1 becomes 1; `page_size` defaults to 10 and is capped at 100. Error: `400` non-numeric parameter.
- `GET /books/{id}`: `200` with the book. Errors: `400` invalid ID, `404`.
- `PUT /books/{id}`: replaces all editable fields (omitting `published_year` sets it to `null`). `200` with the updated book. Errors: `400`, `404`, `422`.
- `DELETE /books/{id}`: `204 No Content` with an empty body. Errors: `400`, `404`.

### Example workflow with curl

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Santoso","email":"budi@example.com","password":"secret123"}'

TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"secret123"}' | jq -r .data.access_token)

curl -X POST http://localhost:8080/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Laskar Pelangi","author":"Andrea Hirata","published_year":2005}'

curl "http://localhost:8080/books?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" -H "Accept-Language: id"
```

---

## Response Format

Every JSON response uses the same envelope. Empty fields are omitted.

```json
{
  "code": 200,
  "message": "OK",
  "data": [ { "id": 1, "title": "Laskar Pelangi", "author": "Andrea Hirata", "published_year": 2005 } ],
  "meta": { "page": 1, "per_page": 10 }
}
```

Validation errors put per-field details in `errors`:

```json
{
  "code": 422,
  "message": "Validation failed",
  "errors": { "title": "is required", "published_year": "must be between 1 and 2026" }
}
```

`message` comes from `config/msg.yaml` in the language from `Accept-Language` (`en` or `id`, default `APP_DEFAULT_LANG`), unless the handler sets a specific message such as `"invalid email or password"`.

| Status | Meaning |
|---|---|
| `400 Bad Request` | Malformed JSON, empty body, unknown fields, wrong types, multiple JSON objects, invalid ID, or non-numeric query parameter. |
| `401 Unauthorized` | Missing, malformed, invalid, or expired token; or invalid login credentials. |
| `404 Not Found` | The resource or path does not exist. |
| `405 Method Not Allowed` | The path exists but does not support the HTTP method. |
| `409 Conflict` | The email address is already registered. |
| `413 Payload Too Large` | The request body exceeds 1 MB. |
| `422 Unprocessable Entity` | The request is well-formed but violates validation rules. |
| `500 Internal Server Error` | Unexpected error. Details are logged, never returned. |
| `503 Service Unavailable` | `/liveness` could not reach a database. |

---

## Testing

Unit tests use generated mocks, so no database or running server is required.

```bash
go test ./...          # all tests
make test              # internal packages with -race and an HTML coverage report (Linux/macOS)
make mock              # regenerate mocks after changing an interface (needs: make install-gomock)
go vet ./...
```

| Package | Coverage |
|---|---|
| `config` | Validation rules; every status code used by the API has an `en` and `id` message. |
| `pkg/auth` | Token round-trip; rejection of expired, tampered, foreign-signed, and malformed tokens. |
| `internal/usecases` | Normalization and validation, pagination bounds, registration, login, duplicate emails, case-insensitive email handling. |
| `internal/adapters/httphandler` | Request decoding edge cases, error-to-status mapping, headers, password hashes never serialized. |
| `internal/middleware` | Bearer token parsing, context propagation, middleware ordering. |
| `internal/di` | JSON envelope, translated messages, headers, `204` without body, panic recovery, middleware errors, JSON `404`/`405`. |

`pkg/kvs/redis` tests are skipped when no Redis server is running on `127.0.0.1:6379`.

---

## Security Considerations

- **Password storage.** bcrypt at cost 12. `password_hash` is excluded from all JSON output.
- **User enumeration protection.** Login returns the same error for unknown emails and wrong passwords, and runs a dummy bcrypt comparison for unknown emails so response times match.
- **Token verification.** Only HS256 is accepted, preventing algorithm-substitution attacks. Issuer and expiration are required.
- **SQL injection.** All queries use parameterized placeholders.
- **Mass assignment.** Separate request DTOs and unknown-field rejection prevent clients from setting `id` or `password_hash`.
- **Race-safe uniqueness.** Duplicate emails are enforced by a database `UNIQUE` constraint.
- **Resource limits.** 1 MB request bodies, page size capped at 100, and read, write, header, and idle timeouts.
- **Information disclosure.** Internal errors are logged server-side only. Database passwords are never logged; connection strings are built with proper escaping.
- **Sensitive responses.** Token responses include `Cache-Control: no-store`.

---

## Known Limitations and Roadmap

- **Token revocation.** Access tokens remain valid until they expire. Production should use short-lived access tokens with revocable refresh tokens.
- **Partial updates.** Only full replacement via `PUT` is supported.
- **Pagination metadata.** List responses do not include a total item count (`dto.MetaDataWithCount` is available).
- **Rate limiting.** No rate limiting on authentication endpoints.
- **Email verification.** Email ownership is not verified after registration.
- **Authorization.** All authenticated users can manage all books. There is no ownership or role-based access control.
- **Linting in CI.** `.golangci.yml` uses the golangci-lint v1 format, which does not support Go 1.26. The lint workflow is present but not wired into `ci.yml` until the config is migrated to golangci-lint v2.
- **Unused boilerplate packages.** `pkg/ws`, `pkg/kvs` and the gRPC server stub come from the boilerplate and are not used by any route.
