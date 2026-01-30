// Copyright (c) 2023-2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"fmt"
	"log/slog"

	serviceextension "extend-tournament-service/pkg/pb"
	"extend-tournament-service/pkg/service"
)

// TournamentServer implements the TournamentService gRPC interface
type TournamentServer struct {
	serviceextension.UnimplementedTournamentServiceServer
	*service.TournamentServiceServer
	*service.ParticipantService
	*service.MatchService
	logger *slog.Logger
}

// NewTournamentServer creates a new tournament server instance
func NewTournamentServer(
	tournamentService *service.TournamentServiceServer,
	participantService *service.ParticipantService,
	matchService *service.MatchService,
	logger *slog.Logger,
) *TournamentServer {
	return &TournamentServer{
		TournamentServiceServer: tournamentService,
		ParticipantService:      participantService,
		MatchService:            matchService,
		logger:                  logger,
	}
}

// RegisterForTournament registers a user for a tournament
func (s *TournamentServer) RegisterForTournament(ctx context.Context, req *serviceextension.RegisterForTournamentRequest) (*serviceextension.RegisterForTournamentResponse, error) {
	return s.ParticipantService.RegisterForTournament(ctx, req)
}

// GetTournamentParticipants retrieves participants for a tournament
func (s *TournamentServer) GetTournamentParticipants(ctx context.Context, req *serviceextension.GetTournamentParticipantsRequest) (*serviceextension.GetTournamentParticipantsResponse, error) {
	return s.ParticipantService.GetTournamentParticipants(ctx, req)
}

// RemoveParticipant removes a participant from a tournament (admin only)
func (s *TournamentServer) RemoveParticipant(ctx context.Context, req *serviceextension.RemoveParticipantRequest) (*serviceextension.RemoveParticipantResponse, error) {
	return s.ParticipantService.RemoveParticipant(ctx, req)
}

// Tournament CRUD operations delegated to TournamentServiceServer

// CreateTournament creates a new tournament
func (s *TournamentServer) CreateTournament(ctx context.Context, req *serviceextension.CreateTournamentRequest) (*serviceextension.CreateTournamentResponse, error) {
	return s.TournamentServiceServer.CreateTournament(ctx, req)
}

// ListTournaments lists tournaments
func (s *TournamentServer) ListTournaments(ctx context.Context, req *serviceextension.ListTournamentsRequest) (*serviceextension.ListTournamentsResponse, error) {
	return s.TournamentServiceServer.ListTournaments(ctx, req)
}

// GetTournament gets a tournament
func (s *TournamentServer) GetTournament(ctx context.Context, req *serviceextension.GetTournamentRequest) (*serviceextension.GetTournamentResponse, error) {
	return s.TournamentServiceServer.GetTournament(ctx, req)
}

// CancelTournament cancels a tournament
func (s *TournamentServer) CancelTournament(ctx context.Context, req *serviceextension.CancelTournamentRequest) (*serviceextension.CancelTournamentResponse, error) {
	return s.TournamentServiceServer.CancelTournament(ctx, req)
}

// ActivateTournament activates a tournament
func (s *TournamentServer) ActivateTournament(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	return s.TournamentServiceServer.ActivateTournament(ctx, req)
}

