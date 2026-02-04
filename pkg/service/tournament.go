// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	extendtournamentservice "extend-tournament-service/pkg/common"
	serviceextension "extend-tournament-service/pkg/pb"
	"extend-tournament-service/pkg/storage"
)

// TournamentServiceServer implements the TournamentService gRPC service
type TournamentServiceServer struct {
	serviceextension.UnimplementedTournamentServiceServer
	tokenRepo          repository.TokenRepository
	configRepo         repository.ConfigRepository
	refreshRepo        repository.RefreshTokenRepository
	tournamentStorage  storage.TournamentStorage
	participantStorage *storage.ParticipantStorage

	authInterceptor *extendtournamentservice.TournamentAuthInterceptor
	logger          *slog.Logger
}

// TournamentStatusTransition represents a valid status transition
type TournamentStatusTransition struct {
	From serviceextension.TournamentStatus
	To   serviceextension.TournamentStatus
}

// GetAllowedStatusTransitions returns the allowed transitions for each tournament status
func (s *TournamentServiceServer) GetAllowedStatusTransitions() map[serviceextension.TournamentStatus][]serviceextension.TournamentStatus {
	return map[serviceextension.TournamentStatus][]serviceextension.TournamentStatus{
		serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT: {
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT,     // Can stay DRAFT (for updates)
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE,    // Can be activated
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED, // Can be cancelled
		},
		serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE: {
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE,    // Can stay ACTIVE (for updates)
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED,   // Can be started
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED, // Can be cancelled
		},
		serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED: {
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED,   // Can stay STARTED
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, // Can be completed
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED, // Can be cancelled
		},
		serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED: {
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, // Terminal state
		},
		serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED: {
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED, // Terminal state
		},
		serviceextension.TournamentStatus_TOURNAMENT_STATUS_UNSPECIFIED: {
			// Should not occur for created tournaments
			serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT,
		},
	}
}

// ValidateStatusTransition validates if a status transition is allowed
func (s *TournamentServiceServer) ValidateStatusTransition(from, to serviceextension.TournamentStatus) error {
	// Get allowed transitions for the current status
	allowedTransitions := s.GetAllowedStatusTransitions()

	// Check if the 'to' status is in the allowed transitions list for the 'from' status
	if allowedTo, exists := allowedTransitions[from]; exists {
		for _, status := range allowedTo {
			if status == to {
				return nil // Transition is allowed
			}
		}
	}

	return grpcStatus.Errorf(codes.InvalidArgument,
		"invalid tournament status transition from %v to %v",
		s.GetStatusName(from),
		s.GetStatusName(to))
}

// GetStatusName returns a human-readable name for tournament status
func (s *TournamentServiceServer) GetStatusName(status serviceextension.TournamentStatus) string {
	switch status {
	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT:
		return "DRAFT"
	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE:
		return "ACTIVE"
	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED:
		return "STARTED"
	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED:
		return "COMPLETED"
	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED:
		return "CANCELLED"
	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_UNSPECIFIED:
		return "UNSPECIFIED"
	default:
		return "UNKNOWN"
	}
}

// CanTransitionTo checks if a tournament can transition to the target status
func (s *TournamentServiceServer) CanTransitionTo(current, target serviceextension.TournamentStatus) bool {
	return s.ValidateStatusTransition(current, target) == nil
}

// IsTerminalStatus checks if the given status is a terminal state
func (s *TournamentServiceServer) IsTerminalStatus(status serviceextension.TournamentStatus) bool {
	return status == serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED ||
		status == serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED
}

// CanBeCancelled checks if a tournament with the given status can be cancelled
func (s *TournamentServiceServer) CanBeCancelled(status serviceextension.TournamentStatus) bool {
	return s.CanTransitionTo(status, serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED)
}

// CanBeStarted checks if a tournament with the given status can be started
func (s *TournamentServiceServer) CanBeStarted(status serviceextension.TournamentStatus) bool {
	return s.CanTransitionTo(status, serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED)
}

