---
phase: 02-participation
plan: 02
type: execute
wave: 2
depends_on: ["02-participation-01"]
files_modified: ["pkg/storage/participant.go", "pkg/storage/tournament.go"]
autonomous: true

must_haves:
  truths:
    - "Participant storage exists with concurrent-safe operations"
    - "MongoDB transactions used for registration atomicity"
    - "Tournament storage enhanced with participant count handling"
    - "Capacity enforcement with atomic operations"
  artifacts:
    - path: "pkg/storage/participant.go"
      provides: "Participant CRUD operations with concurrent safety"
      min_lines: 150
      exports: ["RegisterParticipant", "GetParticipants", "RemoveParticipant"]
    - path: "pkg/storage/tournament.go"
      provides: "Enhanced tournament storage with participant count management"
      contains: "UpdateParticipantCount"
  key_links:
    - from: "pkg/storage/participant.go"
      to: "pkg/storage/tournament.go"
      via: "tournament collection updates"
      pattern: "tournamentCollection\\.UpdateOne"
    - from: "pkg/storage/participant.go"
      to: "MongoDB session"
      via: "transaction handling"
      pattern: "session\\.WithTransaction"
---

<objective>
Implement participant storage layer with concurrent-safe registration operations and MongoDB transaction support for atomic participant management.

Purpose: Provide reliable participant storage with capacity enforcement, duplicate prevention, and transaction-based registration integrity.
Output: Complete participant storage service with concurrent-safe operations and tournament integration.
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/02-participation/02-CONTEXT.md
@.planning/phases/02-participation/02-RESEARCH.md
@.planning/phases/01-foundation/01-foundation-02-SUMMARY.md
</context>

<tasks>

<task type="auto">
  <name>Create participant storage with concurrent registration</name>
  <files>pkg/storage/participant.go</files>
  <action>
Create pkg/storage/participant.go with concurrent-safe participant operations following Phase 1 storage patterns:

