// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"
	"testing"

	serviceextension "extend-tournament-service/pkg/pb"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

// createTestContext creates a context with the required gRPC metadata for participant operations.
func createTestContext(namespace, userID, username string, isAdmin bool) context.Context {
	md := metadata.New(map[string]string{
		"namespace": namespace,
		"x-user-id": userID,
		"x-username": username,
	})
	if isAdmin {
		md.Set("x-is-admin", "true")
	}
	return metadata.NewIncomingContext(context.Background(), md)
}

// --- RegisterForTournament tests ---

func TestRegisterParticipant_NamespaceMismatch(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	ctx := createTestContext("ns1", "user1", "player1", false)
	req := &serviceextension.RegisterForTournamentRequest{
		Namespace:    "ns2", // Mismatch with context namespace "ns1"
		TournamentId: "t1",
	}

	_, err := service.RegisterForTournament(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace mismatch")
}

func TestRegisterParticipant_NoMetadata(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	// Context without gRPC metadata - GetContextUserID will fail
	ctx := context.Background()
	req := &serviceextension.RegisterForTournamentRequest{
		Namespace:    "ns1",
		TournamentId: "t1",
	}

	_, err := service.RegisterForTournament(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestRegisterParticipant_MissingUserID(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	// Context with namespace but no user ID
	md := metadata.New(map[string]string{
		"namespace": "ns1",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &serviceextension.RegisterForTournamentRequest{
		Namespace:    "ns1",
		TournamentId: "t1",
	}

	_, err := service.RegisterForTournament(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

// --- GetTournamentParticipants tests ---

func TestGetTournamentParticipants_NamespaceMismatch(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	ctx := createTestContext("ns1", "user1", "player1", false)
	req := &serviceextension.GetTournamentParticipantsRequest{
		Namespace:    "ns2", // Mismatch
		TournamentId: "t1",
	}

	_, err := service.GetTournamentParticipants(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace mismatch")
}

func TestGetTournamentParticipants_NoMetadata(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	// Without metadata, GetContextNamespace returns default "test-ns",
	// so requesting "ns1" causes namespace mismatch.
	ctx := context.Background()
	req := &serviceextension.GetTournamentParticipantsRequest{
		Namespace:    "ns1",
		TournamentId: "t1",
	}

	_, err := service.GetTournamentParticipants(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace mismatch")
}

// --- RemoveParticipant tests ---

func TestRemoveParticipant_NotAdmin(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	ctx := createTestContext("ns1", "user1", "player1", false) // Not admin
	req := &serviceextension.RemoveParticipantRequest{
		Namespace:    "ns1",
		TournamentId: "t1",
		UserId:       "user2",
	}

	_, err := service.RemoveParticipant(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions")
}

func TestRemoveParticipant_NamespaceMismatch(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	ctx := createTestContext("ns1", "user1", "player1", true) // Admin but wrong namespace
	req := &serviceextension.RemoveParticipantRequest{
		Namespace:    "ns2", // Mismatch
		TournamentId: "t1",
		UserId:       "user2",
	}

	_, err := service.RemoveParticipant(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace mismatch")
}

func TestRemoveParticipant_NoMetadata(t *testing.T) {
	logger := slog.Default()
	service := NewParticipantService(nil, nil, logger)

	ctx := context.Background()
	req := &serviceextension.RemoveParticipantRequest{
		Namespace:    "ns1",
		TournamentId: "t1",
		UserId:       "user2",
	}

	_, err := service.RemoveParticipant(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization failed")
}
