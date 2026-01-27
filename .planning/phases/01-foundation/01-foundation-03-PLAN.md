---
phase: 01-foundation
plan: 03
type: execute
wave: 3
depends_on: ["01-foundation-01", "01-foundation-02"]
files_modified: ["pkg/service/tournament.go"]
autonomous: true
user_setup: []

must_haves:
  truths:
    - "Tournament service implements core CRUD operations with proper validation"
    - "Admin users can create and cancel tournaments"
    - "All users can list tournaments and view tournament details"
    - "Permission checking enforces access control for all operations"
  artifacts:
    - path: "pkg/service/tournament.go"
      provides: "Tournament service implementation"
      contains: "TournamentServiceServer", "CreateTournament", "ListTournaments", "GetTournament", "CancelTournament"
  key_links:
    - from: "pkg/service/tournament.go"
      to: "pkg/storage/tournament.go"
      via: "TournamentStorage dependency"
      pattern: "TournamentStorage"
    - from: "pkg/service/tournament.go"
      to: "pkg/common/auth_interceptors.go"
      via: "Permission validation"
      pattern: "CheckTournamentPermission"
---

<objective>
Implement tournament service core operations with validation and permission checking

Purpose: Create tournament management service with basic CRUD operations, validation, and permission checking
Output: Working tournament service with core functionality ready for server integration
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
@.planning/phases/01-foundation/01-foundation-02-SUMMARY.md

# Existing patterns to follow
@pkg/service/myService.go
@pkg/proto/tournament.proto
</context>

<tasks>

<task type="auto">
  <name>Task 1: Implement tournament service core CRUD operations</name>
  <files>pkg/service/tournament.go</files>
  <action>Create tournament service following myService.go pattern:
1. Define TournamentServiceServer struct with dependencies:
   - TournamentStorage for data persistence
   - TokenRepository, ConfigRepository, RefreshTokenRepository for auth
   - Follow same dependency injection pattern as MyServiceServer
2. Implement CreateTournament method:
   - Validate required fields (name, max_participants)
   - Check admin permissions using CheckTournamentPermission
   - Set initial status to DRAFT
   - Call storage.CreateTournament
   - Return created tournament with generated fields
3. Implement ListTournaments method:
   - Support filtering by status, date range
   - Support pagination with limit/offset
   - No permission check (public read access)
   - Call storage.ListTournaments
   - Return paginated tournament list
4. Implement GetTournament method:
   - Validate tournament_id parameter
   - No permission check (public read access)
   - Call storage.GetTournament
   - Handle not found errors
   - Return tournament details
5. Implement CancelTournament method:
   - Validate tournament exists and is not STARTED
   - Check admin permissions
   - Update status to CANCELLED
   - Return updated tournament
6. Add proper error handling with gRPC status codes
7. Add structured logging for all operations</action>
  <verify>go build ./pkg/service/... compiles without errors</verify>
  <done>Tournament service core CRUD operations implemented with validation and permission checking</done>
</task>

<task type="auto">
  <name>Task 2: Add tournament status transitions and validation</name>
  <files>pkg/service/tournament.go</files>
  <action>Implement tournament status management:
1. Add ValidateStatusTransition helper function:
   - Define allowed transitions (DRAFT→ACTIVE, ACTIVE→STARTED, etc.)
   - Validate status changes before applying
   - Return appropriate errors for invalid transitions
2. Update tournament operations to use status validation:
   - CreateTournament sets status to DRAFT
   - CancelTournament validates not STARTED
   - Add helper methods for status checking
3. Add status-based business logic:
   - Only ACTIVE tournaments can be started
   - Only DRAFT/ACTIVE tournaments can be cancelled
   - COMPLETED tournaments are read-only
4. Add logging for status changes
5. Ensure status persistence in storage layer</action>
  <verify>go build ./pkg/service/... compiles without errors</verify>
  <done>Tournament status transitions implemented with proper validation and business rules</done>
</task>

</tasks>

<verification>
Build the service to verify implementation:
```bash
go build ./pkg/service/...
```

Check that all CRUD operations are implemented
Verify status transition validation logic
Test permission checking integration
</verification>

<success_criteria>
- Tournament service implements core CRUD operations
- Status transitions are properly validated
- Permission checking enforces access control
- Error handling with proper gRPC status codes
- Structured logging for all operations
- Ready for server integration
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-03-SUMMARY.md`
</output>