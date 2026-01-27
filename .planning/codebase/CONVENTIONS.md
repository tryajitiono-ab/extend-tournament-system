# Coding Conventions

**Analysis Date:** 2026-01-27

## Naming Patterns

**Files:**
- Go files use snake_case for packages (e.g., `my_service.go`)
- Package names are lowercase, single words (e.g., `service`, `storage`, `common`)
- Mock files follow pattern: `*_mock.go` in `mocks/` subdirectory

**Functions:**
- Public functions use PascalCase (e.g., `NewMyServiceServer`)
- Private functions use camelCase (e.g., `parseSlogLevel`)
- Interface methods use PascalCase
- Mock functions follow pattern: `Mock[Type]` and `NewMock[Type]`

**Variables:**
- Public variables use PascalCase (e.g., `Validator`)
- Private variables use camelCase (e.g., `logLevelStr`)
- Constants use UPPER_SNAKE_CASE (e.g., `METRICS_ENDPOINT`)
- Package-level vars use camelCase

**Types:**
- Structs use PascalCase (e.g., `MyServiceServerImpl`)
- Interfaces use PascalCase (e.g., `TokenRepository`)
- Type aliases use PascalCase

## Code Style

**Formatting:**
- Tool: golangci-lint with extensive configuration
- Key settings: Silent mode, line-number format, most linters enabled
- Disabled linters include: wsl, gomnd, goerr113, wrapcheck, exhaustivestruct

**Linting:**
- Tool: golangci-lint
- Config: `.golangci.yml`
- Key rules: enable-all: true with specific disabled linters
- Custom header template for copyright notices

## Import Organization

**Order:**
1. Standard library imports (grouped)
2. Third-party imports (grouped)
3. Local imports (grouped with module prefix)

**Path Aliases:**
- Module prefix: `extend-custom-guild-service/pkg/`
- Protocol buffer alias: `pb "extend-custom-guild-service/pkg/pb"`
- SDK imports: `github.com/AccelByte/accelbyte-go-sdk/...`

## Error Handling

**Patterns:**
- Use `status.Errorf` with gRPC codes for service errors
- Return `(result, error)` tuples consistently
- Use `os.Exit(1)` for fatal initialization errors
- Wrap errors with context using `fmt.Errorf("message: %v", err)`

**gRPC Error Codes:**
- `codes.Internal` for service errors
- `codes.Unimplemented` for stub methods
- Use structured error messages

## Logging

**Framework:** slog (structured logging)

**Patterns:**
- JSON handler for structured output: `slog.NewJSONHandler(os.Stdout, opts)`
- Use structured key-value pairs: `logger.Info("message", "key", value)`
- Error logging with context: `logger.Error("message", "error", err)`
- Default logger set via `slog.SetDefault(logger)`

**Log Levels:**
- Custom `parseSlogLevel()` function for string-to-level conversion
- Support for debug, info, warn, error levels
- Default level: info

## Comments

**When to Comment:**
- Copyright headers required on all files
- Function comments for exported functions
- Inline comments for complex business logic

**Copyright Header:**
```go
// Copyright (c) 2023-2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.
```

**JSDoc/TSDoc:**
- Not applicable (Go project)
- Use standard Go doc comments for exported symbols

## Function Design

**Size:** No strict limits, but functions are generally focused

**Parameters:**
- Interface injection for dependencies (e.g., `repository.TokenRepository`)
- Context as first parameter for service methods
- Request structs for gRPC methods

**Return Values:**
- `(result, error)` pattern for service methods
- Response structs for gRPC methods
- Use `status.Errorf` for error returns

## Module Design

**Exports:**
- Exported symbols start with uppercase letter
- Private symbols start with lowercase letter
- Package-level variables for shared state (e.g., `Validator`)

**Barrel Files:**
- Not used (Go doesn't have barrel files)
- Each package imports what it needs directly

---

*Convention analysis: 2026-01-27*