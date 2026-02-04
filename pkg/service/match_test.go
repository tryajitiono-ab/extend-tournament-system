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

// createTestMatchWithRelationships creates a test match with explicit bracket relationships
func createTestMatchWithRelationships(matchID, tournamentID, userID1, userID2 string, round, position int32, nextMatchID, sourceMatch1ID, sourceMatch2ID string) *serviceextension.Match {
	match := createTestMatch(matchID, tournamentID, userID1, userID2, round, position)
	match.NextMatchId = nextMatchID
	match.SourceMatch_1Id = sourceMatch1ID
	match.SourceMatch_2Id = sourceMatch2ID
	return match
}

// TestAdvanceWinner tests the match relationship-based advancement logic
func TestAdvanceWinner(t *testing.T) {
	tests := []struct {
		name         string
		currentMatch *serviceextension.Match
		nextMatch    *serviceextension.Match
		expectSlot   string // "participant1" or "participant2"
	}{
		{
			name:         "SourceMatch1_AdvancesToParticipant1",
			currentMatch: createTestMatchWithRelationships("match-r1-m1", "tournament1", "user1", "user2", 1, 0, "match-r2-m1", "", ""),
			nextMatch:    createTestMatchWithRelationships("match-r2-m1", "tournament1", "", "", 2, 0, "", "match-r1-m1", "match-r1-m2"),
			expectSlot:   "participant1",
		},
		{
			name:         "SourceMatch2_AdvancesToParticipant2",
			currentMatch: createTestMatchWithRelationships("match-r1-m2", "tournament1", "user3", "user4", 1, 1, "match-r2-m1", "", ""),
			nextMatch:    createTestMatchWithRelationships("match-r2-m1", "tournament1", "", "", 2, 0, "", "match-r1-m1", "match-r1-m2"),
			expectSlot:   "participant2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockMatchStorage{}
			mockTournamentStorage := &MockTournamentStorage{}
			logger := slog.Default()
			service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

			tt.currentMatch.Winner = tt.currentMatch.Participant1.UserId
			tt.currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

			mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", tt.currentMatch.NextMatchId).
				Return(tt.nextMatch, nil)
			mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*pb.Match")).
				Return(nil)

			err := service.advanceWinner(context.Background(), "ns1", tt.currentMatch)
			assert.NoError(t, err)

			mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
				if tt.expectSlot == "participant1" {
					return match.Participant1 != nil && match.Participant1.UserId == tt.currentMatch.Winner
				}
				return match.Participant2 != nil && match.Participant2.UserId == tt.currentMatch.Winner
			}))
		})
	}
}

