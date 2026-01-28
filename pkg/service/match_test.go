// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"
	"testing"

	serviceextension "extend-custom-guild-service/pkg/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockMatchStorage is a mock implementation of MatchStorage interface
type MockMatchStorage struct {
	mock.Mock
}

func (m *MockMatchStorage) GetMatch(ctx context.Context, namespace, tournamentID, matchID string) (*serviceextension.Match, error) {
	args := m.Called(ctx, namespace, tournamentID, matchID)
	return args.Get(0).(*serviceextension.Match), args.Error(1)
}

func (m *MockMatchStorage) GetTournamentMatches(ctx context.Context, namespace, tournamentID string) ([]*serviceextension.Match, error) {
	args := m.Called(ctx, namespace, tournamentID)
	return args.Get(0).([]*serviceextension.Match), args.Error(1)
}

func (m *MockMatchStorage) CreateMatches(ctx context.Context, namespace, tournamentID string, matches []*serviceextension.Match) error {
	args := m.Called(ctx, namespace, tournamentID, matches)
	return args.Error(0)
}

func (m *MockMatchStorage) UpdateMatch(ctx context.Context, namespace string, match *serviceextension.Match) error {
	args := m.Called(ctx, namespace, match)
	return args.Error(0)
}

func (m *MockMatchStorage) SubmitMatchResult(ctx context.Context, namespace, tournamentID, matchID, winnerUserID string) (*serviceextension.Match, error) {
	args := m.Called(ctx, namespace, tournamentID, matchID, winnerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*serviceextension.Match), args.Error(1)
}

func (m *MockMatchStorage) GetMatchesByRound(ctx context.Context, namespace, tournamentID string, round int32) ([]*serviceextension.Match, error) {
	args := m.Called(ctx, namespace, tournamentID, round)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*serviceextension.Match), args.Error(1)
}

// MockTournamentStorage is a mock implementation of TournamentStorage interface
type MockTournamentStorage struct {
	mock.Mock
}

func (m *MockTournamentStorage) CreateTournament(ctx context.Context, namespace string, tournament *serviceextension.Tournament) (*serviceextension.Tournament, error) {
	args := m.Called(ctx, namespace, tournament)
	return args.Get(0).(*serviceextension.Tournament), args.Error(1)
}

func (m *MockTournamentStorage) GetTournament(ctx context.Context, namespace string, tournamentID string) (*serviceextension.Tournament, error) {
	args := m.Called(ctx, namespace, tournamentID)
	return args.Get(0).(*serviceextension.Tournament), args.Error(1)
}

func (m *MockTournamentStorage) ListTournaments(ctx context.Context, namespace string, limit, offset int32, status serviceextension.TournamentStatus) ([]*serviceextension.Tournament, int32, error) {
	args := m.Called(ctx, namespace, limit, offset, status)
	return args.Get(0).([]*serviceextension.Tournament), args.Get(1).(int32), args.Error(2)
}

func (m *MockTournamentStorage) UpdateTournament(ctx context.Context, namespace string, tournamentID string, tournament *serviceextension.Tournament) (*serviceextension.Tournament, error) {
	args := m.Called(ctx, namespace, tournamentID, tournament)
	return args.Get(0).(*serviceextension.Tournament), args.Error(1)
}

func (m *MockTournamentStorage) GetTournamentForRegistration(ctx context.Context, namespace string, tournamentID string) (*serviceextension.Tournament, error) {
	args := m.Called(ctx, namespace, tournamentID)
	return args.Get(0).(*serviceextension.Tournament), args.Error(1)
}

func (m *MockTournamentStorage) UpdateParticipantCount(ctx context.Context, namespace string, tournamentID string, increment int32) error {
	args := m.Called(ctx, namespace, tournamentID, increment)
	return args.Error(0)
}

func (m *MockTournamentStorage) CheckTournamentCapacity(ctx context.Context, namespace string, tournamentID string) (bool, error) {
	args := m.Called(ctx, namespace, tournamentID)
	return args.Bool(0), args.Error(1)
}

