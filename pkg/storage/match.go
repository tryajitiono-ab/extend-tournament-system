// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	serviceextension "extend-tournament-service/pkg/pb"
)

// MatchStorage defines the interface for match data operations
type MatchStorage interface {
	GetMatch(ctx context.Context, namespace, tournamentID, matchID string) (*serviceextension.Match, error)
	GetTournamentMatches(ctx context.Context, namespace, tournamentID string) ([]*serviceextension.Match, error)
	CreateMatches(ctx context.Context, namespace, tournamentID string, matches []*serviceextension.Match) error
	UpdateMatch(ctx context.Context, namespace string, match *serviceextension.Match) error
	SubmitMatchResult(ctx context.Context, namespace, tournamentID, matchID, winnerUserID string) (*serviceextension.Match, error)
	GetMatchesByRound(ctx context.Context, namespace, tournamentID string, round int32) ([]*serviceextension.Match, error)
}

// MongoMatchStorage implements MatchStorage using MongoDB
type MongoMatchStorage struct {
	client               *mongo.Client
	database             string
	matchCollection      string
	tournamentCollection string
	logger               *slog.Logger
}

// NewMongoMatchStorage creates a new match storage instance
func NewMongoMatchStorage(client *mongo.Client, database string, logger *slog.Logger) *MongoMatchStorage {
	return &MongoMatchStorage{
		client:               client,
		database:             database,
		matchCollection:      "matches",
		tournamentCollection: "tournaments",
		logger:               logger,
	}
}

// matchDocument represents the MongoDB document structure
type matchDocument struct {
	MatchID      string                                  `bson:"match_id"`
	TournamentID string                                  `bson:"tournament_id"`
	Round        int32                                   `bson:"round"`
	Position     int32                                   `bson:"position"`
	Participant1 *serviceextension.TournamentParticipant `bson:"participant1,omitempty"`
	Participant2 *serviceextension.TournamentParticipant `bson:"participant2,omitempty"`
	Winner       string                                  `bson:"winner,omitempty"`
	Status       serviceextension.MatchStatus            `bson:"status"`
	StartedAt    time.Time                               `bson:"started_at"`
	CompletedAt  *time.Time                              `bson:"completed_at,omitempty"`
	CreatedAt    time.Time                               `bson:"created_at"`
	UpdatedAt    time.Time                               `bson:"updated_at"`
	Namespace    string                                  `bson:"namespace"`
}

// GetMatch retrieves a specific match by ID
func (m *MongoMatchStorage) GetMatch(ctx context.Context, namespace, tournamentID, matchID string) (*serviceextension.Match, error) {
	collection := m.client.Database(m.database).Collection(m.matchCollection)

	filter := bson.M{
		"match_id":      matchID,
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}

	var doc matchDocument
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, grpcStatus.Errorf(codes.NotFound, "match not found: %s", matchID)
		}
		m.logger.Error("failed to get match", "error", err, "match_id", matchID, "tournament_id", tournamentID)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to get match: %v", err)
	}

	return m.documentToProto(&doc), nil
}

// GetTournamentMatches retrieves all matches for a tournament
func (m *MongoMatchStorage) GetTournamentMatches(ctx context.Context, namespace, tournamentID string) ([]*serviceextension.Match, error) {
	collection := m.client.Database(m.database).Collection(m.matchCollection)

	filter := bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}

	// Sort by round then position for bracket organization
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "round", Value: 1}, {Key: "position", Value: 1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		m.logger.Error("failed to query tournament matches", "error", err, "tournament_id", tournamentID)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to query tournament matches: %v", err)
	}
	defer cursor.Close(ctx)

	var docs []matchDocument
	if err := cursor.All(ctx, &docs); err != nil {
		m.logger.Error("failed to decode tournament matches", "error", err, "tournament_id", tournamentID)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to decode tournament matches: %v", err)
	}

	// Convert to protobuf
	matches := make([]*serviceextension.Match, len(docs))
	for i, doc := range docs {
		matches[i] = m.documentToProto(&doc)
	}

	return matches, nil
}