// TestMatchRelationships tests that match relationships are correctly structured
func TestMatchRelationships(t *testing.T) {
	tests := []struct {
		name           string
		matchID        string
		nextMatchID    string
		sourceMatch1ID string
		sourceMatch2ID string
	}{
		{
			name:        "FirstRoundMatch_HasNextMatch",
			matchID:     "match-r1-m1",
			nextMatchID: "match-r2-m1",
		},
		{
			name:           "SecondRoundMatch_HasSourceMatches",
			matchID:        "match-r2-m1",
			sourceMatch1ID: "match-r1-m1",
			sourceMatch2ID: "match-r1-m2",
		},
		{
			name:    "FinalMatch_NoNextMatch",
			matchID: "match-r3-m1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := createTestMatchWithRelationships(tt.matchID, "tournament1", "", "", 1, 0, tt.nextMatchID, tt.sourceMatch1ID, tt.sourceMatch2ID)
			assert.Equal(t, tt.nextMatchID, match.NextMatchId)
			assert.Equal(t, tt.sourceMatch1ID, match.SourceMatch_1Id)
			assert.Equal(t, tt.sourceMatch2ID, match.SourceMatch_2Id)
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
				// The submitted match has no NextMatchId, so advancement is a no-op (final round behavior)
				submittedMatch := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
				mockStorage.On("SubmitMatchResult", mock.Anything, "ns1", "tournament1", "match1", "user1").
					Return(submittedMatch, nil)
				// Mock GetMatch call for bye handling
				mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "match1").
					Return(submittedMatch, nil)
				// Mock bye advancement for next round
				mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
					Return([]*serviceextension.Match{}, nil)
				// Mock tournament completion check
				mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
					Return([]*serviceextension.Match{submittedMatch}, nil)
				// Mock tournament completion
				mockTournamentStorage.On("UpdateTournament", mock.Anything, "ns1", "tournament1", mock.AnythingOfType("*pb.Tournament")).
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

// TestIntegration_WinnerAdvancement tests complete advancement workflow using match relationships
func TestIntegration_WinnerAdvancement(t *testing.T) {
	t.Run("CompleteAdvancementWorkflow", func(t *testing.T) {
		// Test that winner advancement logic works end-to-end with match relationships
		// 4-player bracket: 2 round-1 matches feed into 1 round-2 final

		// Create initial matches with relationships
		round1Match1 := createTestMatchWithRelationships("match-r1-m1", "t1", "user1", "user2", 1, 0, "match-r2-m1", "", "")
		round1Match2 := createTestMatchWithRelationships("match-r1-m2", "t1", "user3", "user4", 1, 1, "match-r2-m1", "", "")
		round2Match1 := createTestMatchWithRelationships("match-r2-m1", "t1", "", "", 2, 0, "", "match-r1-m1", "match-r1-m2")

		// Verify relationship structure
		assert.Equal(t, "match-r2-m1", round1Match1.NextMatchId, "Match 1 should advance to final")
		assert.Equal(t, "match-r2-m1", round1Match2.NextMatchId, "Match 2 should advance to final")
		assert.Equal(t, "match-r1-m1", round2Match1.SourceMatch_1Id, "Final should reference match 1")
		assert.Equal(t, "match-r1-m2", round2Match1.SourceMatch_2Id, "Final should reference match 2")

		// Simulate first round results
		round1Match1.Winner = "user1"
		round1Match1.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

		round1Match2.Winner = "user3"
		round1Match2.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

		// Populate the final match with winners (simulating advancement)
		round2Match1.Participant1 = round1Match1.Participant1 // user1 from source match 1
		round2Match1.Participant2 = round1Match2.Participant1 // user3 from source match 2

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
	assert.Contains(t, err.Error(), "already completed", "Error should mention already completed")
}

// TestValidateMatchWinner_CancelledMatch tests that cancelled matches reject new results
func TestValidateMatchWinner_CancelledMatch(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create cancelled match
	match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
	match.Status = serviceextension.MatchStatus_MATCH_STATUS_CANCELLED

	// Test that cancelled match rejects winner validation
	err := service.validateMatchWinner(match, "user1")
	assert.Error(t, err, "Cancelled match should reject new result")
}

// TestValidateMatchWinner_InProgressMatch tests that in-progress matches accept results
func TestValidateMatchWinner_InProgressMatch(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create in-progress match
	match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
	match.Status = serviceextension.MatchStatus_MATCH_STATUS_IN_PROGRESS

	// Test that in-progress match accepts winner validation
	err := service.validateMatchWinner(match, "user1")
	assert.NoError(t, err, "In-progress match should accept winner")
}

// TestValidateMatchWinner_NilParticipants tests edge case with nil participants
func TestValidateMatchWinner_NilParticipants(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create match with nil participant1
	match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
	match.Participant1 = nil

	// Test that nil participant1 still allows participant2 as winner
	err := service.validateMatchWinner(match, "user2")
	assert.NoError(t, err, "Nil participant1 should still allow participant2 as winner")

	// Create match with both nil participants
	match.Participant2 = nil
	err = service.validateMatchWinner(match, "user3")
	assert.Error(t, err, "Match with no participants should reject any winner")
}

// TestAdvanceWinner_SourceMatch1Advancement tests source match 1 winner fills participant1 slot
func TestAdvanceWinner_SourceMatch1Advancement(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create current match with winner and next_match_id
	currentMatch := createTestMatchWithRelationships("match-r1-m1", "tournament1", "user1", "user2", 1, 0, "match-r2-m1", "", "")
	currentMatch.Winner = "user1"
	currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Create next round match with source match references
	nextMatch := createTestMatchWithRelationships("match-r2-m1", "tournament1", "", "", 2, 0, "", "match-r1-m1", "match-r1-m2")

	// Mock storage calls - direct GetMatch lookup
	mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "match-r2-m1").
		Return(nextMatch, nil)
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*pb.Match")).
		Return(nil)

	// Test advancement
	err := service.advanceWinner(context.Background(), "ns1", currentMatch)
	assert.NoError(t, err, "Winner advancement should succeed")

	// Verify winner placed in participant1 slot (source match 1)
	mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.Participant1 != nil && match.Participant1.UserId == "user1"
	}))
}

