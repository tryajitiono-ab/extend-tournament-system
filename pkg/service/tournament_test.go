// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"log/slog"
	"testing"

	extendtournamentservice "extend-tournament-service/pkg/common"
	serviceextension "extend-tournament-service/pkg/pb"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// newTestTournamentServer creates a TournamentServiceServer for testing GenerateBrackets
// and ValidateStatusTransition (methods that don't need storage or auth).
func newTestTournamentServer() *TournamentServiceServer {
	return &TournamentServiceServer{
		logger: slog.Default(),
	}
}

// newTestAuthInterceptor creates an auth interceptor with nil validator (auth disabled).
func newTestAuthInterceptor() *extendtournamentservice.TournamentAuthInterceptor {
	return extendtournamentservice.NewTournamentAuthInterceptor(iam.OAuth20Service{}, nil, slog.Default())
}

// --- GenerateBrackets tests ---

func TestGenerateBrackets_2Players(t *testing.T) {
	s := newTestTournamentServer()
	participants := makeParticipants(2)

	bracket, err := s.GenerateBrackets(participants)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), bracket.TotalRounds)
	assert.Len(t, bracket.Rounds, 1)
	assert.Len(t, bracket.Rounds[0], 1) // 1 match
	assert.Equal(t, 1, countAllMatches(bracket))
}

func TestGenerateBrackets_4Players(t *testing.T) {
	s := newTestTournamentServer()
	participants := makeParticipants(4)

	bracket, err := s.GenerateBrackets(participants)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), bracket.TotalRounds)
	assert.Len(t, bracket.Rounds, 2)
	assert.Len(t, bracket.Rounds[0], 2) // 2 first-round matches
	assert.Len(t, bracket.Rounds[1], 1) // 1 final match
	assert.Equal(t, 3, countAllMatches(bracket)) // 2+1
}

func TestGenerateBrackets_8Players(t *testing.T) {
	s := newTestTournamentServer()
	participants := makeParticipants(8)

	bracket, err := s.GenerateBrackets(participants)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), bracket.TotalRounds)
	assert.Len(t, bracket.Rounds, 3)
	assert.Len(t, bracket.Rounds[0], 4) // 4 first-round matches
	assert.Len(t, bracket.Rounds[1], 2) // 2 semi-finals
	assert.Len(t, bracket.Rounds[2], 1) // 1 final
	assert.Equal(t, 7, countAllMatches(bracket)) // 4+2+1
}

func TestGenerateBrackets_16Players(t *testing.T) {
	s := newTestTournamentServer()
	participants := makeParticipants(16)

	bracket, err := s.GenerateBrackets(participants)
	assert.NoError(t, err)
	assert.Equal(t, int32(4), bracket.TotalRounds)
	assert.Equal(t, 15, countAllMatches(bracket)) // 8+4+2+1
}

func TestGenerateBrackets_NonPowerOf2(t *testing.T) {
	tests := []struct {
		name              string
		count             int
		expectedRounds    int32
		expectedR1Matches int
		expectedTotal     int
	}{
		{"5 players", 5, 3, 4, 7},
		{"3 players", 3, 2, 2, 3},
		{"6 players", 6, 3, 4, 7},
		{"7 players", 7, 3, 4, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestTournamentServer()
			participants := makeParticipants(tt.count)

			bracket, err := s.GenerateBrackets(participants)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedRounds, bracket.TotalRounds)
			assert.Len(t, bracket.Rounds[0], tt.expectedR1Matches)
			assert.Equal(t, tt.expectedTotal, countAllMatches(bracket))

			// Verify bye matches exist for non-power-of-2
			byeCount := 0
			for _, m := range bracket.Rounds[0] {
				if m.Bye {
					byeCount++
				}
			}
			assert.Greater(t, byeCount, 0, "Non-power-of-2 should have bye matches")
		})
	}
}

