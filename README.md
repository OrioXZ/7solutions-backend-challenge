# 7Solutions Backend Challenge

Backend assignment implementation for the 7Solutions Full-Stack Developer position.

## Current Status

The API currently supports MongoDB health checks, user registration, and JWT login using HS256.

## Requirements

- Go 1.23+
- Docker Desktop

## Run Locally

Start MongoDB:

```bash
docker compose up -d mongo
```

Download Go dependencies:

```bash
go mod tidy
```

Set a JWT secret with at least 32 characters.

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

The API runs on port `8080` by default.

## Available Endpoints

```text
GET  /health
POST /api/v1/auth/register
POST /api/v1/auth/login
```

## Health Check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok","database":"connected"}
```

## Environment Variables

```text
HTTP_PORT=8080
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=seven_solutions
JWT_SECRET=local-development-secret-change-me-1234567890
```

`JWT_SECRET` is required and must contain at least 32 characters. Replace the example value outside local development.

## Project Structure

```text
cmd/api/             Application entrypoint
internal/config/     Environment configuration
internal/database/   MongoDB connection and lifecycle
internal/domain/     Domain entities
internal/httpapi/    HTTP routing and handlers
internal/repository/ Persistence abstraction and MongoDB implementation
internal/security/   Password hashing and JWT signing
internal/service/    Application business logic
postman/             Importable collection and local environment
```

The README will be expanded as the protected user API, middleware, concurrency task, Docker setup, and lottery design are completed.