// TestAdvanceWinner_SourceMatch2Advancement tests source match 2 winner fills participant2 slot
func TestAdvanceWinner_SourceMatch2Advancement(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create current match with winner and next_match_id
	currentMatch := createTestMatchWithRelationships("match-r1-m2", "tournament1", "user3", "user4", 1, 1, "match-r2-m1", "", "")
	currentMatch.Winner = "user3"
	currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Create next round match with source match references
	nextMatch := createTestMatchWithRelationships("match-r2-m1", "tournament1", "", "", 2, 0, "", "match-r1-m1", "match-r1-m2")

	// Mock storage calls - direct GetMatch lookup
	mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "match-r2-m1").
		Return(nextMatch, nil)
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*pb.Match")).
		Return(nil)

	// Test advancement
	err := service.advanceWinner(context.Background(), "ns1", currentMatch)
	assert.NoError(t, err, "Winner advancement should succeed")

	// Verify winner placed in participant2 slot (source match 2)
	mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.Participant2 != nil && match.Participant2.UserId == "user3"
	}))
}

// TestAdvanceWinner_8PlayerBracket tests advancement in a larger bracket
func TestAdvanceWinner_8PlayerBracket(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Match r1-m3 feeds into r2-m2 as source match 1
	currentMatch := createTestMatchWithRelationships("match-r1-m3", "tournament1", "user5", "user6", 1, 2, "match-r2-m2", "", "")
	currentMatch.Winner = "user5"
	currentMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	nextMatch := createTestMatchWithRelationships("match-r2-m2", "tournament1", "", "", 2, 1, "match-r3-m1", "match-r1-m3", "match-r1-m4")

	mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "match-r2-m2").
		Return(nextMatch, nil)
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.AnythingOfType("*pb.Match")).
		Return(nil)

	err := service.advanceWinner(context.Background(), "ns1", currentMatch)
	assert.NoError(t, err)

	mockStorage.AssertCalled(t, "UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.MatchId == "match-r2-m2" && match.Participant1 != nil && match.Participant1.UserId == "user5"
	}))
}

