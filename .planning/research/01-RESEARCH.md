# Phase 1: Foundation - Research

**Researched:** 2026-01-27
**Domain:** Go gRPC Gateway tournament system with AccelByte IAM authentication
**Confidence:** HIGH

## Summary

Phase 1 requires implementing a tournament management foundation using Go with gRPC-Gateway architecture and AccelByte IAM authentication. The standard stack is well-established: Go 1.24+ with gRPC v1.72.0, gRPC-Gateway v2.26.3, and AccelByte Go SDK v0.85.0. Authentication patterns are clearly defined with interceptor-based token validation and permission checking. Tournament data models should use protobuf definitions with standard HTTP annotations for REST API generation.

**Primary recommendation:** Use the existing gRPC-Gateway pattern with AccelByte Go SDK for authentication, following the established interceptor architecture in the codebase.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.24.0 | Backend service implementation | Current stable version with full gRPC support |
| gRPC | v1.72.0 | Core communication protocol | Industry standard for high-performance RPC |
| gRPC-Gateway | v2.26.3 | HTTP/REST API gateway | Official protobuf-to-REST translation |
| AccelByte Go SDK | v0.85.0 | Gaming platform integration | Official SDK for IAM authentication |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Protocol Buffers | v3 | API contract definition | All service definitions |
| OpenTelemetry | v1.35.0 | Observability tracing | Production monitoring |
| Prometheus | v1.22.0 | Metrics collection | Production monitoring |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| gRPC-Gateway | Direct HTTP server | Lose automatic protobuf generation, more boilerplate |
| AccelByte SDK | Direct IAM API calls | Lose token management, error handling, retries |

**Installation:**
```bash
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Architecture Patterns

### Recommended Project Structure
```
src/
├── pkg/
│   ├── proto/           # Protocol buffer definitions
│   ├── pb/             # Generated protobuf code
│   ├── service/        # gRPC service implementations
│   ├── storage/        # Data persistence layer
│   └── common/         # Shared utilities, interceptors
├── cmd/
│   └── server/         # Main application entry point
└── api/
    └── openapi/        # Generated OpenAPI specs
```

### Pattern 1: gRPC Service with HTTP Annotations
**What:** Define services in protobuf with HTTP annotations for REST API generation
**When to use:** All external APIs that need both gRPC and HTTP access
**Example:**
```protobuf
// Source: gRPC-Gateway official documentation
service TournamentService {
  rpc CreateTournament(CreateTournamentRequest) returns (Tournament) {
    option (google.api.http) = {
      post: "/tournaments"
      body: "*"
    };
  }
  
  rpc ListTournaments(ListTournamentsRequest) returns (ListTournamentsResponse) {
    option (google.api.http) = {
      get: "/tournaments"
    };
  }
}
```

### Pattern 2: Authentication Interceptor Chain
**What:** Use gRPC interceptors for authentication, logging, and tracing
**When to use:** All secured endpoints requiring token validation
**Example:**
```go
// Source: AccelByte Go SDK patterns
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Extract token from metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Errorf(codes.Unauthenticated, "metadata not provided")
    }
    
    // Validate with AccelByte SDK
    authHeaders := md.Get("authorization")
    if len(authHeaders) == 0 {
        return nil, status.Errorf(codes.Unauthenticated, "authorization header not provided")
    }
    
    // Token validation logic here
    return handler(ctx, req)
}
```

### Anti-Patterns to Avoid
- **Direct HTTP without gRPC:** Loses protobuf benefits, automatic documentation
- **Custom auth without SDK:** Misses token refresh, error handling, retries
- **Mixed authentication patterns:** Confusing token management, security risks

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Token validation | Custom JWT parsing | AccelByte SDK token validation | Handles refresh, local validation, error cases |
| REST API generation | Manual HTTP handlers | gRPC-Gateway annotations | Automatic OpenAPI, protobuf consistency |
| Permission checking | Custom role mapping | AccelByte SDK permission validation | Handles namespace, resource, action validation |
| Error handling | Custom error codes | gRPC status codes with SDK errors | Standardized, client-friendly |

**Key insight:** The AccelByte Go SDK provides comprehensive token management, local validation, and automatic refresh that would be complex and error-prone to replicate.

## Common Pitfalls

### Pitfall 1: Token Scope Mismatch
**What goes wrong:** Using wrong token type (user vs service) for API calls
**Why it happens:** Confusion between user tokens (from game clients) and service tokens (for backend operations)
**How to avoid:** Clearly separate token flows - validate user tokens, then use service tokens for backend calls
**Warning signs:** 403 Forbidden errors, permission denied for valid operations

### Pitfall 2: Missing Permission Annotations
**What goes wrong:** gRPC methods lack proper permission annotations in protobuf
**Why it happens:** Forgetting to add AccelByte permission requirements to service definitions
**How to avoid:** Always include permission annotations in protobuf for secured endpoints
**Warning signs:** All requests pass authentication but fail authorization

### Pitfall 3: Inconsistent Error Handling
**What goes wrong:** Mixing gRPC status codes with HTTP error codes
**Why it happens:** gRPC-Gateway translation confusion
**How to avoid:** Use gRPC status codes consistently, let gateway handle HTTP translation
**Warning signs:** Clients receive unexpected error formats

## Code Examples

Verified patterns from official sources:

### Tournament Service Definition
```protobuf
// Source: gRPC-Gateway official documentation
syntax = "proto3";