```go
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	tournamentpb "github.com/accelerated-development/tournament-system/pkg/pb"
)

// ParticipantStorage handles participant data operations
type ParticipantStorage struct {
	client             *mongo.Client
	dbName             string
	participantCollection string
	tournamentCollection  string
}

// NewParticipantStorage creates a new participant storage instance
func NewParticipantStorage(client *mongo.Client, dbName string) *ParticipantStorage {
	return &ParticipantStorage{
		client:                 client,
		dbName:                 dbName,
		participantCollection:  "participants",
		tournamentCollection:   "tournaments",
	}
}

// RegisterParticipant registers a user for a tournament with transaction safety
func (p *ParticipantStorage) RegisterParticipant(ctx context.Context, req *tournamentpb.RegisterForTournamentRequest, userID string) (*tournamentpb.RegisterForTournamentResponse, error) {
	session, err := p.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Step 1: Get tournament for update and validate
		tournament, err := p.getTournamentForUpdate(sessCtx, req.GetNamespace(), req.GetTournamentId())
		if err != nil {
			return nil, fmt.Errorf("failed to get tournament: %w", err)
		}

		// Step 2: Validate tournament state and capacity
		if tournament.Status != tournamentpb.TournamentStatus_TOURNAMENT_STATUS_ACTIVE {
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
		participant := &tournamentpb.Participant{
			ParticipantId: uuid.New().String(),
			UserId:        userID,
			TournamentId:  req.GetTournamentId(),
			RegisteredAt:  timestamppb.New(now),
			UpdatedAt:     timestamppb.New(now),
		}

		if _, err := p.participantCollection.InsertOne(sessCtx, participant); err != nil {
			return nil, fmt.Errorf("failed to create participant: %w", err)
		}

		// Step 5: Update tournament participant count
		update := bson.M{"$inc": bson.M{"current_participants": 1}}
		if _, err := p.tournamentCollection.UpdateOne(
			sessCtx,
			bson.M{"tournament_id": req.GetTournamentId(), "namespace": req.GetNamespace()},
			update,
		); err != nil {
			return nil, fmt.Errorf("failed to update tournament count: %w", err)
		}

		return &tournamentpb.RegisterForTournamentResponse{
			ParticipantId: participant.ParticipantId,
			TournamentId:  participant.TournamentId,
			UserId:        participant.UserId,
			RegisteredAt:  participant.RegisteredAt,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*tournamentpb.RegisterForTournamentResponse), nil
}

// GetParticipants retrieves paginated participants for a tournament
func (p *ParticipantStorage) GetParticipants(ctx context.Context, req *tournamentpb.GetTournamentParticipantsRequest) (*tournamentpb.GetTournamentParticipantsResponse, error) {
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
	cursor, err := p.participantCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants: %w", err)
	}
	defer cursor.Close(ctx)

	var participants []*tournamentpb.Participant
	if err := cursor.All(ctx, &participants); err != nil {
		return nil, fmt.Errorf("failed to decode participants: %w", err)
	}

	// Get total count
	total, err := p.participantCollection.CountDocuments(ctx, bson.M{
		"tournament_id": req.GetTournamentId(),
		"namespace":     req.GetNamespace(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count participants: %w", err)
	}

	// Generate next page token
	var nextPageToken string
	if len(participants) > 0 && req.GetPageSize() > 0 && int32(len(participants)) >= req.GetPageSize() {
		nextPageToken = participants[len(participants)-1].ParticipantId
	}

	return &tournamentpb.GetTournamentParticipantsResponse{
		Participants:    participants,
		TotalParticipants: int32(total),
		NextPageToken:   nextPageToken,
	}, nil
}

// RemoveParticipant removes a participant from a tournament (admin only)
func (p *ParticipantStorage) RemoveParticipant(ctx context.Context, req *tournamentpb.RemoveParticipantRequest) (*tournamentpb.RemoveParticipantResponse, error) {
	session, err := p.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Step 1: Find and delete participant
		filter := bson.M{
			"user_id":       req.GetUserId(),
			"tournament_id": req.GetTournamentId(),
			"namespace":     req.GetNamespace(),
		}

		deleteResult, err := p.participantCollection.DeleteOne(sessCtx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete participant: %w", err)
		}

		if deleteResult.DeletedCount == 0 {
			return nil, errors.New("participant not found")
		}

		// Step 2: Update tournament participant count (decrement)
		update := bson.M{"$inc": bson.M{"current_participants": -1}}
		if _, err := p.tournamentCollection.UpdateOne(
			sessCtx,
			bson.M{"tournament_id": req.GetTournamentId(), "namespace": req.GetNamespace()},
			update,
		); err != nil {
			return nil, fmt.Errorf("failed to update tournament count: %w", err)
		}

		return &tournamentpb.RemoveParticipantResponse{
			TournamentId: req.GetTournamentId(),
			UserId:       req.GetUserId(),
			Removed:      true,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*tournamentpb.RemoveParticipantResponse), nil
}

// Helper methods

func (p *ParticipantStorage) getTournamentForUpdate(ctx context.Context, namespace, tournamentID string) (*tournamentpb.Tournament, error) {
	var tournament tournamentpb.Tournament
	err := p.tournamentCollection.FindOne(ctx, bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}).Decode(&tournament)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("tournament not found")
		}
		return nil, err
	}
	return &tournament, nil
}

func (p *ParticipantStorage) findParticipant(ctx context.Context, userID, tournamentID string) (*tournamentpb.Participant, error) {
	var participant tournamentpb.Participant
	err := p.participantCollection.FindOne(ctx, bson.M{
		"user_id":       userID,
		"tournament_id": tournamentID,
	}).Decode(&participant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found is not an error
		}
		return nil, err
	}
	return &participant, nil
}
```