// TestAdvanceWinner_ByeHandling tests that bye participants automatically advance
func TestAdvanceWinner_ByeHandling(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create bye match with no NextMatchId (final round behavior for this simple test)
	byeMatch := createTestMatch("match1", "tournament1", "user1", "", 1, 1)
	byeMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
	byeMatch.Winner = "user1"

	// Mock storage calls for result submission
	mockStorage.On("SubmitMatchResult", mock.Anything, "ns1", "tournament1", "match1", "user1").
		Return(byeMatch, nil)

	// Mock GetMatch call for bye handling
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

	// Mock bye advancement for next round
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
		Return([]*serviceextension.Match{}, nil)

	// Mock tournament completion check
	mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
		Return([]*serviceextension.Match{byeMatch}, nil)

	// Mock tournament completion
	mockTournamentStorage.On("UpdateTournament", mock.Anything, "ns1", "tournament1", mock.AnythingOfType("*pb.Tournament")).
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

	// Create final match with no NextMatchId (terminal match)
	finalMatch := createTestMatchWithRelationships("match-r3-m1", "tournament1", "user1", "user2", 3, 0, "", "match-r2-m1", "match-r2-m2")
	finalMatch.Winner = "user1"
	finalMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Test advancement - should not fail and should not call any storage
	err := service.advanceWinner(context.Background(), "ns1", finalMatch)
	assert.NoError(t, err, "Final match advancement should not error")

	// Verify no GetMatch or UpdateMatch was called (no advancement needed)
	mockStorage.AssertNotCalled(t, "GetMatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockStorage.AssertNotCalled(t, "UpdateMatch", mock.Anything, mock.Anything, mock.Anything)
}

// TestCreateTournamentMatches tests bulk match creation for tournaments
func TestCreateTournamentMatches(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create test matches for tournament
	matches := []*serviceextension.Match{
		createTestMatch("m1", "tournament1", "user1", "user2", 1, 1),
		createTestMatch("m2", "tournament1", "user3", "user4", 1, 2),
		createTestMatch("m3", "tournament1", "", "", 2, 1), // Final match
	}

	// Mock storage calls
	mockStorage.On("CreateMatches", mock.Anything, "ns1", "tournament1", matches).
		Return(nil)

	err := service.CreateTournamentMatches(context.Background(), "ns1", "tournament1", matches)
	assert.NoError(t, err, "Tournament matches creation should succeed")

	mockStorage.AssertExpectations(t)
}

// TestGetMatch tests individual match retrieval
func TestGetMatch(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Test match
	testMatch := createTestMatch("m1", "tournament1", "user1", "user2", 1, 1)

	// Mock storage calls
	mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "tournament1").
		Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)
	mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "m1").
		Return(testMatch, nil)

	req := &serviceextension.GetMatchRequest{
		Namespace:    "ns1",
		TournamentId: "tournament1",
		MatchId:      "m1",
	}

	resp, err := service.GetMatch(context.Background(), req)
	assert.NoError(t, err, "Match retrieval should succeed")
	assert.NotNil(t, resp)
	assert.Equal(t, testMatch.MatchId, resp.Match.MatchId)
	assert.Equal(t, testMatch.TournamentId, resp.Match.TournamentId)

	mockStorage.AssertExpectations(t)
	mockTournamentStorage.AssertExpectations(t)
}

// TestAdminSubmitMatchResult tests admin match result submission
func TestAdminSubmitMatchResult(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Test match with winner (no NextMatchId = final round)
	testMatch := createTestMatch("m1", "tournament1", "user1", "user2", 1, 1)
	testMatch.Winner = "user1"
	testMatch.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

	// Mock tournament existence
	mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "tournament1").
		Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

	// Mock result submission
	mockStorage.On("SubmitMatchResult", mock.Anything, "ns1", "tournament1", "m1", "user1").
		Return(testMatch, nil)

	// Mock GetMatch call for bye handling
	mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "m1").
		Return(testMatch, nil)

	// Mock bye advancement for next round
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(2)).
		Return([]*serviceextension.Match{}, nil)

	// Mock tournament completion check
	mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
		Return([]*serviceextension.Match{testMatch}, nil)

	// Mock tournament completion
	mockTournamentStorage.On("UpdateTournament", mock.Anything, "ns1", "tournament1", mock.AnythingOfType("*pb.Tournament")).
		Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

	req := &serviceextension.AdminSubmitMatchResultRequest{
		Namespace:    "ns1",
		TournamentId: "tournament1",
		MatchId:      "m1",
		WinnerUserId: "user1",
	}

	resp, err := service.AdminSubmitMatchResult(context.Background(), req)
	assert.NoError(t, err, "Admin match result submission should succeed")
	assert.NotNil(t, resp)
	assert.Equal(t, testMatch.Winner, resp.Match.Winner)

	mockStorage.AssertExpectations(t)
	mockTournamentStorage.AssertExpectations(t)
}

