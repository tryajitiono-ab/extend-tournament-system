---
phase: 01-foundation
plan: 04
type: execute
wave: 4
depends_on: ["01-foundation-01", "01-foundation-02", "01-foundation-03"]
files_modified: ["main.go", "pkg/service/tournament.go"]
autonomous: true
user_setup: []

must_haves:
  truths:
    - "Tournament service is registered with gRPC server and available through REST API"
    - "Tournament start operation generates single-elimination brackets"
    - "Service integrates with existing gRPC-Gateway infrastructure"
    - "Server starts successfully and tournament endpoints are available"
  artifacts:
    - path: "main.go"
      provides: "Service registration and server setup"
      contains: "RegisterTournamentServiceServer", "tournamentServiceServer"
    - path: "pkg/service/tournament.go"
      provides: "Tournament start operation with bracket generation"
      contains: "StartTournament", "GenerateBrackets"
  key_links:
    - from: "main.go"
      to: "pkg/service/tournament.go"
      via: "Service registration"
      pattern: "RegisterTournamentServiceServer"
    - from: "pkg/service/tournament.go"
      to: "pkg/storage/tournament.go"
      via: "Bracket data storage"
      pattern: "GenerateBrackets"
---

<objective>
Integrate tournament service with gRPC server and implement bracket generation

Purpose: Complete tournament service integration with server infrastructure and add tournament start functionality with bracket generation
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
@.planning/phases/01-foundation/01-foundation-03-SUMMARY.md

# Existing patterns to follow
@pkg/service/myService.go
@main.go
@pkg/proto/tournament.proto
</context>

<tasks>

<task type="auto">
  <name>Task 1: Integrate tournament service with main server</name>
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
  <name>Task 2: Add basic bracket generation for tournament start</name>
  <files>pkg/service/tournament.go</files>
  <action>Implement basic single-elimination bracket generation:
1. Add GenerateBrackets helper function:
   - Handle power-of-2 and non-power-of-2 participant counts
   - Assign byes for non-power-of-2 participants
   - Create basic bracket structure (rounds array)
   - Store bracket data in tournament record
2. Implement StartTournament method:
   - Validate tournament exists and is in ACTIVE status
   - Check admin permissions
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
Test bracket generation for tournament start
</verification>

<success_criteria>
- Tournament service integrated with existing gRPC-Gateway infrastructure
- Admin users can create and start tournaments with bracket generation
- All users can list tournaments and view details
- Server starts successfully and tournament endpoints available
- Basic bracket generation works for tournament start
- Ready for testing through Swagger UI
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-04-SUMMARY.md`
</output>