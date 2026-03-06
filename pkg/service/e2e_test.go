// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"

	serviceextension "extend-tournament-service/pkg/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ---------------------------------------------------------------------------
// In-memory storage implementations
// ---------------------------------------------------------------------------

// inMemMatchStorage is a thread-safe in-memory MatchStorage for tests.
// It maintains real state so that calls to advanceWinner, HandleByeAdvancement,
// and CheckTournamentCompletion operate correctly without stub wiring.
type inMemMatchStorage struct {
	mu      sync.Mutex
	matches map[string]*serviceextension.Match // key: matchID
}

func newInMemMatchStorage() *inMemMatchStorage {
	return &inMemMatchStorage{matches: make(map[string]*serviceextension.Match)}
}

func (s *inMemMatchStorage) get(matchID string) (*serviceextension.Match, bool) {
	m, ok := s.matches[matchID]
	return m, ok
}

func (s *inMemMatchStorage) GetMatch(_ context.Context, _, _, matchID string) (*serviceextension.Match, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.get(matchID)
	if !ok {
		return nil, grpcStatus.Errorf(codes.NotFound, "match not found: %s", matchID)
	}
	return m, nil
}

func (s *inMemMatchStorage) GetTournamentMatches(_ context.Context, _, _ string) ([]*serviceextension.Match, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*serviceextension.Match, 0, len(s.matches))
	for _, m := range s.matches {
		result = append(result, m)
	}
	return result, nil
}

func (s *inMemMatchStorage) CreateMatches(_ context.Context, _, _ string, matches []*serviceextension.Match) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range matches {
		s.matches[m.MatchId] = m
	}
	return nil
}

func (s *inMemMatchStorage) UpdateMatch(_ context.Context, _ string, match *serviceextension.Match) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.get(match.MatchId); !ok {
		return grpcStatus.Errorf(codes.NotFound, "match not found: %s", match.MatchId)
	}
	s.matches[match.MatchId] = match
	return nil
}

func (s *inMemMatchStorage) SubmitMatchResult(_ context.Context, _, _, matchID, winnerUserID string) (*serviceextension.Match, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.get(matchID)
	if !ok {
		return nil, grpcStatus.Errorf(codes.NotFound, "match not found: %s", matchID)
	}
	m.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
	m.Winner = winnerUserID
	m.CompletedAt = timestamppb.Now()
	return m, nil
}

func (s *inMemMatchStorage) GetMatchesByRound(_ context.Context, _, _ string, round int32) ([]*serviceextension.Match, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []*serviceextension.Match
	for _, m := range s.matches {
		if m.Round == round {
			result = append(result, m)
		}
	}
	return result, nil
}

// inMemTournamentStorage is an in-memory TournamentStorage for tests.
type inMemTournamentStorage struct {
	mu          sync.Mutex
	tournaments map[string]*serviceextension.Tournament // key: tournamentID
}

func newInMemTournamentStorage(initial *serviceextension.Tournament) *inMemTournamentStorage {
	s := &inMemTournamentStorage{tournaments: make(map[string]*serviceextension.Tournament)}
	if initial != nil {
		s.tournaments[initial.TournamentId] = initial
	}
	return s
}

func (s *inMemTournamentStorage) GetTournament(_ context.Context, _, tournamentID string) (*serviceextension.Tournament, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tournaments[tournamentID]
	if !ok {
		return nil, grpcStatus.Errorf(codes.NotFound, "tournament not found: %s", tournamentID)
	}
	return t, nil
}

func (s *inMemTournamentStorage) UpdateTournament(_ context.Context, _, tournamentID string, tournament *serviceextension.Tournament) (*serviceextension.Tournament, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tournaments[tournamentID] = tournament
	return tournament, nil
}

func (s *inMemTournamentStorage) CreateTournament(_ context.Context, _ string, t *serviceextension.Tournament) (*serviceextension.Tournament, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tournaments[t.TournamentId] = t
	return t, nil
}

func (s *inMemTournamentStorage) ListTournaments(context.Context, string, int32, int32, serviceextension.TournamentStatus) ([]*serviceextension.Tournament, int32, error) {
	return nil, 0, nil
}

func (s *inMemTournamentStorage) GetTournamentForRegistration(ctx context.Context, ns, id string) (*serviceextension.Tournament, error) {
	return s.GetTournament(ctx, ns, id)
}

