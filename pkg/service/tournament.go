// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	extendcustomguildservice "extend-custom-guild-service/pkg/common"
	serviceextension "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/storage"
)

// TournamentServiceServer implements the TournamentService gRPC service
type TournamentServiceServer struct {
	serviceextension.UnimplementedTournamentServiceServer
	tokenRepo         repository.TokenRepository
	configRepo        repository.ConfigRepository
	refreshRepo       repository.RefreshTokenRepository
	tournamentStorage storage.TournamentStorage
	authInterceptor   *extendcustomguildservice.TournamentAuthInterceptor
	logger            *slog.Logger
}

// NewTournamentServiceServer creates a new TournamentServiceServer instance
func NewTournamentServiceServer(
	tokenRepo repository.TokenRepository,
	configRepo repository.ConfigRepository,
	refreshRepo repository.RefreshTokenRepository,
	tournamentStorage storage.TournamentStorage,
	authInterceptor *extendcustomguildservice.TournamentAuthInterceptor,
	logger *slog.Logger,
) *TournamentServiceServer {
	return &TournamentServiceServer{
		tokenRepo:         tokenRepo,
		configRepo:        configRepo,
		refreshRepo:       refreshRepo,
		tournamentStorage: tournamentStorage,
		authInterceptor:   authInterceptor,
		logger:            logger,
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

	// Validate that tournament can be cancelled
	if tournament.Status == serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED {
		return nil, grpcStatus.Errorf(codes.FailedPrecondition, "cannot cancel tournament that has already started")
	}
	if tournament.Status == serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED {
		return nil, grpcStatus.Errorf(codes.FailedPrecondition, "cannot cancel tournament that has already completed")
	}
	if tournament.Status == serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED {
		return nil, grpcStatus.Errorf(codes.FailedPrecondition, "tournament is already cancelled")
	}

	// Update tournament status to cancelled
	tournament.Status = serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED
	tournament.UpdatedAt = timestamppb.New(time.Now())

	// Update tournament in storage
	updatedTournament, err := s.tournamentStorage.UpdateTournament(ctx, req.Namespace, req.TournamentId, tournament)
	if err != nil {
		s.logger.Error("failed to cancel tournament", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	s.logger.Info("tournament cancelled successfully",
		"tournament_id", updatedTournament.TournamentId,
		"namespace", req.Namespace,
		"name", updatedTournament.Name)

	return &serviceextension.CancelTournamentResponse{
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

	// Validate that tournament can be started
	if tournament.Status != serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE {
		return nil, grpcStatus.Errorf(codes.FailedPrecondition, "can only start tournaments with ACTIVE status, current status: %v", tournament.Status)
	}

	// Update tournament status to started
	tournament.Status = serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED
	tournament.UpdatedAt = timestamppb.New(time.Now())

	// Update tournament in storage
	updatedTournament, err := s.tournamentStorage.UpdateTournament(ctx, req.Namespace, req.TournamentId, tournament)
	if err != nil {
		s.logger.Error("failed to start tournament", "error", err, "namespace", req.Namespace, "tournament_id", req.TournamentId)
		return nil, err
	}

	s.logger.Info("tournament started successfully",
		"tournament_id", updatedTournament.TournamentId,
		"namespace", req.Namespace,
		"name", updatedTournament.Name)

	return &serviceextension.StartTournamentResponse{
		Tournament: updatedTournament,
	}, nil
}
