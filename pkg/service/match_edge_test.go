// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"
	"testing"

	serviceextension "extend-tournament-service/pkg/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Test helper for edge tests
func createTestMatchForEdge(matchID, tournamentID, userID1, userID2 string, round, position int32) *serviceextension.Match {
	match := &serviceextension.Match{
		MatchId:      matchID,
		TournamentId: tournamentID,
		Round:        round,
		Position:     position,
		Status:       serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED,
		StartedAt:    timestamppb.Now(),
	}

	// Only add participant1 if userID1 is provided
	if userID1 != "" {
		match.Participant1 = &serviceextension.TournamentParticipant{
			UserId:      userID1,
			Username:    "player1",
			DisplayName: "Player One",
		}
	}

	// Only add participant2 if userID2 is provided
	if userID2 != "" {
		match.Participant2 = &serviceextension.TournamentParticipant{
			UserId:      userID2,
			Username:    "player2",
			DisplayName: "Player Two",
		}
	}

	return match
}

// Additional edge case tests for REFACTOR phase to improve coverage

// TestAdvanceWinner_NoNextMatchId tests that advancement is skipped when NextMatchId is empty
func TestAdvanceWinner_NoNextMatchId(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Match with no NextMatchId (final round)
	match := createTestMatchForEdge("m1", "tournament1", "user1", "user2", 1, 0)
	match.Winner = "user1"
	match.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	err := service.advanceWinner(context.Background(), "ns1", match)
	assert.NoError(t, err, "Should succeed without error when no next match")

	// No storage calls should be made
	mockStorage.AssertNotCalled(t, "GetMatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockStorage.AssertNotCalled(t, "UpdateMatch", mock.Anything, mock.Anything, mock.Anything)
}

// TestValidateMatchWinner_BoundaryConditions tests edge cases for validation
func TestValidateMatchWinner_BoundaryConditions(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	t.Run("BothParticipantsNil", func(t *testing.T) {
		match := createTestMatchForEdge("m1", "tournament1", "", "", 1, 1) // Both nil

		err := service.validateMatchWinner(match, "any_user")
		assert.Error(t, err, "Should reject when both participants are nil")
		assert.Contains(t, err.Error(), "not a participant")
	})

	t.Run("EmptyWinner", func(t *testing.T) {
		match := createTestMatchForEdge("m1", "tournament1", "user1", "user2", 1, 1)

		err := service.validateMatchWinner(match, "")
		assert.Error(t, err, "Should reject empty winner")
		assert.Contains(t, err.Error(), "winner_user_id is required")
	})

	t.Run("InMatchStatusProgress", func(t *testing.T) {
		match := createTestMatchForEdge("m1", "tournament1", "user1", "user2", 1, 1)
		match.Status = serviceextension.MatchStatus_MATCH_STATUS_IN_PROGRESS

		err := service.validateMatchWinner(match, "user1")
		assert.NoError(t, err, "Should accept winner for in-progress match")
	})
}
