---
phase: 02-participation
plan: 03
type: execute
wave: 2
depends_on: ["02-participation-02"]
files_modified: ["pkg/service/participant.go", "pkg/service/tournament.go"]
autonomous: true

must_haves:
  truths:
    - "Participant service implements registration business logic"
    - "User authentication context properly extracted and validated"
    - "Registration capacity enforcement with race condition handling"
    - "Admin authorization for participant removal operations"
  artifacts:
    - path: "pkg/service/participant.go"
      provides: "Participant registration service with authorization"
      min_lines: 200
      exports: ["RegisterForTournament", "GetTournamentParticipants", "RemoveParticipant"]
    - path: "pkg/service/tournament.go"
      provides: "Enhanced tournament service with registration integration"
      contains: "participant integration"
  key_links:
    - from: "pkg/service/participant.go"
      to: "pkg/storage/participant.go"
      via: "storage layer calls"
      pattern: "participantStorage\\."
    - from: "pkg/service/participant.go"
      to: "Phase 1 auth patterns"
      via: "context extraction"
      pattern: "getContextUser|getContextNamespace"
---

<objective>
Implement participant registration service with user authentication, authorization, and business logic for tournament registration management.

Purpose: Provide complete registration service with user context extraction, permission validation, and integration with participant storage.
Output: Working participant service with proper authentication, authorization, and registration business logic.
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
@.planning/phases/01-foundation/01-foundation-03-SUMMARY.md
@.planning/phases/01-foundation/01-foundation-02-SUMMARY.md
</context>

<tasks>

<task type="auto">
  <name>Create participant service with authentication</name>
  <files>pkg/service/participant.go</files>
  <action>
Create pkg/service/participant.go with registration service logic following Phase 1 service patterns:

