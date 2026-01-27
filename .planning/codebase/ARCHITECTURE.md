# Architecture

**Analysis Date:** 2026-01-27

## Pattern Overview

**Overall:** gRPC Service Extension with REST Gateway

**Key Characteristics:**
- gRPC server with HTTP/REST gateway for external access
- Protocol Buffers for service definition and code generation
- Interceptor-based authentication and authorization
- Cloud storage abstraction for data persistence
- Observability with OpenTelemetry tracing and Prometheus metrics

## Layers

**Application Layer:**
- Purpose: Main application entry point and server orchestration
- Location: `main.go`
- Contains: Server initialization, middleware setup, service registration
- Depends on: Service layer, Common layer, Storage layer
- Used by: Container runtime

**Service Layer:**
- Purpose: Business logic implementation for gRPC services
- Location: `pkg/service/`
- Contains: Service implementations, request/response handling
- Depends on: Storage layer, Protocol buffers
- Used by: gRPC server

**Storage Layer:**
- Purpose: Data persistence abstraction
- Location: `pkg/storage/`
- Contains: Storage interfaces and implementations
- Depends on: AccelByte CloudSave SDK
- Used by: Service layer

**Common Layer:**
- Purpose: Shared utilities and cross-cutting concerns
- Location: `pkg/common/`
- Contains: Authentication interceptors, logging, tracing, gateway setup
- Depends on: AccelByte SDK, OpenTelemetry, gRPC middleware
- Used by: All layers

**Protocol Layer:**
- Purpose: Service contract definitions and generated code
- Location: `pkg/proto/`, `pkg/pb/`
- Contains: Protocol buffer definitions, generated gRPC code
- Depends on: Google API annotations, OpenAPI annotations
- Used by: Service layer, Common layer

## Data Flow

**Request Flow:**

1. HTTP request arrives at gRPC-Gateway (port 8000)
2. Gateway translates REST request to gRPC call
3. Authentication interceptor validates token and permissions
4. Logging interceptor records request metadata
5. Tracing interceptor creates/continues trace span
6. Service implementation processes business logic
7. Storage layer persists/retrieves data via CloudSave
8. Response flows back through interceptors to gateway
9. Gateway translates gRPC response to HTTP response

**State Management:**
- Stateless service architecture
- External state managed via AccelByte CloudSave
- Authentication state managed via AccelByte IAM

## Key Abstractions

**Service Interface:**
- Purpose: gRPC service contract
- Examples: `pkg/proto/service.proto`
- Pattern: Protocol buffer service definitions with HTTP annotations

**Storage Interface:**
- Purpose: Data persistence abstraction
- Examples: `pkg/storage/storage.go`
- Pattern: Interface-based storage with CloudSave implementation

**Auth Interceptor:**
- Purpose: Request authentication and authorization
- Examples: `pkg/common/authServerInterceptor.go`
- Pattern: gRPC interceptor with token validation and permission checking

## Entry Points

**Main Application:**
- Location: `main.go`
- Triggers: Application startup
- Responsibilities: Server initialization, middleware setup, service registration

**gRPC Server:**
- Location: Started in `main.go`
- Triggers: gRPC requests from gateway
- Responsibilities: Service method execution, interceptor chain

**HTTP Gateway:**
- Location: Started in `main.go`
- Triggers: External HTTP requests
- Responsibilities: REST-to-gRPC translation, Swagger UI serving

**Metrics Server:**
- Location: Started in `main.go`
- Triggers: Prometheus scraping
- Responsibilities: Metrics endpoint exposure

## Error Handling

**Strategy:** gRPC status codes with structured error messages

**Patterns:**
- Service layer returns gRPC status errors with descriptive messages
- Storage layer errors wrapped in gRPC Internal status
- Authentication errors return Unauthenticated or PermissionDenied status
- Validation errors return InvalidArgument status

## Cross-Cutting Concerns

**Logging:** Structured JSON logging with slog, request/response logging via interceptors
**Validation:** Token validation via AccelByte IAM SDK, permission checking via proto annotations
**Authentication:** Bearer token validation with AccelByte IAM, permission-based authorization
**Observability:** OpenTelemetry tracing with B3 propagation, Prometheus metrics for gRPC operations
**Documentation:** OpenAPI/Swagger generation from proto annotations

---

*Architecture analysis: 2026-01-27*