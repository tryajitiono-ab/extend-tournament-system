// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	extendcustomguildservice "extend-custom-guild-service/pkg/common"
	serviceextension "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/storage"
)

// MatchService implements match management business logic
type MatchService struct {
	matchStorage      storage.MatchStorage
	tournamentStorage storage.TournamentStorage
	authInterceptor   *extendcustomguildservice.TournamentAuthInterceptor
	logger            *slog.Logger
}

// NewMatchService creates a new match service instance
func NewMatchService(
	matchStorage storage.MatchStorage,
	tournamentStorage storage.TournamentStorage,
	authInterceptor *extendcustomguildservice.TournamentAuthInterceptor,
	logger *slog.Logger,
) *MatchService {
	return &MatchService{
		matchStorage:      matchStorage,
		tournamentStorage: tournamentStorage,
		authInterceptor:   authInterceptor,
		logger:            logger,
	}
}

// validateMatchWinner validates that the winner is one of the participants
func (m *MatchService) validateMatchWinner(match *serviceextension.Match, winnerUserID string) error {
	if winnerUserID == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, "winner_user_id is required")
	}

	// Check if winner matches participant1
	if match.Participant1 != nil && match.Participant1.UserId == winnerUserID {
		return nil
	}

	// Check if winner matches participant2
	if match.Participant2 != nil && match.Participant2.UserId == winnerUserID {
		return nil
	}

	return grpcStatus.Errorf(codes.InvalidArgument, "winner %s is not a participant in match %s", winnerUserID, match.MatchId)
}

// calculateNextPosition calculates next round position based on current position
// Formula: nextPosition = (currentPosition - 1) / 2 + 1
func calculateNextPosition(currentPos int32) int32 {
	return (currentPos-1)/2 + 1
}

// calculateTournamentProgress calculates total rounds and current round based on matches
func (m *MatchService) calculateTournamentProgress(matches []*serviceextension.Match) (int32, int32) {
	if len(matches) == 0 {
		return 0, 0
	}

	// Find highest round number to determine total rounds
	maxRound := int32(0)
	currentRound := int32(0)
	hasActiveMatches := false

	for _, match := range matches {
		if match.Round > maxRound {
			maxRound = match.Round
		}

		// Check if this round has any in-progress or scheduled matches
		// to determine the current active round
		if match.Status == serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED ||
			match.Status == serviceextension.MatchStatus_MATCH_STATUS_IN_PROGRESS {
			hasActiveMatches = true
			if currentRound == 0 || match.Round < currentRound {
				currentRound = match.Round
			}
		}
	}

	// For integration tests: if we have completed matches in a round and no active matches in next round,
	// the current round should still be the completed round
	if !hasActiveMatches && currentRound == maxRound {
		// No active matches found, current round is the highest round with matches
	} else {
		// If all matches are completed, current round is the last round
		if currentRound == 0 {
			currentRound = maxRound
		}
	}

	return maxRound, currentRound
}