// CreateMatches creates multiple matches for a tournament (bulk insert)
func (m *MongoMatchStorage) CreateMatches(ctx context.Context, namespace, tournamentID string, matches []*serviceextension.Match) error {
	if len(matches) == 0 {
		return nil
	}

	collection := m.client.Database(m.database).Collection(m.matchCollection)

	now := time.Now()
	documents := make([]interface{}, len(matches))

	for i, match := range matches {
		// Generate UUID if not provided
		if match.MatchId == "" {
			match.MatchId = uuid.New().String()
		}

		doc := &matchDocument{
			MatchID:      match.MatchId,
			TournamentID: tournamentID,
			Round:        match.Round,
			Position:     match.Position,
			Participant1: match.Participant1,
			Participant2: match.Participant2,
			Winner:       match.Winner,
			Status:       match.Status,
			StartedAt:    match.StartedAt.AsTime(),
			Namespace:    namespace,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if !match.CompletedAt.AsTime().IsZero() {
			doc.CompletedAt = &[]time.Time{match.CompletedAt.AsTime()}[0]
		}

		documents[i] = doc
	}

	// Bulk insert for performance
	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		m.logger.Error("failed to create matches", "error", err, "tournament_id", tournamentID, "count", len(matches))
		return grpcStatus.Errorf(codes.Internal, "failed to create matches: %v", err)
	}

	m.logger.Info("matches created successfully", "tournament_id", tournamentID, "count", len(matches))
	return nil
}

// UpdateMatch updates an existing match
func (m *MongoMatchStorage) UpdateMatch(ctx context.Context, namespace string, match *serviceextension.Match) error {
	collection := m.client.Database(m.database).Collection(m.matchCollection)

	// Build update document with current timestamp
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"round":        match.Round,
			"position":     match.Position,
			"participant1": match.Participant1,
			"participant2": match.Participant2,
			"winner":       match.Winner,
			"status":       match.Status,
			"started_at":   match.StartedAt.AsTime(),
			"updated_at":   now,
		},
	}

	// Handle completed_at field
	if !match.CompletedAt.AsTime().IsZero() {
		update["$set"].(bson.M)["completed_at"] = match.CompletedAt.AsTime()
	} else {
		update["$unset"] = bson.M{"completed_at": ""}
	}

	filter := bson.M{
		"match_id":      match.MatchId,
		"tournament_id": match.TournamentId,
		"namespace":     namespace,
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		m.logger.Error("failed to update match", "error", err, "match_id", match.MatchId)
		return grpcStatus.Errorf(codes.Internal, "failed to update match: %v", err)
	}

	if result.MatchedCount == 0 {
		return grpcStatus.Errorf(codes.NotFound, "match not found: %s", match.MatchId)
	}

	m.logger.Info("match updated successfully", "match_id", match.MatchId, "status", match.Status.String())
	return nil
}

// GetMatchesByRound retrieves matches for a specific round
func (m *MongoMatchStorage) GetMatchesByRound(ctx context.Context, namespace, tournamentID string, round int32) ([]*serviceextension.Match, error) {
	collection := m.client.Database(m.database).Collection(m.matchCollection)

	filter := bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
		"round":         round,
	}

	// Sort by position for bracket rendering
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "position", Value: 1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		m.logger.Error("failed to query matches by round", "error", err, "tournament_id", tournamentID, "round", round)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to query matches by round: %v", err)
	}
	defer cursor.Close(ctx)

	var docs []matchDocument
	if err := cursor.All(ctx, &docs); err != nil {
		m.logger.Error("failed to decode matches by round", "error", err, "tournament_id", tournamentID, "round", round)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to decode matches by round: %v", err)
	}

	// Convert to protobuf
	matches := make([]*serviceextension.Match, len(docs))
	for i, doc := range docs {
		matches[i] = m.documentToProto(&doc)
	}

	return matches, nil
}

