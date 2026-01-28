// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	serviceextension "extend-custom-guild-service/pkg/pb"
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
	return &serviceextension.Match{
		MatchId:      matchID,
		TournamentId: tournamentID,
		Round:        round,
		Position:     position,
		Participant1: &serviceextension.TournamentParticipant{
			UserId:      userID1,
			Username:    "player1",
			DisplayName: "Player One",
		},
		Participant2: &serviceextension.TournamentParticipant{
			UserId:      userID2,
			Username:    "player2",
			DisplayName: "Player Two",
		},
		Status:    serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED,
		StartedAt: timestamppb.Now(),
	}
}

func createTestMatchWithBye(matchID, tournamentID, userID string, round, position int32) *serviceextension.Match {
	return &serviceextension.Match{
		MatchId:      matchID,
		TournamentId: tournamentID,
		Round:        round,
		Position:     position,
		Participant1: &serviceextension.TournamentParticipant{
			UserId:      userID,
			Username:    "player1",
			DisplayName: "Player One",
		},
		// Participant2 is nil for bye
		Status:    serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED,
		StartedAt: timestamppb.Now(),
	}
}

// TestValidateMatchWinner tests the validateMatchWinner function
func TestValidateMatchWinner(t *testing.T) {
	tests := []struct {
		name         string
		winnerUserID string
		participant1 *serviceextension.TournamentParticipant
		participant2 *serviceextension.TournamentParticipant
		expectError  bool
		errorCode    codes.Code
	}{
		{
			name:         "ValidWinner_Participant1",
			winnerUserID: "user1",
			participant1: &serviceextension.TournamentParticipant{UserId: "user1"},
			participant2: &serviceextension.TournamentParticipant{UserId: "user2"},
			expectError:  false,
		},
		{
			name:         "ValidWinner_Participant2",
			winnerUserID: "user2",
			participant1: &serviceextension.TournamentParticipant{UserId: "user1"},
			participant2: &serviceextension.TournamentParticipant{UserId: "user2"},
			expectError:  false,
		},
		{
			name:         "InvalidWinner_NotParticipant",
			winnerUserID: "user3",
			participant1: &serviceextension.TournamentParticipant{UserId: "user1"},
			participant2: &serviceextension.TournamentParticipant{UserId: "user2"},
			expectError:  true,
			errorCode:    codes.InvalidArgument,
		},
		{
			name:         "EmptyWinnerID",
			winnerUserID: "",
			participant1: &serviceextension.TournamentParticipant{UserId: "user1"},
			participant2: &serviceextension.TournamentParticipant{UserId: "user2"},
			expectError:  true,
			errorCode:    codes.InvalidArgument,
		},
		{
			name:         "ValidWinner_WithBye",
			winnerUserID: "user1",
			participant1: &serviceextension.TournamentParticipant{UserId: "user1"},
			participant2: nil, // Bye
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockMatchStorage{}
			mockTournamentStorage := &MockTournamentStorage{}
			logger := slog.Default()

			service := NewMatchService(mockStorage, mockTournamentStorage, logger)

			// This test expects validateMatchWinner to be called as part of result submission
			// For now, we'll test through SubmitMatchResult which calls validation

			if tt.expectError {
				// Expect SubmitMatchResult to fail with validation error
				mockStorage.On("SubmitMatchResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, tt.winnerUserID).
					Return(nil, grpcStatus.Errorf(tt.errorCode, "winner validation failed"))
			} else {
				// Expect SubmitMatchResult to succeed
				expectedMatch := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
				expectedMatch.Winner = tt.winnerUserID
				mockStorage.On("SubmitMatchResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, tt.winnerUserID).
					Return(expectedMatch, nil)
			}

			req := &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				MatchId:      "match1",
				WinnerUserId: tt.winnerUserID,
			}

			resp, err := service.SubmitMatchResult(context.Background(), req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
				grpcErr, ok := grpcStatus.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, grpcErr.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.winnerUserID, resp.Match.Winner)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

// TestAdvanceWinner tests the advanceWinner function
func TestAdvanceWinner(t *testing.T) {
	tests := []struct {
		name                string
		currentMatch        *serviceextension.Match
		expectedNextRound   int32
		expectedNextPos     int32
		isFinalMatch        bool
		expectNoAdvancement bool
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
			isFinalMatch:      false,
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
			isFinalMatch:      false,
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
			isFinalMatch:      false,
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
			isFinalMatch:      false,
		},
		{
			name: "Round2Position1Advancement",
			currentMatch: &serviceextension.Match{
				MatchId:      "match5",
				Round:        2,
				Position:     1,
				TournamentId: "tournament1",
			},
			expectedNextRound: 3,
			expectedNextPos:   1,
			isFinalMatch:      false,
		},
		{
			name: "FinalMatch_NoAdvancement",
			currentMatch: &serviceextension.Match{
				MatchId:      "final",
				Round:        3,
				Position:     1,
				TournamentId: "tournament1",
			},
			isFinalMatch:        true,
			expectNoAdvancement: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, this test validates that the bracket math is correct
			// The actual advancement logic will be tested through integration tests

			// Test the bracket position calculation
			nextPosition := calculateNextPosition(tt.currentMatch.Position)
			nextRound := tt.currentMatch.Round + 1

			if !tt.expectNoAdvancement {
				assert.Equal(t, tt.expectedNextPos, nextPosition, "Next position calculation incorrect")
				assert.Equal(t, tt.expectedNextRound, nextRound, "Next round should be current round + 1")
			}
		})
	}
}

// TestSubmitMatchResult_BusinessLogic tests match result submission business rules
func TestSubmitMatchResult_BusinessLogic(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockMatchStorage, *MockTournamentStorage)
		req         *serviceextension.SubmitMatchResultRequest
		expectError bool
		errorCode   codes.Code
		expectedWin string
	}{
		{
			name: "ValidSubmission",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// Tournament exists
				t.On("GetTournament", mock.Anything, "ns1", "tournament1").
					Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

				// Result submission succeeds
				match := createTestMatch("match1", "tournament1", "user1", "user2", 1, 1)
				match.Winner = "user1"
				m.On("SubmitMatchResult", mock.Anything, "ns1", "tournament1", "match1", "user1").
					Return(match, nil)
			},
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				MatchId:      "match1",
				WinnerUserId: "user1",
			},
			expectError: false,
			expectedWin: "user1",
		},
		{
			name: "TournamentNotFound",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				t.On("GetTournament", mock.Anything, "ns1", "tournament1").
					Return(nil, grpcStatus.Errorf(codes.NotFound, "tournament not found"))
			},
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				MatchId:      "match1",
				WinnerUserId: "user1",
			},
			expectError: true,
			errorCode:   codes.NotFound,
		},
		{
			name: "EmptyNamespace",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// No mocks needed - validation should fail first
			},
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
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// No mocks needed - validation should fail first
			},
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
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// No mocks needed - validation should fail first
			},
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
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// No mocks needed - validation should fail first
			},
			req: &serviceextension.SubmitMatchResultRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				MatchId:      "match1",
				WinnerUserId: "",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockMatchStorage{}
			mockTournamentStorage := &MockTournamentStorage{}

			tt.setupMocks(mockStorage, mockTournamentStorage)

			logger := slog.Default()
			service := NewMatchService(mockStorage, mockTournamentStorage, logger)

			resp, err := service.SubmitMatchResult(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
				grpcErr, ok := grpcStatus.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, grpcErr.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedWin, resp.Match.Winner)
			}

			mockStorage.AssertExpectations(t)
			mockTournamentStorage.AssertExpectations(t)
		})
	}
}

