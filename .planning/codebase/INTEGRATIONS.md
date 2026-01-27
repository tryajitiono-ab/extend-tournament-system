# External Integrations

**Analysis Date:** 2026-01-27

## APIs & External Services

**Gaming Platform:**
- AccelByte Gaming Platform - Core backend services
  - SDK/Client: github.com/AccelByte/accelbyte-go-sdk v0.85.0
  - Auth: AB_CLIENT_ID, AB_CLIENT_SECRET (env vars)
  - Base URL: AB_BASE_URL (env var)
  - Namespace: AB_NAMESPACE (env var)

**Sub-services:**
- AccelByte IAM (Identity & Access Management) - User authentication
- AccelByte CloudSave - Data persistence for guild progress
- AccelByte Permission System - Resource-based authorization

## Data Storage

**Databases:**
- AccelByte CloudSave - Primary data storage
  - Connection: Via AccelByte SDK
  - Client: cloudsave.AdminGameRecordService
  - Pattern: Key-value storage for guild progress

**File Storage:**
- Local filesystem only - Swagger UI and static assets

**Caching:**
- None detected - All data retrieved from CloudSave API

## Authentication & Identity

**Auth Provider:**
- AccelByte IAM - OAuth 2.0 client credentials flow
  - Implementation: Custom token validator with refresh
  - Location: `pkg/common/authServerInterceptor.go`
  - Config: PLUGIN_GRPC_SERVER_AUTH_ENABLED (env var)

## Monitoring & Observability

**Error Tracking:**
- None detected - No external error tracking service

**Logs:**
- Structured JSON logging (slog)
- OpenTelemetry tracing integration
- Export: Zipkin (configurable via OTEL_EXPORTER_ZIPKIN_ENDPOINT)

**Metrics:**
- Prometheus metrics collection
- gRPC server metrics
- Go runtime metrics
- Endpoint: `/metrics` on port 8080

## CI/CD & Deployment

**Hosting:**
- Docker containers (multi-stage builds)
- Platform: Not specified (container-agnostic)

**CI Pipeline:**
- None detected - No CI/CD configuration files

## Environment Configuration

**Required env vars:**
- AB_CLIENT_ID - AccelByte client identifier
- AB_CLIENT_SECRET - AccelByte client secret
- AB_BASE_URL - AccelByte platform base URL
- AB_NAMESPACE - AccelByte namespace
- PLUGIN_GRPC_SERVER_AUTH_ENABLED - Authentication toggle
- BASE_PATH - API base path configuration

**Optional env vars:**
- LOG_LEVEL - Logging verbosity (default: info)
- OTEL_SERVICE_NAME - OpenTelemetry service name
- OTEL_EXPORTER_ZIPKIN_ENDPOINT - Zipkin collector endpoint
- REFRESH_INTERVAL - Token refresh interval (default: 600s)

**Secrets location:**
- Environment variables only
- No secret management service detected

## Webhooks & Callbacks

**Incoming:**
- None detected - No webhook endpoints

**Outgoing:**
- None detected - No external webhook calls

---

*Integration audit: 2026-01-27*