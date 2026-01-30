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

// ParticipantStorage handles participant data operations
type ParticipantStorage struct {
	client                *mongo.Client
	database              string
	participantCollection string
	tournamentCollection  string
	logger                *slog.Logger
}

// NewParticipantStorage creates a new participant storage instance
func NewParticipantStorage(client *mongo.Client, database string, logger *slog.Logger) *ParticipantStorage {
	return &ParticipantStorage{
		client:                client,
		database:              database,
		participantCollection: "participants",
		tournamentCollection:  "tournaments",
		logger:                logger,
	}
}

// participantDocument represents the MongoDB document structure
type participantDocument struct {
	ParticipantID string    `bson:"participant_id"`
	UserID        string    `bson:"user_id"`
	Username      string    `bson:"username"`
	DisplayName   string    `bson:"display_name"`
	TournamentID  string    `bson:"tournament_id"`
	Namespace     string    `bson:"namespace"`
	RegisteredAt  time.Time `bson:"registered_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
}

// RegisterParticipant registers a user for a tournament with transaction safety
func (p *ParticipantStorage) RegisterParticipant(ctx context.Context, req *serviceextension.RegisterForTournamentRequest, userID string) (*serviceextension.RegisterForTournamentResponse, error) {
	session, err := p.client.StartSession()
	if err != nil {
		p.logger.Error("failed to start session", "error", err)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Step 1: Get tournament for update and validate
		tournament, err := p.getTournamentForUpdate(sessCtx, req.GetNamespace(), req.GetTournamentId())
		if err != nil {
			return nil, fmt.Errorf("failed to get tournament: %w", err)
		}

		// Step 2: Validate tournament state and capacity
		if tournament.Status != serviceextension.TournamentStatus_TOURNAMENT_STATUS_ACTIVE {
			return nil, errors.New("tournament not open for registration")
		}

		if tournament.CurrentParticipants >= tournament.MaxParticipants {
			return nil, errors.New("tournament is full")
		}

		// Step 3: Check for duplicate registration
		existing, err := p.findParticipant(sessCtx, userID, req.GetTournamentId())
		if err == nil && existing != nil {
			return nil, errors.New("already registered for this tournament")
		}

		// Step 4: Create participant record
		now := time.Now()
		participant := &serviceextension.Participant{
			ParticipantId: uuid.New().String(),
			UserId:        userID,
			TournamentId:  req.GetTournamentId(),
			RegisteredAt:  timestamppb.New(now),
			UpdatedAt:     timestamppb.New(now),
		}

		// Convert to MongoDB document
		doc := &participantDocument{
			ParticipantID: participant.ParticipantId,
			UserID:        participant.UserId,
			Username:      participant.Username,    // Will be populated from user context
			DisplayName:   participant.DisplayName, // Will be populated from user context
			TournamentID:  participant.TournamentId,
			Namespace:     req.GetNamespace(),
			RegisteredAt:  now,
			UpdatedAt:     now,
		}

		participantCollection := p.client.Database(p.database).Collection(p.participantCollection)
		if _, err := participantCollection.InsertOne(sessCtx, doc); err != nil {
			p.logger.Error("failed to create participant", "error", err, "participant_id", participant.ParticipantId)
			return nil, fmt.Errorf("failed to create participant: %w", err)
		}

		// Step 5: Update tournament participant count
		tournamentCollection := p.client.Database(p.database).Collection(p.tournamentCollection)
		update := bson.M{"$inc": bson.M{"current_participants": 1}}
		if _, err := tournamentCollection.UpdateOne(
			sessCtx,
			bson.M{"tournament_id": req.GetTournamentId(), "namespace": req.GetNamespace()},
			update,
		); err != nil {
			p.logger.Error("failed to update tournament count", "error", err, "tournament_id", req.GetTournamentId())
			return nil, fmt.Errorf("failed to update tournament count: %w", err)
		}

		p.logger.Info("participant registered", "participant_id", participant.ParticipantId, "tournament_id", participant.TournamentId, "user_id", userID)

		return &serviceextension.RegisterForTournamentResponse{
			ParticipantId: participant.ParticipantId,
			TournamentId:  participant.TournamentId,
			UserId:        participant.UserId,
			RegisteredAt:  participant.RegisteredAt,
		}, nil
	})

	if err != nil {
		p.logger.Error("transaction failed", "error", err, "tournament_id", req.GetTournamentId(), "user_id", userID)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to register participant: %v", err)
	}

	return result.(*serviceextension.RegisterForTournamentResponse), nil
}

// GetParticipants retrieves paginated participants for a tournament
func (p *ParticipantStorage) GetParticipants(ctx context.Context, req *serviceextension.GetTournamentParticipantsRequest) (*serviceextension.GetTournamentParticipantsResponse, error) {
	collection := p.client.Database(p.database).Collection(p.participantCollection)

	// Build query filter
	filter := bson.M{
		"tournament_id": req.GetTournamentId(),
		"namespace":     req.GetNamespace(),
	}

	// Set up pagination
	findOptions := options.Find()
	if req.GetPageSize() > 0 {
		findOptions.SetLimit(int64(req.GetPageSize()))
	}
	if req.GetPageToken() != "" {
		// Simple pagination using participant_id as cursor
		filter["participant_id"] = bson.M{"$gt": req.GetPageToken()}
	}
	findOptions.SetSort(bson.M{"registered_at": 1}) // Registration order

	// Query participants
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		p.logger.Error("failed to query participants", "error", err, "tournament_id", req.GetTournamentId())
		return nil, grpcStatus.Errorf(codes.Internal, "failed to query participants: %v", err)
	}
	defer cursor.Close(ctx)

	var docs []participantDocument
	if err := cursor.All(ctx, &docs); err != nil {
		p.logger.Error("failed to decode participants", "error", err, "tournament_id", req.GetTournamentId())
		return nil, grpcStatus.Errorf(codes.Internal, "failed to decode participants: %v", err)
	}

	// Convert to protobuf
	participants := make([]*serviceextension.Participant, len(docs))
	for i, doc := range docs {
		participants[i] = p.documentToProto(&doc)
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, bson.M{
		"tournament_id": req.GetTournamentId(),
		"namespace":     req.GetNamespace(),
	})
	if err != nil {
		p.logger.Error("failed to count participants", "error", err, "tournament_id", req.GetTournamentId())
		return nil, grpcStatus.Errorf(codes.Internal, "failed to count participants: %v", err)
	}

	// Generate next page token
	var nextPageToken string
	if len(participants) > 0 && req.GetPageSize() > 0 && int32(len(participants)) >= req.GetPageSize() {
		nextPageToken = participants[len(participants)-1].ParticipantId
	}

	return &serviceextension.GetTournamentParticipantsResponse{
		Participants:      participants,
		TotalParticipants: int32(total),
		NextPageToken:     nextPageToken,
	}, nil
}

// RemoveParticipant removes a participant from a tournament (admin only)
func (p *ParticipantStorage) RemoveParticipant(ctx context.Context, req *serviceextension.RemoveParticipantRequest) (*serviceextension.RemoveParticipantResponse, error) {
	session, err := p.client.StartSession()
	if err != nil {
		p.logger.Error("failed to start session", "error", err)
		return nil, grpcStatus.Errorf(codes.Internal, "failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Step 1: Find and delete participant
		filter := bson.M{
			"user_id":       req.GetUserId(),
			"tournament_id": req.GetTournamentId(),
			"namespace":     req.GetNamespace(),
		}

		participantCollection := p.client.Database(p.database).Collection(p.participantCollection)
		deleteResult, err := participantCollection.DeleteOne(sessCtx, filter)
		if err != nil {
			p.logger.Error("failed to delete participant", "error", err, "user_id", req.GetUserId(), "tournament_id", req.GetTournamentId())
			return nil, fmt.Errorf("failed to delete participant: %w", err)
		}

		if deleteResult.DeletedCount == 0 {
			return nil, errors.New("participant not found")
		}

		// Step 2: Update tournament participant count (decrement)
		tournamentCollection := p.client.Database(p.database).Collection(p.tournamentCollection)
		update := bson.M{"$inc": bson.M{"current_participants": -1}}
		if _, err := tournamentCollection.UpdateOne(
			sessCtx,
			bson.M{"tournament_id": req.GetTournamentId(), "namespace": req.GetNamespace()},
			update,
		); err != nil {
			p.logger.Error("failed to update tournament count", "error", err, "tournament_id", req.GetTournamentId())
			return nil, fmt.Errorf("failed to update tournament count: %w", err)
		}

		p.logger.Info("participant removed", "tournament_id", req.GetTournamentId(), "user_id", req.GetUserId())

		return &serviceextension.RemoveParticipantResponse{
			TournamentId: req.GetTournamentId(),
			UserId:       req.GetUserId(),
			Removed:      true,
		}, nil
	})

	if err != nil {
		p.logger.Error("transaction failed", "error", err, "tournament_id", req.GetTournamentId(), "user_id", req.GetUserId())
		return nil, grpcStatus.Errorf(codes.Internal, "failed to remove participant: %v", err)
	}

	return result.(*serviceextension.RemoveParticipantResponse), nil
}

// Helper methods

func (p *ParticipantStorage) getTournamentForUpdate(ctx context.Context, namespace, tournamentID string) (*serviceextension.Tournament, error) {
	collection := p.client.Database(p.database).Collection(p.tournamentCollection)

	var tournamentDoc struct {
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

	err := collection.FindOne(ctx, bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}).Decode(&tournamentDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("tournament not found")
		}
		return nil, err
	}

	return &serviceextension.Tournament{
		TournamentId:        tournamentDoc.TournamentID,
		Name:                tournamentDoc.Name,
		Description:         tournamentDoc.Description,
		MaxParticipants:     tournamentDoc.MaxParticipants,
		CurrentParticipants: tournamentDoc.CurrentParticipants,
		Status:              tournamentDoc.Status,
		CreatedAt:           timestamppb.New(tournamentDoc.CreatedAt),
		UpdatedAt:           timestamppb.New(tournamentDoc.UpdatedAt),
		StartTime:           timestamppb.New(tournamentDoc.StartTime),
		EndTime:             timestamppb.New(tournamentDoc.EndTime),
	}, nil
}

func (p *ParticipantStorage) findParticipant(ctx context.Context, userID, tournamentID string) (*serviceextension.Participant, error) {
	collection := p.client.Database(p.database).Collection(p.participantCollection)

	var doc participantDocument
	err := collection.FindOne(ctx, bson.M{
		"user_id":       userID,
		"tournament_id": tournamentID,
	}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found is not an error
		}
		return nil, err
	}
	return p.documentToProto(&doc), nil
}

// documentToProto converts MongoDB document to protobuf message
func (p *ParticipantStorage) documentToProto(doc *participantDocument) *serviceextension.Participant {
	return &serviceextension.Participant{
		ParticipantId: doc.ParticipantID,
		UserId:        doc.UserID,
		Username:      doc.Username,
		DisplayName:   doc.DisplayName,
		TournamentId:  doc.TournamentID,
		RegisteredAt:  timestamppb.New(doc.RegisteredAt),
		UpdatedAt:     timestamppb.New(doc.UpdatedAt),
	}
}