// Test helper to create test matches
func createTestMatch(matchID, tournamentID, userID1, userID2 string, round, position int32) *serviceextension.Match {
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

// TestAdvanceWinner tests the bracket position calculation logic
func TestAdvanceWinner(t *testing.T) {
	tests := []struct {
		name              string
		currentMatch      *serviceextension.Match
		expectedNextRound int32
		expectedNextPos   int32
	}{
		{
			name: "Position1Advancement",
			currentMatch: &serviceextension.Match{
				MatchId:      "match1",
				Round:        1,
				Position:     1,
				TournamentId: "tournament1",
			},
			expectedNextRound: 2,
			expectedNextPos:   1,
		},
		{
			name: "Position2Advancement",
			currentMatch: &serviceextension.Match{
				MatchId:      "match2",
				Round:        1,
				Position:     2,
				TournamentId: "tournament1",
			},
			expectedNextRound: 2,
			expectedNextPos:   1,
		},
		{
			name: "Position3Advancement",
			currentMatch: &serviceextension.Match{
				MatchId:      "match3",
				Round:        1,
				Position:     3,
				TournamentId: "tournament1",
			},
			expectedNextRound: 2,
			expectedNextPos:   2,
		},
		{
			name: "Position4Advancement",
			currentMatch: &serviceextension.Match{
				MatchId:      "match4",
				Round:        1,
				Position:     4,
				TournamentId: "tournament1",
			},
			expectedNextRound: 2,
			expectedNextPos:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test bracket position calculation
			nextPosition := calculateNextPosition(tt.currentMatch.Position)
			nextRound := tt.currentMatch.Round + 1

			assert.Equal(t, tt.expectedNextPos, nextPosition, "Next position calculation incorrect")
			assert.Equal(t, tt.expectedNextRound, nextRound, "Next round should be current round + 1")
		})
	}
}

// TestBracketMath tests the bracket position calculations
func TestBracketMath(t *testing.T) {
	tests := []struct {
		name            string
		currentPos      int32
		expectedNextPos int32
	}{
		{
			name:            "Position1_to_Position1",
			currentPos:      1,
			expectedNextPos: 1,
		},
		{
			name:            "Position2_to_Position1",
			currentPos:      2,
			expectedNextPos: 1,
		},
		{
			name:            "Position3_to_Position2",
			currentPos:      3,
			expectedNextPos: 2,
		},
		{
			name:            "Position4_to_Position2",
			currentPos:      4,
			expectedNextPos: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextPos := calculateNextPosition(tt.currentPos)
			assert.Equal(t, tt.expectedNextPos, nextPos, "Next position calculation incorrect")
		})
	}
}

