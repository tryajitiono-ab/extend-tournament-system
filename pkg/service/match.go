// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	extendtournamentservice "extend-tournament-service/pkg/common"
	serviceextension "extend-tournament-service/pkg/pb"
	"extend-tournament-service/pkg/storage"
)

// Error constants for match service operations
const (
	errWinnerRequired        = "winner_user_id is required"
	errMatchAlreadyCompleted = "match %s is already completed"
	errMatchCancelled        = "match %s is cancelled"
	errNotParticipant        = "winner %s is not a participant in match %s"
	errMatchNotFound         = "match not found: %s"
	errTournamentNotFound    = "tournament not found: %s"
	errPermissionDenied      = "permission denied for tournament operation"
)

// MatchService implements match management business logic
type MatchService struct {
	matchStorage      storage.MatchStorage
	tournamentStorage storage.TournamentStorage
	authInterceptor   *extendtournamentservice.TournamentAuthInterceptor
	logger            *slog.Logger
}

// NewMatchService creates a new match service instance
func NewMatchService(
	matchStorage storage.MatchStorage,
	tournamentStorage storage.TournamentStorage,
	authInterceptor *extendtournamentservice.TournamentAuthInterceptor,
	logger *slog.Logger,
) *MatchService {
	return &MatchService{
		matchStorage:      matchStorage,
		tournamentStorage: tournamentStorage,
		authInterceptor:   authInterceptor,
		logger:            logger,
	}
}

// CreateTournamentMatches creates all matches for a tournament from generated bracket data
func (m *MatchService) CreateTournamentMatches(ctx context.Context, namespace, tournamentID string, matches []*serviceextension.Match) error {
	m.logger.Info("creating tournament matches", "tournament_id", tournamentID, "match_count", len(matches))

	// Create all matches in storage using bulk insert
	err := m.matchStorage.CreateMatches(ctx, namespace, tournamentID, matches)
	if err != nil {
		m.logger.Error("failed to create tournament matches", "error", err, "tournament_id", tournamentID)
		return grpcStatus.Errorf(codes.Internal, "failed to create tournament matches: %v", err)
	}

	m.logger.Info("tournament matches created successfully", "tournament_id", tournamentID, "match_count", len(matches))
	return nil
}

// validateMatchWinner validates that the winner is one of the participants
// validateMatchWinner validates that winner is one of the participants
// Returns nil if winner is valid participant in non-completed/cancelled match
// Returns error if winner is empty, not a participant, or match is finished
func (m *MatchService) validateMatchWinner(match *serviceextension.Match, winnerUserID string) error {
	if winnerUserID == "" {
		return grpcStatus.Errorf(codes.InvalidArgument, errWinnerRequired)
	}

	// Check if match is already completed
	if match.Status == serviceextension.MatchStatus_MATCH_STATUS_COMPLETED {
		return grpcStatus.Errorf(codes.FailedPrecondition, errMatchAlreadyCompleted, match.MatchId)
	}

	// Check if match is cancelled
	if match.Status == serviceextension.MatchStatus_MATCH_STATUS_CANCELLED {
		return grpcStatus.Errorf(codes.FailedPrecondition, errMatchCancelled, match.MatchId)
	}

	// Check if winner matches participant1
	if match.Participant1 != nil && match.Participant1.UserId == winnerUserID {
		return nil
	}

	// Check if winner matches participant2
	if match.Participant2 != nil && match.Participant2.UserId == winnerUserID {
		return nil
	}

	return grpcStatus.Errorf(codes.InvalidArgument, errNotParticipant, winnerUserID, match.MatchId)
}

