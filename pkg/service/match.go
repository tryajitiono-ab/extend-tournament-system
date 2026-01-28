// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	serviceextension "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/storage"
)

// MatchService implements match management business logic
type MatchService struct {
	matchStorage      storage.MatchStorage
	tournamentStorage storage.TournamentStorage
	logger            *slog.Logger
}

// NewMatchService creates a new match service instance
func NewMatchService(
	matchStorage storage.MatchStorage,
	tournamentStorage storage.TournamentStorage,
	logger *slog.Logger,
) *MatchService {
	return &MatchService{
		matchStorage:      matchStorage,
		tournamentStorage: tournamentStorage,
		logger:            logger,
	}
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
	_, err := m.tournamentStorage.GetTournament(ctx, req.Namespace, req.TournamentId)
	if err != nil {
		m.logger.Error("failed to verify tournament exists", "error", err, "tournament_id", req.TournamentId)
		return nil, err
	}

	var matches []*serviceextension.Match
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

	// TODO: Calculate total_rounds and current_round based on actual matches
	// For now, return basic response
	return &serviceextension.GetTournamentMatchesResponse{
		Matches:      matches,
		TotalRounds:  1, // Will be calculated in service implementation
		CurrentRound: 1, // Will be calculated in service implementation
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

	// Submit result with transaction safety
	match, err := m.matchStorage.SubmitMatchResult(ctx, req.Namespace, req.TournamentId, req.MatchId, req.WinnerUserId)
	if err != nil {
		m.logger.Error("failed to submit match result", "error", err, "match_id", req.MatchId, "winner_user_id", req.WinnerUserId)
		return nil, err
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

	// Submit result with transaction safety (same as game server)
	match, err := m.matchStorage.SubmitMatchResult(ctx, req.Namespace, req.TournamentId, req.MatchId, req.WinnerUserId)
	if err != nil {
		m.logger.Error("failed to submit admin match result", "error", err, "match_id", req.MatchId, "winner_user_id", req.WinnerUserId)
		return nil, err
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