// advanceWinner advances winner to next round match
func (m *MatchService) advanceWinner(ctx context.Context, namespace string, match *serviceextension.Match) error {
	// Calculate next round and position
	nextRound := match.Round + 1
	nextPosition := calculateNextPosition(match.Position)

	m.logger.Info("advancing winner to next round",
		"match_id", match.MatchId,
		"winner", match.Winner,
		"from_round", match.Round,
		"from_position", match.Position,
		"to_round", nextRound,
		"to_position", nextPosition)

	// Find the winner participant from the current match
	var winnerParticipant *serviceextension.TournamentParticipant
	if match.Participant1 != nil && match.Participant1.UserId == match.Winner {
		winnerParticipant = match.Participant1
	} else if match.Participant2 != nil && match.Participant2.UserId == match.Winner {
		winnerParticipant = match.Participant2
	}

	if winnerParticipant == nil {
		return grpcStatus.Errorf(codes.Internal, "winner participant not found in match %s", match.MatchId)
	}

	// Get next round matches to find where to place the winner
	nextRoundMatches, err := m.matchStorage.GetMatchesByRound(ctx, namespace, match.TournamentId, nextRound)
	if err != nil {
		m.logger.Error("failed to get next round matches", "error", err, "next_round", nextRound)
		return grpcStatus.Errorf(codes.Internal, "failed to get next round matches: %v", err)
	}

	// Find the match at the calculated next position
	var nextRoundMatch *serviceextension.Match
	for _, nextMatch := range nextRoundMatches {
		if nextMatch.Position == nextPosition {
			nextRoundMatch = nextMatch
			break
		}
	}

	// If no match exists for the next position, this might be the final round
	if nextRoundMatch == nil {
		m.logger.Info("no next round match found, tournament might be in final round",
			"match_id", match.MatchId,
			"current_round", match.Round,
			"next_round", nextRound,
			"next_position", nextPosition)
		return nil
	}

	// Update the next round match with the advancing participant
	updated := false
	if nextRoundMatch.Participant1 == nil {
		nextRoundMatch.Participant1 = winnerParticipant
		updated = true
	} else if nextRoundMatch.Participant2 == nil {
		nextRoundMatch.Participant2 = winnerParticipant
		updated = true
	}

	if !updated {
		return grpcStatus.Errorf(codes.Internal, "next round match %s already has both participants", nextRoundMatch.MatchId)
	}

	// Save the updated next round match
	if err := m.matchStorage.UpdateMatch(ctx, match.TournamentId, nextRoundMatch); err != nil {
		m.logger.Error("failed to update next round match with advancing participant",
			"error", err,
			"next_match_id", nextRoundMatch.MatchId,
			"winner_user_id", match.Winner)
		return grpcStatus.Errorf(codes.Internal, "failed to update next round match: %v", err)
	}

	m.logger.Info("winner advanced successfully",
		"match_id", match.MatchId,
		"winner", match.Winner,
		"next_match_id", nextRoundMatch.MatchId,
		"next_round", nextRound,
		"next_position", nextPosition)

	return nil
}

// GetTournamentMatches retrieves all matches for a tournament
func (m *MatchService) GetTournamentMatches(ctx context.Context, req *serviceextension.GetTournamentMatchesRequest) (*serviceextension.GetTournamentMatchesResponse, error) {
	m.logger.Info("GetTournamentMatches called", "namespace", req.Namespace, "tournament_id", req.TournamentId, "round", req.Round)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}

	// Verify tournament exists
	if m.tournamentStorage != nil {
		_, err := m.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
		if err != nil {
			m.logger.Error("failed to verify tournament exists", "error", err, "tournament_id", req.TournamentId)
			return nil, err
		}
	}

	var matches []*serviceextension.Match
	var err error

	if req.Round > 0 {
		// Get matches for specific round
		matches, err = m.matchStorage.GetMatchesByRound(ctx, req.Namespace, req.TournamentId, req.Round)
		if err != nil {
			m.logger.Error("failed to get matches by round", "error", err, "round", req.Round)
			return nil, err
		}
	} else {
		// Get all tournament matches
		matches, err = m.matchStorage.GetTournamentMatches(ctx, req.Namespace, req.TournamentId)
		if err != nil {
			m.logger.Error("failed to get tournament matches", "error", err, "tournament_id", req.TournamentId)
			return nil, err
		}
	}

	m.logger.Info("tournament matches retrieved successfully", "tournament_id", req.TournamentId, "count", len(matches))

	// Calculate total rounds and current round based on matches
	totalRounds, currentRound := m.calculateTournamentProgress(matches)

	return &serviceextension.GetTournamentMatchesResponse{
		Matches:      matches,
		TotalRounds:  totalRounds,
		CurrentRound: currentRound,
	}, nil
}

// GetMatch retrieves a specific match by ID
func (m *MatchService) GetMatch(ctx context.Context, req *serviceextension.GetMatchRequest) (*serviceextension.GetMatchResponse, error) {
	m.logger.Info("GetMatch called", "namespace", req.Namespace, "tournament_id", req.TournamentId, "match_id", req.MatchId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}
	if req.MatchId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "match_id is required")
	}

	// Verify tournament exists
	if m.tournamentStorage != nil {
		_, err := m.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
		if err != nil {
			m.logger.Error("failed to verify tournament exists for match lookup", "error", err, "tournament_id", req.TournamentId)
			return nil, err
		}
	}

	// Get match from storage
	match, err := m.matchStorage.GetMatch(ctx, req.Namespace, req.TournamentId, req.MatchId)
	if err != nil {
		m.logger.Error("failed to get match", "error", err, "match_id", req.MatchId)
		return nil, err
	}

	m.logger.Info("match retrieved successfully", "match_id", req.MatchId, "tournament_id", req.TournamentId)

	return &serviceextension.GetMatchResponse{
		Match: match,
	}, nil
}

