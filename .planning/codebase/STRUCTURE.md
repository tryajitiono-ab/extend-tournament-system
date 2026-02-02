# Codebase Structure

**Analysis Date:** 2026-01-27

## Directory Layout

```
extend-tournament-service/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── Makefile                   # Build commands
├── Dockerfile                 # Container build definition
├── docker-compose.yaml        # Local development setup
├── .env.template             # Environment variables template
├── .golangci.yml             # Linting configuration
├── .gitignore                # Git ignore rules
├── README.md                 # Project documentation
├── pkg/                      # Main application packages
│   ├── common/               # Shared utilities and cross-cutting concerns
│   ├── pb/                   # Generated protocol buffer code
│   ├── proto/                # Protocol buffer definitions
│   ├── service/              # Business logic implementations
│   └── storage/              # Data persistence layer
├── gateway/                  # gRPC-Gateway configuration
├── docs/                     # Documentation files
├── demo/                     # Demo and testing collections
├── third_party/              # Third-party dependencies
└── .planning/                # Planning and analysis documents
```

## Directory Purposes

**pkg/common/:
- Purpose: Shared utilities, interceptors, and cross-cutting concerns
- Contains: Authentication, logging, tracing, gateway setup, utility functions
- Key files: `authServerInterceptor.go`, `gateway.go`, `logging.go`, `tracerProvider.go`, `utils.go`

**pkg/pb/:
- Purpose: Generated protocol buffer code for gRPC services
- Contains: Service definitions, message types, gateway code
- Key files: `service.pb.go`, `service_grpc.pb.go`, `service.pb.gw.go`, `permission.pb.go`

**pkg/proto/:
- Purpose: Protocol buffer definition files
- Contains: Service contracts, message schemas, API annotations
- Key files: `service.proto`, `permission.proto`, Google API annotations

**pkg/service/:
- Purpose: Business logic implementation for gRPC services
- Contains: Service implementations, request/response handling
- Key files: `myService.go`, mocks for testing

**pkg/storage/:
- Purpose: Data persistence abstraction and implementations
- Contains: Storage interfaces, MongoDB implementation
- Key files: `tournament.go`, `participant.go`, `match.go`

**gateway/:
- Purpose: gRPC-Gateway API documentation
- Contains: Swagger/OpenAPI specifications
- Key files: API documentation JSON files

**demo/:
- Purpose: Demo collections for API testing
- Contains: Postman collections for authentication and service testing
- Key files: `get-access-token.postman_collection.json`, `service-extension-demo.postman_collection.json`

**docs/:
- Purpose: Project documentation
- Contains: Migration guides, development setup instructions
- Key files: Development container setup, migration documentation

**third_party/:
- Purpose: Third-party static assets
- Contains: Swagger UI static files
- Key files: Embedded Swagger UI resources

## Key File Locations

**Entry Points:**
- `main.go`: Application entry point with server initialization
- `docker-compose.yaml`: Local development environment setup
- `Dockerfile`: Container build configuration

**Configuration:**
- `.env.template`: Environment variables template
- `.golangci.yml`: Go linting configuration
- `Makefile`: Build and development commands

**Core Logic:**
- `pkg/service/myService.go`: Main service implementation
- `pkg/storage/storage.go`: Data persistence layer
- `pkg/proto/service.proto`: Service contract definition

**Testing:**
- `pkg/service/mocks/`: Generated mock files for testing

## Naming Conventions

**Files:**
- Go source files: `camelCase.go` (e.g., `myService.go`, `authServerInterceptor.go`)
- Protocol files: `snake_case.proto` (e.g., `service.proto`, `permission.proto`)
- Generated files: `snake_case.pb.go`, `snake_case_grpc.pb.go`
- Configuration files: `kebab-case.yaml` or `camelCase.yml` (e.g., `docker-compose.yaml`, `.golangci.yml`)

**Directories:**
- Package directories: `camelCase` (e.g., `common`, `service`, `storage`)
- Configuration directories: `kebab-case` or lowercase (e.g., `third_party`)

## Where to Add New Code

**New Feature:**
- Primary code: `pkg/service/[featureName].go`
- Proto definitions: `pkg/proto/[featureName].proto`
- Tests: `pkg/service/mocks/[featureName]_mock.go`

**New Component/Module:**
- Implementation: `pkg/[componentName]/[componentName].go`
- Interface: `pkg/[componentName]/interface.go` (if needed)

**Utilities:**
- Shared helpers: `pkg/common/[utilityName].go`
- Storage implementations: `pkg/storage/[storageType].go`

## Special Directories

**pkg/pb/:**
- Purpose: Generated code from protocol buffers
- Generated: Yes
- Committed: Yes

**gateway/:**
- Purpose: API documentation for gRPC-Gateway
- Generated: Partially (from proto annotations)
- Committed: Yes

**third_party/:**
- Purpose: Third-party static dependencies
- Generated: No
- Committed: Yes

**docs/:**
- Purpose: Project documentation
- Generated: No
- Committed: Yes

---

*Structure analysis: 2026-01-27*