// TournamentParticipant represents a participant in the tournament
type TournamentParticipant struct {
	UserId      string `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// Bracket represents a single match in a tournament bracket
type Bracket struct {
	MatchId      string                 `json:"matchId"`
	Round        int32                  `json:"round"`
	Position     int32                  `json:"position"`
	Participant1 *TournamentParticipant `json:"participant1,omitempty"`
	Participant2 *TournamentParticipant `json:"participant2,omitempty"`
	Winner       string                 `json:"winner,omitempty"`
	Bye          bool                   `json:"bye"`
}

// BracketData represents the complete tournament bracket structure
type BracketData struct {
	TotalRounds int32       `json:"totalRounds"`
	Rounds      [][]Bracket `json:"rounds"`
	StartedAt   string      `json:"startedAt"`
}

// GenerateBrackets generates a single-elimination bracket for the tournament
func (s *TournamentServiceServer) GenerateBrackets(participants []TournamentParticipant) (*BracketData, error) {
	s.logger.Info("generating tournament brackets", "participant_count", len(participants))

	if len(participants) < 2 {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "at least 2 participants required for bracket generation")
	}

	// Calculate required rounds (next power of 2)
	participantCount := len(participants)
	totalRounds := int32(math.Ceil(math.Log2(float64(participantCount))))

	// Find next power of 2 for bracket size
	bracketSize := int(math.Pow(2, float64(totalRounds)))

	// Calculate number of byes needed
	byeCount := bracketSize - participantCount

	s.logger.Info("bracket calculation",
		"participants", participantCount,
		"bracket_size", bracketSize,
		"total_rounds", totalRounds,
		"bye_count", byeCount)

	// Initialize bracket rounds
	bracketData := &BracketData{
		TotalRounds: totalRounds,
		Rounds:      make([][]Bracket, totalRounds),
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	// Generate first round with participants and byes
	firstRound := make([]Bracket, bracketSize/2)

	// Shuffle participants for random seeding (for now, use order as provided)
	// In a real implementation, you might want to seed based on rankings
	currentParticipantIndex := 0

	for i := 0; i < len(firstRound); i++ {
		match := Bracket{
			MatchId:  fmt.Sprintf("match-r1-m%d", i+1),
			Round:    1,
			Position: int32(i),
			Bye:      false,
		}

		// Add first participant
		if currentParticipantIndex < len(participants) {
			participant := participants[currentParticipantIndex]
			match.Participant1 = &TournamentParticipant{
				UserId:      participant.UserId,
				Username:    participant.Username,
				DisplayName: participant.DisplayName,
			}
			currentParticipantIndex++
		}

		// Add second participant or assign bye
		if currentParticipantIndex < len(participants) {
			participant := participants[currentParticipantIndex]
			match.Participant2 = &TournamentParticipant{
				UserId:      participant.UserId,
				Username:    participant.Username,
				DisplayName: participant.DisplayName,
			}
			currentParticipantIndex++
		} else if byeCount > 0 {
			// Assign bye to participant 1
			match.Bye = true
			byeCount--
		}

		firstRound[i] = match
	}

	bracketData.Rounds[0] = firstRound

	// Generate subsequent rounds (empty slots to be filled as tournament progresses)
	for round := 1; round < int(totalRounds); round++ {
		matchesInRound := int(math.Pow(2, float64(totalRounds-int32(round)-1)))
		roundMatches := make([]Bracket, matchesInRound)

		for i := 0; i < matchesInRound; i++ {
			match := Bracket{
				MatchId:  fmt.Sprintf("match-r%d-m%d", round+1, i+1),
				Round:    int32(round + 1),
				Position: int32(i),
				Bye:      false,
			}
			roundMatches[i] = match
		}

		bracketData.Rounds[round] = roundMatches
	}

	s.logger.Info("bracket generation completed",
		"total_rounds", bracketData.TotalRounds,
		"first_round_matches", len(bracketData.Rounds[0]))

	return bracketData, nil
}

// CompleteTournament completes a tournament with winner declaration
func (s *TournamentServiceServer) CompleteTournament(ctx context.Context, namespace, tournamentID, winnerUserID string) (*serviceextension.Tournament, error) {
	s.logger.Info("CompleteTournament called", "namespace", namespace, "tournament_id", tournamentID, "winner", winnerUserID)

	// Validate required fields
	if namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if tournamentID == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}

	// Check admin permissions for tournament completion
	permission := s.authInterceptor.GetTournamentPermission("UPDATE", namespace)
	if err := s.authInterceptor.CheckTournamentPermission(ctx, permission, namespace); err != nil {
		s.logger.Warn("complete tournament permission denied", "error", err, "namespace", namespace, "tournament_id", tournamentID)
		return nil, err
	}

	// Get current tournament to validate status
	tournament, err := s.tournamentStorage.GetTournament(ctx, namespace, tournamentID)
	if err != nil {
		s.logger.Error("failed to get tournament for completion", "error", err, "namespace", namespace, "tournament_id", tournamentID)
		return nil, err
	}

	// Validate status transition using centralized validation
	newStatus := serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED
	if err := s.ValidateStatusTransition(tournament.Status, newStatus); err != nil {
		s.logger.Warn("invalid tournament status transition for completion",
			"error", err,
			"tournament_id", tournamentID,
			"current_status", s.GetStatusName(tournament.Status),
			"target_status", s.GetStatusName(newStatus))
		return nil, err
	}

	// Store previous status for logging
	previousStatus := tournament.Status

	// Update tournament status to completed and set winner
	tournament.Status = newStatus
	tournament.UpdatedAt = timestamppb.New(time.Now())
	// TODO: Add winner field to Tournament protobuf when available
	// For now, we'll log the winner but can't store it in the tournament record

	// Update tournament in storage
	updatedTournament, err := s.tournamentStorage.UpdateTournament(ctx, namespace, tournamentID, tournament)
	if err != nil {
		s.logger.Error("failed to complete tournament", "error", err, "namespace", namespace, "tournament_id", tournamentID)
		return nil, err
	}

	// Log status change for auditing
	s.LogStatusChange(ctx, updatedTournament.TournamentId, namespace, previousStatus, newStatus, "Tournament completed")

	s.logger.Info("tournament completed successfully",
		"tournament_id", updatedTournament.TournamentId,
		"namespace", namespace,
		"name", updatedTournament.Name,
		"previous_status", s.GetStatusName(previousStatus),
		"winner_user_id", winnerUserID)

	// Log that this is a terminal state
	s.logger.Info("tournament reached terminal state",
		"tournament_id", updatedTournament.TournamentId,
		"namespace", namespace,
		"terminal_status", s.GetStatusName(newStatus),
		"winner", winnerUserID)

	return updatedTournament, nil
}

// CompleteTournamentByAdmin completes a tournament via admin request (changes from STARTED to COMPLETED)
func (s *TournamentServiceServer) CompleteTournamentByAdmin(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	s.logger.Info("CompleteTournamentByAdmin called", "namespace", req.Namespace, "tournament_id", req.TournamentId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}

	// Use the actual completion logic without requiring winner since it's admin action
	updatedTournament, err := s.CompleteTournament(ctx, req.Namespace, req.TournamentId, "")
	if err != nil {
		return nil, err
	}

	return &serviceextension.StartTournamentResponse{
		Tournament: updatedTournament,
	}, nil
}

// LogStatusChange logs a tournament status change for auditing
func (s *TournamentServiceServer) LogStatusChange(ctx context.Context, tournamentID, namespace string, from, to serviceextension.TournamentStatus, reason string) {
	s.logger.Info("tournament status changed",
		"tournament_id", tournamentID,
		"namespace", namespace,
		"from_status", s.GetStatusName(from),
		"to_status", s.GetStatusName(to),
		"reason", reason,
		"timestamp", time.Now().UTC())
}

// NewTournamentServiceServer creates a new TournamentServiceServer instance
func NewTournamentServiceServer(
	tokenRepo repository.TokenRepository,
	configRepo repository.ConfigRepository,
	refreshRepo repository.RefreshTokenRepository,
	tournamentStorage storage.TournamentStorage,
	participantStorage *storage.ParticipantStorage,
	authInterceptor *extendtournamentservice.TournamentAuthInterceptor,
	logger *slog.Logger,
) *TournamentServiceServer {
	return &TournamentServiceServer{
		tokenRepo:          tokenRepo,
		configRepo:         configRepo,
		refreshRepo:        refreshRepo,
		tournamentStorage:  tournamentStorage,
		participantStorage: participantStorage,

		authInterceptor: authInterceptor,
		logger:          logger,
	}
}

// CreateTournament creates a new tournament
func (s *TournamentServiceServer) CreateTournament(ctx context.Context, req *serviceextension.CreateTournamentRequest) (*serviceextension.CreateTournamentResponse, error) {
	s.logger.Info("CreateTournament called", "namespace", req.Namespace, "name", req.Name)

	// Validate required fields
	if req.Name == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament name is required")
	}
	if req.MaxParticipants <= 0 {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "max_participants must be greater than 0")
	}
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}

	// Validate time range if both are provided
	if !req.StartTime.AsTime().IsZero() && !req.EndTime.AsTime().IsZero() {
		if req.StartTime.AsTime().After(req.EndTime.AsTime()) {
			return nil, grpcStatus.Errorf(codes.InvalidArgument, "start_time cannot be after end_time")
		}
	}

	// Check admin permissions for tournament creation
	permission := s.authInterceptor.GetTournamentPermission("CREATE", req.Namespace)
	if err := s.authInterceptor.CheckTournamentPermission(ctx, permission, req.Namespace); err != nil {
		s.logger.Warn("create tournament permission denied", "error", err, "namespace", req.Namespace)
		return nil, err
	}

	// Create tournament object
	tournament := &serviceextension.Tournament{
		Name:            req.Name,
		Description:     req.Description,
		MaxParticipants: req.MaxParticipants,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
	}

	// Store tournament
	createdTournament, err := s.tournamentStorage.CreateTournament(ctx, req.Namespace, tournament)
	if err != nil {
		s.logger.Error("failed to create tournament", "error", err, "namespace", req.Namespace, "name", req.Name)
		return nil, err
	}

	s.logger.Info("tournament created successfully",
		"tournament_id", createdTournament.TournamentId,
		"namespace", req.Namespace,
		"name", createdTournament.Name)

	return &serviceextension.CreateTournamentResponse{
		Tournament: createdTournament,
	}, nil
}

// ListTournaments lists tournaments with filtering and pagination
func (s *TournamentServiceServer) ListTournaments(ctx context.Context, req *serviceextension.ListTournamentsRequest) (*serviceextension.ListTournamentsResponse, error) {
	s.logger.Info("ListTournaments called", "namespace", req.Namespace, "limit", req.Limit, "offset", req.Offset)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}

	// Set default pagination values
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// No permission check for public read access

	// Get tournaments from storage
	tournaments, totalCount, err := s.tournamentStorage.ListTournaments(ctx, req.Namespace, limit, offset, req.Status)
	if err != nil {
		s.logger.Error("failed to list tournaments", "error", err, "namespace", req.Namespace)
		return nil, err
	}

	s.logger.Info("tournaments listed successfully",
		"namespace", req.Namespace,
		"count", len(tournaments),
		"total_count", totalCount)

	return &serviceextension.ListTournamentsResponse{
		Tournaments: tournaments,
		TotalCount:  totalCount,
	}, nil
}

// GetTournament retrieves a specific tournament by ID
func (s *TournamentServiceServer) GetTournament(ctx context.Context, req *serviceextension.GetTournamentRequest) (*serviceextension.GetTournamentResponse, error) {
	s.logger.Info("GetTournament called", "namespace", req.Namespace, "tournament_id", req.TournamentId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}

	// No permission check for public read access

	// Get tournament from storage
	tournament, err := s.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
	if err != nil {
		s.logger.Error("failed to get tournament", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	s.logger.Info("tournament retrieved successfully",
		"tournament_id", tournament.TournamentId,
		"namespace", req.Namespace,
		"name", tournament.Name)

	return &serviceextension.GetTournamentResponse{
		Tournament: tournament,
	}, nil
}

// CancelTournament cancels a tournament
func (s *TournamentServiceServer) CancelTournament(ctx context.Context, req *serviceextension.CancelTournamentRequest) (*serviceextension.CancelTournamentResponse, error) {
	s.logger.Info("CancelTournament called", "namespace", req.Namespace, "tournament_id", req.TournamentId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}

	// Check admin permissions for tournament cancellation
	permission := s.authInterceptor.GetTournamentPermission("CANCEL", req.Namespace)
	if err := s.authInterceptor.CheckTournamentPermission(ctx, permission, req.Namespace); err != nil {
		s.logger.Warn("cancel tournament permission denied", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Get current tournament to validate status
	tournament, err := s.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
	if err != nil {
		s.logger.Error("failed to get tournament for cancellation", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Validate status transition using centralized validation
	newStatus := serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED
	if err := s.ValidateStatusTransition(tournament.Status, newStatus); err != nil {
		s.logger.Warn("invalid tournament status transition for cancellation",
			"error", err,
			"tournament_id", req.TournamentId,
			"current_status", s.GetStatusName(tournament.Status),
			"target_status", s.GetStatusName(newStatus))
		return nil, err
	}

	// Store previous status for logging
	previousStatus := tournament.Status

	// Update tournament status to cancelled
	tournament.Status = newStatus
	tournament.UpdatedAt = timestamppb.New(time.Now())

	// Update tournament in storage
	updatedTournament, err := s.tournamentStorage.UpdateTournament(ctx, req.Namespace, req.TournamentId, tournament)
	if err != nil {
		s.logger.Error("failed to cancel tournament", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Log status change for auditing
	s.LogStatusChange(ctx, updatedTournament.TournamentId, req.Namespace, previousStatus, newStatus, "Tournament cancelled by admin")

	s.logger.Info("tournament cancelled successfully",
		"tournament_id", updatedTournament.TournamentId,
		"namespace", req.Namespace,
		"name", updatedTournament.Name,
		"previous_status", s.GetStatusName(previousStatus))

	return &serviceextension.CancelTournamentResponse{
		Tournament: updatedTournament,
	}, nil
}

// ActivateTournament activates a tournament (changes from DRAFT to ACTIVE)
func (s *TournamentServiceServer) ActivateTournament(ctx context.Context, req *serviceextension.ActivateTournamentRequest) (*serviceextension.ActivateTournamentResponse, error) {
	s.logger.Info("ActivateTournament called", "namespace", req.Namespace, "tournament_id", req.TournamentId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}

	// Check admin permissions for tournament activation
	permission := s.authInterceptor.GetTournamentPermission("UPDATE", req.Namespace)
	if err := s.authInterceptor.CheckTournamentPermission(ctx, permission, req.Namespace); err != nil {
		s.logger.Warn("activate tournament permission denied", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Get current tournament to validate status
	tournament, err := s.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
	if err != nil {
		s.logger.Error("failed to get tournament for activation", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Validate status transition using centralized validation
	newStatus := serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE
	if err := s.ValidateStatusTransition(tournament.Status, newStatus); err != nil {
		s.logger.Warn("invalid tournament status transition for activation",
			"error", err,
			"tournament_id", req.TournamentId,
			"current_status", s.GetStatusName(tournament.Status),
			"target_status", s.GetStatusName(newStatus))
		return nil, err
	}

	// Store previous status for logging
	previousStatus := tournament.Status

	// Update tournament status to active
	tournament.Status = newStatus
	tournament.UpdatedAt = timestamppb.New(time.Now())

	// Update tournament in storage
	updatedTournament, err := s.tournamentStorage.UpdateTournament(ctx, req.Namespace, req.TournamentId, tournament)
	if err != nil {
		s.logger.Error("failed to activate tournament", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Log status change for auditing
	s.LogStatusChange(ctx, updatedTournament.TournamentId, req.Namespace, previousStatus, newStatus, "Tournament activated by admin")

	s.logger.Info("tournament activated successfully",
		"tournament_id", updatedTournament.TournamentId,
		"namespace", req.Namespace,
		"name", updatedTournament.Name,
		"previous_status", s.GetStatusName(previousStatus))

	return &serviceextension.ActivateTournamentResponse{
		Tournament: updatedTournament,
	}, nil
}

// StartTournament starts a tournament
func (s *TournamentServiceServer) StartTournament(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	s.logger.Info("StartTournament called", "namespace", req.Namespace, "tournament_id", req.TournamentId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}

	// Check admin permissions for tournament start
	permission := s.authInterceptor.GetTournamentPermission("START", req.Namespace)
	if err := s.authInterceptor.CheckTournamentPermission(ctx, permission, req.Namespace); err != nil {
		s.logger.Warn("start tournament permission denied", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Get current tournament to validate status
	tournament, err := s.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
	if err != nil {
		s.logger.Error("failed to get tournament for starting", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Validate status transition using centralized validation
	newStatus := serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED
	if err := s.ValidateStatusTransition(tournament.Status, newStatus); err != nil {
		s.logger.Warn("invalid tournament status transition for starting",
			"error", err,
			"tournament_id", req.TournamentId,
			"current_status", s.GetStatusName(tournament.Status),
			"target_status", s.GetStatusName(newStatus))
		return nil, err
	}

	// Validate tournament has sufficient participants for bracket generation
	if tournament.CurrentParticipants < 2 {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "at least 2 participants required to start tournament (current: %d)", tournament.CurrentParticipants)
	}

	// Store previous status for logging
	previousStatus := tournament.Status

	// Generate brackets before changing status
	s.logger.Info("generating brackets for tournament start", "tournament_id", req.TournamentId, "participants", tournament.CurrentParticipants)

	// Note: Match creation is now handled at server level in enhanced StartTournament method
	// This method only handles to basic status transition now
	// Get participants for validation
	participantsReq := &serviceextension.GetTournamentParticipantsRequest{
		Namespace:    req.Namespace,
		TournamentId: req.TournamentId,
		PageSize:     100, // Get all participants
	}

	participantsResp, err := s.participantStorage.GetParticipants(ctx, participantsReq)
	if err != nil {
		s.logger.Error("failed to get participants for bracket generation", "error", err, "tournament_id", req.TournamentId)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to get tournament participants: %v", err)
	}

	if len(participantsResp.Participants) < 2 {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "at least 2 participants required to start tournament (current: %d)", len(participantsResp.Participants))
	}

	// Convert registered participants to TournamentParticipant format for bracket generation
	tournamentParticipants := make([]TournamentParticipant, len(participantsResp.Participants))
	for i, participant := range participantsResp.Participants {
		tournamentParticipants[i] = TournamentParticipant{
			UserId:      participant.UserId,
			Username:    participant.Username,
			DisplayName: participant.DisplayName,
		}
	}

	// Generate bracket structure
	bracketData, err := s.GenerateBrackets(tournamentParticipants)
	if err != nil {
		s.logger.Error("failed to generate brackets", "error", err, "tournament_id", req.TournamentId)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to generate tournament brackets: %v", err)
	}

	// Convert bracket data to protobuf Match objects for storage
	var allMatches []*serviceextension.Match
	for roundIdx, round := range bracketData.Rounds {
		for matchIdx, bracket := range round {
			match := &serviceextension.Match{
				MatchId:   fmt.Sprintf("match-r%d-m%d", roundIdx+1, matchIdx+1),
				Round:     int32(roundIdx + 1),
				Position:  int32(matchIdx),
				Status:    serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED,
				StartedAt: timestamppb.Now(),
			}

			// Set next_match_id (where the winner advances to)
			if roundIdx < len(bracketData.Rounds)-1 {
				nextMatchPos := matchIdx/2 + 1
				match.NextMatchId = fmt.Sprintf("match-r%d-m%d", roundIdx+2, nextMatchPos)
			}

			// Set source match IDs (which matches feed into this one)
			if roundIdx > 0 {
				sourcePos1 := matchIdx*2 + 1
				sourcePos2 := matchIdx*2 + 2
				match.SourceMatch_1Id = fmt.Sprintf("match-r%d-m%d", roundIdx, sourcePos1)
				match.SourceMatch_2Id = fmt.Sprintf("match-r%d-m%d", roundIdx, sourcePos2)
			}

			// Add participant 1 if exists
			if bracket.Participant1 != nil {
				match.Participant1 = &serviceextension.TournamentParticipant{
					UserId:      bracket.Participant1.UserId,
					Username:    bracket.Participant1.Username,
					DisplayName: bracket.Participant1.DisplayName,
				}
			}

			// Add participant 2 if exists
			if bracket.Participant2 != nil {
				match.Participant2 = &serviceextension.TournamentParticipant{
					UserId:      bracket.Participant2.UserId,
					Username:    bracket.Participant2.Username,
					DisplayName: bracket.Participant2.DisplayName,
				}
			}

			// Handle bye participants (automatic advancement)
			if bracket.Bye && match.Participant1 != nil {
				match.Winner = match.Participant1.UserId
				match.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
				match.CompletedAt = timestamppb.Now()
			}

			allMatches = append(allMatches, match)
		}
	}

	// TODO: Create all matches in storage using MatchService
	// This will be implemented when MatchService is properly integrated
	// For now, match creation needs to be handled at the server level
	if len(allMatches) > 0 {
		s.logger.Info("tournament matches ready to be created",
			"tournament_id", req.TournamentId,
			"total_rounds", bracketData.TotalRounds,
			"first_round_matches", len(bracketData.Rounds[0]),
			"total_matches", len(allMatches))
	}

	// Update tournament status to started
	tournament.Status = newStatus
	tournament.UpdatedAt = timestamppb.New(time.Now())

	// Update tournament in storage
	updatedTournament, err := s.tournamentStorage.UpdateTournament(ctx, req.Namespace, req.TournamentId, tournament)
	if err != nil {
		s.logger.Error("failed to start tournament", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Log status change for auditing
	s.LogStatusChange(ctx, updatedTournament.TournamentId, req.Namespace, previousStatus, newStatus, "Tournament started by admin")

	s.logger.Info("tournament started successfully",
		"tournament_id", updatedTournament.TournamentId,
		"namespace", req.Namespace,
		"name", updatedTournament.Name,
		"previous_status", s.GetStatusName(previousStatus))

	return &serviceextension.StartTournamentResponse{
		Tournament: updatedTournament,
	}, nil
}

// GetTournamentWithParticipants retrieves tournament details with current participant count
func (s *TournamentServiceServer) GetTournamentWithParticipants(ctx context.Context, req *serviceextension.GetTournamentRequest) (*serviceextension.GetTournamentResponse, error) {
	// Use existing GetTournament logic first
	response, err := s.GetTournament(ctx, req)
	if err != nil {
		return nil, err
	}

	// Get participant count for more accurate info
	participantsReq := &serviceextension.GetTournamentParticipantsRequest{
		Namespace:    req.GetNamespace(),
		TournamentId: req.GetTournamentId(),
		PageSize:     1, // We only need count
	}

	participantsResp, err := s.participantStorage.GetParticipants(ctx, participantsReq)
	if err != nil {
		s.logger.Warn("failed to get participant count for tournament",
			"tournament_id", req.GetTournamentId(),
			"error", err)
		// Continue with tournament data, just log the error
	} else {
		// Update tournament's current participants with actual count
		if response.Tournament != nil {
			response.Tournament.CurrentParticipants = participantsResp.GetTotalParticipants()
		}
	}

	return response, nil
}

// StartTournamentWithValidation starts tournament with participant validation
func (s *TournamentServiceServer) StartTournamentWithValidation(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	// Check minimum participant requirements
	participantsReq := &serviceextension.GetTournamentParticipantsRequest{
		Namespace:    req.GetNamespace(),
		TournamentId: req.GetTournamentId(),
	}

	participantsResp, err := s.participantStorage.GetParticipants(ctx, participantsReq)
	if err != nil {
		s.logger.Error("failed to get participants for tournament start validation",
			"tournament_id", req.GetTournamentId(),
			"error", err)
		return nil, fmt.Errorf("failed to validate participants: %w", err)
	}

	if participantsResp.GetTotalParticipants() < 2 {
		return nil, fmt.Errorf("tournament requires at least 2 participants to start")
	}

	// Continue with existing StartTournament logic
	return s.StartTournament(ctx, req)
}