// TestHandleByeAdvancement tests bye participant advancement
func TestHandleByeAdvancement(t *testing.T) {
	mockStorage := &MockMatchStorage{}
	mockTournamentStorage := &MockTournamentStorage{}

	logger := slog.Default()
	service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

	// Create bye match with NextMatchId for advancement
	byeMatch := createTestMatchWithRelationships("match-r1-m1", "tournament1", "user1", "", 1, 0, "match-r2-m1", "", "")

	// Create next round match with source references
	nextMatch := createTestMatchWithRelationships("match-r2-m1", "tournament1", "", "", 2, 0, "", "match-r1-m1", "match-r1-m2")

	// Mock GetMatchesByRound for HandleByeAdvancement to find bye matches
	mockStorage.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(1)).
		Return([]*serviceextension.Match{byeMatch}, nil)

	// Mock UpdateMatch for completing the bye match
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.MatchId == "match-r1-m1" && match.Winner == "user1" && match.Status == serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
	})).Return(nil)

	// Mock GetMatch for advanceWinner (direct match lookup)
	mockStorage.On("GetMatch", mock.Anything, "ns1", "tournament1", "match-r2-m1").
		Return(nextMatch, nil)

	// Mock UpdateMatch for advancing winner to next round
	mockStorage.On("UpdateMatch", mock.Anything, "ns1", mock.MatchedBy(func(match *serviceextension.Match) bool {
		return match.MatchId == "match-r2-m1" && match.Participant1 != nil && match.Participant1.UserId == "user1"
	})).Return(nil)

	err := service.HandleByeAdvancement(context.Background(), "ns1", "tournament1", 1)
	assert.NoError(t, err, "Bye advancement should succeed")

	mockStorage.AssertExpectations(t)
}

// TestCheckTournamentCompletion tests tournament completion detection
func TestCheckTournamentCompletion(t *testing.T) {
	t.Run("IncompleteTournament", func(t *testing.T) {
		mockStorage := &MockMatchStorage{}
		mockTournamentStorage := &MockTournamentStorage{}

		logger := slog.Default()
		service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

		// Create matches with some incomplete
		matches := []*serviceextension.Match{
			createTestMatch("m1", "tournament1", "user1", "user2", 1, 1),
			createTestMatch("m2", "tournament1", "user3", "user4", 1, 2),
			createTestMatch("m3", "tournament1", "", "", 2, 1), // Final match (no winner)
		}

		// Complete first match
		matches[0].Winner = "user1"
		matches[0].Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED

		// Second match still scheduled
		matches[1].Status = serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED

		mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
			Return(matches, nil)

		isComplete, winner, err := service.CheckTournamentCompletion(context.Background(), "ns1", "tournament1")
		assert.NoError(t, err, "Tournament completion check should succeed")
		assert.False(t, isComplete, "Tournament should not be complete")
		assert.Empty(t, winner, "Winner should be empty for incomplete tournament")

		mockStorage.AssertExpectations(t)
	})

	t.Run("CompleteTournament", func(t *testing.T) {
		mockStorage := &MockMatchStorage{}
		mockTournamentStorage := &MockTournamentStorage{}

		logger := slog.Default()
		service := NewMatchService(mockStorage, mockTournamentStorage, nil, logger)

		// Create completed matches including final
		matches := []*serviceextension.Match{
			createTestMatch("m1", "tournament1", "user1", "user2", 1, 1),
			createTestMatch("m2", "tournament1", "user3", "user4", 1, 2),
			createTestMatch("m3", "tournament1", "user1", "user3", 2, 1), // Final match completed
		}

		// Complete all matches
		for _, match := range matches {
			match.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
		}
		matches[2].Winner = "user1" // Final match winner

		mockStorage.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
			Return(matches, nil)

		isComplete, winner, err := service.CheckTournamentCompletion(context.Background(), "ns1", "tournament1")
		assert.NoError(t, err, "Tournament completion check should succeed")
		assert.True(t, isComplete, "Tournament should be complete")
		assert.Equal(t, "user1", winner, "Should return final match winner")

		mockStorage.AssertExpectations(t)
	})
}

