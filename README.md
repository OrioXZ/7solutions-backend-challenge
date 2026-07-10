# 7Solutions Backend Challenge

Backend assignment implementation for the 7Solutions Full-Stack Developer position.

## Current Status

Go HTTP server connected to MongoDB with a database-aware health check.

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

Start the API:

```bash
go run ./cmd/api
```

The API runs on port `8080` by default.

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
```

## Project Structure

```text
cmd/api/             Application entrypoint
internal/config/     Environment configuration
internal/database/   MongoDB connection and lifecycle
internal/httpapi/    HTTP routing and handlers
```

The README will be expanded as the user API, authentication, tests, Docker setup, and lottery design are completed.
