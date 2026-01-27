---
phase: 01-foundation
plan: 03
type: execute
wave: 3
depends_on: ["01-foundation-01", "01-foundation-02"]
files_modified: ["pkg/service/tournament.go", "main.go"]
autonomous: true
user_setup: []

must_haves:
  truths:
    - "Tournament service implements all CRUD operations with proper validation"
    - "Admin users can create, start, and cancel tournaments"
    - "All users can list tournaments and view tournament details"
    - "Permission checking enforces access control for all operations"
    - "Service integrates with existing gRPC-Gateway infrastructure"
  artifacts:
    - path: "pkg/service/tournament.go"
      provides: "Tournament service implementation"
      contains: "TournamentServiceServer", "CreateTournament", "ListTournaments", "GetTournament"
    - path: "main.go"
      provides: "Service registration and server setup"
      contains: "RegisterTournamentServiceServer", "tournamentServiceServer"
  key_links:
    - from: "pkg/service/tournament.go"
      to: "pkg/storage/tournament.go"
      via: "TournamentStorage dependency"
      pattern: "TournamentStorage"
    - from: "pkg/service/tournament.go"
      to: "pkg/common/auth_interceptors.go"
      via: "Permission validation"
      pattern: "CheckTournamentPermission"
    - from: "main.go"
      to: "pkg/service/tournament.go"
      via: "Service registration"
      pattern: "RegisterTournamentServiceServer"
---

<objective>
Implement tournament service with full CRUD operations and integrate with gRPC server

Purpose: Create complete tournament management service with validation, permission checking, and integration with existing infrastructure
Output: Working tournament service ready for API access through Swagger UI and REST endpoints
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
@main.go
@pkg/proto/tournament.proto
</context>

<tasks>

<task type="auto">
  <name>Task 1: Implement tournament service server</name>
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
5. Implement StartTournament method:
   - Validate tournament exists and is in ACTIVE status
   - Check admin permissions
   - Generate single-elimination brackets (basic implementation)
   - Update status to STARTED
   - Return updated tournament
6. Implement CancelTournament method:
   - Validate tournament exists and is not STARTED
   - Check admin permissions
   - Update status to CANCELLED
   - Return updated tournament
7. Add proper error handling with gRPC status codes
8. Add structured logging for all operations</action>
  <verify>go build ./pkg/service/... compiles without errors</verify>
  <done>Tournament service server implemented with all CRUD operations and validation</done>
</task>

<task type="auto">
  <name>Task 2: Integrate tournament service with main server</name>
  <files>main.go</files>
  <action>Update main.go to register tournament service:
1. Import tournament service package
2. Create tournamentServiceServer instance following myServiceServer pattern:
   - Pass tokenRepo, configRepo, refreshRepo, cloudSaveStorage
   - Use same dependency injection approach
3. Register tournament service with gRPC server:
   - Add: pb.RegisterTournamentServiceServer(s, tournamentServiceServer)
   - Place after existing service registration
4. Ensure tournament service uses existing interceptor chain:
   - Auth interceptors will handle permission checking
   - Logging interceptors will handle operation logging
   - Tracing interceptors will handle distributed tracing
5. Test that server starts successfully with tournament service registered
6. Verify that tournament endpoints are available through gRPC-Gateway</action>
  <verify>go build . compiles successfully and server starts without errors</verify>
  <done>Tournament service integrated with gRPC server and available through REST API</done>
</task>

<task type="auto">
  <name>Task 3: Add basic bracket generation for tournament start</name>
  <files>pkg/service/tournament.go</files>
  <action>Implement basic single-elimination bracket generation:
1. Add GenerateBrackets helper function:
   - Handle power-of-2 and non-power-of-2 participant counts
   - Assign byes for non-power-of-2 participants
   - Create basic bracket structure (rounds array)
   - Store bracket data in tournament record
2. Update StartTournament method to call GenerateBrackets:
   - Generate brackets before changing status to STARTED
   - Store bracket structure in tournament data
   - Return tournament with generated brackets
3. Add validation for bracket generation:
   - Ensure tournament has sufficient participants
   - Validate bracket structure consistency
4. Follow research recommendations for standard single-elimination format
5. Add logging for bracket generation process</action>
  <verify>go build ./pkg/service/... compiles without errors</verify>
  <done>Basic bracket generation implemented for tournament start operation</done>
</task>

</tasks>

<verification>
Build and test the complete service:
```bash
go build .
./extend-custom-guild-service &
# Test that server starts and tournament endpoints are available
curl http://localhost:8000/v1/tournaments
```

Verify tournament service registration in main.go
Check that all CRUD operations are implemented
Test permission checking with different user roles
</verification>

<success_criteria>
- Tournament service implements all required CRUD operations
- Admin users can create, start, and cancel tournaments
- All users can list tournaments and view details
- Permission checking enforces access control
- Basic bracket generation works for tournament start
- Service integrated with existing gRPC-Gateway infrastructure
- Server starts successfully and tournament endpoints available
- Ready for testing through Swagger UI
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-03-SUMMARY.md`
</output>