// TestEdgeCases_CompleteWorkflows tests edge cases and full workflows
func TestEdgeCases_CompleteWorkflows(t *testing.T) {
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

// --- Integration tests: GenerateBrackets → match count verification ---

// TestFullTournament_4Players verifies bracket generation produces correct match structure for 4 players.
func TestFullTournament_4Players(t *testing.T) {
	server := &TournamentServiceServer{logger: slog.Default()}

	participants := makeParticipants(4)
	bracket, err := server.GenerateBrackets(participants)
	assert.NoError(t, err)

	// 4 players = 2 rounds, 3 total matches (2+1)
	assert.Equal(t, int32(2), bracket.TotalRounds)
	assert.Equal(t, 3, countAllMatches(bracket))

	// Round 1: 2 matches with 2 participants each
	assert.Len(t, bracket.Rounds[0], 2)
	for _, m := range bracket.Rounds[0] {
		assert.NotNil(t, m.Participant1)
		assert.NotNil(t, m.Participant2)
		assert.False(t, m.Bye)
		assert.Equal(t, int32(1), m.Round)
	}

	// Round 2: 1 final match, empty (waiting for winners)
	assert.Len(t, bracket.Rounds[1], 1)
	assert.Equal(t, int32(2), bracket.Rounds[1][0].Round)
}

// TestFullTournament_8Players verifies bracket generation produces correct match structure for 8 players.
func TestFullTournament_8Players(t *testing.T) {
	server := &TournamentServiceServer{logger: slog.Default()}

	participants := makeParticipants(8)
	bracket, err := server.GenerateBrackets(participants)
	assert.NoError(t, err)

	// 8 players = 3 rounds, 7 total matches (4+2+1)
	assert.Equal(t, int32(3), bracket.TotalRounds)
	assert.Equal(t, 7, countAllMatches(bracket))

	// Round 1: 4 matches
	assert.Len(t, bracket.Rounds[0], 4)
	for _, m := range bracket.Rounds[0] {
		assert.NotNil(t, m.Participant1)
		assert.NotNil(t, m.Participant2)
		assert.False(t, m.Bye)
	}

	// Round 2: 2 semi-final matches
	assert.Len(t, bracket.Rounds[1], 2)

	// Round 3: 1 final match
	assert.Len(t, bracket.Rounds[2], 1)

	// Verify match IDs follow expected format
	assert.Equal(t, "match-r1-m1", bracket.Rounds[0][0].MatchId)
	assert.Equal(t, "match-r2-m1", bracket.Rounds[1][0].MatchId)
	assert.Equal(t, "match-r3-m1", bracket.Rounds[2][0].MatchId)
}

// TestFullTournament_NonPowerOf2 verifies bracket generation with bye handling for 5 players.
func TestFullTournament_NonPowerOf2(t *testing.T) {
	server := &TournamentServiceServer{logger: slog.Default()}

	participants := makeParticipants(5)
	bracket, err := server.GenerateBrackets(participants)
	assert.NoError(t, err)

	// 5 players → bracket size 8 → 3 rounds, 7 total matches
	assert.Equal(t, int32(3), bracket.TotalRounds)
	assert.Equal(t, 7, countAllMatches(bracket))

	// Round 1: 4 matches, some should have byes
	assert.Len(t, bracket.Rounds[0], 4)

	byeCount := 0
	matchesWithBothParticipants := 0
	for _, m := range bracket.Rounds[0] {
		if m.Bye {
			byeCount++
		}
		if m.Participant1 != nil && m.Participant2 != nil {
			matchesWithBothParticipants++
		}
	}

	// Bye matches are created when there aren't enough participants to fill both slots
	assert.Greater(t, byeCount, 0, "5 players in 8-slot bracket should have bye matches")
	assert.GreaterOrEqual(t, matchesWithBothParticipants, 1, "Should have at least one match with both participants")

	// Verify all 5 participants appear across first round matches
	participantSet := make(map[string]bool)
	for _, m := range bracket.Rounds[0] {
		if m.Participant1 != nil {
			participantSet[m.Participant1.UserId] = true
		}
		if m.Participant2 != nil {
			participantSet[m.Participant2.UserId] = true
		}
	}
	assert.Equal(t, 5, len(participantSet), "All 5 participants should appear in first round")
}