// StartTournament starts a tournament with bracket generation
func (s *TournamentServer) StartTournament(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	// First, handle bracket generation using MatchService
	// Get participants to generate bracket
	participantsReq := &serviceextension.GetTournamentParticipantsRequest{
		Namespace:    req.Namespace,
		TournamentId: req.TournamentId,
		PageSize:     1000, // Get all participants
	}

	participantsResp, err := s.ParticipantService.GetTournamentParticipants(ctx, participantsReq)
	if err != nil {
		s.logger.Error("failed to get participants for bracket generation", "error", err, "tournament_id", req.TournamentId)
		return nil, err
	}

	if len(participantsResp.Participants) < 2 {
		return nil, fmt.Errorf("at least 2 participants required to start tournament (current: %d)", len(participantsResp.Participants))
	}

	// Generate bracket structure using existing tournament service logic
	// Convert participants to TournamentParticipant format for bracket generation
	tournamentParticipants := make([]service.TournamentParticipant, len(participantsResp.Participants))
	for i, participant := range participantsResp.Participants {
		tournamentParticipants[i] = service.TournamentParticipant{
			UserId:      participant.UserId,
			Username:    participant.Username,
			DisplayName: participant.DisplayName,
		}
	}

	// Generate bracket data using tournament service's bracket generation
	bracketData, err := s.TournamentServiceServer.GenerateBrackets(tournamentParticipants)
	if err != nil {
		s.logger.Error("failed to generate brackets", "error", err, "tournament_id", req.TournamentId)
		return nil, err
	}

	// Convert bracket data to protobuf Match objects for storage
	var allMatches []*serviceextension.Match
	for roundIdx, round := range bracketData.Rounds {
		for matchIdx, bracket := range round {
			match := &serviceextension.Match{
				MatchId:      fmt.Sprintf("match-r%d-m%d", roundIdx+1, matchIdx+1),
				TournamentId: req.TournamentId,
				Round:        int32(roundIdx + 1),
				Position:     int32(matchIdx),
				Status:       serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED,
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
			}

			allMatches = append(allMatches, match)
		}
	}

	// Create all matches in storage using MatchService
	if len(allMatches) > 0 {
		err := s.MatchService.CreateTournamentMatches(ctx, req.Namespace, req.TournamentId, allMatches)
		if err != nil {
			s.logger.Error("failed to create tournament matches", "error", err, "tournament_id", req.TournamentId, "match_count", len(allMatches))
			return nil, err
		}

		s.logger.Info("tournament matches created successfully",
			"tournament_id", req.TournamentId,
			"total_rounds", bracketData.TotalRounds,
			"first_round_matches", len(bracketData.Rounds[0]),
			"total_matches", len(allMatches))

		// Handle bye advancement for rounds 2 and beyond
		// Since round 1 bye participants are already marked as completed,
		// we need to advance them to round 2
		for round := int32(2); round <= bracketData.TotalRounds; round++ {
			if err := s.MatchService.HandleByeAdvancement(ctx, req.Namespace, req.TournamentId, round); err != nil {
				s.logger.Warn("failed to handle bye advancement for round", "error", err, "round", round)
				// Don't fail tournament start, just log the warning
			}
		}
	}

	// Now delegate to tournament service for status change
	return s.TournamentServiceServer.StartTournament(ctx, req)
}

// CompleteTournament completes a tournament
func (s *TournamentServer) CompleteTournament(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	return s.TournamentServiceServer.CompleteTournamentByAdmin(ctx, req)
}

// Match operations delegated to MatchService

// GetTournamentMatches retrieves all matches for a tournament
func (s *TournamentServer) GetTournamentMatches(ctx context.Context, req *serviceextension.GetTournamentMatchesRequest) (*serviceextension.GetTournamentMatchesResponse, error) {
	return s.MatchService.GetTournamentMatches(ctx, req)
}

// GetMatch retrieves a specific match by ID
func (s *TournamentServer) GetMatch(ctx context.Context, req *serviceextension.GetMatchRequest) (*serviceextension.GetMatchResponse, error) {
	return s.MatchService.GetMatch(ctx, req)
}

// SubmitMatchResult submits a match result (game server)
func (s *TournamentServer) SubmitMatchResult(ctx context.Context, req *serviceextension.SubmitMatchResultRequest) (*serviceextension.SubmitMatchResultResponse, error) {
	return s.MatchService.SubmitMatchResult(ctx, req)
}

// AdminSubmitMatchResult submits a match result (admin override)
func (s *TournamentServer) AdminSubmitMatchResult(ctx context.Context, req *serviceextension.AdminSubmitMatchResultRequest) (*serviceextension.AdminSubmitMatchResultResponse, error) {
	return s.MatchService.AdminSubmitMatchResult(ctx, req)
}
