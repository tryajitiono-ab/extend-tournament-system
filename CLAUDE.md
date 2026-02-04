# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AccelByte Extend Service Extension app implementing a tournament management service in Go. It exposes a gRPC API (port 6565) with a REST gateway (port 8000) and Prometheus metrics (port 8080). Data is stored in MongoDB (replica set required). Authentication is handled via AccelByte IAM OAuth2 tokens.

## Build & Run Commands

```bash
# Build the Go binary (runs inside Docker)
make build

# Build and run with MongoDB
docker compose up --build

# Regenerate protobuf stubs after editing service.proto
make proto

# Run tests (unit tests only, no external dependencies needed)
go test ./pkg/service/...

# Run a single test
go test ./pkg/service/ -run TestAdvanceWinner

# Lint (golangci-lint with enable-all, see .golangci.yml for disabled linters)
golangci-lint run

# Inside devcontainer, make targets run directly without Docker wrappers
```

## Environment Setup

Copy `.env.template` to `.env` and fill in `AB_CLIENT_ID`, `AB_CLIENT_SECRET`, `AB_BASE_URL`, `AB_NAMESPACE`. Set `PLUGIN_GRPC_SERVER_AUTH_ENABLED=false` for local development without IAM. The Swagger UI is at `http://localhost:8000/tournament/apidocs/`.

## Architecture

### Request Flow

```
Client → REST (gRPC Gateway :8000) → Auth Interceptor → gRPC Server (:6565) → Service Layer → MongoDB
```

### Key Packages

- **`main.go`** — Bootstraps gRPC server, REST gateway, MongoDB connection, IAM client, and OpenTelemetry tracing
- **`pkg/proto/service.proto`** — Single proto file defining all API endpoints, messages, and REST mappings. Edit this first when adding/changing endpoints, then run `make proto`
- **`pkg/pb/`** — Auto-generated from proto. Never edit directly
- **`pkg/server/tournament.go`** — gRPC handler implementations (thin layer delegating to services)
- **`pkg/service/`** — Business logic: `tournament.go` (lifecycle), `participant.go` (registration), `match.go` (bracket generation, result submission, winner advancement)
- **`pkg/storage/`** — MongoDB persistence layer with interfaces (`TournamentStorage`, `ParticipantStorage`, `MatchStorage`)
- **`pkg/common/`** — Auth interceptors, gateway setup, tracing, logging utilities

### Tournament Lifecycle

`DRAFT → ACTIVE → STARTED → COMPLETED` (or `CANCELLED` from any state). Starting a tournament triggers automatic single-elimination bracket generation with bye handling for non-power-of-2 participant counts.

### Frontend

`web/templates/` has HTML templates and `web/static/` has JS/CSS for bracket visualization. Served by the gRPC Gateway's static file handler.

## Proto / Code Generation

The proto file at `pkg/proto/service.proto` defines both gRPC services and REST endpoint mappings via google.api.http annotations. Custom `permission` annotations control AccelByte IAM authorization. After editing, run `make proto` to regenerate `pkg/pb/`.

## File Header Requirement

All Go source files must have the AccelByte copyright header (enforced by golangci-lint's goheader):
```
// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.
```

## Testing

Tests are in `pkg/service/` using mock storage implementations from `pkg/service/mocks/repo_mock.go`. Tests run without MongoDB or any external service. Auth is disabled in test configuration (`.env.test`).

## Deployment

Production deployment uses `extend-helper-cli image-upload` to push the Docker image to the AccelByte registry, then deploy via the AccelByte console.
