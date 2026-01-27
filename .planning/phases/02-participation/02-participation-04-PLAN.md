---
phase: 02-participation
plan: 04
type: execute
wave: 2
depends_on: ["02-participation-03"]
files_modified: ["cmd/server/main.go", "pkg/pb/tournament_grpc.pb.go"]
autonomous: true

must_haves:
  truths:
    - "Participant registration endpoints integrated with gRPC server"
    - "Authentication interceptor chain includes participant services"
    - "REST endpoints available through gRPC-Gateway"
    - "OpenAPI documentation includes participant endpoints"
  artifacts:
    - path: "cmd/server/main.go"
      provides: "Server integration with participant services"
      contains: "NewParticipantService|RegisterForTournament|GetTournamentParticipants"
    - path: "pkg/pb/tournament_grpc.pb.go"
      provides: "Generated gRPC server interface with participant methods"
      contains: "RegisterForTournament.*interface"
  key_links:
    - from: "cmd/server/main.go"
      to: "pkg/service/participant.go"
      via: "service instantiation"
      pattern: "NewParticipantService"
    - from: "cmd/server/main.go"
      to: "Phase 1 server setup"
      via: "gRPC server registration"
      pattern: "RegisterTournamentServiceServer"
---

<objective>
Integrate participant registration service with gRPC server and enable REST endpoints through gRPC-Gateway with proper authentication and documentation.

Purpose: Complete participant registration functionality by connecting service layer to server infrastructure with authentication, REST endpoints, and API documentation.
Output: Fully functional tournament registration system with gRPC and REST endpoints, authentication, and OpenAPI documentation.
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
@.planning/phases/01-foundation/01-foundation-04-SUMMARY.md
@.planning/phases/01-foundation/01-foundation-03-SUMMARY.md
</context>

<tasks>

<task type="auto">
  <name>Integrate participant service with gRPC server</name>
  <files>cmd/server/main.go</files>
  <action>
Enhance existing cmd/server/main.go to integrate participant registration service:

1. Add participant storage initialization (after tournament storage):
```go
// After tournamentStorage initialization
participantStorage := storage.NewParticipantStorage(mongoClient, dbName)
```

2. Add participant service initialization (after tournament service):
```go
// After tournamentService initialization
participantService := service.NewParticipantService(
    participantStorage,
    tournamentStorage,
    logger,
)
```

3. Add participant service to gRPC server registration (enhance existing RegisterTournamentServiceServer):
```go
// Find the existing tournament service registration and enhance it
tournamentServer := &server.TournamentServer{
    TournamentService:   tournamentService,
    ParticipantService: participantService,  // Add this field to server struct
}

// Register with gRPC server
pb.RegisterTournamentServiceServer(s, tournamentServer)
```

4. If server.TournamentServer doesn't exist or needs enhancement, modify it to include participant service:
```go
// In pkg/server/tournament.go (create if doesn't exist)
type TournamentServer struct {
    pb.UnimplementedTournamentServiceServer
    *service.TournamentService
    *service.ParticipantService  // Add participant service
}

// Implement participant methods by delegating to service
func (s *TournamentServer) RegisterForTournament(ctx context.Context, req *pb.RegisterForTournamentRequest) (*pb.RegisterForTournamentResponse, error) {
    return s.ParticipantService.RegisterForTournament(ctx, req)
}

func (s *TournamentServer) GetTournamentParticipants(ctx context.Context, req *pb.GetTournamentParticipantsRequest) (*pb.GetTournamentParticipantsResponse, error) {
    return s.ParticipantService.GetTournamentParticipants(ctx, req)
}

func (s *TournamentServer) RemoveParticipant(ctx context.Context, req *pb.RemoveParticipantRequest) (*pb.RemoveParticipantResponse, error) {
    return s.ParticipantService.RemoveParticipant(ctx, req)
}
```

5. Ensure authentication interceptors are applied to participant endpoints (should be automatic if using existing interceptor chain):
```go
// Verify interceptor setup includes participant endpoints
s := grpc.NewServer(
    grpc.UnaryInterceptor(chain.Chain),
)
```

Follow Phase 1 server integration patterns exactly for consistency.
  </action>
  <verify>grep -n "NewParticipantService" cmd/server/main.go && grep -n "RegisterForTournament" cmd/server/main.go</verify>
  <done>Participant service integrated with gRPC server and authentication chain</done>
</task>

<task type="auto">
  <name>Enhance server struct for participant methods</name>
  <files>pkg/server/tournament.go</files>
  <action>
Create or enhance pkg/server/tournament.go to include participant method implementations:

