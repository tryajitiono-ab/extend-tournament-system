// Copyright (c) 2023-2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"

	serviceextension "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/service"
)

// TournamentServer implements the TournamentService gRPC interface
type TournamentServer struct {
	serviceextension.UnimplementedTournamentServiceServer
	*service.TournamentServiceServer
	*service.ParticipantService
	*service.MatchService
}

// NewTournamentServer creates a new tournament server instance
func NewTournamentServer(
	tournamentService *service.TournamentServiceServer,
	participantService *service.ParticipantService,
	matchService *service.MatchService,
) *TournamentServer {
	return &TournamentServer{
		TournamentServiceServer: tournamentService,
		ParticipantService:      participantService,
		MatchService:            matchService,
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

// StartTournament starts a tournament
func (s *TournamentServer) StartTournament(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	return s.TournamentServiceServer.StartTournament(ctx, req)
}

// CompleteTournament completes a tournament
func (s *TournamentServer) CompleteTournament(ctx context.Context, req *serviceextension.StartTournamentRequest) (*serviceextension.StartTournamentResponse, error) {
	return s.TournamentServiceServer.CompleteTournament(ctx, req)
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
