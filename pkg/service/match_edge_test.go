// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"log/slog"
	"testing"

	serviceextension "extend-tournament-service/pkg/pb"

	"github.com/stretchr/testify/assert"
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

// TestCalculateNextPosition_EdgeCases tests edge cases for position calculation
func TestCalculateNextPosition_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    int32
		expected int32
	}{
		{"Position1", 1, 1},
		{"Position2", 2, 1},
		{"Position3", 3, 2},
		{"Position4", 4, 2},
		{"Position5", 5, 3},
		{"Position8", 8, 4},
		{"ZeroPosition", 0, 0}, // Invalid case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNextPosition(tt.input)
			assert.Equal(t, tt.expected, result, "Position calculation incorrect for input %d", tt.input)
		})
	}
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