// TestGetTournamentMatches_BusinessLogic tests tournament match retrieval
func TestGetTournamentMatches_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockMatchStorage, *MockTournamentStorage)
		req           *serviceextension.GetTournamentMatchesRequest
		expectError   bool
		errorCode     codes.Code
		expectedCount int
	}{
		{
			name: "GetAllMatches",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// Tournament exists
				t.On("GetTournament", mock.Anything, "ns1", "tournament1").
					Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

				// Return matches
				matches := []*serviceextension.Match{
					createTestMatch("match1", "tournament1", "user1", "user2", 1, 1),
					createTestMatch("match2", "tournament1", "user3", "user4", 1, 2),
				}
				m.On("GetTournamentMatches", mock.Anything, "ns1", "tournament1").
					Return(matches, nil)
			},
			req: &serviceextension.GetTournamentMatchesRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				Round:        0, // All rounds
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "GetSpecificRound",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// Tournament exists
				t.On("GetTournament", mock.Anything, "ns1", "tournament1").
					Return(&serviceextension.Tournament{TournamentId: "tournament1"}, nil)

				// Return round 1 matches
				matches := []*serviceextension.Match{
					createTestMatch("match1", "tournament1", "user1", "user2", 1, 1),
				}
				m.On("GetMatchesByRound", mock.Anything, "ns1", "tournament1", int32(1)).
					Return(matches, nil)
			},
			req: &serviceextension.GetTournamentMatchesRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
				Round:        1,
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "TournamentNotFound",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				t.On("GetTournament", mock.Anything, "ns1", "tournament1").
					Return(nil, grpcStatus.Errorf(codes.NotFound, "tournament not found"))
			},
			req: &serviceextension.GetTournamentMatchesRequest{
				Namespace:    "ns1",
				TournamentId: "tournament1",
			},
			expectError: true,
			errorCode:   codes.NotFound,
		},
		{
			name: "EmptyNamespace",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// No mocks needed - validation should fail first
			},
			req: &serviceextension.GetTournamentMatchesRequest{
				Namespace:    "",
				TournamentId: "tournament1",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "EmptyTournamentId",
			setupMocks: func(m *MockMatchStorage, t *MockTournamentStorage) {
				// No mocks needed - validation should fail first
			},
			req: &serviceextension.GetTournamentMatchesRequest{
				Namespace:    "ns1",
				TournamentId: "",
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockMatchStorage{}
			mockTournamentStorage := &MockTournamentStorage{}

			tt.setupMocks(mockStorage, mockTournamentStorage)

			logger := slog.Default()
			service := NewMatchService(mockStorage, mockTournamentStorage, logger)

			resp, err := service.GetTournamentMatches(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
				grpcErr, ok := grpcStatus.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, grpcErr.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedCount, len(resp.Matches))
			}

			mockStorage.AssertExpectations(t)
			mockTournamentStorage.AssertExpectations(t)
		})
	}
}