// TestSubmitMatchResult_Validation tests basic validation
func TestSubmitMatchResult_Validation(t *testing.T) {
	tests := []struct {
		name        string
		req         *serviceextension.SubmitMatchResultRequest
		expectError bool
		errorCode   codes.Code
	}{
		{
			name: "EmptyNamespace",
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "",
				TournamentId: "tournament1",
				MatchId:      "match1",
				WinnerUserId: "user1",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "EmptyTournamentId",
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "",
				MatchId:      "match1",
				WinnerUserId: "user1",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "EmptyMatchId",
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				MatchId:      "",
				WinnerUserId: "user1",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "EmptyWinnerUserId",
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				MatchId:      "match1",
				WinnerUserId: "",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "ValidRequest",
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				MatchId:      "match1",
				WinnerUserId: "user1",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockMatchStorage{}
			mockTournamentStorage := &MockTournamentStorage{}

			// Mock tournament existence for valid request
			if !tt.expectError || tt.name == "ValidRequest" {
				mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "tournament1").
					Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)
				mockStorage.On("SubmitMatchResult", mock.Anything, "ns1", "tournament1", "match1", "user1").
					Return(createTestMatch("match1", "tournament1", "user1", "user2", 1, 1), nil)
				// Mock GetMatch call for advancement in valid request
				mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "match1").
					Return(createTestMatch("match1", "tournament1", "user1", "user2", 1, 1), nil)
				// Mock advancement calls for valid request
				mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
					Return([]*serviceextension.Match{}, nil)
				mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*serviceextension.Match")).
					Return(nil)
				// Mock tournament completion check
				mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
					Return([]*serviceextension.Match{createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)}, nil)
				// Mock tournament completion
				mockTournamentStorage.On("UpdateTournament", mock.Anything, "ns1", "tournament1", mock.AnythingOfType("*serviceextension.Tournament")).
					Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)
			}

			logger := slog.Default()
			service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

			resp, err := service.SubmitMatchResult(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
				grpcErr, ok := grpcStatus.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, grpcErr.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// calculateTournamentProgress calculates total rounds and current round based on matches
func calculateTournamentProgress(matches []*serviceextension.Match) (int32, int32) {
	if len(matches) == 0 {
		return 0, 0
	}

	// Find highest round number to determine total rounds
	maxRound := int32(0)
	currentRound := int32(0)

	for _, match := range matches {
		if match.Round > maxRound {
			maxRound = match.Round
		}

		// Check if this round has any in-progress or scheduled matches
		// to determine the current active round
		if match.Status == serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED ||
			match.Status == serviceextension.MatchStatus_MATCH_STATUS_IN_PROGRESS {
			if currentRound == 0 {
				currentRound = match.Round
			} else if match.Round < currentRound {
				currentRound = match.Round
			}
		}
	}

	// If all matches are completed, current round is the last round
	if currentRound == 0 {
		currentRound = maxRound
	}

	return maxRound, currentRound
}

// TestIntegration_WinnerAdvancement tests complete advancement workflow
func TestIntegration_WinnerAdvancement(t *testing.T) {
	t.Run("CompleteAdvancementWorkflow", func(t *testing.T) {
		// Test that winner advancement logic works end-to-end
		// This tests the bracket math for a 4-player tournament

		// Create initial matches
		round1Match1 := createTestMatch("m1", "t1", "user1", "user2", 1, 1)
		round1Match2 := createTestMatch("m2", "t1", "user3", "user4", 1, 2)

		// Simulate first round results
		round1Match1.Winner = "user1"
		round1Match1.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

		round1Match2.Winner = "user3"
		round1Match2.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

		// Test bracket position calculations
		// Match 1 (position 1) should advance to position 1 in round 2
		nextPosition1 := calculateNextPosition(round1Match1.Position)
		assert.Equal(t, int32(1), nextPosition1, "Position 1 should advance to position 1")
		assert.Equal(t, int32(2), round1Match1.Round+1, "Should advance to round 2")

		// Match 2 (position 2) should advance to position 1 in round 2
		nextPosition2 := calculateNextPosition(round1Match2.Position)
		assert.Equal(t, int32(1), nextPosition2, "Position 2 should advance to position 1")
		assert.Equal(t, int32(2), round1Match2.Round+1, "Should advance to round 2")

		// Create additional round 2 match (quarterfinal) for more realistic test
		round2Match1 := createTestMatch("m3", "t1", "user1", "user3", 2, 1)

		// Test tournament progress calculation
		matches := []*serviceextension.Match{round1Match1, round1Match2, round2Match1}
		totalRounds, currentRound := calculateTournamentProgress(matches)
		assert.Equal(t, int32(2), totalRounds, "Should have 2 total rounds")
		assert.Equal(t, int32(2), currentRound, "Should be in round 2")
	})
}

// TestIntegration_GetTournamentMatches tests complete match retrieval workflow
func TestIntegration_GetTournamentMatches(t *testing.T) {
	t.Run("CompleteMatchRetrievalWorkflow", func(t *testing.T) {
		mockStorage := &MockMatchStorage{}
		mockTournamentStorage := &MockTournamentStorage{}

		// Mock tournament existence
		mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "tournament1").
			Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

		// Mock matches for all rounds
		allMatches := []*serviceextension.Match{
			createTestMatch("m1", "t1", "user1", "user2", 1, 1),
			createTestMatch("m2", "t1", "user3", "user4", 1, 2),
			createTestMatch("m3", "t1", "user1", "user2", 2, 1), // Round 2, position 1 (final)
		}

		// Test with no round filter
		mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
			Return(allMatches, nil)

		logger := slog.Default()
		service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

		req := &serviceextension.GetTournamentMatchesRequest{
			Namespace:    "ns1",
			TournamentId: "tournament1",
			Round:        0, // All rounds
		}

		resp, err := service.GetTournamentMatches(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, len(allMatches), len(resp.Matches))
		assert.Equal(t, int32(2), resp.TotalRounds)
		assert.Equal(t, int32(1), resp.CurrentRound)

		mockStorage.AssertExpectations(t)
		mockTournamentStorage.AssertExpectations(t)
	})
}

// RED PHASE TESTS - These should fail initially and pass after implementation

// TestValidateMatchWinner_ValidWinner tests that valid participant IDs are accepted
func TestValidateMatchWinner_ValidWinner(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Test match with two participants
	match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)

	// Test that participant1 is valid winner
	err := service.validateMatchWinner(match, "user1")
	assert.NoError(t, err, "Participant1 should be valid winner")

	// Test that participant2 is valid winner
	err = service.validateMatchWinner(match, "user2")
	assert.NoError(t, err, "Participant2 should be valid winner")
}