If file doesn't exist, create it:
```go
package server

import (
    "context"

    "github.com/accelerated-development/tournament-system/pkg/pb"
    "github.com/accelerated-development/tournament-system/pkg/service"
)

// TournamentServer implements the TournamentService gRPC interface
type TournamentServer struct {
    pb.UnimplementedTournamentServiceServer
    *service.TournamentService
    *service.ParticipantService
}

// NewTournamentServer creates a new tournament server instance
func NewTournamentServer(
    tournamentService *service.TournamentService,
    participantService *service.ParticipantService,
) *TournamentServer {
    return &TournamentServer{
        TournamentService:   tournamentService,
        ParticipantService: participantService,
    }
}

// Tournament CRUD methods (existing from Phase 1)
// GetTournament, ListTournaments, CreateTournament, etc.

// Participant registration methods

// RegisterForTournament registers a user for a tournament
func (s *TournamentServer) RegisterForTournament(ctx context.Context, req *pb.RegisterForTournamentRequest) (*pb.RegisterForTournamentResponse, error) {
    return s.ParticipantService.RegisterForTournament(ctx, req)
}

// GetTournamentParticipants retrieves participants for a tournament
func (s *TournamentServer) GetTournamentParticipants(ctx context.Context, req *pb.GetTournamentParticipantsRequest) (*pb.GetTournamentParticipantsResponse, error) {
    return s.ParticipantService.GetTournamentParticipants(ctx, req)
}

// RemoveParticipant removes a participant from a tournament (admin only)
func (s *TournamentServer) RemoveParticipant(ctx context.Context, req *pb.RemoveParticipantRequest) (*pb.RemoveParticipantResponse, error) {
    return s.ParticipantService.RemoveParticipant(ctx, req)
}
```

If file exists, add ParticipantService field and participant method implementations:
```go
// Add to existing TournamentServer struct
*service.ParticipantService

// Add new methods
func (s *TournamentServer) RegisterForTournament(ctx context.Context, req *pb.RegisterForTournamentRequest) (*pb.RegisterForTournamentResponse, error) {
    return s.ParticipantService.RegisterForTournament(ctx, req)
}

func (s *TournamentServer) GetTournamentParticipants(ctx context.Context, req *pb.GetTournamentParticipantsRequest) (*pb.GetTournamentParticipantsResponse, error) {
    return s.ParticipantService.GetTournamentParticipants(ctx, req)
}

func (s *TournamentServer) RemoveParticipant(ctx context.Context, req *pb.RemoveParticipantRequest) (*pb.RemoveParticipantResponse, error) {
    return s.ParticipantService.RemoveParticipant(ctx, req)
}
```

This completes the gRPC server integration for participant registration.
  </action>
  <verify>grep -n "ParticipantService" pkg/server/tournament.go && grep -n "RegisterForTournament" pkg/server/tournament.go</verify>
  <done>Server struct enhanced with participant service and method implementations</done>
</task>

<task type="auto">
  <name>Verify REST endpoints and documentation generation</name>
  <files>pkg/pb/tournament.pb.gw.go</files>
  <action>
Verify that REST endpoints and OpenAPI documentation are properly generated for participant registration:

1. Check that REST endpoints exist in generated code:
```bash
grep -n "register\|participants" pkg/pb/tournament.pb.gw.go
```

2. Verify HTTP mappings are correct:
```bash
grep -A 5 -B 5 "/v1/public/namespace" pkg/pb/tournament.pb.gw.go
```

3. Check OpenAPI documentation includes participant endpoints:
```bash
grep -A 10 "RegisterForTournament\|GetTournamentParticipants" pkg/pb/tournament.swagger.json 2>/dev/null || echo "Swagger file not found, will be generated"
```

4. If swagger file doesn't exist, run documentation generation:
```bash
protoc --openapiv2_out=. --openapiv2_opt=logtostderr=true pkg/proto/*.proto
```

5. Verify generated swagger includes participant endpoints:
```bash
grep -A 5 "Register for Tournament\|Get Tournament Participants" pkg/proto/tournament.swagger.json
```

6. Test compilation of entire codebase:
```bash
go build ./cmd/server
```

The gRPC-Gateway should automatically generate REST endpoints from the HTTP annotations added in plan 01. The OpenAPI documentation should include all participant endpoints with proper security definitions.
  </action>
  <verify>grep -n "register\|participants" pkg/pb/tournament.pb.gw.go && go build ./cmd/server</verify>
  <done>REST endpoints verified working and OpenAPI documentation includes participant registration</done>
</task>

</tasks>

<verification>
- Participant service instantiated and injected into gRPC server
- Tournament server struct enhanced with participant service methods
- All participant RPC methods implemented with proper delegation
- REST endpoints available through gRPC-Gateway with correct URL patterns
- OpenAPI documentation includes participant registration endpoints
- Authentication interceptors properly applied to participant endpoints
- Codebase compiles without errors
- Integration follows Phase 1 server patterns exactly
</verification>

<success_criteria>
- Complete server integration for participant registration functionality
- gRPC endpoints working for RegisterForTournament, GetTournamentParticipants, RemoveParticipant
- REST endpoints available at /v1/public/namespace/{namespace}/tournaments/{tournament_id}/register
- REST endpoints available at /v1/public/namespace/{namespace}/tournaments/{tournament_id}/participants
- Admin REST endpoint available at /v1/admin/namespace/{namespace}/tournaments/{tournament_id}/participants/{user_id}
- OpenAPI documentation includes all participant endpoints with security definitions
- Authentication and authorization properly enforced
- System ready for end-to-end testing of tournament registration
</success_criteria>

<output>
After completion, create `.planning/phases/02-participation/02-participation-04-SUMMARY.md`
</output>