func (s *inMemTournamentStorage) UpdateParticipantCount(context.Context, string, string, int32) error {
	return nil
}

func (s *inMemTournamentStorage) CheckTournamentCapacity(context.Context, string, string) (bool, error) {
	return true, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildMatches creates the Match slice for an N-player single-elimination bracket,
// applying the same ID/relationship scheme used by StartTournament.
func buildMatches(tournamentID string, participants []TournamentParticipant) []*serviceextension.Match {
	server := &TournamentServiceServer{logger: slog.Default()}
	bracket, err := server.GenerateBrackets(participants)
	if err != nil {
		panic(fmt.Sprintf("buildMatches: GenerateBrackets failed: %v", err))
	}

	var matches []*serviceextension.Match
	for rIdx, round := range bracket.Rounds {
		for mIdx, b := range round {
			m := &serviceextension.Match{
				MatchId:      fmt.Sprintf("match-r%d-m%d", rIdx+1, mIdx+1),
				TournamentId: tournamentID,
				Round:        int32(rIdx + 1),
				Position:     int32(mIdx),
				Status:       serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED,
				StartedAt:    timestamppb.Now(),
			}
			// Link forward (winner advancement)
			if rIdx < len(bracket.Rounds)-1 {
				m.NextMatchId = fmt.Sprintf("match-r%d-m%d", rIdx+2, mIdx/2+1)
			}
			// Link backward (source matches)
			if rIdx > 0 {
				m.SourceMatch_1Id = fmt.Sprintf("match-r%d-m%d", rIdx, mIdx*2+1)
				m.SourceMatch_2Id = fmt.Sprintf("match-r%d-m%d", rIdx, mIdx*2+2)
			}
			if b.Participant1 != nil {
				m.Participant1 = &serviceextension.TournamentParticipant{
					UserId:      b.Participant1.UserId,
					Username:    b.Participant1.Username,
					DisplayName: b.Participant1.DisplayName,
				}
			}
			if b.Participant2 != nil {
				m.Participant2 = &serviceextension.TournamentParticipant{
					UserId:      b.Participant2.UserId,
					Username:    b.Participant2.Username,
					DisplayName: b.Participant2.DisplayName,
				}
			}
			matches = append(matches, m)
		}
	}
	return matches
}

func submitResult(t *testing.T, svc *MatchService, ns, tournamentID, matchID, winnerID string) {
	t.Helper()
	_, err := svc.SubmitMatchResult(context.Background(), &serviceextension.SubmitMatchResultRequest{
		Namespace:    ns,
		TournamentId: tournamentID,
		MatchId:      matchID,
		WinnerUserId: winnerID,
	})
	require.NoError(t, err, "SubmitMatchResult(%s, winner=%s)", matchID, winnerID)
}

// ---------------------------------------------------------------------------
// End-to-end tests
// ---------------------------------------------------------------------------

// TestE2E_4Players runs a complete 4-player single-elimination tournament from
// bracket generation through final match, verifying winner advancement and
// tournament completion.
//
//	 Round 1          Round 2
//	p1 ─┐
//	     ├─ p1 ─┐
//	p2 ─┘       │
//	             ├─ winner (p1)
//	p3 ─┐       │
//	     ├─ p3 ─┘
//	p4 ─┘
func TestE2E_4Players(t *testing.T) {
	const (
		ns           = "test-ns"
		tournamentID = "tournament-1"
	)

	participants := makeParticipants(4) // p1..p4

	tournament := &serviceextension.Tournament{
		TournamentId: tournamentID,
		Status:       serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED,
	}
	tournamentStore := newInMemTournamentStorage(tournament)
	matchStore := newInMemMatchStorage()

	svc := NewMatchService(matchStore, tournamentStore, nil, slog.Default())

	// Seed matches
	matches := buildMatches(tournamentID, participants)
	require.NoError(t, matchStore.CreateMatches(context.Background(), ns, tournamentID, matches))

	// --- Round 1 ---
	// p1 beats p2
	submitResult(t, svc, ns, tournamentID, "match-r1-m1", participants[0].UserId)
	// p3 beats p4
	submitResult(t, svc, ns, tournamentID, "match-r1-m2", participants[2].UserId)

	// Verify final match now has both participants populated
	final, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r2-m1")
	require.NoError(t, err)
	assert.Equal(t, participants[0].UserId, final.Participant1.UserId, "p1 should advance to final slot 1")
	assert.Equal(t, participants[2].UserId, final.Participant2.UserId, "p3 should advance to final slot 2")
	assert.Equal(t, serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED, final.Status, "final should still be scheduled")

	// --- Round 2 (Final) ---
	// p1 beats p3
	submitResult(t, svc, ns, tournamentID, "match-r2-m1", participants[0].UserId)

	// Verify tournament is completed
	completedTournament, err := tournamentStore.GetTournament(context.Background(), ns, tournamentID)
	require.NoError(t, err)
	assert.Equal(t, serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, completedTournament.Status)

	// Verify final match winner
	final, err = matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r2-m1")
	require.NoError(t, err)
	assert.Equal(t, participants[0].UserId, final.Winner, "p1 should be the champion")
	assert.Equal(t, serviceextension.MatchStatus_MATCH_STATUS_COMPLETED, final.Status)
}

// TestE2E_8Players runs a complete 8-player tournament verifying winner
// advancement across 3 rounds and correct tournament completion.
func TestE2E_8Players(t *testing.T) {
	const (
		ns           = "test-ns"
		tournamentID = "tournament-2"
	)

	participants := makeParticipants(8) // p1..p8

	tournamentStore := newInMemTournamentStorage(&serviceextension.Tournament{
		TournamentId: tournamentID,
		Status:       serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED,
	})
	matchStore := newInMemMatchStorage()
	svc := NewMatchService(matchStore, tournamentStore, nil, slog.Default())

	require.NoError(t, matchStore.CreateMatches(context.Background(), ns, tournamentID,
		buildMatches(tournamentID, participants)))

	// Round 1: odd-indexed player always loses (p1, p3, p5, p7 advance)
	submitResult(t, svc, ns, tournamentID, "match-r1-m1", participants[0].UserId) // p1 beats p2
	submitResult(t, svc, ns, tournamentID, "match-r1-m2", participants[2].UserId) // p3 beats p4
	submitResult(t, svc, ns, tournamentID, "match-r1-m3", participants[4].UserId) // p5 beats p6
	submitResult(t, svc, ns, tournamentID, "match-r1-m4", participants[6].UserId) // p7 beats p8

	// Verify semi-finals are populated
	sf1, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r2-m1")
	require.NoError(t, err)
	assert.NotNil(t, sf1.Participant1)
	assert.NotNil(t, sf1.Participant2)

	sf2, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r2-m2")
	require.NoError(t, err)
	assert.NotNil(t, sf2.Participant1)
	assert.NotNil(t, sf2.Participant2)

	// Round 2: p1 and p5 advance
	submitResult(t, svc, ns, tournamentID, "match-r2-m1", participants[0].UserId) // p1 beats p3
	submitResult(t, svc, ns, tournamentID, "match-r2-m2", participants[4].UserId) // p5 beats p7

	// Verify final is populated
	final, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r3-m1")
	require.NoError(t, err)
	assert.NotNil(t, final.Participant1)
	assert.NotNil(t, final.Participant2)
	assert.Equal(t, serviceextension.MatchStatus_MATCH_STATUS_SCHEDULED, final.Status)

	// Round 3 (Final): p1 wins
	submitResult(t, svc, ns, tournamentID, "match-r3-m1", participants[0].UserId)

	completedTournament, err := tournamentStore.GetTournament(context.Background(), ns, tournamentID)
	require.NoError(t, err)
	assert.Equal(t, serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, completedTournament.Status)

	final, err = matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r3-m1")
	require.NoError(t, err)
	assert.Equal(t, participants[0].UserId, final.Winner)
}

// TestE2E_7Players runs a 7-player tournament (one bye in round 1), verifying
// that the bye participant is auto-advanced and the bracket completes correctly.
//
//	 Round 1     Round 2   Round 3
//	p1 ─┐
//	     ├─ p1 ─┐
//	p2 ─┘       │
//	             ├─ p1 ─┐
//	p3 ─┐       │       │
//	     ├─ p3 ─┘       ├─ p1 (champion)
//	p4 ─┘               │
//	                     │
//	p5 ─┐               │
//	     ├─ p5 ─┐       │
//	p6 ─┘       ├─ p5 ──┘
//	            │
//	p7 ─(bye)──┘
func TestE2E_7Players(t *testing.T) {
	const (
		ns           = "test-ns"
		tournamentID = "tournament-3"
	)

	participants := makeParticipants(7) // p1..p7; 1 bye in round 1

	tournamentStore := newInMemTournamentStorage(&serviceextension.Tournament{
		TournamentId: tournamentID,
		Status:       serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED,
	})
	matchStore := newInMemMatchStorage()
	svc := NewMatchService(matchStore, tournamentStore, nil, slog.Default())

	matches := buildMatches(tournamentID, participants)
	require.NoError(t, matchStore.CreateMatches(context.Background(), ns, tournamentID, matches))

	// Advance round-1 bye matches immediately after match creation, mirroring what a
	// properly integrated StartTournament would trigger on tournament start.
	require.NoError(t, svc.HandleByeAdvancement(context.Background(), ns, tournamentID, 1))

	// Identify the one real (non-bye) round-1 match and play it.
	var realMatchID string
	for _, m := range matches {
		if m.Participant1 != nil && m.Participant2 != nil {
			realMatchID = m.MatchId
			break
		}
	}
	require.NotEmpty(t, realMatchID, "expected at least one real (non-bye) round-1 match")

	// Play all three real round-1 matches; Participant1 wins each.
	for _, m := range matches {
		if m.Participant1 != nil && m.Participant2 != nil {
			submitResult(t, svc, ns, tournamentID, m.MatchId, m.Participant1.UserId)
		}
	}

	// After all round-1 matches, round 2 should be fully populated.
	r2m1, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r2-m1")
	require.NoError(t, err)
	assert.NotNil(t, r2m1.Participant1, "r2-m1 slot 1 should be filled")
	assert.NotNil(t, r2m1.Participant2, "r2-m1 slot 2 should be filled")

	r2m2, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r2-m2")
	require.NoError(t, err)
	assert.NotNil(t, r2m2.Participant1, "r2-m2 slot 1 should be filled (by p7 bye)")
	assert.NotNil(t, r2m2.Participant2, "r2-m2 slot 2 should be filled (by winner of r1-m3)")

	// Round 2
	submitResult(t, svc, ns, tournamentID, "match-r2-m1", r2m1.Participant1.UserId)
	submitResult(t, svc, ns, tournamentID, "match-r2-m2", r2m2.Participant1.UserId)

	// Final
	final, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r3-m1")
	require.NoError(t, err)
	assert.NotNil(t, final.Participant1)
	assert.NotNil(t, final.Participant2)
	submitResult(t, svc, ns, tournamentID, "match-r3-m1", final.Participant1.UserId)

	// Tournament must be completed with the correct champion
	completedTournament, err := tournamentStore.GetTournament(context.Background(), ns, tournamentID)
	require.NoError(t, err)
	assert.Equal(t, serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED, completedTournament.Status)

	finalResult, err := matchStore.GetMatch(context.Background(), ns, tournamentID, "match-r3-m1")
	require.NoError(t, err)
	assert.NotEmpty(t, finalResult.Winner)
	assert.Equal(t, serviceextension.MatchStatus_MATCH_STATUS_COMPLETED, finalResult.Status)
}

// TestE2E_OutOfOrderSubmission verifies that submitting a result for a later
// round while an earlier round is still incomplete is rejected.
func TestE2E_OutOfOrderSubmission(t *testing.T) {
	const (
		ns           = "test-ns"
		tournamentID = "tournament-4"
	)

	participants := makeParticipants(4)

	matchStore := newInMemMatchStorage()
	tournamentStore := newInMemTournamentStorage(&serviceextension.Tournament{
		TournamentId: tournamentID,
		Status:       serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED,
	})
	svc := NewMatchService(matchStore, tournamentStore, nil, slog.Default())

	require.NoError(t, matchStore.CreateMatches(context.Background(), ns, tournamentID,
		buildMatches(tournamentID, participants)))

	// Attempt to submit the final before any round-1 match is done
	_, err := svc.SubmitMatchResult(context.Background(), &serviceextension.SubmitMatchResultRequest{
		Namespace:    ns,
		TournamentId: tournamentID,
		MatchId:      "match-r2-m1",
		WinnerUserId: participants[0].UserId,
	})
	require.Error(t, err)
	st, _ := grpcStatus.FromError(err)
	assert.Equal(t, codes.FailedPrecondition, st.Code(), "expected FailedPrecondition when round 1 is incomplete")
}
