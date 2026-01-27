---
phase: 01-foundation
plan: 02
type: execute
wave: 2
depends_on: ["01-foundation-01"]
files_modified: ["pkg/storage/tournament.go", "pkg/common/auth_interceptors.go"]
autonomous: true
user_setup: []

must_haves:
  truths:
    - "Tournament storage persists and retrieves tournament data using CloudSave"
    - "Authentication interceptors validate AccelByte IAM tokens for tournament operations"
    - "Permission checking enforces admin vs user access controls"
    - "Storage layer handles tournament lifecycle transitions correctly"
  artifacts:
    - path: "pkg/storage/tournament.go"
      provides: "Tournament storage implementation"
      contains: "TournamentStorage", "CreateTournament", "GetTournament", "ListTournaments"
    - path: "pkg/common/auth_interceptors.go"
      provides: "Authentication and authorization middleware"
      contains: "TournamentAuthInterceptor", "CheckTournamentPermission"
  key_links:
    - from: "pkg/storage/tournament.go"
      to: "AccelByte CloudSave"
      via: "AdminGameRecordService"
      pattern: "AdminGameRecordService"
    - from: "pkg/common/auth_interceptors.go"
      to: "AccelByte IAM"
      via: "Token validation"
      pattern: "oauthService\\."
    - from: "pkg/common/auth_interceptors.go"
      to: "Tournament permissions"
      via: "Permission checking"
      pattern: "TOURNAMENT:"
---

<objective>
Implement tournament storage layer and authentication interceptors with AccelByte IAM integration

Purpose: Create data persistence layer using CloudSave and enforce permission-based access control for tournament operations
Output: Working storage for tournament data and authentication middleware ready for service integration
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/01-foundation/01-CONTEXT.md
@.planning/research/01-RESEARCH.md
@.planning/phases/01-foundation/01-foundation-01-SUMMARY.md

# Existing patterns to follow
@pkg/storage/storage.go
@pkg/common/authServerInterceptor.go
@pkg/proto/tournament.proto
</context>

<tasks>

<task type="auto">
  <name>Task 1: Implement tournament storage layer</name>
  <files>pkg/storage/tournament.go</files>
  <action>Create tournament storage following existing storage.go pattern:
1. Define TournamentStorage struct with AdminGameRecordService dependency
2. Implement CreateTournament method:
   - Generate UUID for tournament_id
   - Set initial status to DRAFT
   - Set created_at and updated_at timestamps
   - Store using AdminGameRecordService with proper namespace
   - Return tournament with generated fields
3. Implement GetTournament method:
   - Retrieve tournament record by ID from CloudSave
   - Handle not found errors appropriately
   - Convert CloudSave format to Tournament proto
4. Implement ListTournaments method:
   - Support filtering by status using CloudSave queries
   - Support pagination using limit/offset
   - Convert results to Tournament proto list
5. Implement UpdateTournament method:
   - Support status transitions (DRAFT→ACTIVE, ACTIVE→STARTED, etc.)
   - Validate state transitions
   - Update updated_at timestamp
   - Return updated tournament
6. Add error handling for CloudSave operations with proper gRPC status codes
7. Follow the existing CloudSaveStorage pattern from storage.go</action>
  <verify>go build ./pkg/storage/... compiles without errors</verify>
  <done>Tournament storage layer implemented with CRUD operations using AccelByte CloudSave</done>
</task>

<task type="auto">
  <name>Task 2: Create tournament authentication interceptors</name>
  <files>pkg/common/auth_interceptors.go</files>
  <action>Create tournament-specific authentication following authServerInterceptor.go pattern:
1. Define TournamentAuthInterceptor struct with OAuth20Service dependency
2. Implement CheckTournamentPermission function:
   - Extract user info from validated token
   - Check required permission against user's permissions
   - Support both admin and user permission levels
   - Return error for insufficient permissions
3. Implement unary interceptor for tournament operations:
   - Skip auth for health check endpoints
   - Validate Bearer token from metadata
   - Call OAuth20Service for token validation
   - Check tournament-specific permissions based on operation
   - Add user context to request for authorization decisions
4. Implement stream interceptor for tournament operations with same logic
5. Add helper functions for permission mapping:
   - Map tournament operations to required permissions
   - Handle namespace-based permission checking
   - Support service token validation for game server access
6. Follow existing interceptor patterns for consistency
7. Add proper logging for authentication failures</action>
  <verify>go build ./pkg/common/... compiles without errors</verify>
  <done>Tournament authentication interceptors created with AccelByte IAM integration and permission checking</done>
</task>

<task type="auto">
  <name>Task 3: Integrate storage and auth with existing infrastructure</name>
  <files>pkg/storage/tournament.go, pkg/common/auth_interceptors.go</files>
  <action>Integrate new components with existing infrastructure:
1. Update tournament.go to use existing CloudSave patterns:
   - Import and use same CloudSave client as storage.go
   - Follow same error handling patterns
   - Use same logging format with slog
2. Update auth_interceptors.go to use existing auth patterns:
   - Import same OAuth20Service setup as main.go
   - Follow same token validation flow
   - Use same logging interceptor patterns
   - Ensure compatibility with existing interceptor chain
3. Add proper namespace handling for multi-tenant support
4. Test integration by building main.go with new components</action>
  <verify>go build . compiles successfully including new tournament components</verify>
  <done>Tournament storage and auth integrated with existing AccelByte infrastructure</done>
</task>

</tasks>

<verification>
Build all components to verify integration:
```bash
go build ./pkg/storage/...
go build ./pkg/common/...
go build .
```

Check that CloudSave storage methods follow existing patterns
Verify authentication interceptors integrate with existing auth flow
Test that namespace handling works correctly
</verification>

<success_criteria>
- Tournament storage implemented using AccelByte CloudSave
- CRUD operations for tournaments working with proper error handling
- Authentication interceptors validate AccelByte IAM tokens
- Permission checking enforces admin vs user access
- Integration with existing infrastructure successful
- Code follows established patterns and compiles without errors
- Ready for tournament service implementation
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-02-SUMMARY.md`
</output>