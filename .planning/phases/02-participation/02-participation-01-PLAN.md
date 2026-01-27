---
phase: 02-participation
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: ["pkg/proto/tournament.proto", "pkg/pb/tournament.pb.go", "pkg/pb/tournament_grpc.pb.go", "pkg/pb/tournament.pb.gw.go"]
autonomous: true

must_haves:
  truths:
    - "Participant protobuf messages exist with proper field definitions"
    - "Registration RPC endpoints defined with HTTP annotations"
    - "Participant listing RPC endpoint with pagination support"
    - "Security definitions require Bearer token authentication"
  artifacts:
    - path: "pkg/proto/tournament.proto"
      provides: "Participant messages and registration service definitions"
      contains: "message Participant", "message RegisterForTournamentRequest"
      min_lines: 50
    - path: "pkg/pb/tournament.pb.go"
      provides: "Generated Go structs for participant registration"
      contains: "type Participant struct", "type RegisterForTournamentRequest struct"
    - path: "pkg/pb/tournament_grpc.pb.go"
      provides: "gRPC service interface for registration operations"
      contains: "RegisterForTournament", "GetTournamentParticipants"
  key_links:
    - from: "pkg/proto/tournament.proto"
      to: "Phase 1 tournament messages"
      via: "import tournament definitions"
      pattern: "import.*tournament.*proto"
---

<objective>
Extend tournament protobuf with participant registration data model and service interface for player registration management.

Purpose: Add participant messages and registration endpoints to existing tournament service with proper authentication and REST API generation.
Output: Generated protobuf definitions for participant registration with gRPC and REST endpoints.
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/02-participation/02-CONTEXT.md
@.planning/phases/02-participation/02-RESEARCH.md
@.planning/phases/01-foundation/01-foundation-01-SUMMARY.md
</context>

<tasks>

<task type="auto">
  <name>Add participant protobuf messages</name>
  <files>pkg/proto/tournament.proto</files>
  <action>
Add participant messages to existing tournament.proto file:

1. Add Participant message after Tournament enum:
```protobuf
message Participant {
  string participant_id = 1 [(google.api.field_behavior) = REQUIRED];
  string user_id = 2 [(google.api.field_behavior) = REQUIRED];
  string username = 3 [(google.api.field_behavior) = OPTIONAL];
  string display_name = 4 [(google.api.field_behavior) = OPTIONAL];
  string tournament_id = 5 [(google.api.field_behavior) = REQUIRED];
  google.protobuf.Timestamp registered_at = 6 [(google.api.field_behavior) = REQUIRED];
  google.protobuf.Timestamp updated_at = 7 [(google.api.field_behavior) = OPTIONAL];
}
```

2. Add registration request/response messages:
```protobuf
message RegisterForTournamentRequest {
  string namespace = 1 [(google.api.field_behavior) = REQUIRED];
  string tournament_id = 2 [(google.api.field_behavior) = REQUIRED];
}

message RegisterForTournamentResponse {
  string participant_id = 1 [(google.api.field_behavior) = REQUIRED];
  string tournament_id = 2 [(google.api.field_behavior) = REQUIRED];
  string user_id = 3 [(google.api.field_behavior) = REQUIRED];
  google.protobuf.Timestamp registered_at = 4 [(google.api.field_behavior) = REQUIRED];
}

message GetTournamentParticipantsRequest {
  string namespace = 1 [(google.api.field_behavior) = REQUIRED];
  string tournament_id = 2 [(google.api.field_behavior) = REQUIRED];
  int32 page_size = 3 [(google.api.field_behavior) = OPTIONAL];
  string page_token = 4 [(google.api.field_behavior) = OPTIONAL];
}

message GetTournamentParticipantsResponse {
  repeated Participant participants = 1 [(google.api.field_behavior) = REQUIRED];
  int32 total_participants = 2 [(google.api.field_behavior) = REQUIRED];
  string next_page_token = 3 [(google.api.field_behavior) = OPTIONAL];
}
```

3. Add RemoveParticipantRequest for admin use (REG-02):
```protobuf
message RemoveParticipantRequest {
  string namespace = 1 [(google.api.field_behavior) = REQUIRED];
  string tournament_id = 2 [(google.api.field_behavior) = REQUIRED];
  string user_id = 3 [(google.api.field_behavior) = REQUIRED];
}

message RemoveParticipantResponse {
  string tournament_id = 1 [(google.api.field_behavior) = REQUIRED];
  string user_id = 2 [(google.api.field_behavior) = REQUIRED];
  bool removed = 3 [(google.api.field_behavior) = REQUIRED];
}
```