```go
package service

import (
	"context"
	"fmt"

	"github.com/accelerated-development/tournament-system/pkg/common"
	"github.com/accelerated-development/tournament-system/pkg/pb"
	"github.com/accelerated-development/tournament-system/pkg/storage"
	"go.uber.org/zap"
)

// ParticipantService handles participant registration operations
type ParticipantService struct {
	participantStorage *storage.ParticipantStorage
	tournamentStorage  *storage.TournamentStorage
	logger             *zap.Logger
}

// NewParticipantService creates a new participant service instance
func NewParticipantService(
	participantStorage *storage.ParticipantStorage,
	tournamentStorage *storage.TournamentStorage,
	logger *zap.Logger,
) *ParticipantService {
	return &ParticipantService{
		participantStorage: participantStorage,
		tournamentStorage:  tournamentStorage,
		logger:             logger,
	}
}

// RegisterForTournament registers a user for a tournament
func (p *ParticipantService) RegisterForTournament(ctx context.Context, req *pb.RegisterForTournamentRequest) (*pb.RegisterForTournamentResponse, error) {
	// Extract user context
	namespace, err := common.GetContextNamespace(ctx)
	if err != nil {
		p.logger.Error("failed to get namespace from context", zap.Error(err))
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	userID, err := common.GetContextUserID(ctx)
	if err != nil {
		p.logger.Error("failed to get user ID from context", zap.Error(err))
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	username, err := common.GetContextUsername(ctx)
	if err != nil {
		p.logger.Warn("failed to get username from context", zap.Error(err))
		// Username is optional, continue without it
	}

	// Validate request namespace matches context namespace
	if req.GetNamespace() != namespace {
		p.logger.Error("namespace mismatch", 
			zap.String("req_namespace", req.GetNamespace()),
			zap.String("ctx_namespace", namespace))
		return nil, fmt.Errorf("namespace mismatch")
	}

	p.logger.Info("user registering for tournament",
		zap.String("user_id", userID),
		zap.String("username", username),
		zap.String("tournament_id", req.GetTournamentId()),
		zap.String("namespace", namespace),
	)

	// Call storage with user context
	response, err := p.participantStorage.RegisterParticipant(ctx, req, userID)
	if err != nil {
		p.logger.Error("failed to register participant",
			zap.String("user_id", userID),
			zap.String("tournament_id", req.GetTournamentId()),
			zap.Error(err))
		return nil, err
	}

	p.logger.Info("user successfully registered for tournament",
		zap.String("user_id", userID),
		zap.String("participant_id", response.GetParticipantId()),
		zap.String("tournament_id", req.GetTournamentId()),
	)

	return response, nil
}

// GetTournamentParticipants retrieves participants for a tournament
func (p *ParticipantService) GetTournamentParticipants(ctx context.Context, req *pb.GetTournamentParticipantsRequest) (*pb.GetTournamentParticipantsResponse, error) {
	// Extract user context
	namespace, err := common.GetContextNamespace(ctx)
	if err != nil {
		p.logger.Error("failed to get namespace from context", zap.Error(err))
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	// Validate request namespace matches context namespace
	if req.GetNamespace() != namespace {
		p.logger.Error("namespace mismatch", 
			zap.String("req_namespace", req.GetNamespace()),
			zap.String("ctx_namespace", namespace))
		return nil, fmt.Errorf("namespace mismatch")
	}

	p.logger.Info("retrieving tournament participants",
		zap.String("tournament_id", req.GetTournamentId()),
		zap.String("namespace", namespace),
		zap.Int32("page_size", req.GetPageSize()),
	)

	// Get participants from storage
	response, err := p.participantStorage.GetParticipants(ctx, req)
	if err != nil {
		p.logger.Error("failed to get tournament participants",
			zap.String("tournament_id", req.GetTournamentId()),
			zap.Error(err))
		return nil, err
	}

	p.logger.Info("successfully retrieved tournament participants",
		zap.String("tournament_id", req.GetTournamentId()),
		zap.Int32("participant_count", response.GetTotalParticipants()),
	)

	return response, nil
}

// RemoveParticipant removes a participant from a tournament (admin only)
func (p *ParticipantService) RemoveParticipant(ctx context.Context, req *pb.RemoveParticipantRequest) (*pb.RemoveParticipantResponse, error) {
	// Extract user context and verify admin permissions
	namespace, err := common.GetContextNamespace(ctx)
	if err != nil {
		p.logger.Error("failed to get namespace from context", zap.Error(err))
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	// Check admin permissions
	isAdmin, err := common.IsAdminUser(ctx)
	if err != nil {
		p.logger.Error("failed to check admin permissions", zap.Error(err))
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	if !isAdmin {
		p.logger.Warn("unauthorized attempt to remove participant", 
			zap.String("user_id", "<redacted>"),
			zap.String("target_user_id", req.GetUserId()),
			zap.String("tournament_id", req.GetTournamentId()))
		return nil, fmt.Errorf("insufficient permissions: admin role required")
	}

	// Validate request namespace matches context namespace
	if req.GetNamespace() != namespace {
		p.logger.Error("namespace mismatch", 
			zap.String("req_namespace", req.GetNamespace()),
			zap.String("ctx_namespace", namespace))
		return nil, fmt.Errorf("namespace mismatch")
	}

	adminUserID, _ := common.GetContextUserID(ctx)
	
	p.logger.Info("admin removing participant from tournament",
		zap.String("admin_user_id", adminUserID),
		zap.String("target_user_id", req.GetUserId()),
		zap.String("tournament_id", req.GetTournamentId()),
		zap.String("namespace", namespace),
	)

	// Remove participant via storage
	response, err := p.participantStorage.RemoveParticipant(ctx, req)
	if err != nil {
		p.logger.Error("failed to remove participant",
			zap.String("target_user_id", req.GetUserId()),
			zap.String("tournament_id", req.GetTournamentId()),
			zap.Error(err))
		return nil, err
	}

	p.logger.Info("successfully removed participant from tournament",
		zap.String("admin_user_id", adminUserID),
		zap.String("target_user_id", req.GetUserId()),
		zap.String("tournament_id", req.GetTournamentId()),
	)

	return response, nil
}
```

Follow Phase 1 service patterns for error handling, logging, and context extraction.
  </action>
  <verify>grep -n "RegisterForTournament" pkg/service/participant.go && grep -n "GetContextUserID" pkg/service/participant.go</verify>
  <done>Participant service created with authentication, authorization, and business logic</done>
</task>

<task type="auto">
  <name>Enhance tournament service for participant integration</name>
  <files>pkg/service/tournament.go</files>
  <action>
Enhance existing tournament.go service to integrate with participant functionality:

