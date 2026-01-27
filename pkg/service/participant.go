// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"fmt"
	"log/slog"

	extendcustomguildservice "extend-custom-guild-service/pkg/common"
	serviceextension "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/storage"
)

// ParticipantService handles participant registration operations
type ParticipantService struct {
	participantStorage *storage.ParticipantStorage
	tournamentStorage  *storage.TournamentStorage
	logger             *slog.Logger
}

// NewParticipantService creates a new participant service instance
func NewParticipantService(
	participantStorage *storage.ParticipantStorage,
	tournamentStorage *storage.TournamentStorage,
	logger *slog.Logger,
) *ParticipantService {
	return &ParticipantService{
		participantStorage: participantStorage,
		tournamentStorage:  tournamentStorage,
		logger:             logger,
	}
}

// RegisterForTournament registers a user for a tournament
func (p *ParticipantService) RegisterForTournament(ctx context.Context, req *serviceextension.RegisterForTournamentRequest) (*serviceextension.RegisterForTournamentResponse, error) {
	// Extract user context
	namespace, err := extendcustomguildservice.GetContextNamespace(ctx)
	if err != nil {
		p.logger.Error("failed to get namespace from context", "error", err)
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	userID, err := extendcustomguildservice.GetContextUserID(ctx)
	if err != nil {
		p.logger.Error("failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	username, err := extendcustomguildservice.GetContextUsername(ctx)
	if err != nil {
		p.logger.Warn("failed to get username from context", "error", err)
		// Username is optional, continue without it
	}

	// Validate request namespace matches context namespace
	if req.GetNamespace() != namespace {
		p.logger.Error("namespace mismatch",
			"req_namespace", req.GetNamespace(),
			"ctx_namespace", namespace)
		return nil, fmt.Errorf("namespace mismatch")
	}

	p.logger.Info("user registering for tournament",
		"user_id", userID,
		"username", username,
		"tournament_id", req.GetTournamentId(),
		"namespace", namespace,
	)

	// Call storage with user context
	response, err := p.participantStorage.RegisterParticipant(ctx, req, userID)
	if err != nil {
		p.logger.Error("failed to register participant",
			"user_id", userID,
			"tournament_id", req.GetTournamentId(),
			"error", err)
		return nil, err
	}

	p.logger.Info("user successfully registered for tournament",
		"user_id", userID,
		"participant_id", response.GetParticipantId(),
		"tournament_id", req.GetTournamentId(),
	)

	return response, nil
}

// GetTournamentParticipants retrieves participants for a tournament
func (p *ParticipantService) GetTournamentParticipants(ctx context.Context, req *serviceextension.GetTournamentParticipantsRequest) (*serviceextension.GetTournamentParticipantsResponse, error) {
	// Extract user context
	namespace, err := extendcustomguildservice.GetContextNamespace(ctx)
	if err != nil {
		p.logger.Error("failed to get namespace from context", "error", err)
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	// Validate request namespace matches context namespace
	if req.GetNamespace() != namespace {
		p.logger.Error("namespace mismatch",
			"req_namespace", req.GetNamespace(),
			"ctx_namespace", namespace)
		return nil, fmt.Errorf("namespace mismatch")
	}

	p.logger.Info("retrieving tournament participants",
		"tournament_id", req.GetTournamentId(),
		"namespace", namespace,
		"page_size", req.GetPageSize(),
	)

	// Get participants from storage
	response, err := p.participantStorage.GetParticipants(ctx, req)
	if err != nil {
		p.logger.Error("failed to get tournament participants",
			"tournament_id", req.GetTournamentId(),
			"error", err)
		return nil, err
	}

	p.logger.Info("successfully retrieved tournament participants",
		"tournament_id", req.GetTournamentId(),
		"participant_count", response.GetTotalParticipants(),
	)

	return response, nil
}

// RemoveParticipant removes a participant from a tournament (admin only)
func (p *ParticipantService) RemoveParticipant(ctx context.Context, req *serviceextension.RemoveParticipantRequest) (*serviceextension.RemoveParticipantResponse, error) {
	// Extract user context and verify admin permissions
	namespace, err := extendcustomguildservice.GetContextNamespace(ctx)
	if err != nil {
		p.logger.Error("failed to get namespace from context", "error", err)
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	// Check admin permissions
	isAdmin, err := extendcustomguildservice.IsAdminUser(ctx)
	if err != nil {
		p.logger.Error("failed to check admin permissions", "error", err)
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	if !isAdmin {
		p.logger.Warn("unauthorized attempt to remove participant",
			"user_id", "<redacted>",
			"target_user_id", req.GetUserId(),
			"tournament_id", req.GetTournamentId())
		return nil, fmt.Errorf("insufficient permissions: admin role required")
	}

	// Validate request namespace matches context namespace
	if req.GetNamespace() != namespace {
		p.logger.Error("namespace mismatch",
			"req_namespace", req.GetNamespace(),
			"ctx_namespace", namespace)
		return nil, fmt.Errorf("namespace mismatch")
	}

	adminUserID, _ := extendcustomguildservice.GetContextUserID(ctx)

	p.logger.Info("admin removing participant from tournament",
		"admin_user_id", adminUserID,
		"target_user_id", req.GetUserId(),
		"tournament_id", req.GetTournamentId(),
		"namespace", namespace,
	)

	// Remove participant via storage
	response, err := p.participantStorage.RemoveParticipant(ctx, req)
	if err != nil {
		p.logger.Error("failed to remove participant",
			"target_user_id", req.GetUserId(),
			"tournament_id", req.GetTournamentId(),
			"error", err)
		return nil, err
	}

	p.logger.Info("successfully removed participant from tournament",
		"admin_user_id", adminUserID,
		"target_user_id", req.GetUserId(),
		"tournament_id", req.GetTournamentId(),
	)

	return response, nil
}