Follow Phase 1 storage patterns for MongoDB connection handling and error management.
  </action>
  <verify>grep -n "RegisterParticipant" pkg/storage/participant.go && grep -n "WithTransaction" pkg/storage/participant.go</verify>
  <done>Participant storage created with transaction-based registration and concurrent safety</done>
</task>

<task type="auto">
  <name>Enhance tournament storage for participant integration</name>
  <files>pkg/storage/tournament.go</files>
  <action>
Enhance existing tournament.go storage to support participant count management:

1. Add method to get tournament with proper locking for registration:
```go
// GetTournamentForRegistration gets tournament with additional validation for registration
func (t *TournamentStorage) GetTournamentForRegistration(ctx context.Context, namespace, tournamentID string) (*tournamentpb.Tournament, error) {
	var tournament tournamentpb.Tournament
	err := t.collection.FindOne(ctx, bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}).Decode(&tournament)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("tournament not found")
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}
	
	// Additional validation for registration
	if tournament.Status != tournamentpb.TournamentStatus_TOURNAMENT_STATUS_ACTIVE {
		return nil, errors.New("tournament not open for registration")
	}
	
	return &tournament, nil
}
```

2. Add method to update participant count (used by participant storage):
```go
// UpdateParticipantCount atomically updates tournament participant count
func (t *TournamentStorage) UpdateParticipantCount(ctx context.Context, namespace, tournamentID string, increment int32) error {
	filter := bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}
	
	update := bson.M{"$inc": bson.M{"current_participants": increment}}
	
	result, err := t.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update participant count: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return errors.New("tournament not found")
	}
	
	// Validate count doesn't go below 0 or above max
	var updatedTournament tournamentpb.Tournament
	err = t.collection.FindOne(ctx, filter).Decode(&updatedTournament)
	if err != nil {
		return fmt.Errorf("failed to verify updated tournament: %w", err)
	}
	
	if updatedTournament.CurrentParticipants < 0 {
		return errors.New("participant count cannot be negative")
	}
	
	if updatedTournament.CurrentParticipants > updatedTournament.MaxParticipants {
		return errors.New("participant count exceeds maximum")
	}
	
	return nil
}
```

3. Add method to check tournament capacity:
```go
// CheckTournamentCapacity returns whether tournament has space for more participants
func (t *TournamentStorage) CheckTournamentCapacity(ctx context.Context, namespace, tournamentID string) (bool, error) {
	var tournament tournamentpb.Tournament
	err := t.collection.FindOne(ctx, bson.M{
		"tournament_id": tournamentID,
		"namespace":     namespace,
	}).Decode(&tournament)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("tournament not found")
		}
		return false, fmt.Errorf("failed to get tournament: %w", err)
	}
	
	return tournament.CurrentParticipants < tournament.MaxParticipants, nil
}
```

These enhancements support the participant storage transaction logic.
  </action>
  <verify>grep -n "UpdateParticipantCount" pkg/storage/tournament.go && grep -n "CheckTournamentCapacity" pkg/storage/tournament.go</verify>
  <done>Tournament storage enhanced with participant count management methods</done>
</task>

</tasks>

<verification>
- Participant storage implements transaction-based registration
- MongoDB sessions properly used for atomic operations
- Tournament capacity enforcement with database-level validation
- Participant listing with pagination support
- Admin participant removal with count adjustment
- Duplicate registration prevention implemented
- Error handling follows Phase 1 patterns
- All database operations use proper namespace filtering
</verification>

<success_criteria>
- Complete participant storage with concurrent-safe registration
- MongoDB transaction support for atomic participant/tournament updates
- Capacity enforcement with proper error messages
- Participant listing with pagination and sorting by registration order
- Admin participant removal with tournament count adjustment
- Duplicate registration prevention
- Integration with existing tournament storage patterns
</success_criteria>

<output>
After completion, create `.planning/phases/02-participation/02-participation-02-SUMMARY.md`
</output>