Follow existing protobuf patterns from Phase 1 for consistency.
  </action>
  <verify>grep -n "message Participant" pkg/proto/tournament.proto && grep -n "RegisterForTournamentRequest" pkg/proto/tournament.proto</verify>
  <done>Participant messages added to protobuf with proper field behavior annotations</done>
</task>

<task type="auto">
  <name>Add registration RPCs to TournamentService</name>
  <files>pkg/proto/tournament.proto</files>
  <action>
Add registration endpoints to existing TournamentService in tournament.proto:

1. Add RegisterForTournament RPC after existing tournament RPCs:
```protobuf
  rpc RegisterForTournament (RegisterForTournamentRequest) returns (RegisterForTournamentResponse) {
    option (google.api.http) = {
      post: "/v1/public/namespace/{namespace}/tournaments/{tournament_id}/register"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Register for Tournament"
      description: "Register user for tournament with capacity enforcement"
      security: {
        security_requirement: {
          key: "Bearer"
          value: {}
        }
      }
    };
  }
```

2. Add GetTournamentParticipants RPC:
```protobuf
  rpc GetTournamentParticipants (GetTournamentParticipantsRequest) returns (GetTournamentParticipantsResponse) {
    option (google.api.http) = {
      get: "/v1/public/namespace/{namespace}/tournaments/{tournament_id}/participants"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Get Tournament Participants"
      description: "List all participants for a tournament with pagination"
      security: {
        security_requirement: {
          key: "Bearer"
          value: {}
        }
      }
    };
  }
```

3. Add RemoveParticipant RPC for admin use (REG-02):
```protobuf
  rpc RemoveParticipant (RemoveParticipantRequest) returns (RemoveParticipantResponse) {
    option (google.api.http) = {
      delete: "/v1/admin/namespace/{namespace}/tournaments/{tournament_id}/participants/{user_id}"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Remove Tournament Participant"
      description: "Admin-only: Remove participant from tournament"
      security: {
        security_requirement: {
          key: "Bearer"
          value: {}
        }
      }
    };
  }
```

Follow existing HTTP annotation patterns from Phase 1 tournament RPCs.
  </action>
  <verify>grep -n "RegisterForTournament" pkg/proto/tournament.proto && grep -n "GetTournamentParticipants" pkg/proto/tournament.proto</verify>
  <done>Registration RPCs added to TournamentService with proper REST endpoints and security</done>
</task>

<task type="auto">
  <name>Regenerate protobuf Go code</name>
  <files>pkg/pb/tournament.pb.go, pkg/pb/tournament_grpc.pb.go, pkg/pb/tournament.pb.gw.go</files>
  <action>
Regenerate protobuf Go files to include new participant messages and RPCs:

1. Navigate to project root and run protobuf generation:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
    --grpc-gateway_opt=generate_unbound_methods=true \
    pkg/proto/*.proto
```

2. Verify generated files contain new participant types:
   - Participant struct in tournament.pb.go
   - RegisterForTournament methods in tournament_grpc.pb.go
   - HTTP handlers in tournament.pb.gw.go

3. Check for any compilation errors or import issues.

Follow Phase 1 protobuf generation pattern exactly.
  </action>
  <verify>grep -n "type Participant struct" pkg/pb/tournament.pb.go && grep -n "RegisterForTournament" pkg/pb/tournament_grpc.pb.go</verify>
  <done>Protobuf Go code regenerated with participant registration types and gRPC/REST handlers</done>
</task>

</tasks>

<verification>
- Participant protobuf messages exist with proper field behavior annotations
- Registration RPCs defined with HTTP annotations and Bearer token security
- Generated Go files compile and include new participant types
- REST endpoints follow Phase 1 URL patterns (/v1/public/namespace/... for public, /v1/admin/namespace/... for admin)
- All messages have required namespace and tournament_id fields for consistency
</verification>

<success_criteria>
- Complete participant data model in protobuf with proper field types
- Registration endpoints with RESTful HTTP mapping
- Participant listing with pagination support
- Admin-only participant removal endpoint
- Generated Go code compiles without errors
- OpenAPI documentation includes new endpoints with security definitions
</success_criteria>

<output>
After completion, create `.planning/phases/02-participation/02-participation-01-SUMMARY.md`
</output>