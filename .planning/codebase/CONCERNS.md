# Codebase Concerns

**Analysis Date:** 2026-01-27

## Tech Debt

**Monolithic Main Function:**
- Issue: The `main.go` file contains 350 lines with server initialization, configuration, and startup logic all mixed together
- Files: `[main.go]`
- Impact: Makes the application difficult to test, configure, and maintain. Server setup is tightly coupled and hard to modify independently.
- Fix approach: Extract server initialization into separate packages (server, config, bootstrap) and use dependency injection patterns.

**Generated Code Commitment:**
- Issue: Large protobuf-generated files (380+ lines) are committed to version control
- Files: `[pkg/pb/service.pb.go, pkg/pb/service_grpc.pb.go, pkg/pb/service.pb.gw.go]`
- Impact: Bloates repository, makes diffs noisy, potential merge conflicts
- Fix approach: Add .gitignore rules for generated files and set up code generation in build pipeline

**Hardcoded Configuration:**
- Issue: Port numbers and service paths are hardcoded as constants
- Files: `[main.go]`
- Impact: Reduces deployment flexibility, requires code changes for environment-specific configs
- Fix approach: Move all configurable values to environment variables or configuration files

## Known Bugs

**Authentication Error Handling:**
- Symptoms: Authentication errors are logged but service continues to start up
- Files: `[main.go:132-135]`
- Trigger: When token validator initialization fails
- Workaround: None - service runs in potentially broken state

**Swagger File Discovery:**
- Symptoms: Swagger JSON serving fails silently if no matching files found
- Files: `[main.go:308-314]`
- Trigger: Missing swagger files in gateway/apidocs directory
- Workaround: Ensure swagger files are present before starting server

## Security Considerations

**Client Credentials Exposure:**
- Risk: Client ID and secret are handled in plaintext in main function
- Files: `[main.go:153-159]`
- Current mitigation: Basic error handling on login failure
- Recommendations: Use secure credential management, avoid logging sensitive data

**Token Validation Bypass:**
- Risk: Authentication can be disabled via environment variable
- Files: `[main.go:129]`
- Current mitigation: Default is "true" (enabled)
- Recommendations: Add additional safeguards for production deployments

**Namespace Default:**
- Risk: Hardcoded namespace fallback could lead to data access issues
- Files: `[pkg/common/authServerInterceptor.go:158]`
- Current mitigation: Uses "accelbyte" as fallback
- Recommendations: Make namespace mandatory or use more secure defaults

## Performance Bottlenecks

**JSON Marshaling Round-trip:**
- Problem: Storage layer marshals to JSON then unmarshals for protobuf conversion
- Files: `[pkg/storage/tournament.go, pkg/storage/participant.go, pkg/storage/match.go]`
- Cause: Inefficient data transformation between MongoDB and protobuf
- Improvement path: Direct protobuf-to-model conversion without JSON intermediate

**Synchronous File Operations:**
- Problem: Swagger JSON parsing happens on every HTTP request
- Files: `[main.go:307-350]`
- Cause: File I/O in request handler without caching
- Improvement path: Pre-parse and cache swagger files at startup

## Fragile Areas

**Service Initialization:**
- Files: `[main.go:83-261]`
- Why fragile: Long initialization sequence with many failure points that cause os.Exit(1)
- Safe modification: Break into smaller, testable functions with proper error propagation
- Test coverage: No unit tests for initialization logic

**Authentication Interceptor:**
- Files: `[pkg/common/authServerInterceptor.go]`
- Why fragile: Complex reflection-based permission extraction with multiple failure modes
- Safe modification: Add comprehensive error handling and validation
- Test coverage: No tests for auth interceptor edge cases

**Storage Interface:**
- Files: `[pkg/storage/tournament.go, pkg/storage/participant.go, pkg/storage/match.go]`
- Why fragile: No retry logic or connection error handling for MongoDB operations
- Safe modification: Add circuit breaker pattern and retry mechanisms
- Test coverage: No integration tests for storage failure scenarios

## Scaling Limits

**Single Instance Deployment:**
- Current capacity: Single-process gRPC server
- Limit: No horizontal scaling capability
- Scaling path: Add support for multiple instances with load balancing

**In-Memory Token Validation:**
- Current capacity: Local token validation with in-memory caches
- Limit: Memory-bound cache sizes, no distributed invalidation
- Scaling path: Implement distributed cache (Redis) for token validation

## Dependencies at Risk

**AccelByte SDK:**
- Risk: Heavy dependency on AccelByte ecosystem makes code less portable
- Impact: Service cannot be easily migrated to other platforms
- Migration plan: Abstract storage and authentication interfaces behind clean abstractions

**Protobuf Version Lock:**
- Risk: Specific protobuf versions may cause compatibility issues
- Impact: gRPC client/server version mismatches
- Migration plan: Use version-compatible gRPC gateway and maintain compatibility matrix

## Missing Critical Features

**Health Check Implementation:**
- Problem: Basic health check registered but no custom health logic
- Blocks: Proper monitoring and load balancer integration
- Files: `[main.go:177]`

**Graceful Shutdown:**
- Problem: No graceful connection draining on shutdown
- Blocks: Zero-downtime deployments
- Files: `[main.go:257-260]`

**Configuration Validation:**
- Problem: No validation of required environment variables at startup
- Blocks: Early detection of misconfigurations
- Files: `[pkg/common/utils.go:32-44]`

## Test Coverage Gaps

**No Unit Tests:**
- What's not tested: All business logic, authentication, storage operations
- Files: `[pkg/service/myService.go, pkg/storage/storage.go, pkg/common/*.go]`
- Risk: Regressions and bugs in core functionality
- Priority: High

**No Integration Tests:**
- What's not tested: gRPC endpoints, HTTP gateway, authentication flow
- Files: `[main.go server setup, pb generated endpoints]`
- Risk: Configuration and deployment issues
- Priority: High

**No Error Path Testing:**
- What's not tested: Error handling, timeouts, connection failures
- Files: All error-prone areas identified above
- Risk: Poor error handling in production
- Priority: Medium

---

*Concerns audit: 2026-01-27*