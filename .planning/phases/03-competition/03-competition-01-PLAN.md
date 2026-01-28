---
phase: 03-competition
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: [pkg/proto/tournament.proto, pkg/pb/tournament.pb.go, pkg/pb/tournament_grpc.pb.go, pkg/pb/tournament.pb.gw.go]
autonomous: true
user_setup: []

must_haves:
  truths:
    - "Match protobuf messages defined with proper tournament association"
    - "Match service endpoints defined for viewing and result submission"
    - "REST endpoints generated with proper authentication patterns"
    - "gRPC code generation includes all match-related types and services"
  artifacts:
    - path: "pkg/proto/tournament.proto"
      provides: "Match message definitions and service methods"
      contains: "message Match", "message SubmitMatchResultRequest", "rpc GetTournamentMatches"
      min_lines: 50
    - path: "pkg/pb/tournament.pb.go"
      provides: "Generated Go match types and service interfaces"
      exports: ["Match", "SubmitMatchResultRequest", "GetTournamentMatchesRequest"]
      min_lines: 200
    - path: "pkg/pb/tournament_grpc.pb.go"
      provides: "Generated gRPC service interface with match methods"
      exports: ["TournamentServiceServer.SubmitMatchResult", "TournamentServiceServer.GetTournamentMatches"]
      min_lines: 100
    - path: "pkg/pb/tournament.pb.gw.go"
      provides: "Generated REST endpoints for match operations"
      exports: ["RegisterTournamentServiceHandler", "RegisterTournamentServiceHandlerFromEndpoint"]
      min_lines: 150
  key_links:
    - from: "pkg/proto/tournament.proto"
      to: "pkg/pb/tournament.pb.go"
      via: "protoc compilation"
      pattern: "protoc.*tournament.proto"
    - from: "message Match"
      to: "existing Tournament message"
      via: "tournament_id field reference"
      pattern: "string tournament_id"
    - from: "match service methods"
      to: "existing authentication patterns"
      via: "HTTP annotations with security requirements"
      pattern: "security_requirement.*Bearer.*ServiceToken"
---

<objective>
Extend tournament protobuf with match data model and service endpoints

Purpose: Define the data contracts for match management, result submission, and bracket viewing that integrate with existing tournament and participant systems
Output: Complete protobuf extension with match messages, service methods, and generated gRPC/REST code
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/03-competition/03-CONTEXT.md
@.planning/phases/03-competition/03-RESEARCH.md
@pkg/proto/tournament.proto
@pkg/service/tournament.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Define match protobuf messages and service methods</name>
  <files>pkg/proto/tournament.proto</files>
  <action>
    Extend tournament.proto with match-related definitions:
    
    1. Add MatchStatus enum (SCHEDULED, IN_PROGRESS, COMPLETED, CANCELLED)
    2. Add Match message with fields:
       - match_id (string)
       - tournament_id (string) 
       - round (int32)
       - position (int32)
       - participant1 (TournamentParticipant)
       - participant2 (TournamentParticipant)
       - winner (string - AccelByte user ID)
       - status (MatchStatus)
       - started_at (google.protobuf.Timestamp)
       - completed_at (google.protobuf.Timestamp)
    3. Add service methods:
       - GetTournamentMatches (returns bracket organized by round)
       - GetMatch (individual match details)
       - SubmitMatchResult (for game servers with Service token auth)
       - AdminSubmitMatchResult (admin override with Bearer token auth)
    4. Follow existing HTTP annotation patterns:
       - Public endpoints use /v1/public/ namespace
       - Admin endpoints use /v1/admin/ namespace  
       - Include Bearer and ServiceToken security requirements
       - Add proper OpenAPI summaries and descriptions
    5. Import existing TournamentParticipant and TournamentStatus to maintain consistency
  </action>
  <verify>protoc compilation succeeds without errors or warnings</verify>
  <done>Match protobuf messages and service methods properly defined with existing tournament integration</done>
</task>

<task type="auto">
  <name>Task 2: Generate gRPC code and REST endpoints</name>
  <files>pkg/pb/tournament.pb.go, pkg/pb/tournament_grpc.pb.go, pkg/pb/tournament.pb.gw.go</files>
  <action>
    Run protobuf code generation to create Go types and gRPC/REST endpoints:
    
    1. Execute make proto or equivalent protoc command:
       ```
       protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. pkg/proto/tournament.proto
       ```
    2. Verify generated files include:
       - Match struct with all fields from protobuf definition
       - TournamentServiceServer interface with new match methods
       - REST endpoint registration for match operations
       - OpenAPI security definitions inherited from existing tournament service
    3. Check that generated code follows existing patterns in tournament.pb.go
    4. Ensure no compilation errors in generated files
  </action>
  <verify>ls -la pkg/pb/tournament*.go shows updated files with recent timestamps and go build ./pkg/pb/ succeeds</verify>
  <done>Generated gRPC and REST code includes all match types and service methods with proper authentication integration</done>
</task>

</tasks>

<verification>
- [ ] Match protobuf definition includes all required fields and status enum
- [ ] Service methods defined with proper HTTP annotations and security
- [ ] Generated Go code compiles without errors
- [ ] REST endpoints follow existing namespace patterns (/v1/public/, /v1/admin/)
- [ ] Authentication patterns match existing tournament service (Bearer + ServiceToken)
- [ ] Integration with existing tournament and participant types maintained
</verification>

<success_criteria>
Complete match data model and service contract ready for storage layer implementation
- Match message definition with tournament association and result tracking
- Service methods for viewing brackets and submitting results  
- Generated gRPC/REST endpoints with proper authentication
- Integration with existing tournament service patterns maintained
</success_criteria>

<output>
After completion, create `.planning/phases/03-competition/03-competition-01-SUMMARY.md`
</output>