1. Add participant storage field to TournamentService struct:
```go
type TournamentService struct {
	storage           *storage.TournamentStorage
	participantStorage *storage.ParticipantStorage  // Add this line
	authInterceptor   *auth.TournamentAuthInterceptor
	logger           *zap.Logger
}
```

2. Update constructor to accept participant storage:
```go
func NewTournamentService(
	storage *storage.TournamentStorage,
	participantStorage *storage.ParticipantStorage,  // Add this parameter
	authInterceptor *auth.TournamentAuthInterceptor,
	logger *zap.Logger,
) *TournamentService {
	return &TournamentService{
		storage:           storage,
		participantStorage: participantStorage,  // Add this line
		authInterceptor:   authInterceptor,
		logger:           logger,
	}
}
```

3. Add method to get tournament with participant info (enhanced tournament details):
```go
// GetTournamentWithParticipants retrieves tournament details with current participant count
func (t *TournamentService) GetTournamentWithParticipants(ctx context.Context, req *pb.GetTournamentRequest) (*pb.GetTournamentResponse, error) {
	// Use existing GetTournament logic first
	response, err := t.GetTournament(ctx, req)
	if err != nil {
		return nil, err
	}

	// Get participant count for more accurate info
	participantsReq := &pb.GetTournamentParticipantsRequest{
		Namespace:    req.GetNamespace(),
		TournamentId: req.GetTournamentId(),
		PageSize:     1,  // We only need the count
	}

	participantsResp, err := t.participantStorage.GetParticipants(ctx, participantsReq)
	if err != nil {
		t.logger.Warn("failed to get participant count for tournament",
			zap.String("tournament_id", req.GetTournamentId()),
			zap.Error(err))
		// Continue with tournament data, just log the error
	} else {
		// Update tournament's current participants with actual count
		if response.Tournament != nil {
			response.Tournament.CurrentParticipants = participantsResp.GetTotalParticipants()
		}
	}

	return response, nil
}
```

4. Add participant count validation to tournament start (enhanced safety):
```go
// StartTournamentWithValidation starts tournament with participant validation
func (t *TournamentService) StartTournamentWithValidation(ctx context.Context, req *pb.StartTournamentRequest) (*pb.StartTournamentResponse, error) {
	// Check minimum participant requirements
	participantsReq := &pb.GetTournamentParticipantsRequest{
		Namespace:    req.GetNamespace(),
		TournamentId: req.GetTournamentId(),
	}

	participantsResp, err := t.participantStorage.GetParticipants(ctx, participantsReq)
	if err != nil {
		t.logger.Error("failed to get participants for tournament start validation",
			zap.String("tournament_id", req.GetTournamentId()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to validate participants: %w", err)
	}

	if participantsResp.GetTotalParticipants() < 2 {
		return nil, fmt.Errorf("tournament requires at least 2 participants to start")
	}

	// Continue with existing StartTournament logic
	return t.StartTournament(ctx, req)
}
```

These enhancements integrate participant functionality with existing tournament operations.
  </action>
  <verify>grep -n "participantStorage" pkg/service/tournament.go && grep -n "GetTournamentWithParticipants" pkg/service/tournament.go</verify>
  <done>Tournament service enhanced with participant storage integration and validation</done>
</task>

</tasks>

<verification>
- Participant service implements all registration RPCs with proper authentication
- User context extraction follows Phase 1 patterns (GetContextUserID, GetContextNamespace)
- Admin authorization implemented for participant removal operations
- Tournament service enhanced with participant storage integration
- Logging includes user IDs (redacted where appropriate) and tournament details
- Error handling consistent with Phase 1 service patterns
- Namespace validation implemented in all methods
- Participant count validation for tournament start operations
</verification>

<success_criteria>
- Complete participant service with authentication and authorization
- Registration endpoints with user context extraction and validation
- Participant listing with public access (authenticated users only)
- Admin-only participant removal with proper permission checks
- Tournament service integration with participant count management
- Enhanced tournament details with accurate participant counts
- Comprehensive logging for audit and debugging
- Error handling and validation following existing patterns
</success_criteria>

<output>
After completion, create `.planning/phases/02-participation/02-participation-03-SUMMARY.md`
</output>