// SubmitMatchResult submits a match result with transaction safety
func (m *MongoMatchStorage) SubmitMatchResult(ctx context.Context, namespace, tournamentID, matchID, winnerUserID string) (*serviceextension.Match, error) {
	session, err := m.client.StartSession()
	if err != nil {
		m.logger.Error("failed to start session", "error", err)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Step 1: Get match for update and validate
		match, err := m.getMatchForUpdate(sessCtx, namespace, tournamentID, matchID)
		if err != nil {
			return nil, fmt.Errorf("failed to get match: %w", err)
		}

		// Step 2: Validate match status allows result submission
		if match.Status == serviceextension.MatchStatus_MATCH_STATUS_COMPLETED {
			return nil, errors.New("match already completed")
		}
		if match.Status == serviceextension.MatchStatus_MATCH_STATUS_CANCELLED {
			return nil, errors.New("match is cancelled")
		}

		// Step 3: Validate winner is one of the participants
		if err := m.validateMatchWinner(match, winnerUserID); err != nil {
			return nil, err
		}

		// Step 4: Update match with result
		now := time.Now()
		match.Winner = winnerUserID
		match.Status = serviceextension.MatchStatus_MATCH_STATUS_COMPLETED
		match.CompletedAt = timestamppb.New(now)

		// Update in database
		if err := m.UpdateMatch(sessCtx, namespace, match); err != nil {
			return nil, fmt.Errorf("failed to update match: %w", err)
		}

		// Step 5: Log the result submission
		m.logger.Info("match result submitted",
			"match_id", matchID,
			"tournament_id", tournamentID,
			"winner_user_id", winnerUserID,
			"namespace", namespace,
			"completed_at", now)

		return match, nil
	})

	if err != nil {
		m.logger.Error("transaction failed", "error", err, "match_id", matchID, "winner_user_id", winnerUserID)

		// Return appropriate gRPC status codes
		if errors.Is(err, errors.New("match already completed")) {
			return nil, grpcStatus.Errorf(codes.AlreadyExists, "match already completed: %s", matchID)
		}
		if errors.Is(err, errors.New("match is cancelled")) {
			return nil, grpcStatus.Errorf(codes.FailedPrecondition, "match is cancelled: %s", matchID)
		}

		return nil, grpcStatus.Errorf(codes.Internal, "failed to submit match result: %v", err)
	}

	return result.(*serviceextension.Match), nil
}

// documentToProto converts MongoDB document to protobuf message
func (m *MongoMatchStorage) documentToProto(doc *matchDocument) *serviceextension.Match {
	match := &serviceextension.Match{
		MatchId:      doc.MatchID,
		TournamentId: doc.TournamentID,
		Round:        doc.Round,
		Position:     doc.Position,
		Participant1: doc.Participant1,
		Participant2: doc.Participant2,
		Winner:       doc.Winner,
		Status:       doc.Status,
		StartedAt:    timestamppb.New(doc.StartedAt),
	}

	if doc.CompletedAt != nil {
		match.CompletedAt = timestamppb.New(*doc.CompletedAt)
	}

	return match
}

// getMatchForUpdate retrieves a match within a transaction context for update
func (m *MongoMatchStorage) getMatchForUpdate(ctx context.Context, namespace, tournamentID, matchID string) (*serviceextension.Match, error) {
	collection := m.client.Database(m.database).Collection(m.matchCollection)

	var doc matchDocument
	err := collection.FindOne(ctx, bson.M{
		"match_id":      matchID,
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, grpcStatus.Errorf(codes.NotFound, "match not found: %s", matchID)
		}
		return nil, err
	}

	return m.documentToProto(&doc), nil
}

// validateMatchWinner validates that the winner is one of the participants
func (m *MongoMatchStorage) validateMatchWinner(match *serviceextension.Match, winnerUserID string) error {
	if match.Participant1 != nil && match.Participant1.UserId == winnerUserID {
		return nil
	}
	if match.Participant2 != nil && match.Participant2.UserId == winnerUserID {
		return nil
	}

	return grpcStatus.Errorf(codes.InvalidArgument, "winner %s is not a participant in match %s", winnerUserID, match.MatchId)
}

// EnsureIndexes creates database indexes for the matches collection
func (m *MongoMatchStorage) EnsureIndexes(ctx context.Context) error {
	collection := m.client.Database(m.database).Collection(m.matchCollection)

	// Compound index for tournament and round queries
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "tournament_id", Value: 1},
			{Key: "namespace", Value: 1},
			{Key: "round", Value: 1},
			{Key: "position", Value: 1},
		},
		Options: options.Index().SetName("tournament_round_position_idx"),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create tournament_round_position_idx: %w", err)
	}

	// Index for match lookups
	matchIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "match_id", Value: 1},
			{Key: "namespace", Value: 1},
		},
		Options: options.Index().SetName("match_namespace_idx").SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(ctx, matchIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create match_namespace_idx: %w", err)
	}

	m.logger.Info("match storage indexes created successfully")
	return nil
}
