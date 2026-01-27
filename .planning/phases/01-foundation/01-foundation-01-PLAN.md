---
phase: 01-foundation
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: ["pkg/proto/tournament.proto", "pkg/pb/tournament.pb.go", "pkg/pb/tournament_grpc.pb.go", "pkg/pb/tournament.pb.gw.go"]
autonomous: true
user_setup:
  - service: accelbyte
    why: "Permission configuration for tournament operations"
    env_vars:
      - name: AB_CLIENT_ID
        source: "AccelByte Admin Portal -> Game Services -> Extend -> Your Service -> Credentials"
      - name: AB_CLIENT_SECRET  
        source: "AccelByte Admin Portal -> Game Services -> Extend -> Your Service -> Credentials"
    dashboard_config:
      - task: "Create tournament permissions"
        location: "AccelByte Admin Portal -> IAM -> Permissions -> Add Permission"

must_haves:
  truths:
    - "Tournament data model supports required fields (name, description, max participants, status)"
    - "Tournament status enum covers all lifecycle states (draft, active, started, completed, cancelled)"
    - "HTTP annotations enable REST API generation for all tournament operations"
    - "Permission annotations integrate with AccelByte IAM for authorization"
  artifacts:
    - path: "pkg/proto/tournament.proto"
      provides: "Tournament data model and service definition"
      contains: "message Tournament", "enum TournamentStatus", "service TournamentService"
    - path: "pkg/pb/tournament.pb.go"
      provides: "Generated Go structs for tournament data"
      min_lines: 100
    - path: "pkg/pb/tournament_grpc.pb.go"
      provides: "Generated gRPC service interface"
      exports: ["TournamentServiceServer", "RegisterTournamentServiceServer"]
    - path: "pkg/pb/tournament.pb.gw.go"
      provides: "Generated REST gateway handlers"
      exports: ["RegisterTournamentServiceHandlerFromEndpoint"]
  key_links:
    - from: "pkg/proto/tournament.proto"
      to: "AccelByte IAM"
      via: "permission annotations"
      pattern: "option \\(permission\\.action\\)"
    - from: "pkg/proto/tournament.proto"
      to: "REST API"
      via: "HTTP annotations"
      pattern: "option \\(google\\.api\\.http\\)"
---

<objective>
Create tournament data model and service definition with AccelByte IAM integration

Purpose: Establish the foundation for tournament management with proper data structures, REST API generation, and permission-based access control
Output: Complete protobuf definitions with generated Go code for tournament service
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

# Existing patterns to follow
@pkg/proto/service.proto
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create tournament protobuf definition</name>
  <files>pkg/proto/tournament.proto</files>
  <action>Create tournament.proto following existing service.proto patterns:
1. Define Tournament message with fields: tournament_id, name, description, max_participants, current_participants, status, created_at, updated_at, start_time, end_time
2. Define TournamentStatus enum: DRAFT, ACTIVE, STARTED, COMPLETED, CANCELLED
3. Define TournamentService with CRUD operations:
   - CreateTournament (admin only)
   - ListTournaments (public read)
   - GetTournament (public read)  
   - StartTournament (admin only)
   - CancelTournament (admin only)
4. Add HTTP annotations for REST API generation following pattern in service.proto
5. Add permission annotations for AccelByte IAM integration:
   - Admin operations: CREATE, UPDATE on "ADMIN:NAMESPACE:{namespace}:TOURNAMENT"
   - Read operations: READ on "NAMESPACE:{namespace}:TOURNAMENT"
6. Add OpenAPI operation summaries and descriptions following service.proto pattern
7. Include proper security requirements for Bearer token authentication</action>
  <verify>protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. pkg/proto/tournament.proto runs without errors</verify>
  <action>After creating .proto file, run protoc generation to create Go files:
```bash
protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. \
  --proto_path=third_party \
  --proto_path=third_party/googleapis \
  --proto_path=pkg/proto \
  pkg/proto/tournament.proto
```</action>
  <done>Generated tournament.pb.go, tournament_grpc.pb.go, and tournament.pb.gw.go files with complete data model and service interface</done>
</task>

<task type="auto">
  <name>Task 2: Configure AccelByte tournament permissions</name>
  <files>pkg/proto/tournament.proto</files>
  <action>Update permission annotations in tournament.proto to use correct AccelByte permission format:
1. Verify permission resource format matches AccelByte namespace pattern
2. Ensure admin operations require ADMIN:NAMESPACE:{namespace}:TOURNAMENT:CREATE/UPDATE
3. Ensure read operations require NAMESPACE:{namespace}:TOURNAMENT:READ  
4. Add permission validation comments for future reference
5. Test that generated code includes permission metadata</action>
  <verify>grep -n "permission\." pkg/proto/tournament.proto shows proper permission annotations</verify>
  <done>AccelByte IAM permissions properly configured for tournament operations with admin/user distinction</done>
</task>

</tasks>

<verification>
Run protoc command to generate Go code from protobuf:
```bash
protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. \
  --proto_path=third_party \
  --proto_path=third_party/googleapis \
  --proto_path=pkg/proto \
  pkg/proto/tournament.proto
```

Verify generated files exist and contain tournament definitions:
- pkg/pb/tournament.pb.go (data structures)
- pkg/pb/tournament_grpc.pb.go (gRPC interface)
- pkg/pb/tournament.pb.gw.go (REST gateway)

Check that permission annotations are present in service definition
</verification>

<success_criteria>
- Complete tournament data model with all required fields
- Tournament status enum covering all lifecycle states
- gRPC service definition with all CRUD operations
- HTTP annotations enabling REST API generation
- AccelByte IAM permission annotations properly configured
- Generated Go code compiles without errors
- Ready for service implementation in next plan
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation/01-foundation-01-SUMMARY.md`
</output>