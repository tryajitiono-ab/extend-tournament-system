// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package storage

import (
	"context"
	"encoding/json"
	pb "extend-tournament-service/pkg/pb"
	"log/slog"

	"github.com/AccelByte/accelbyte-go-sdk/cloudsave-sdk/pkg/cloudsaveclientmodels"

	"github.com/AccelByte/accelbyte-go-sdk/cloudsave-sdk/pkg/cloudsaveclient/admin_game_record"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/cloudsave"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Storage interface for cloudsave operations (guild service legacy)
type Storage interface {
	GetGuildProgress(ctx context.Context, namespace string, key string) (*pb.GuildProgress, error)
	SaveGuildProgress(ctx context.Context, namespace string, key string, value *pb.GuildProgress) (*pb.GuildProgress, error)
}

// StorageRegistry provides factory functions for all storage types following established patterns
type StorageRegistry struct {
	client   *mongo.Client
	database string
	logger   *slog.Logger
}

// NewStorageRegistry creates a new storage registry with MongoDB client
func NewStorageRegistry(client *mongo.Client, database string, logger *slog.Logger) *StorageRegistry {
	return &StorageRegistry{
		client:   client,
		database: database,
		logger:   logger,
	}
}

// NewTournamentStorage creates a MongoDB tournament storage instance
func (r *StorageRegistry) NewTournamentStorage() TournamentStorage {
	return NewMongoTournamentStorage(r.client, r.database, r.logger)
}

// NewParticipantStorage creates a MongoDB participant storage instance
func (r *StorageRegistry) NewParticipantStorage() *ParticipantStorage {
	return NewParticipantStorage(r.client, r.database, r.logger)
}

// NewMatchStorage creates a MongoDB match storage instance following established patterns
func (r *StorageRegistry) NewMatchStorage() MatchStorage {
	return NewMongoMatchStorage(r.client, r.database, r.logger)
}

// EnsureAllIndexes creates all necessary database indexes for all storage types
func (r *StorageRegistry) EnsureAllIndexes(ctx context.Context) error {
	r.logger.Info("ensuring database indexes for all storage types")

	// Create match storage indexes using concrete type
	matchStorage := NewMongoMatchStorage(r.client, r.database, r.logger)
	if err := matchStorage.EnsureIndexes(ctx); err != nil {
		return err
	}

	r.logger.Info("all storage indexes created successfully")
	return nil
}

type CloudsaveStorage struct {
	csStorage *cloudsave.AdminGameRecordService
}

func NewCloudSaveStorage(csStorage *cloudsave.AdminGameRecordService) *CloudsaveStorage {
	return &CloudsaveStorage{
		csStorage: csStorage,
	}
}

func (c *CloudsaveStorage) SaveGuildProgress(ctx context.Context, namespace string, key string, value *pb.GuildProgress) (*pb.GuildProgress, error) {
	input := &admin_game_record.AdminPostGameRecordHandlerV1Params{
		Body:      value,
		Key:       key,
		Namespace: namespace,
		Context:   ctx,
	}
	response, err := c.csStorage.AdminPostGameRecordHandlerV1Short(input)
	if err != nil {
		return nil, err
	}

	guildProgress, err := parseResponseToGuildProgress(response)
	if err != nil {
		return nil, err
	}

	return guildProgress, nil
}

func (c *CloudsaveStorage) GetGuildProgress(ctx context.Context, namespace string, key string) (*pb.GuildProgress, error) {
	input := &admin_game_record.AdminGetGameRecordHandlerV1Params{
		Key:       key,
		Namespace: namespace,
		Context:   ctx,
	}
	response, err := c.csStorage.AdminGetGameRecordHandlerV1Short(input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error getting guild progress: %v", err)
	}

	guildProgress, err := parseResponseToGuildProgress(response)
	if err != nil {
		return nil, err
	}

	return guildProgress, nil
}

func parseResponseToGuildProgress(response *cloudsaveclientmodels.ModelsGameRecordAdminResponse) (*pb.GuildProgress, error) {
	// Convert the response value to a JSON string
	valueJSON, err := json.Marshal(response.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error marshalling value into JSON: %v", err)
	}

	// Unmarshal the JSON string into a pb.GuildProgress
	var guildProgress pb.GuildProgress
	err = json.Unmarshal(valueJSON, &guildProgress)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error unmarshalling value into GuildProgress: %v", err)
	}

	return &guildProgress, nil
}