package tournament.v1;

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

service TournamentService {
  rpc CreateTournament(CreateTournamentRequest) returns (Tournament) {
    option (google.api.http) = {
      post: "/v1/tournaments"
      body: "*"
    };
  }
  
  rpc ListTournaments(ListTournamentsRequest) returns (ListTournamentsResponse) {
    option (google.api.http) = {
      get: "/v1/tournaments"
    };
  }
  
  rpc GetTournament(GetTournamentRequest) returns (Tournament) {
    option (google.api.http) = {
      get: "/v1/tournaments/{tournament_id}"
    };
  }
}

message Tournament {
  string tournament_id = 1;
  string name = 2;
  string description = 3;
  int32 max_participants = 4;
  TournamentStatus status = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}

enum TournamentStatus {
  TOURNAMENT_STATUS_UNSPECIFIED = 0;
  TOURNAMENT_STATUS_DRAFT = 1;
  TOURNAMENT_STATUS_ACTIVE = 2;
  TOURNAMENT_STATUS_STARTED = 3;
  TOURNAMENT_STATUS_COMPLETED = 4;
  TOURNAMENT_STATUS_CANCELLED = 5;
}
```

### Authentication with AccelByte SDK
```go
// Source: AccelByte Go SDK documentation
package service

import (
    "context"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
    "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam/iam_client"
)

type TournamentService struct {
    iamClient iam.Client
    tokenRepo auth.TokenRepository
}

func (s *TournamentService) ValidateToken(ctx context.Context, accessToken string) (*iam.TokenValidationV3, error) {
    // Use AccelByte SDK for token validation
    input := &iam_client.TokenValidationV3Params{
        Token: &accessToken,
    }
    
    validated, err := s.iamClient.TokenValidationV3Short(input)
    if err != nil {
        return nil, err
    }
    
    return validated.GetPayload(), nil
}
```

### Interceptor Chain Setup
```go
// Source: gRPC-Gateway best practices
func setupServer() *grpc.Server {
    // Create interceptor chain
    chain := grpc.ChainUnaryInterceptor(
        loggingInterceptor,
        authInterceptor,
        tracingInterceptor,
    )
    
    return grpc.NewServer(
        grpc.UnaryInterceptor(chain),
        grpc.StreamInterceptor(streamInterceptor),
    )
}

func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Skip auth for health checks
    if info.FullMethod == "/grpc.health.v1.Health/Check" {
        return handler(ctx, req)
    }
    
    // Extract and validate token
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Errorf(codes.Unauthenticated, "metadata not provided")
    }
    
    authHeaders := md.Get("authorization")
    if len(authHeaders) == 0 {
        return nil, status.Errorf(codes.Unauthenticated, "authorization header not provided")
    }
    
    // Validate with AccelByte SDK
    token := strings.TrimPrefix(authHeaders[0], "Bearer ")
    if err := validateToken(token); err != nil {
        return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
    }
    
    return handler(ctx, req)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| gRPC-Gateway v1 | gRPC-Gateway v2 | 2022 | Better protobuf annotations, improved performance |
| Manual token refresh | SDK automatic refresh | 2023 | Simplified token management, better reliability |
| Basic HTTP logging | Structured logging with tracing | 2024 | Better observability, debugging capabilities |

**Deprecated/outdated:**
- gRPC-Gateway v1: Use v2 for better protobuf support
- Manual token validation: Use AccelByte SDK local validation
- Custom error handling: Use gRPC status codes

## Open Questions

Things that couldn't be fully resolved:

1. **Tournament bracket generation algorithm**
   - What we know: Single-elimination format required
   - What's unclear: Specific seeding algorithm, bracket structure for non-power-of-2 participants
   - Recommendation: Implement standard single-elimination with byes for non-power-of-2

2. **Permission granularity for tournament operations**
   - What we know: Admin vs user permissions required
   - What's unclear: Specific AccelByte permission strings for tournament CRUD
   - Recommendation: Define custom permissions in AccelByte namespace (e.g., TOURNAMENT:CREATE, TOURNAMENT:READ)

3. **Long-lived session implementation**
   - What we know: 24-48 hour sessions desired for gaming
   - What's unclear: Session storage mechanism, refresh strategy
   - Recommendation: Use AccelByte SDK automatic token refresh with 80% lifetime threshold

## Sources

### Primary (HIGH confidence)
- gRPC-Gateway v2 documentation - HTTP annotations, interceptor patterns
- AccelByte Go SDK v0.85.0 - Token validation, authentication flows
- Protocol Buffers v3 - Message definitions, service syntax

### Secondary (MEDIUM confidence)
- Go gRPC interceptor patterns - Verified with official gRPC docs
- AccelByte Extend SDK documentation - Authentication best practices
- gRPC-Gateway tutorial examples - Project structure patterns

### Tertiary (LOW confidence)
- Tournament system examples - Need verification with actual implementation
- WebSearch patterns for gaming APIs - Require official documentation verification

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Based on official documentation and current versions
- Architecture: HIGH - Verified with gRPC-Gateway and AccelByte SDK docs
- Pitfalls: MEDIUM - Based on common patterns, need validation with actual implementation

**Research date:** 2026-01-27
**Valid until:** 2026-02-26 (30 days for stable stack)