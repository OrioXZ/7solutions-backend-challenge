# 7Solutions Backend Challenge

Backend assignment implementation for the 7Solutions Full-Stack Developer position.

## Features

- RESTful user management API written in Go
- MongoDB persistence using the official Go driver
- User registration and login
- JWT authentication signed with HMAC-SHA256 (`HS256`)
- Protected user CRUD endpoints
- Input validation and bcrypt password hashing
- Structured HTTP request logging
- Background user-count logging every 10 seconds
- Graceful shutdown using `context.Context`
- Unit and integration tests using Go's standard `testing` package
- Docker and Docker Compose support for the API and MongoDB

## Lottery Search System

The design-only Lottery Search System proposal is available here:

- [Lottery Search System Design Proposal](./LOTTERY_DESIGN.md)

## Requirements

Choose one of the following setups:

- Docker Desktop, or
- Go 1.23+ and Docker Desktop for MongoDB only

The API listens on `http://localhost:8080` by default.

## Quick Start with Docker Compose

Set a JWT secret containing at least 32 characters.

PowerShell:

```powershell
$env:JWT_SECRET="local-development-secret-change-me-1234567890"
docker compose up --build
```

Bash:

```bash
export JWT_SECRET="local-development-secret-change-me-1234567890"
docker compose up --build
```

Docker Compose starts both services:

- API: `http://localhost:8080`
- MongoDB: `mongodb://localhost:27017`

Verify the application:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "ok",
  "database": "connected"
}
```

Stop the containers:

```bash
docker compose down
```

To also delete the local MongoDB data volume:

```bash
docker compose down -v
```

## Run the API Locally

Start MongoDB only:

```bash
docker compose up -d mongo
```

Download dependencies:

```bash
go mod download
```

Set the JWT secret.

PowerShell:

```powershell
$env:JWT_SECRET="local-development-secret-change-me-1234567890"
```

Bash:

```bash
export JWT_SECRET="local-development-secret-change-me-1234567890"
```

Start the API:

```bash
go run ./cmd/api
```

## Environment Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `HTTP_PORT` | No | `8080` | HTTP server port |
| `MONGODB_URI` | No | `mongodb://localhost:27017` | MongoDB connection URI |
| `MONGODB_DATABASE` | No | `seven_solutions` | MongoDB database name |
| `JWT_SECRET` | Yes | None | HS256 signing secret; minimum 32 characters |

Do not use the example JWT secret outside local development.

## API Endpoints

Public endpoints:

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Check API and MongoDB availability |
| `POST` | `/api/v1/auth/register` | Register a user |
| `POST` | `/api/v1/auth/login` | Authenticate and receive a JWT |

Protected endpoints require `Authorization: Bearer <token>`:

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/users` | Create a user |
| `GET` | `/api/v1/users` | List all users |
| `GET` | `/api/v1/users/{id}` | Fetch a user by ID |
| `PATCH` | `/api/v1/users/{id}` | Update a user's name and/or email |
| `DELETE` | `/api/v1/users/{id}` | Delete a user |

## JWT Guide

### 1. Register a user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

Response: `201 Created`

```json
{
  "data": {
    "id": "64f000000000000000000001",
    "name": "Alice",
    "email": "alice@example.com",
    "created_at": "2026-07-11T01:00:00Z"
  }
}
```

### 2. Login and receive a token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123"
  }'
```

Response: `200 OK`

```json
{
  "data": {
    "access_token": "<jwt-token>",
    "token_type": "Bearer",
    "expires_at": "2026-07-12T01:00:00Z"
  }
}
```

The token is signed with `HS256` and contains:

- `sub`: authenticated user ID
- `iat`: issued-at timestamp
- `exp`: expiration timestamp

Tokens expire after 24 hours.

### 3. Use the token

```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <token>"
```

A missing, malformed, invalid, or expired token returns `401 Unauthorized`.

## Sample User Requests and Responses

### Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob",
    "email": "bob@example.com",
    "password": "password123"
  }'
```

Response: `201 Created`

```json
{
  "data": {
    "id": "64f000000000000000000002",
    "name": "Bob",
    "email": "bob@example.com",
    "created_at": "2026-07-11T01:05:00Z"
  }
}
```

### List Users

```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <jwt-token>"
```

Response: `200 OK`

```json
{
  "data": [
    {
      "id": "64f000000000000000000002",
      "name": "Bob",
      "email": "bob@example.com",
      "created_at": "2026-07-11T01:05:00Z"
    }
  ]
}
```

### Get User by ID

```bash
curl http://localhost:8080/api/v1/users/64f000000000000000000002 \
  -H "Authorization: Bearer <jwt-token>"
