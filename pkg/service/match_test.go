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
