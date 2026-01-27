// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package storage

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	serviceextension "extend-custom-guild-service/pkg/pb"
)

// TournamentStorage defines the interface for tournament data persistence
type TournamentStorage interface {
	CreateTournament(ctx context.Context, namespace string, tournament *serviceextension.Tournament) (*serviceextension.Tournament, error)
	GetTournament(ctx context.Context, namespace string, tournamentID string) (*serviceextension.Tournament, error)
	ListTournaments(ctx context.Context, namespace string, limit, offset int32, status serviceextension.TournamentStatus) ([]*serviceextension.Tournament, int32, error)
	UpdateTournament(ctx context.Context, namespace string, tournamentID string, tournament *serviceextension.Tournament) (*serviceextension.Tournament, error)
}

// MongoTournamentStorage implements TournamentStorage using MongoDB
type MongoTournamentStorage struct {
	client     *mongo.Client
	database   string
	collection string
	logger     *slog.Logger
}

// NewMongoTournamentStorage creates a new MongoDB tournament storage instance
func NewMongoTournamentStorage(client *mongo.Client, database string, logger *slog.Logger) *MongoTournamentStorage {
	return &MongoTournamentStorage{
		client:     client,
		database:   database,
		collection: "tournaments",
		logger:     logger,
	}
}

// tournamentDocument represents the MongoDB document structure
type tournamentDocument struct {
	TournamentID        string                            `bson:"tournament_id"`
	Namespace           string                            `bson:"namespace"`
	Name                string                            `bson:"name"`
	Description         string                            `bson:"description"`
	MaxParticipants     int32                             `bson:"max_participants"`
	CurrentParticipants int32                             `bson:"current_participants"`
	Status              serviceextension.TournamentStatus `bson:"status"`
	CreatedAt           time.Time                         `bson:"created_at"`
	UpdatedAt           time.Time                         `bson:"updated_at"`
	StartTime           time.Time                         `bson:"start_time"`
	EndTime             time.Time                         `bson:"end_time"`
}

// CreateTournament creates a new tournament in MongoDB
func (m *MongoTournamentStorage) CreateTournament(ctx context.Context, namespace string, tournament *serviceextension.Tournament) (*serviceextension.Tournament, error) {
	// Generate UUID if not provided
	if tournament.TournamentId == "" {
		tournament.TournamentId = uuid.New().String()
	}

	// Set initial status and timestamps
	now := time.Now()
	tournament.Status = serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT
	tournament.CreatedAt = timestamppb.New(now)
	tournament.UpdatedAt = timestamppb.New(now)
	tournament.CurrentParticipants = 0

	// Convert to MongoDB document
	doc := &tournamentDocument{
		TournamentID:        tournament.TournamentId,
		Namespace:           namespace,
		Name:                tournament.Name,
		Description:         tournament.Description,
		MaxParticipants:     tournament.MaxParticipants,
		CurrentParticipants: tournament.CurrentParticipants,
		Status:              tournament.Status,
		CreatedAt:           now,
		UpdatedAt:           now,
		StartTime:           tournament.StartTime.AsTime(),
		EndTime:             tournament.EndTime.AsTime(),
	}

	// Get collection
	collection := m.client.Database(m.database).Collection(m.collection)

	// Insert document
	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		m.logger.Error("failed to create tournament", "error", err, "tournament_id", tournament.TournamentId, "namespace", namespace)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to create tournament: %v", err)
	}

	m.logger.Info("tournament created", "tournament_id", tournament.TournamentId, "namespace", namespace, "name", tournament.Name)
	return tournament, nil
}

// GetTournament retrieves a tournament by ID
func (m *MongoTournamentStorage) GetTournament(ctx context.Context, namespace string, tournamentID string) (*serviceextension.Tournament, error) {
	collection := m.client.Database(m.database).Collection(m.collection)

	// Find tournament by ID and namespace
	filter := bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}

	var doc tournamentDocument
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, grpcStatus.Errorf(codes.NotFound, "tournament not found: %s", tournamentID)
		}
		m.logger.Error("failed to get tournament", "error", err, "tournament_id", tournamentID, "namespace", namespace)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to get tournament: %v", err)
	}

	return m.documentToProto(&doc), nil
}