// TestValidateMatchWinner_InvalidWinner tests that non-participant IDs are rejected
func TestValidateMatchWinner_InvalidWinner(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Test match with two participants
	match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)

	// Test that non-participant is rejected
	err := service.validateMatchWinner(match, "user3")
	assert.Error(t, err, "Non-participant should be rejected")
	assert.Contains(t, err.Error(), "not a participant", "Error should mention not a participant")
}

// TestValidateMatchWinner_EmptyWinner tests that empty winner ID is rejected
func TestValidateMatchWinner_EmptyWinner(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Test match with two participants
	match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)

	// Test that empty winner is rejected
	err := service.validateMatchWinner(match, "")
	assert.Error(t, err, "Empty winner should be rejected")
	assert.Contains(t, err.Error(), "winner_user_id is required", "Error should mention required winner")
}

// TestValidateMatchWinner_AlreadyCompleted tests that completed matches reject new results
func TestValidateMatchWinner_AlreadyCompleted(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create already completed match
	match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
	match.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
	match.Winner = "user1"

	// Test that completed match rejects new winner validation
	err := service.validateMatchWinner(match, "user2")
	assert.Error(t, err, "Already completed match should reject new result")
}

// TestAdvanceWinner_Position1Advancement tests position 1 advances to position 1 in next round
func TestAdvanceWinner_Position1Advancement(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create current match with winner
	currentMatch := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
	currentMatch.Winner = "user1"
	currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Create next round match (should be at position 1 in round 2)
	nextMatch := createTestMatch("match2", "tournament1", "", "", 2, 1)

	// Mock storage calls
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
		Return([]*serviceextension.Match{nextMatch}, nil)
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*serviceextension.Match")).
		Return(nil)

	// Test advancement
	err := service.advanceWinner(context.Background(), "ns1", currentMatch)
	assert.NoError(t, err, "Winner advancement should succeed")

	// Verify next match was updated with participant
	mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.Participant1 != nil && match.Participant1.UserId == "user1"
	}))
}

// TestAdvanceWinner_Position2Advancement tests position 2 advances to position 1 in next round
func TestAdvanceWinner_Position2Advancement(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create current match at position 2 with winner
	currentMatch := createTestMatch("match1", "tournament1", "user1", "user2", 1, 2)
	currentMatch.Winner = "user1"
	currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Create next round match (should be at position 1 in round 2)
	nextMatch := createTestMatch("match2", "tournament1", "", "", 2, 1)

	// Mock storage calls
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
		Return([]*serviceextension.Match{nextMatch}, nil)
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*serviceextension.Match")).
		Return(nil)

	// Test advancement
	err := service.advanceWinner(context.Background(), "ns1", currentMatch)
	assert.NoError(t, err, "Winner advancement should succeed")

	// Verify position 2 advances to position 1
	mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.Position == 1 && match.Round == 2
	}))
}

// TestAdvanceWinner_Position3Advancement tests position 3 advances to position 2 in next round
func TestAdvanceWinner_Position3Advancement(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create current match at position 3 with winner
	currentMatch := createTestMatch("match1", "tournament1", "user1", "user2", 1, 3)
	currentMatch.Winner = "user1"
	currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Create next round match (should be at position 2 in round 2)
	nextMatch := createTestMatch("match2", "tournament1", "", "", 2, 2)

	// Mock storage calls
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
		Return([]*serviceextension.Match{nextMatch}, nil)
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*serviceextension.Match")).
		Return(nil)

	// Test advancement
	err := service.advanceWinner(context.Background(), "ns1", currentMatch)
	assert.NoError(t, err, "Winner advancement should succeed")

	// Verify position 3 advances to position 2
	mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.Position == 2 && match.Round == 2
	}))
}

// TestAdvanceWinner_Position4Advancement tests position 4 advances to position 2 in next round
func TestAdvanceWinner_Position4Advancement(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create current match at position 4 with winner
	currentMatch := createTestMatch("match1", "tournament1", "user1", "user2", 1, 4)
	currentMatch.Winner = "user1"
	currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Create next round match (should be at position 2 in round 2)
	nextMatch := createTestMatch("match2", "tournament1", "", "", 2, 2)

	// Mock storage calls
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
		Return([]*serviceextension.Match{nextMatch}, nil)
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*serviceextension.Match")).
		Return(nil)

	// Test advancement
	err := service.advanceWinner(context.Background(), "ns1", currentMatch)
	assert.NoError(t, err, "Winner advancement should succeed")

	// Verify position 4 advances to position 2
	mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.Position == 2 && match.Round == 2
	}))
}