func TestGenerateBrackets_MinPlayers(t *testing.T) {
	s := newTestTournamentServer()
	participants := makeParticipants(2)

	bracket, err := s.GenerateBrackets(participants)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), bracket.TotalRounds)
	assert.NotNil(t, bracket.Rounds[0][0].Participant1)
	assert.NotNil(t, bracket.Rounds[0][0].Participant2)
	assert.False(t, bracket.Rounds[0][0].Bye)
}

func TestGenerateBrackets_LessThan2(t *testing.T) {
	s := newTestTournamentServer()

	t.Run("1 participant", func(t *testing.T) {
		_, err := s.GenerateBrackets(makeParticipants(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 2 participants")
	})

	t.Run("0 participants", func(t *testing.T) {
		_, err := s.GenerateBrackets([]TournamentParticipant{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 2 participants")
	})
}

// --- ValidateStatusTransition tests ---

func TestValidateStatusTransition(t *testing.T) {
	s := newTestTournamentServer()

	validTransitions := []struct {
		name string
		from serviceextension.TournamentStatus
		to   serviceextension.TournamentStatus
	}{
		{"DRAFT->ACTIVE", serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT, serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE},
		{"ACTIVE->STARTED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE, serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED},
		{"STARTED->COMPLETED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED, serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED},
		{"DRAFT->CANCELLED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT, serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED},
		{"ACTIVE->CANCELLED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE, serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED},
		{"STARTED->CANCELLED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED, serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED},
		{"DRAFT->DRAFT", serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT, serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT},
		{"ACTIVE->ACTIVE", serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE, serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE},
	}

	for _, tt := range validTransitions {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateStatusTransition(tt.from, tt.to)
			assert.NoError(t, err)
		})
	}
}

func TestValidateStatusTransition_Invalid(t *testing.T) {
	s := newTestTournamentServer()

	invalidTransitions := []struct {
		name string
		from serviceextension.TournamentStatus
		to   serviceextension.TournamentStatus
	}{
		{"COMPLETED->ACTIVE", serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE},
		{"COMPLETED->STARTED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED},
		{"COMPLETED->DRAFT", serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT},
		{"CANCELLED->ACTIVE", serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED, serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE},
		{"CANCELLED->STARTED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED, serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED},
		{"DRAFT->STARTED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT, serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED},
		{"DRAFT->COMPLETED", serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT, serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED},
	}

	for _, tt := range invalidTransitions {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateStatusTransition(tt.from, tt.to)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid tournament status transition")
		})
	}
}

// --- Helper status tests ---

func TestIsTerminalStatus(t *testing.T) {
	s := newTestTournamentServer()

	assert.True(t, s.IsTerminalStatus(serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED))
	assert.True(t, s.IsTerminalStatus(serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED))
	assert.False(t, s.IsTerminalStatus(serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT))
	assert.False(t, s.IsTerminalStatus(serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE))
	assert.False(t, s.IsTerminalStatus(serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED))
}

func TestCanBeCancelled(t *testing.T) {
	s := newTestTournamentServer()

	assert.True(t, s.CanBeCancelled(serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT))
	assert.True(t, s.CanBeCancelled(serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE))
	assert.True(t, s.CanBeCancelled(serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED))
	assert.False(t, s.CanBeCancelled(serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED))
}

func TestGetStatusName(t *testing.T) {
	s := newTestTournamentServer()

	assert.Equal(t, "DRAFT", s.GetStatusName(serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT))
	assert.Equal(t, "ACTIVE", s.GetStatusName(serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE))
	assert.Equal(t, "STARTED", s.GetStatusName(serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED))
	assert.Equal(t, "COMPLETED", s.GetStatusName(serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED))
	assert.Equal(t, "CANCELLED", s.GetStatusName(serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED))
	assert.Equal(t, "UNSPECIFIED", s.GetStatusName(serviceextension.TournamentStatus_TOURNAMENT_STATUS_UNSPECIFIED))
}

// --- CompleteTournament tests ---

func TestCompleteTournament_StatusTransition(t *testing.T) {
	mockTournamentStorage := &MockTournamentStorage{}

	s := &TournamentServiceServer{
		tournamentStorage: mockTournamentStorage,
		authInterceptor:   newTestAuthInterceptor(),
		logger:            slog.Default(),
	}

	tournament := &serviceextension.Tournament{
		TournamentId: "t1",
		Name:         "Test Tournament",
		Status:       serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED,
		CreatedAt:    timestamppb.Now(),
	}

	mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "t1").Return(tournament, nil)
	mockTournamentStorage.On("UpdateTournament", mock.Anything, "ns1", "t1", mock.AnythingOfType("*pb.Tournament")).
		Return(&serviceextension.Tournament{
			TournamentId: "t1",
			Name:         "Test Tournament",
			Status:       serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED,
		}, nil)

	result, err := s.CompleteTournament(context.Background(), "ns1", "t1", "winner-user")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, result.Status)
	mockTournamentStorage.AssertExpectations(t)
}

func TestCompleteTournament_InvalidTransition(t *testing.T) {
	mockTournamentStorage := &MockTournamentStorage{}

	s := &TournamentServiceServer{
		tournamentStorage: mockTournamentStorage,
		authInterceptor:   newTestAuthInterceptor(),
		logger:            slog.Default(),
	}

	// Tournament in DRAFT state - cannot transition to COMPLETED
	tournament := &serviceextension.Tournament{
		TournamentId: "t1",
		Status:       serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT,
	}

	mockTournamentStorage.On("GetTournament", mock.Anything, "ns1", "t1").Return(tournament, nil)

	_, err := s.CompleteTournament(context.Background(), "ns1", "t1", "winner-user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tournament status transition")
}

func TestCompleteTournament_ValidationErrors(t *testing.T) {
	s := &TournamentServiceServer{
		authInterceptor: newTestAuthInterceptor(),
		logger:          slog.Default(),
	}

	t.Run("empty namespace", func(t *testing.T) {
		_, err := s.CompleteTournament(context.Background(), "", "t1", "winner")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "namespace is required")
	})

	t.Run("empty tournament_id", func(t *testing.T) {
		_, err := s.CompleteTournament(context.Background(), "ns1", "", "winner")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tournament_id is required")
	})
}

// --- CreateTournament validation tests ---

func TestCreateTournament_Validation(t *testing.T) {
	mockTournamentStorage := &MockTournamentStorage{}

	s := &TournamentServiceServer{
		tournamentStorage: mockTournamentStorage,
		authInterceptor:   newTestAuthInterceptor(),
		logger:            slog.Default(),
	}

	t.Run("empty name", func(t *testing.T) {
		req := &serviceextension.CreateTournamentRequest{
			Name:            "",
			MaxParticipants: 8,
			Namespace:       "ns1",
		}
		_, err := s.CreateTournament(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tournament name is required")
	})

	t.Run("zero max_participants", func(t *testing.T) {
		req := &serviceextension.CreateTournamentRequest{
			Name:            "Test",
			MaxParticipants: 0,
			Namespace:       "ns1",
		}
		_, err := s.CreateTournament(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_participants must be greater than 0")
	})

	t.Run("empty namespace", func(t *testing.T) {
		req := &serviceextension.CreateTournamentRequest{
			Name:            "Test",
			MaxParticipants: 8,
			Namespace:       "",
		}
		_, err := s.CreateTournament(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "namespace is required")
	})
}

// --- Test helpers ---

func makeParticipants(count int) []TournamentParticipant {
	participants := make([]TournamentParticipant, count)
	for i := 0; i < count; i++ {
		id := string(rune('A' + i))
		participants[i] = TournamentParticipant{
			UserId:      "user" + id,
			Username:    "player" + id,
			DisplayName: "Player " + id,
		}
	}
	return participants
}

func countAllMatches(bracket *BracketData) int {
	total := 0
	for _, round := range bracket.Rounds {
		total += len(round)
	}
	return total
}