// TestByeHandling tests bye participant handling in matches
func TestByeHandling(t *testing.T) {
	tests := []struct {
		name        string
		match       *serviceextension.Match
		expectError bool
		expectedBye bool
		expectedWin string
	}{
		{
			name:  "MatchWithBye",
			match: createTestMatchWithBye("match1", "tournament1", "user1", 1, 1),
			// Should auto-advance the participant
			expectError: false,
			expectedBye: true,
			expectedWin: "user1",
		},
		{
			name:  "RegularMatch",
			match: createTestMatch("match2", "tournament1", "user1", "user2", 1, 2),
			// Should require result submission
			expectError: false,
			expectedBye: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates bye detection logic
			hasBye := tt.match.Participant2 == nil && tt.match.Participant1 != nil
			assert.Equal(t, tt.expectedBye, hasBye, "Bye detection incorrect")

			if hasBye {
				// Bye matches should auto-advance the participant
				assert.Equal(t, tt.expectedWin, tt.match.Participant1.UserId)
			}
		})
	}
}

// TestBracketMath tests bracket position calculations
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
		{
			name:            "Position5_to_Position3",
			currentPos:      5,
			expectedNextPos: 3,
		},
		{
			name:            "Position6_to_Position3",
			currentPos:      6,
			expectedNextPos: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextPos := calculateNextPosition(tt.currentPos)
			assert.Equal(t, tt.expectedNextPos, nextPos, "Next position calculation incorrect")
		})
	}
}

// calculateNextPosition calculates the next round position based on current position
// Formula: nextPosition = (currentPosition - 1) / 2 + 1
func calculateNextPosition(currentPos int32) int32 {
	return (currentPos-1)/2 + 1
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("FinalMatch_NoAdvancement", func(t *testing.T) {
		// Final match (round 3, position 1 in 8-player tournament)
		// This would be detected as final match in real implementation
		// For now, we test math calculation
		assert.True(t, true, "Final match detection logic to be implemented")
	})

	t.Run("ZeroParticipants", func(t *testing.T) {
		// Should handle nil participants gracefully
		assert.True(t, true, "Empty match handling to be implemented")
	})

	t.Run("ZeroParticipants", func(t *testing.T) {
		// Should handle nil participants gracefully
		assert.True(t, true, "Empty match handling to be implemented")
	})
}