// SubmitMatchResult submits a match result (game server)
func (m *MatchService) SubmitMatchResult(ctx context.Context, req *serviceextension.SubmitMatchResultRequest) (*serviceextension.SubmitMatchResultResponse, error) {
	m.logger.Info("SubmitMatchResult called", "namespace", req.Namespace, "tournament_id", req.TournamentId, "match_id", req.MatchId, "winner_user_id", req.WinnerUserId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}
	if req.MatchId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "match_id is required")
	}
	if req.WinnerUserId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "winner_user_id is required")
	}

	// Check game server permissions (service token authentication)
	if m.authInterceptor != nil {
		permission := m.authInterceptor.GetTournamentPermission("UPDATE", req.Namespace)
		if err := m.authInterceptor.CheckTournamentPermission(ctx, permission, req.Namespace); err != nil {
			m.logger.Warn("submit match result permission denied", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
			return nil, err
		}
	}

	// Verify tournament exists
	if m.tournamentStorage != nil {
		_, err := m.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
		if err != nil {
			m.logger.Error("failed to verify tournament exists", "error", err, "tournament_id", req.TournamentId)
			return nil, err
		}
	}

	// Submit result with transaction safety
	match, err := m.matchStorage.SubmitMatchResult(ctx, req.Namespace, req.TournamentId, req.MatchId, req.WinnerUserId)
	if err != nil {
		m.logger.Error("failed to submit match result", "error", err, "match_id", req.MatchId, "winner_user_id", req.WinnerUserId)
		return nil, err
	}

	// Advance winner to next round
	if err := m.advanceWinner(ctx, req.Namespace, match); err != nil {
		m.logger.Error("failed to advance winner", "error", err, "match_id", match.MatchId, "winner", match.Winner)
		// Don't fail the result submission if advancement fails, just log the error
	}

	m.logger.Info("match result submitted successfully",
		"match_id", req.MatchId,
		"tournament_id", req.TournamentId,
		"winner_user_id", req.WinnerUserId)

	return &serviceextension.SubmitMatchResultResponse{
		Match: match,
	}, nil
}

// AdminSubmitMatchResult submits a match result (admin override)
func (m *MatchService) AdminSubmitMatchResult(ctx context.Context, req *serviceextension.AdminSubmitMatchResultRequest) (*serviceextension.AdminSubmitMatchResultResponse, error) {
	m.logger.Info("AdminSubmitMatchResult called", "namespace", req.Namespace, "tournament_id", req.TournamentId, "match_id", req.MatchId, "winner_user_id", req.WinnerUserId)

	// Validate required fields
	if req.Namespace == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "namespace is required")
	}
	if req.TournamentId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "tournament_id is required")
	}
	if req.MatchId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "match_id is required")
	}
	if req.WinnerUserId == "" {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "winner_user_id is required")
	}

	// Check admin permissions (bearer token authentication)
	if m.authInterceptor != nil {
		permission := m.authInterceptor.GetTournamentPermission("UPDATE", req.Namespace)
		if err := m.authInterceptor.CheckTournamentPermission(ctx, permission, req.Namespace); err != nil {
			m.logger.Warn("admin submit match result permission denied", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
			return nil, err
		}
	}

	// Verify tournament exists
	if m.tournamentStorage != nil {
		_, err := m.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
		if err != nil {
			m.logger.Error("failed to verify tournament exists for admin submission", "error", err, "tournament_id", req.TournamentId)
			return nil, err
		}
	}

	// Submit result with transaction safety (same as game server)
	match, err := m.matchStorage.SubmitMatchResult(ctx, req.Namespace, req.TournamentId, req.MatchId, req.WinnerUserId)
	if err != nil {
		m.logger.Error("failed to submit admin match result", "error", err, "match_id", req.MatchId, "winner_user_id", req.WinnerUserId)
		return nil, err
	}

	// Advance winner to next round
	if err := m.advanceWinner(ctx, req.Namespace, match); err != nil {
		m.logger.Error("failed to advance winner after admin submission", "error", err, "match_id", match.MatchId, "winner", match.Winner)
		// Don't fail the result submission if advancement fails, just log the error
	}

	m.logger.Info("admin match result submitted successfully",
		"match_id", req.MatchId,
		"tournament_id", req.TournamentId,
		"winner_user_id", req.WinnerUserId,
		"admin_override", true)

	return &serviceextension.AdminSubmitMatchResultResponse{
		Match: match,
	}, nil
}
