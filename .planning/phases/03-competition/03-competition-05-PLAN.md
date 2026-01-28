---
phase: 03-competition
plan: 05
type: execute
wave: 1
depends_on: [03-competition-02]
files_modified: [pkg/storage/storage.go, pkg/service/tournament.go, main.go]
autonomous: true
gap_closure: true

must_haves:
  truths:
    - "Match storage interface fully integrated with storage registry"
    - "System generates single-elimination brackets when tournament starts"
    - "Match creation integrates with tournament start workflow"
    - "Storage layer follows established patterns for session management"
  artifacts:
    - path: "pkg/storage/storage.go"
      provides: "Unified storage registry with MatchStorage integration"
      contains: "MatchStorage interface and MongoDB connection sharing"
      min_lines: 100
    - path: "pkg/service/tournament.go"
      provides: "Enhanced tournament start with bracket generation"
      contains: "GenerateBrackets integration and match creation"
      min_lines: 50
    - path: "main.go"
      provides: "Complete storage initialization including match storage"
      contains: "MatchStorage registration and MongoDB setup"
      min_lines: 200
  key_links:
    - from: "MatchStorage"
      to: "Storage registry patterns"
      via: "following TournamentStorage/ParticipantStorage patterns"
      pattern: "MatchStorage.*MongoDB.*session.*transaction"
    - from: "Tournament start"
      to: "Match creation"
      via: "automatic bracket generation when status changes to STARTED"
      pattern: "StartTournament.*GenerateBrackets.*CreateMatches"
    - from: "main.go"
      to: "storage initialization"
      via: "complete storage setup following existing patterns"
      pattern: "NewMongoMatchStorage.*storage.*MatchStorage"
---

<objective>
Complete storage integration and bracket generation for tournament workflow automation

Purpose: Fix the storage registry integration gaps and ensure tournament start automatically generates match brackets, enabling the complete tournament workflow
Output: Fully integrated storage system where MatchStorage follows established patterns and tournament start triggers bracket generation
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/03-competition/03-competition-VERIFICATION.md
@.planning/phases/03-competition/03-competition-01-SUMMARY.md
@.planning/phases/03-competition/03-competition-02-SUMMARY.md
</context>

<tasks>

<task type="auto">
  <name>Integrate MatchStorage with storage registry patterns</name>
  <files>pkg/storage/storage.go</files>
  <action>
    1. Add MatchStorage interface definition to storage.go following existing patterns:
       - MatchStorage interface with all methods from existing implementation
       - Consistent with TournamentStorage and ParticipantStorage patterns
       - Include proper documentation and context parameter handling
    
    2. Add storage registry functions:
       - NewMatchStorage function that creates MongoMatchStorage with MongoDB client
       - Integration with existing MongoDB connection management
       - Session and transaction support following established patterns
    
    3. Update storage initialization to include MatchStorage:
       - Add MatchStorage to any storage registry or initialization functions
       - Ensure MongoDB connection sharing between storage types
       - Follow same error handling and logging patterns
    
    Reference existing code: From STATE.md, existing storage patterns are established in Phase 1 and 2
    Gap reason: MatchStorage not registered in storage.go, no integration with existing patterns (from VERIFICATION.md gap 2)
  </action>
  <verify>Check pkg/storage/storage.go contains MatchStorage interface and initialization functions matching existing patterns</verify>
  <done>MatchStorage follows same storage registry patterns as TournamentStorage and ParticipantStorage</done>
</task>

<task type="auto">
  <name>Integrate bracket generation with tournament start workflow</name>
  <files>pkg/service/tournament.go</files>
  <action>
    1. Enhance StartTournament method to include bracket generation:
       - After tournament status changes to STARTED, call bracket generation
       - Use existing GenerateBrackets function (from Phase 1 foundation)
       - Create matches using MatchStorage.CreateMatches with generated bracket data
       - Handle odd participant counts with proper bye assignments
    
    2. Add GenerateTournamentBrackets helper method:
       - Query registered participants using ParticipantStorage
       - Generate single-elimination bracket structure with round/position calculations
       - Handle bye participants for odd numbers
       - Return list of matches with participant assignments
    
    3. Integrate with existing tournament validation:
       - Ensure minimum participant requirements are met
       - Validate tournament is in correct state for bracket generation
       - Handle errors and rollback tournament status if bracket generation fails
    
    Reference existing code: From 01-foundation-04-SUMMARY, bracket generation algorithm exists but integration with matches is missing
    Gap reason: Integration between tournament start and match creation not verified (from VERIFICATION.md gap 4)
  </action>
  <verify>Review StartTournament method - should call bracket generation and create matches when tournament starts</verify>
  <done>Tournament start automatically creates all bracket matches with proper participant assignments</done>
</task>

<task type="auto">
  <name>Update main.go with complete storage initialization</name>
  <files>main.go</files>
  <action>
    1. Update storage initialization in main.go:
       - Add MatchStorage initialization alongside existing TournamentStorage and ParticipantStorage
       - Use NewMatchStorage from storage.go registry
       - Ensure MongoDB client is shared across all storage types
       - Add proper error handling for storage initialization
    
    2. Update service initialization:
       - Pass MatchStorage to MatchService constructor
       - Ensure all storage dependencies are properly injected
       - Fix any LSP errors related to missing authentication interceptor parameters
    
    3. Add database index initialization:
       - Ensure MatchStorage indexes are created on startup
       - Follow same patterns as tournament and participant index creation
       - Add logging for index creation status
    
    Reference existing code: From 03-competition-02-SUMMARY, MatchStorage exists but main.go integration is incomplete
    Gap reason: Storage initialization incomplete, causing LSP errors and missing integration
  </action>
  <verify>Check main.go compiles without errors and includes MatchStorage initialization following existing patterns</verify>
  <done>Application starts successfully with all storage types properly initialized and integrated</done>
</task>

</tasks>

<verification>
After completing all tasks, verify the complete storage integration and tournament workflow:

1. Build the application: `go build ./...` should succeed without errors
2. Run the application and verify storage initialization completes successfully
3. Create a tournament with registered participants
4. Start the tournament and verify brackets are automatically generated
5. Check that matches are created in storage with correct round/position structure
6. Verify storage operations (create, read, update) work for all storage types
7. Test transaction support across storage types
</verification>

<success_criteria>
- MatchStorage interface fully integrated with storage registry (addresses gap 2)
- Tournament start automatically generates single-elimination brackets (addresses gap 4)
- Match creation integrates seamlessly with tournament start workflow
- Storage layer follows established MongoDB session and transaction patterns
- Application builds and runs without storage-related errors
- All storage types share MongoDB client connection properly
</success_criteria>

<output>
After completion, create `.planning/phases/03-competition/03-competition-05-SUMMARY.md`
</output>