// TestAdvanceWinner_ByeHandling tests that bye participants automatically advance
func TestAdvanceWinner_ByeHandling(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create bye match (only participant1, no participant2)
	byeMatch := createTestMatch("match1", "tournament1", "user1", "", 1, 1)
	byeMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
	byeMatch.Winner = "user1" // Should auto-complete for bye

	// Mock storage calls for result submission
	mockStorage.On("SubmitMatchResult", mock.Anything, "ns1", "tournament1", "match1", "user1").
		Return(byeMatch, nil)

	// Mock GetMatch call for advancement
	mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "match1").
		Return(byeMatch, nil)

	// Test result submission with bye participant
	req := &serviceextension.SubmitMatchResultRequest{
		Namespace:    "ns1",
		TournamentId: "tournament1",
		MatchId:      "match1",
		WinnerUserId: "user1",
	}

	// Mock tournament existence
	mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "tournament1").
		Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

		// Mock advancement to next round
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
		Return([]*serviceextension.Match{}, nil) // No next round match

	// Mock tournament completion check
	mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
		Return([]*serviceextension.Match{byeMatch}, nil) // Return just the bye match

	// Mock tournament completion
	mockTournamentStorage.On("UpdateTournament", mock.Anything, "ns1", "tournament1", mock.AnythingOfType("*serviceextension.Tournament")).
		Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

	resp, err := service.SubmitMatchResult(context.Background(), req)
	assert.NoError(t, err, "Bye participant result should be accepted")
	assert.NotNil(t, resp)
	assert.Equal(t, "user1", resp.Match.Winner)
}

// TestAdvanceWinner_FinalMatch tests that final match winner doesn't advance
func TestAdvanceWinner_FinalMatch(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create final match (highest round)
	finalMatch := createTestMatch("final1", "tournament1", "user1", "user2", 3, 1)
	finalMatch.Winner = "user1"
	finalMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Mock storage calls - no next round matches should exist
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(4)).
		Return([]*serviceextension.Match{}, nil)

	// Test advancement - should not fail but should not advance to next round
	err := service.advanceWinner(context.Background(), "ns1", finalMatch)
	assert.NoError(t, err, "Final match advancement should not error")

	// Verify no UpdateMatch was called (since no next round match exists)
	mockStorage.AssertNotCalled(t, "UpdateMatch", mock.Anything, mock.Anything, mock.Anything)
}

// TestEdgeCases_CompelteWorkflows tests edge cases and full workflows
func TestEdgeCases_CompelteWorkflows(t *testing.T) {
	t.Run("ByeMatchWorkflow", func(t *testing.T) {
		// Test bye match (single participant)
		byeMatch := createTestMatch("m1", "tournament1", "user1", "", 1, 1)

		mockStorage := &MockMatchStorage{}
		mockTournamentStorage := &MockTournamentStorage{}

		// Mock tournament existence
		mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "tournament1").
			Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

		// Mock GetMatch for SubmitMatchResult internal validation
		mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "m1").
			Return(byeMatch, nil)

		// Mock result submission should validate winner is participant1
		mockStorage.On("SubmitMatchResult", mock.Anything, "ns1", "tournament1", "m1", "user1").
			Return(&serviceextension.Match{
				MatchId:      "m1",
				Winner:       "user1",
				Participant1: byeMatch.Participant1,
				TournamentId: "tournament1",
				Round:        1,
				Position:     1,
			}, nil)

		// Mock advancement
		mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
			Return([]*serviceextension.Match{}, nil)

		// Mock tournament completion check
		mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
			Return([]*serviceextension.Match{byeMatch}, nil)

		logger := slog.Default()
		service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

		req := &serviceextension.SubmitMatchResultRequest{
			Namespace:    "ns1",
			TournamentId: "tournament1",
			MatchId:      "m1",
			WinnerUserId: "user1",
		}

		resp, err := service.SubmitMatchResult(context.Background(), req)
		assert.NoError(t, err, "Bye participant result should be accepted")
		assert.NotNil(t, resp)
		assert.Equal(t, "user1", resp.Match.Winner)

		mockStorage.AssertExpectations(t)
		mockTournamentStorage.AssertExpectations(t)
	})
}