```

Response: `200 OK` with the user object inside `data`.

### Update User

At least one of `name` or `email` is required.

```bash
curl -X PATCH http://localhost:8080/api/v1/users/64f000000000000000000002 \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Robert",
    "email": "robert@example.com"
  }'
```

Response: `200 OK`

```json
{
  "data": {
    "id": "64f000000000000000000002",
    "name": "Robert",
    "email": "robert@example.com",
    "created_at": "2026-07-11T01:05:00Z"
  }
}
```

### Delete User

```bash
curl -X DELETE http://localhost:8080/api/v1/users/64f000000000000000000002 \
  -H "Authorization: Bearer <jwt-token>"
```

Response: `204 No Content`

## Error Response Format

Errors use a consistent structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is invalid"
  }
}
```

Common status codes:

| Status | Example |
| --- | --- |
| `400` | Invalid JSON, invalid ID, or validation error |
| `401` | Missing or invalid JWT, or invalid login credentials |
| `404` | User not found |
| `409` | Email already exists |
| `500` | Unexpected server error |
| `503` | MongoDB unavailable during health check |

## Validation Rules

- `name` is required and whitespace is trimmed
- `email` must be valid, is trimmed, and is normalized to lowercase
- `email` is unique through a MongoDB unique index
- `password` must contain at least 8 characters
- passwords are stored only as bcrypt hashes
- update requests must contain `name`, `email`, or both
- path IDs must be valid MongoDB ObjectIDs
- unknown JSON fields and multiple JSON objects are rejected

## Testing

Run all tests:

```bash
go test ./...
```

The tests use Go's standard `testing` package and cover:

- registration and login business logic
- user CRUD business logic
- JWT creation and validation
- authentication middleware
- HTTP handlers and route protection
- request logging middleware
- periodic user-count worker

MongoDB interactions are abstracted through `UserRepository`. Service unit tests replace the real MongoDB repository with mocks or stubs, so MongoDB is not required to run `go test ./...`.

## Postman

Import the collection and local environment files from the `postman/` directory.

The collection runs the following flow in order:

1. Health Check
2. Register
3. Login
4. Create User
5. List Users
6. Get User by ID
7. Update User
8. Delete User

The scripts automatically pass the JWT and created user IDs between requests.

## Runtime Behavior

HTTP requests are logged as structured JSON containing method, path, status, and duration.

A background goroutine queries and logs the total user count every 10 seconds. It stops when the application context is cancelled.

The application handles `Ctrl+C` and `SIGTERM` gracefully by:

1. stopping background work
2. stopping new HTTP requests and allowing active requests to finish
3. closing the MongoDB connection
4. exiting after a maximum shutdown timeout

## Architecture

The project uses a lightweight ports-and-adapters approach:

```text
HTTP request
    -> HTTP handler (inbound adapter)
    -> service (application and business rules)
    -> UserRepository interface (port)
    -> MongoUserRepository (outbound adapter)
    -> MongoDB
```

Project structure:

```text
cmd/api/             Application entrypoint and dependency wiring
internal/background/ Periodic background tasks
internal/config/     Environment configuration
internal/database/   MongoDB connection and lifecycle
internal/domain/     Domain entities
internal/httpapi/    HTTP routing, handlers, and middleware
internal/repository/ Persistence interface and MongoDB adapter
internal/security/   Password hashing and JWT signing/validation
internal/service/    Application business rules
postman/             Importable Postman collection and environment
```

Business logic depends on interfaces rather than the MongoDB driver or HTTP framework, which keeps the service layer testable and replaceable.

## Assumptions and Design Decisions

- MongoDB ObjectIDs are used as user IDs.
- Email matching is case-insensitive because emails are normalized to lowercase before persistence and lookup.
- A unique MongoDB index is the source of truth for preventing duplicate emails under concurrent requests.
- Passwords are never returned by the API and only bcrypt hashes are persisted.
- JWTs are stateless, valid for 24 hours, and are not refreshed or revoked in this assignment scope.
- Registration, login, and health checks are public; all user CRUD operations are protected.
- The user list is not paginated because the assignment requests a simple list operation. Pagination should be added for a production system with large user datasets.
- The API uses the Go standard library HTTP router to keep dependencies small.
- gRPC was not implemented because it is optional; the REST API remains the primary transport.
- The architecture applies ports-and-adapters principles pragmatically without adding unnecessary layers for the assignment size.