// calculateNextPosition calculates next round position based on current position
// Uses standard single-elimination bracket math:
// Position 1 & 2 -> Position 1 (next round)
// Calculate next round position for winner advancement (0-indexed positions)
// Position 0 & 1 -> Position 0 (next round)
// Position 2 & 3 -> Position 1 (next round)
// Formula: nextPosition = currentPosition / 2 (integer division)
func calculateNextPosition(currentPos int32) int32 {
	return currentPos / 2
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
	if err := m.matchStorage.UpdateMatch(ctx, namespace, nextRoundMatch); err != nil {
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

// CheckTournamentCompletion checks if a tournament is complete and returns the winner
func (m *MatchService) CheckTournamentCompletion(ctx context.Context, namespace, tournamentID string) (bool, string, error) {
	m.logger.Info("checking tournament completion", "namespace", namespace, "tournament_id", tournamentID)

	// Get all matches for the tournament
	matches, err := m.matchStorage.GetTournamentMatches(ctx, namespace, tournamentID)
	if err != nil {
		m.logger.Error("failed to get tournament matches for completion check", "error", err, "tournament_id", tournamentID)
		return false, "", grpcStatus.Errorf(codes.Internal, "failed to get tournament matches: %v", err)
	}

	if len(matches) == 0 {
		m.logger.Warn("no matches found for tournament completion check", "tournament_id", tournamentID)
		return false, "", nil
	}

	// Check if all matches are completed or cancelled
	allFinished := true
	var finalMatchWinner string
	maxRound := int32(0)

	// First pass: find the highest round
	for _, match := range matches {
		if match.Round > maxRound {
			maxRound = match.Round
		}
	}

	// Second pass: check completion and find final winner
	for _, match := range matches {
		// Check if match is not finished
		if match.Status != serviceextension.MatchStatus_MATCH_STATUS_COMPLETED &&
			match.Status != serviceextension.MatchStatus_MATCH_STATUS_CANCELLED {
			allFinished = false
		}

		// Track winner from the highest round (final match)
		if match.Round == maxRound && match.Status == serviceextension.MatchStatus_MATCH_STATUS_COMPLETED {
			finalMatchWinner = match.Winner
		}
	}

	if allFinished && finalMatchWinner != "" {
		m.logger.Info("tournament completion detected",
			"tournament_id", tournamentID,
			"total_matches", len(matches),
			"max_round", maxRound,
			"winner", finalMatchWinner)
		return true, finalMatchWinner, nil
	}

	if allFinished {
		m.logger.Warn("tournament matches finished but no winner found", "tournament_id", tournamentID)
		return true, "", nil
	}

	m.logger.Info("tournament not yet complete",
		"tournament_id", tournamentID,
		"total_matches", len(matches),
		"max_round", maxRound)

	return false, "", nil
}

// HandleByeAdvancement automatically advances participants with byes
func (m *MatchService) HandleByeAdvancement(ctx context.Context, namespace, tournamentID string, round int32) error {
	m.logger.Info("handling bye advancement", "namespace", namespace, "tournament_id", tournamentID, "round", round)

	// Get all matches for the specified round
	matches, err := m.matchStorage.GetMatchesByRound(ctx, namespace, tournamentID, round)
	if err != nil {
		m.logger.Error("failed to get matches for bye advancement", "error", err, "round", round)
		return grpcStatus.Errorf(codes.Internal, "failed to get matches: %v", err)
	}

	if len(matches) == 0 {
		m.logger.Info("no matches found for bye advancement", "round", round)
		return nil
	}

	// Process each match for bye advancement
	for _, match := range matches {
		// Check if this match has a bye (only one participant)
		if (match.Participant1 != nil && match.Participant2 == nil) ||
			(match.Participant1 == nil && match.Participant2 != nil) {

			// Identify the single participant
			var soloParticipant *serviceextension.TournamentParticipant
			if match.Participant1 != nil {
				soloParticipant = match.Participant1
			} else {
				soloParticipant = match.Participant2
			}

			if soloParticipant == nil {
				m.logger.Warn("match has no participants", "match_id", match.MatchId)
				continue
			}

			m.logger.Info("advancing bye participant",
				"match_id", match.MatchId,
				"participant_user_id", soloParticipant.UserId,
				"round", round)

			// Update match with result (bye participant advances)
			match.Winner = soloParticipant.UserId
			match.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
			match.CompletedAt = timestamppb.New(time.Now())

			// Save the updated match
			if err := m.matchStorage.UpdateMatch(ctx, namespace, match); err != nil {
				m.logger.Error("failed to update bye match", "error", err, "match_id", match.MatchId)
				return grpcStatus.Errorf(codes.Internal, "failed to update bye match: %v", err)
			}

			// Advance the participant to next round
			if err := m.advanceWinner(ctx, namespace, match); err != nil {
				m.logger.Error("failed to advance bye participant to next round", "error", err, "participant_user_id", soloParticipant.UserId)
				// Don't fail the entire operation, just log the error
			}
		}
	}

	m.logger.Info("bye advancement completed successfully", "round", round, "processed_matches", len(matches))
	return nil
}

// completeTournament completes a tournament using tournament storage
func (m *MatchService) completeTournament(ctx context.Context, namespace, tournamentID, winnerUserID string) error {
	m.logger.Info("completing tournament", "namespace", namespace, "tournament_id", tournamentID, "winner", winnerUserID)

	// Get current tournament
	tournament, err := m.tournamentStorage.GetTournament(ctx, namespace, tournamentID)
	if err != nil {
		m.logger.Error("failed to get tournament for completion", "error", err, "tournament_id", tournamentID)
		return grpcStatus.Errorf(codes.Internal, "failed to get tournament: %v", err)
	}

	// Update tournament status to completed
	tournament.Status = serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED
	tournament.UpdatedAt = timestamppb.New(time.Now())
	// TODO: Add winner field to Tournament protobuf when available

	// Update tournament in storage
	_, err = m.tournamentStorage.UpdateTournament(ctx, namespace, tournamentID, tournament)
	if err != nil {
		m.logger.Error("failed to complete tournament", "error", err, "tournament_id", tournamentID)
		return grpcStatus.Errorf(codes.Internal, "failed to complete tournament: %v", err)
	}

	m.logger.Info("tournament completed successfully",
		"tournament_id", tournamentID,
		"namespace", namespace,
		"name", tournament.Name,
		"winner_user_id", winnerUserID)

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

	// Get the target match to check its round
	targetMatch, err := m.matchStorage.GetMatch(ctx, req.Namespace, req.TournamentId, req.MatchId)
	if err != nil {
		m.logger.Error("failed to get match for validation", "error", err, "match_id", req.MatchId)
		return nil, grpcStatus.Errorf(codes.NotFound, "match not found: %s", req.MatchId)
	}

	// Safeguard: Prevent submitting results for future rounds if previous rounds are incomplete
	if targetMatch.Round > 1 {
		// Check if all matches in previous rounds are completed
		for round := int32(1); round < targetMatch.Round; round++ {
			previousRoundMatches, err := m.matchStorage.GetMatchesByRound(ctx, req.Namespace, req.TournamentId, round)
			if err != nil {
				m.logger.Error("failed to check previous round completion", "error", err, "round", round)
				return nil, grpcStatus.Errorf(codes.Internal, "failed to validate round completion: %v", err)
			}

			// Check if all matches in this round are completed
			for _, prevMatch := range previousRoundMatches {
				if prevMatch.Status != serviceextension.MatchStatus_MATCH_STATUS_COMPLETED {
					m.logger.Warn("attempted to submit result for future round while previous round incomplete",
						"match_id", req.MatchId,
						"target_round", targetMatch.Round,
						"incomplete_match_id", prevMatch.MatchId,
						"incomplete_round", round)
					return nil, grpcStatus.Errorf(codes.FailedPrecondition,
						"cannot submit result for round %d: match %s in round %d is not completed",
						targetMatch.Round, prevMatch.MatchId, round)
				}
			}
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

	// Handle bye advancement for subsequent rounds after each match result
	// Get the match to determine what round we're in
	currentMatch, err := m.matchStorage.GetMatch(ctx, req.Namespace, req.TournamentId, req.MatchId)
	if err != nil {
		m.logger.Error("failed to get current match for bye handling", "error", err, "match_id", req.MatchId)
	} else {
		// Handle bye advancement for the next round
		nextRound := currentMatch.Round + 1
		if err := m.HandleByeAdvancement(ctx, req.Namespace, req.TournamentId, nextRound); err != nil {
			m.logger.Error("failed to handle bye advancement for next round", "error", err, "round", nextRound)
			// Don't fail the result submission, just log the error
		}
	}

	// Check if tournament is complete after this match result
	isComplete, winner, err := m.CheckTournamentCompletion(ctx, req.Namespace, req.TournamentId)
	if err != nil {
		m.logger.Error("failed to check tournament completion", "error", err, "tournament_id", req.TournamentId)
		// Don't fail the result submission, just log the error
	} else if isComplete {
		m.logger.Info("tournament completed, attempting to finalize", "tournament_id", req.TournamentId, "winner", winner)

		// Complete the tournament with winner
		err := m.completeTournament(ctx, req.Namespace, req.TournamentId, winner)
		if err != nil {
			m.logger.Error("failed to complete tournament", "error", err, "tournament_id", req.TournamentId, "winner", winner)
			// Don't fail the result submission, just log the error
		} else {
			m.logger.Info("tournament finalized successfully", "tournament_id", req.TournamentId, "winner", winner)
		}
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

	// Get the target match to check its round
	targetMatch, err := m.matchStorage.GetMatch(ctx, req.Namespace, req.TournamentId, req.MatchId)
	if err != nil {
		m.logger.Error("failed to get match for admin validation", "error", err, "match_id", req.MatchId)
		return nil, grpcStatus.Errorf(codes.NotFound, "match not found: %s", req.MatchId)
	}

	// Safeguard: Prevent submitting results for future rounds if previous rounds are incomplete
	if targetMatch.Round > 1 {
		// Check if all matches in previous rounds are completed
		for round := int32(1); round < targetMatch.Round; round++ {
			previousRoundMatches, err := m.matchStorage.GetMatchesByRound(ctx, req.Namespace, req.TournamentId, round)
			if err != nil {
				m.logger.Error("failed to check previous round completion for admin submission", "error", err, "round", round)
				return nil, grpcStatus.Errorf(codes.Internal, "failed to validate round completion: %v", err)
			}

			// Check if all matches in this round are completed
			for _, prevMatch := range previousRoundMatches {
				if prevMatch.Status != serviceextension.MatchStatus_MATCH_STATUS_COMPLETED {
					m.logger.Warn("admin attempted to submit result for future round while previous round incomplete",
						"match_id", req.MatchId,
						"target_round", targetMatch.Round,
						"incomplete_match_id", prevMatch.MatchId,
						"incomplete_round", round)
					return nil, grpcStatus.Errorf(codes.FailedPrecondition,
						"cannot submit result for round %d: match %s in round %d is not completed",
						targetMatch.Round, prevMatch.MatchId, round)
				}
			}
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

	// Handle bye advancement for subsequent rounds after each admin match result
	// Get the match to determine what round we're in
	currentMatch, err := m.matchStorage.GetMatch(ctx, req.Namespace, req.TournamentId, req.MatchId)
	if err != nil {
		m.logger.Error("failed to get current match for bye handling in admin submission", "error", err, "match_id", req.MatchId)
	} else {
		// Handle bye advancement for the next round
		nextRound := currentMatch.Round + 1
		if err := m.HandleByeAdvancement(ctx, req.Namespace, req.TournamentId, nextRound); err != nil {
			m.logger.Error("failed to handle bye advancement for next round after admin submission", "error", err, "round", nextRound)
			// Don't fail the result submission, just log the error
		}
	}

	// Check if tournament is complete after this admin result submission
	isComplete, winner, err := m.CheckTournamentCompletion(ctx, req.Namespace, req.TournamentId)
	if err != nil {
		m.logger.Error("failed to check tournament completion after admin submission", "error", err, "tournament_id", req.TournamentId)
		// Don't fail the result submission, just log the error
	} else if isComplete {
		m.logger.Info("tournament completed after admin submission, attempting to finalize", "tournament_id", req.TournamentId, "winner", winner)

		// Complete the tournament with winner
		err := m.completeTournament(ctx, req.Namespace, req.TournamentId, winner)
		if err != nil {
			m.logger.Error("failed to complete tournament after admin submission", "error", err, "tournament_id", req.TournamentId, "winner", winner)
			// Don't fail the result submission, just log the error
		} else {
			m.logger.Info("tournament finalized successfully after admin submission", "tournament_id", req.TournamentId, "winner", winner)
		}
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
