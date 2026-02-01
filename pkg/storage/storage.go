// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package storage

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
)

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