// ListTournaments retrieves tournaments with pagination and optional status filter
func (m *MongoTournamentStorage) ListTournaments(ctx context.Context, namespace string, limit, offset int32, status serviceextension.TournamentStatus) ([]*serviceextension.Tournament, int32, error) {
	collection := m.client.Database(m.database).Collection(m.collection)

	// Build filter
	filter := bson.M{"namespace": namespace}
	if status != serviceextension.TournamentStatus_TOURNAMENT_STATUS_UNSPECIFIED {
		filter["status"] = status
	}

	// Count total documents
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		m.logger.Error("failed to count tournaments", "error", err, "namespace", namespace)
		return nil, 0, grpcStatus.Errorf(codes.Internal, "failed to count tournaments: %v", err)
	}

	// Find documents with pagination
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by creation date descending

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		m.logger.Error("failed to list tournaments", "error", err, "namespace", namespace)
		return nil, 0, grpcStatus.Errorf(codes.Internal, "failed to list tournaments: %v", err)
	}
	defer cursor.Close(ctx)

	var docs []tournamentDocument
	if err = cursor.All(ctx, &docs); err != nil {
		m.logger.Error("failed to decode tournaments", "error", err, "namespace", namespace)
		return nil, 0, grpcStatus.Errorf(codes.Internal, "failed to decode tournaments: %v", err)
	}

	// Convert to protobuf
	tournaments := make([]*serviceextension.Tournament, len(docs))
	for i, doc := range docs {
		tournaments[i] = m.documentToProto(&doc)
	}

	return tournaments, int32(totalCount), nil
}

// UpdateTournament updates an existing tournament
func (m *MongoTournamentStorage) UpdateTournament(ctx context.Context, namespace string, tournamentID string, tournament *serviceextension.Tournament) (*serviceextension.Tournament, error) {
	collection := m.client.Database(m.database).Collection(m.collection)

	// First, get existing tournament to validate status transitions
	existing, err := m.GetTournament(ctx, namespace, tournamentID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !isValidStatusTransition(existing.Status, tournament.Status) {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "invalid status transition from %v to %v", existing.Status, tournament.Status)
	}

	// Update timestamp
	tournament.UpdatedAt = timestamppb.New(time.Now())

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"name":                 tournament.Name,
			"description":          tournament.Description,
			"max_participants":     tournament.MaxParticipants,
			"current_participants": tournament.CurrentParticipants,
			"status":               tournament.Status,
			"updated_at":           tournament.UpdatedAt.AsTime(),
			"start_time":           tournament.StartTime.AsTime(),
			"end_time":             tournament.EndTime.AsTime(),
		},
	}

	// Update tournament
	filter := bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		m.logger.Error("failed to update tournament", "error", err, "tournament_id", tournamentID, "namespace", namespace)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to update tournament: %v", err)
	}

	if result.MatchedCount == 0 {
		return nil, grpcStatus.Errorf(codes.NotFound, "tournament not found: %s", tournamentID)
	}

	m.logger.Info("tournament updated", "tournament_id", tournamentID, "namespace", namespace, "status", tournament.Status)
	return tournament, nil
}

// isValidStatusTransition validates tournament status transitions
func isValidStatusTransition(from, to serviceextension.TournamentStatus) bool {
	switch from {
	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT:
		return to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_DRAFT ||
			to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE ||
			to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED

	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE:
		return to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE ||
			to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED ||
			to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED

	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED:
		return to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_STARTED ||
			to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED ||
			to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED

	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED:
		return to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_COMPLETED // Terminal state

	case serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED:
		return to == serviceextension.TournamentStatus_TOURNAMENT_STATUS_CANCELLED // Terminal state

	default:
		return false
	}
}

// documentToProto converts MongoDB document to protobuf message
func (m *MongoTournamentStorage) documentToProto(doc *tournamentDocument) *serviceextension.Tournament {
	return &serviceextension.Tournament{
		TournamentId:        doc.TournamentID,
		Name:                doc.Name,
		Description:         doc.Description,
		MaxParticipants:     doc.MaxParticipants,
		CurrentParticipants: doc.CurrentParticipants,
		Status:              doc.Status,
		CreatedAt:           timestamppb.New(doc.CreatedAt),
		UpdatedAt:           timestamppb.New(doc.UpdatedAt),
		StartTime:           timestamppb.New(doc.StartTime),
		EndTime:             timestamppb.New(doc.EndTime),
	}
}
