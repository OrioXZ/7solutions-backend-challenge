# 7Solutions Backend Challenge

Backend assignment implementation for the 7Solutions Full-Stack Developer position.

## Current Status

Initial Go HTTP server setup with a health-check endpoint.

## Requirements

- Go 1.24+

## Run Locally

```bash
go run ./cmd/api
```

The API runs on port `8080` by default. To use another port:

```bash
HTTP_PORT=9090 go run ./cmd/api
```

## Health Check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## Project Structure

```text
cmd/api/            Application entrypoint
internal/config/    Environment configuration
internal/httpapi/   HTTP routing and handlers
```

The README will be expanded as the API, MongoDB integration, tests, Docker setup, and lottery design are completed.
