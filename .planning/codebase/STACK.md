# Technology Stack

**Analysis Date:** 2026-01-27

## Languages

**Primary:**
- Go 1.24.0 - Backend service implementation

**Secondary:**
- Protocol Buffers v3 - API contract definition
- Shell Script - Build automation (proto.sh)
- Dockerfile - Container configuration

## Runtime

**Environment:**
- Go 1.24.0
- Alpine Linux 3.22 (runtime container)

**Package Manager:**
- Go modules
- Lockfile: go.sum (present)

## Frameworks

**Core:**
- gRPC v1.72.0 - Core communication protocol
- gRPC-Gateway v2.26.3 - HTTP/REST API gateway
- AccelByte Go SDK v0.85.0 - Gaming platform integration

**Testing:**
- Go mock v0.2.0 - Mock generation
- Not detected: Specific test framework configuration

**Build/Dev:**
- Protocol Buffers compiler v21.9 - Code generation
- Docker multi-stage builds - Containerization
- Make - Build automation

## Key Dependencies

**Critical:**
- github.com/AccelByte/accelbyte-go-sdk v0.85.0 - Gaming platform SDK
- google.golang.org/grpc v1.72.0 - gRPC framework
- google.golang.org/protobuf v1.36.6 - Protocol buffers

**Infrastructure:**
- go.opentelemetry.io/otel v1.35.0 - Observability
- github.com/prometheus/client_golang v1.22.0 - Metrics
- github.com/go-openapi/loads v0.22.0 - OpenAPI spec handling

## Configuration

**Environment:**
- Environment variables (.env.template)
- Key configs required: AB_CLIENT_ID, AB_CLIENT_SECRET, AB_BASE_URL, AB_NAMESPACE

**Build:**
- Dockerfile (multi-stage)
- Makefile (build automation)
- proto.sh (protocol buffer generation)

## Platform Requirements

**Development:**
- Go 1.24.0+
- Docker (for containerized builds)
- Protocol Buffers compiler

**Production:**
- Linux container runtime
- AccelByte platform access
- OpenTelemetry collector (optional)
- Prometheus server (optional)

---

*Stack analysis: 2026-01-27*