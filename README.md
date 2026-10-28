# Bookstore API

A RESTful API for managing a book catalog, built with Go and PostgreSQL. It provides user registration and login with JWT-based authentication, and full CRUD operations on books for authenticated users.

The project follows a layered architecture with clear separation of concerns, relies primarily on the Go standard library, and is fully covered by unit tests that run without a database.

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Database Migrations](#database-migrations)
- [API Reference](#api-reference)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Security Considerations](#security-considerations)
- [Known Limitations and Roadmap](#known-limitations-and-roadmap)

---

## Features

- **User authentication**: registration, login, and a profile endpoint secured with JWT Bearer tokens.
- **Book management**: create, list (with pagination), retrieve, update, and delete books.
- **Input validation**: field-level validation errors returned in a consistent, client-friendly format.
- **Strict request parsing**: rejects malformed JSON, unknown fields, wrong types, and oversized bodies.
- **Secure password storage**: passwords hashed with bcrypt (cost 12).
- **Structured logging**: every request logged with method, path, status, size, and duration using `log/slog`.
- **Panic recovery**: unexpected panics are converted into clean `500` responses instead of dropped connections.
- **Hardened HTTP server**: read, write, header, and idle timeouts configured to mitigate slow-client attacks.
- **Versioned schema migrations**: SQL migrations managed with golang-migrate.

---

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.22+ |
| HTTP routing | `net/http` standard library (`ServeMux` with method and wildcard patterns) |
| Database | PostgreSQL 16 |
| Database driver | [pgx v5](https://github.com/jackc/pgx) with `pgxpool` connection pooling |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Password hashing | [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) |
| Authentication | [golang-jwt v5](https://github.com/golang-jwt/jwt) (HS256) |
| Configuration | Environment variables, with [godotenv](https://github.com/joho/godotenv) for local development |
| Logging | `log/slog` standard library |
| Local infrastructure | Docker Compose |

---

## Architecture

The application is organized into distinct layers. Each layer has a single responsibility and depends only on the domain model and on interfaces it defines itself.

```
HTTP Request
     │
     ▼
┌─────────────────────────────┐
│ Global middleware           │  Recover → Logging
├─────────────────────────────┤
│ Router (ServeMux)           │  Matches method + path
├─────────────────────────────┤
│ Route middleware            │  RequireAuth (protected routes only)
├─────────────────────────────┤
│ Handler                     │  HTTP concerns: decode request, encode response, map errors to status codes
├─────────────────────────────┤
│ Service                     │  Business rules: normalization, validation, pagination, authentication flow
├─────────────────────────────┤
│ Repository                  │  Data access: parameterized SQL queries via pgx
└─────────────────────────────┘
     │
     ▼
 PostgreSQL
```

### Design principles

- **Consumer-defined interfaces.** Each layer declares the minimal interface it needs from the layer below (for example, `service.BookRepository` and `handler.BookService`). Lower layers expose concrete structs. This keeps packages decoupled and makes every layer testable with lightweight fakes.
- **Explicit dependency injection.** All components receive their dependencies through constructors. `cmd/api/main.go` acts as the composition root and is the only place that wires concrete implementations together. There are no global variables or `init()` side effects.
- **Domain errors.** Infrastructure errors are translated into domain errors at the repository boundary (for example, `pgx.ErrNoRows` becomes `model.ErrNotFound`). Handlers map domain errors to HTTP status codes in a single place, so transport concerns never leak into business logic.
- **Context propagation.** The request context flows from the handler through the service to the database driver, so client disconnects and timeouts cancel in-flight queries.
- **Fail fast.** Configuration is validated and the database connection is verified at startup. The application refuses to start with invalid settings.

---

## Project Structure

```
bookstore-api/
├── cmd/
│   └── api/
│       ├── main.go                 # Composition root: loads config, wires dependencies, starts server
│       └── routes.go               # Route registration and middleware setup
├── internal/
│   ├── auth/
│   │   ├── context.go              # Helpers to store and retrieve the user ID in request context
│   │   ├── password.go             # bcrypt hashing and verification
│   │   └── token.go                # JWT generation and verification
│   ├── config/
│   │   └── config.go               # Environment-based configuration with validation
│   ├── database/
│   │   └── postgres.go             # PostgreSQL connection pool setup
│   ├── handler/
│   │   ├── auth_handler.go         # /auth endpoints
│   │   ├── book_handler.go         # /books endpoints
│   │   ├── health.go               # /health endpoint
│   │   ├── request.go              # JSON decoding, path and query parameter parsing
│   │   └── response.go             # JSON responses and error-to-status mapping
│   ├── middleware/
│   │   ├── auth.go                 # Bearer token authentication
│   │   ├── logging.go              # Structured request logging
│   │   ├── middleware.go           # Middleware type and chaining helper
│   │   └── recovery.go             # Panic recovery
│   ├── model/
│   │   ├── book.go                 # Book entity, input, and list result
│   │   ├── errors.go               # Domain errors and validation error type
│   │   └── user.go                 # User entity, inputs, and auth token
│   ├── repository/
│   │   ├── book_repository.go      # Book persistence
│   │   ├── repository.go           # Shared scanning and error helpers
│   │   └── user_repository.go      # User persistence
│   └── service/
│       ├── auth_service.go         # Registration and login logic
│       └── book_service.go         # Book business rules
├── migrations/                     # Versioned SQL migrations (up and down)
├── .env.example                    # Template for local environment variables
├── docker-compose.yml              # PostgreSQL and migration tooling
├── go.mod
└── go.sum
```

Test files (`*_test.go`) live alongside the code they test.

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.22 or later
- [Docker](https://www.docker.com/) with Docker Compose
- Optional: [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) for running migrations from the host

> The `go` directive in `go.mod` must be 1.22 or later. The method-based route patterns used by the router (for example, `GET /books/{id}`) are only enabled from that version onward.

### 1. Clone the repository

```bash
git clone https://github.com/<your-username>/bookstore-api.git
cd bookstore-api
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Generate a secure random secret for `JWT_SECRET` and add it to `.env`:

```bash
# macOS / Linux
openssl rand -base64 32
```

```powershell
# Windows PowerShell
$b = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b)
[Convert]::ToBase64String($b)
```

See [Configuration](#configuration) for all available variables.

### 3. Start PostgreSQL

```bash
docker compose up -d
docker compose ps    # wait until the db service is "healthy"
```

### 4. Run database migrations

```bash
docker compose run --rm migrate up
```

### 5. Start the server

```bash
go run ./cmd/api
```

Expected output:

```
INFO config loaded env=development port=8080 jwt_ttl=24h0m0s
INFO database connected max_conns=10
INFO server starting addr=:8080
```

### 6. Verify

```bash
curl -i http://localhost:8080/health
```

> **Windows users:** in PowerShell, `curl` is an alias for `Invoke-WebRequest`. Use `curl.exe` to call the real curl binary.

---

## Configuration

All configuration is read from environment variables. In development, values are loaded from a `.env` file if present. Variables already set in the environment take precedence over values in `.env`.

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_ENV` | No | `development` | Environment name (`development`, `production`). |
| `HTTP_PORT` | No | `8080` | Port the HTTP server listens on. |
| `DATABASE_URL` | **Yes** | — | PostgreSQL connection string, e.g. `postgres://user:pass@127.0.0.1:5432/bookstore?sslmode=disable`. |
| `JWT_SECRET` | **Yes** | — | Secret used to sign tokens. Must be at least 32 characters. |
| `JWT_TTL` | No | `24h` | Token lifetime, in Go duration format (`15m`, `1h`, `24h`). |

> Never commit `.env` to version control. It is listed in `.gitignore`.

> On Windows, prefer `127.0.0.1` over `localhost` in `DATABASE_URL`. `localhost` may resolve to IPv6 (`::1`), which Docker Desktop port forwarding does not always handle reliably.

---

## Database Migrations

Migrations are plain SQL files in `migrations/`, named `{version}_{description}.up.sql` and `{version}_{description}.down.sql`.

### Using Docker Compose (no local installation required)

```bash
docker compose run --rm migrate up         # apply all pending migrations
docker compose run --rm migrate down 1     # roll back the last migration
docker compose run --rm migrate version    # show current schema version
```

### Using the golang-migrate CLI

Install:

```bash
go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Run:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Create a new migration:

```bash
migrate create -ext sql -dir migrations -seq add_isbn_to_books
```

> Never edit a migration that has already been applied to a shared environment. Create a new migration instead.

### Schema

**users**

| Column | Type | Constraints |
|---|---|---|
| `id` | `BIGSERIAL` | Primary key |
| `name` | `VARCHAR(100)` | Not null |
| `email` | `VARCHAR(255)` | Not null, unique |
| `password_hash` | `VARCHAR(255)` | Not null |
| `created_at` | `TIMESTAMPTZ` | Not null, default `NOW()` |
| `updated_at` | `TIMESTAMPTZ` | Not null, default `NOW()` |

**books**

| Column | Type | Constraints |
|---|---|---|
| `id` | `BIGSERIAL` | Primary key |
| `title` | `VARCHAR(255)` | Not null |
| `author` | `VARCHAR(255)` | Not null |
| `published_year` | `INTEGER` | Nullable |
| `created_at` | `TIMESTAMPTZ` | Not null, default `NOW()` |
| `updated_at` | `TIMESTAMPTZ` | Not null, default `NOW()` |

---

## API Reference

**Base URL:** `http://localhost:8080`

All request and response bodies use JSON. Timestamps are formatted as RFC 3339.

### Endpoints overview

| Method | Path | Description | Auth |
|---|---|---|---|
| `GET` | `/health` | Health check | No |
| `POST` | `/auth/register` | Register a new user | No |
| `POST` | `/auth/login` | Log in and obtain an access token | No |
| `GET` | `/auth/me` | Get the authenticated user | Yes |
| `GET` | `/books` | List books (paginated) | Yes |
| `POST` | `/books` | Create a book | Yes |
| `GET` | `/books/{id}` | Get a book by ID | Yes |
| `PUT` | `/books/{id}` | Replace a book | Yes |
| `DELETE` | `/books/{id}` | Delete a book | Yes |

### Authentication

Protected endpoints require an access token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Obtain a token from `POST /auth/login`. The scheme name is case-insensitive. Requests with a missing, malformed, invalid, or expired token receive `401 Unauthorized` with a `WWW-Authenticate` header.

---

### Health

#### `GET /health`

Returns `200 OK` when the server is running.

```json
{ "status": "ok" }
```

---

### Auth

#### `POST /auth/register`

Creates a new user account.

**Request**

```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "password": "secret123"
}
```

| Field | Rules |
|---|---|
| `name` | Required. Max 100 characters. Leading and trailing whitespace is trimmed. |
| `email` | Required. Valid email address. Max 255 characters. Normalized to lowercase. |
| `password` | Required. Min 8 characters, max 72 bytes (bcrypt limit). Not trimmed. |

**Response: `201 Created`**

```json
{
  "id": 1,
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "created_at": "2026-10-29T01:30:00+07:00",
  "updated_at": "2026-10-29T01:30:00+07:00"
}
```

**Errors:** `400` malformed request, `409` email already registered, `422` validation failed.

---

#### `POST /auth/login`

Authenticates a user and returns an access token.

**Request**

```json
{
  "email": "budi@example.com",
  "password": "secret123"
}
```

**Response: `200 OK`**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_at": "2026-10-30T01:30:00+07:00"
}
```

The response includes `Cache-Control: no-store`.

**Errors:** `400` malformed request, `401` invalid email or password.

---

#### `GET /auth/me`

Returns the currently authenticated user.

**Response: `200 OK`**

```json
{
  "id": 1,
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "created_at": "2026-10-29T01:30:00+07:00",
  "updated_at": "2026-10-29T01:30:00+07:00"
}
```

**Errors:** `401` missing or invalid token, or the account no longer exists.

---

### Books

All book endpoints require authentication.

#### Book object

| Field | Type | Description |
|---|---|---|
| `id` | integer | Unique identifier, assigned by the server. |
| `title` | string | Book title. |
| `author` | string | Author name. |
| `published_year` | integer or `null` | Year of publication. Always present in responses, `null` when unknown. |
| `created_at` | string | Creation timestamp. |
| `updated_at` | string | Last update timestamp. |

#### Book input rules

| Field | Rules |
|---|---|
| `title` | Required. Max 255 characters. Trimmed. |
| `author` | Required. Max 255 characters. Trimmed. |
| `published_year` | Optional. Between 1 and the current year. |

---

#### `POST /books`

Creates a book.

**Request**

```json
{
  "title": "Laskar Pelangi",
  "author": "Andrea Hirata",
  "published_year": 2005
}
```

**Response: `201 Created`**

Includes a `Location: /books/{id}` header.

```json
{
  "id": 1,
  "title": "Laskar Pelangi",
  "author": "Andrea Hirata",
  "published_year": 2005,
  "created_at": "2026-10-29T01:35:00+07:00",
  "updated_at": "2026-10-29T01:35:00+07:00"
}
```

**Errors:** `400` malformed request or unknown fields, `401` unauthorized, `413` body larger than 1 MB, `422` validation failed.

---

#### `GET /books`

Returns a paginated list of books ordered by ID.

**Query parameters**

| Parameter | Default | Description |
|---|---|---|
| `page` | `1` | Page number. Values below 1 are treated as 1. |
| `page_size` | `10` | Items per page. Values below 1 use the default; values above 100 are capped at 100. |

**Response: `200 OK`**

```json
{
  "books": [
    {
      "id": 1,
      "title": "Laskar Pelangi",
      "author": "Andrea Hirata",
      "published_year": 2005,
      "created_at": "2026-10-29T01:35:00+07:00",
      "updated_at": "2026-10-29T01:35:00+07:00"
    }
  ],
  "page": 1,
  "page_size": 10
}
```

`books` is always an array, empty when there are no results. `page` and `page_size` reflect the values actually applied by the server.

**Errors:** `400` non-numeric query parameter, `401` unauthorized.

---

#### `GET /books/{id}`

Returns a single book.

**Response: `200 OK`** with a book object.

**Errors:** `400` invalid ID, `401` unauthorized, `404` book not found.

---

#### `PUT /books/{id}`

Replaces all editable fields of a book. Fields omitted from the request are cleared (for example, omitting `published_year` sets it to `null`).

**Request**

```json
{
  "title": "Laskar Pelangi (Revised Edition)",
  "author": "Andrea Hirata",
  "published_year": 2008
}
```

**Response: `200 OK`** with the updated book object.

**Errors:** `400` invalid ID or malformed request, `401` unauthorized, `404` book not found, `422` validation failed.

---

#### `DELETE /books/{id}`

Deletes a book.

**Response: `204 No Content`** with an empty body.

**Errors:** `400` invalid ID, `401` unauthorized, `404` book not found.

---

### Example workflow with curl

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Santoso","email":"budi@example.com","password":"secret123"}'

# Log in and store the token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"secret123"}' | jq -r .access_token)

# Create a book
curl -X POST http://localhost:8080/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Laskar Pelangi","author":"Andrea Hirata","published_year":2005}'

# List books
curl "http://localhost:8080/books?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

A Postman collection (`bookstore-api.postman_collection.json`) with automated assertions for every endpoint and error case is also available. Import it into Postman and use **Run collection** to execute the full flow.

---

## Error Handling

All application errors share a consistent JSON format:

```json
{ "error": "resource not found" }
```

Validation errors include per-field details:

```json
{
  "error": "validation failed",
  "fields": {
    "title": "is required",
    "published_year": "must be between 1 and 2026"
  }
}
```

### Status codes

| Status | Meaning |
|---|---|
| `400 Bad Request` | The request could not be parsed: malformed JSON, empty body, unknown fields, wrong field types, multiple JSON objects, invalid ID, or non-numeric query parameter. |
| `401 Unauthorized` | Missing, malformed, invalid, or expired token; or invalid login credentials. |
| `404 Not Found` | The requested resource does not exist. |
| `405 Method Not Allowed` | The path exists but does not support the HTTP method. The `Allow` header lists supported methods. |
| `409 Conflict` | The email address is already registered. |
| `413 Payload Too Large` | The request body exceeds 1 MB. |
| `422 Unprocessable Entity` | The request is well-formed but violates validation rules. |
| `500 Internal Server Error` | An unexpected server-side error. Details are logged but never exposed to the client. |

> `404` responses for unknown paths and `405` responses are generated by the standard library router and are returned as plain text rather than JSON.

---

## Testing

Unit tests cover every layer and use in-memory fakes, so no database or running server is required.

```bash
# Run all tests
go test ./...

# Verbose output for a specific package
go test ./internal/service/ -v

# Static analysis
go vet ./...
```

| Package | Coverage |
|---|---|
| `internal/auth` | Token round-trip; rejection of expired, tampered, foreign-signed, and malformed tokens. |
| `internal/service` | Input normalization and validation, pagination bounds, registration, login, duplicate emails, case-insensitive email handling. |
| `internal/handler` | Request decoding edge cases, error-to-status mapping, response headers, and confirmation that password hashes are never serialized. |
| `internal/middleware` | Bearer token parsing, context propagation, panic recovery, logging transparency, and middleware ordering. |

Authentication tests take a few seconds because bcrypt is intentionally slow.

---

## Security Considerations

- **Password storage.** Passwords are hashed with bcrypt at cost 12 and never stored or logged in plain text. The `password_hash` field is excluded from all JSON output at the model level.
- **User enumeration protection.** Login returns the same error for unknown emails and wrong passwords. When an email is not found, a dummy bcrypt comparison is still performed so that response times do not reveal whether an account exists.
- **Token verification.** Only HS256 is accepted, regardless of the algorithm declared in the token header, which prevents algorithm-substitution attacks. The issuer and expiration claims are required and validated.
- **SQL injection.** All queries use parameterized placeholders. User input is never concatenated into SQL.
- **Mass assignment.** Separate input types (`BookInput`, `RegisterInput`) and strict JSON decoding with unknown-field rejection prevent clients from setting server-controlled fields such as `id` or `password_hash`.
- **Race-safe uniqueness.** Duplicate emails are enforced by a database `UNIQUE` constraint rather than a check-then-insert pattern.
- **Resource limits.** Request bodies are limited to 1 MB, page sizes are capped at 100, and the HTTP server enforces read, write, header, and idle timeouts.
- **Information disclosure.** Internal error details are logged server-side only. Clients receive generic messages for `500` errors. Secrets and query strings are never written to logs.
- **Sensitive responses.** Token responses include `Cache-Control: no-store` to prevent caching by browsers and intermediaries.

---

## Known Limitations and Roadmap

- **Token revocation.** Access tokens are stateless and remain valid until they expire. A production deployment should use short-lived access tokens combined with revocable refresh tokens.
- **Graceful shutdown.** The server does not yet drain in-flight requests on `SIGINT`/`SIGTERM`.
- **Partial updates.** Only full replacement via `PUT` is supported. `PATCH` would require pointer fields to distinguish omitted values from empty ones.
- **Pagination metadata.** List responses do not include a total item count.
- **Router responses.** `404` and `405` responses from the router are plain text rather than JSON.
- **Rate limiting.** No rate limiting is applied to authentication endpoints.
- **Email verification.** Email ownership is not verified after registration.
- **Authorization.** All authenticated users can manage all books. There is no ownership or role-based access control.