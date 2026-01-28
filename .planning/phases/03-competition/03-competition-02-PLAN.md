---
phase: 03-competition
plan: 02
type: execute
wave: 2
depends_on: [03-competition-01]
files_modified: [pkg/storage/match.go, pkg/storage/storage.go]
autonomous: true
user_setup: []

must_haves:
  truths:
    - "Match storage interface defined with CRUD operations for MongoDB"
    - "Transaction support for atomic match result submission"
    - "Match retrieval methods for bracket viewing and individual match details"
    - "Integration with existing MongoDB connection and session management"
    - "Concurrent-safe match updates with proper error handling"
  artifacts:
    - path: "pkg/storage/match.go"
      provides: "MatchStorage interface and MongoDB implementation"
      exports: ["MatchStorage", "Match", "GetMatch", "UpdateMatch", "SubmitMatchResult"]
      min_lines: 300
    - path: "pkg/storage/storage.go"
      provides: "Storage interface registry and MongoDB connection sharing"
      contains: "MongoDB", "collection", "session", "transaction"
      min_lines: 100
  key_links:
    - from: "MatchStorage"
      to: "existing TournamentStorage patterns"
      via: "MongoDB session and transaction patterns"
      pattern: "StartSession.*WithTransaction"
    - from: "match storage operations"
      to: "tournament storage"
      via: "shared MongoDB client and session management"
      pattern: "mongo.Collection.*match.*tournament"
    - from: "MatchStorage.SubmitMatchResult"
      to: "ParticipantStorage transaction patterns"
      via: "atomic multi-document operations"
      pattern: "transaction.*callback.*UpdateMatch"
---

<objective>
Implement match storage layer with MongoDB persistence and transaction support

Purpose: Provide atomic, concurrent-safe match data operations that integrate with existing tournament and participant storage patterns
Output: Complete MatchStorage implementation with CRUD operations, transaction support, and bracket retrieval methods
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/03-competition/03-CONTEXT.md
@.planning/phases/03-competition/03-RESEARCH.md
@03-competition-01-SUMMARY.md
@pkg/storage/tournament.go
@pkg/storage/participant.go
@pkg/service/tournament.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Define MatchStorage interface and MongoDB implementation</name>
  <files>pkg/storage/match.go</files>
  <action>
    Create MatchStorage following existing patterns from tournament.go and participant.go:
    
    1. Define MatchStorage interface with methods:
       - GetMatch(ctx, namespace, tournament_id, match_id) -> (*Match, error)
       - GetTournamentMatches(ctx, namespace, tournament_id) -> ([]Match, error)
       - CreateMatches(ctx, namespace, tournament_id, matches) -> error
       - UpdateMatch(ctx, namespace, match) -> error
       - SubmitMatchResult(ctx, namespace, tournament_id, match_id, winner_user_id) -> error
       - GetMatchesByRound(ctx, namespace, tournament_id, round) -> ([]Match, error)
    
    2. Implement MongoDB MatchStorage struct with:
       - MongoDB client and database references (reuse from existing storage)
       - Match collection management with proper indexing
       - Connection health checks and error handling
    
    3. Match document structure matching protobuf Match message:
       - match_id (string, indexed)
       - tournament_id (string, indexed) 
       - round (int32)
       - position (int32)
       - participant1, participant2 (embedded documents)
       - winner (string)
       - status (string)
       - started_at, completed_at (timestamps)
       - created_at, updated_at (auto-managed)
    
    4. Follow existing naming conventions and error handling patterns
    5. Use same MongoDB session management as TournamentStorage
  </action>
  <verify>go build ./pkg/storage/ succeeds and MatchStorage interface methods are properly defined</verify>
  <done>MatchStorage interface and MongoDB implementation created with all required CRUD operations</done>
</task>

<task type="auto">
  <name>Task 2: Implement atomic match result submission with transactions</name>
  <files>pkg/storage/match.go</files>
  <action>
    Implement transaction-based match result submission following participant storage patterns:
    
    1. Create SubmitMatchResult method with MongoDB transaction:
       - Start session using existing StartSession pattern
       - Define transaction callback function
       - Validate match exists and is in correct status
       - Validate winner is one of the participants
       - Update match with winner and completion timestamp
       - Handle round progression logic if needed
       - Commit transaction atomically
    
    2. Error handling for transaction scenarios:
       - Match not found -> return specific error
       - Invalid winner -> return validation error
       - Match already completed -> return conflict error
       - Transaction rollback on any failure
    
    3. Use existing gRPC status error patterns from tournament storage
    4. Add structured logging for audit trail
    5. Implement duplicate submission prevention using status validation
    
    4. Integration points:
       - Reuse session management from existing storage patterns
       - Match same error formatting and status codes
       - Follow same context propagation patterns
  </action>
  <verify>Unit tests show transaction rollback works and match result submission is atomic</verify>
  <done>Atomic match result submission implemented with proper validation, error handling, and audit logging</done>
</task>

<task type="auto">
  <name>Task 3: Implement match retrieval and bracket viewing methods</name>
  <files>pkg/storage/match.go, pkg/storage/storage.go</files>
  <action>
    Implement match viewing and bracket organization methods:
    
    1. Create GetTournamentMatches method:
       - Query all matches for tournament sorted by round and position
       - Return matches organized by round for bracket display
       - Handle empty tournament or missing matches gracefully
       - Support pagination for large tournaments
    
    2. Create GetMatchesByRound method:
       - Filter matches by tournament_id and round
       - Return matches in position order for bracket rendering
       - Include participant and winner information
       - Handle round queries beyond tournament bounds
    
    3. Create CreateMatches method:
       - Bulk insert matches for tournament initialization
       - Called when tournament starts to populate bracket
       - Use MongoDB insertMany for performance
       - Include proper error handling for partial failures
    
    4. Update storage.go to register MatchStorage:
       - Add MatchStorage to existing storage registry pattern
       - Share MongoDB connection and session management
       - Initialize match collection indexes
       - Add health check endpoint for match storage
    
    5. Performance optimizations:
       - Add database indexes for common queries (tournament_id, match_id, round)
       - Use appropriate MongoDB query projections
       - Implement efficient sorting and pagination
  </action>
  <verify>Integration tests show matches can be retrieved and organized correctly by round</verify>
  <done>Match retrieval and bracket viewing methods implemented with proper sorting, pagination, and performance optimization</done>
</task>

</tasks>

<verification>
- [ ] MatchStorage interface defined with all required CRUD operations
- [ ] MongoDB implementation follows existing tournament/participant storage patterns
- [ ] Transaction-based match result submission with proper validation
- [ ] Atomic operations prevent race conditions and ensure data consistency
- [ ] Match retrieval methods support bracket viewing and individual match details
- [ ] Database indexes added for common query patterns
- [ ] Error handling matches existing gRPC status patterns
- [ ] Integration with existing MongoDB connection and session management
</verification>

<success_criteria>
Complete match storage layer ready for service implementation
- Atomic match result submission with transaction safety
- Match retrieval methods for bracket viewing and individual details
- Integration with existing MongoDB infrastructure
- Concurrent-safe operations with proper error handling
- Performance optimizations for tournament bracket queries
</success_criteria>

<output>
After completion, create `.planning/phases/03-competition/03-competition-02-SUMMARY.md`
</output>