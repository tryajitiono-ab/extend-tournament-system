---
phase: 03-competition
plan: 03
type: tdd
wave: 3
depends_on: [03-competition-02]
files_modified: [pkg/service/match.go]
autonomous: true
user_setup: []

must_haves:
  truths:
    - "Match result validation logic prevents invalid winner submissions"
    - "Winner advancement algorithm correctly progresses tournament brackets"
    - "Match viewing endpoints return properly organized bracket data"
    - "Match result submission works for all authorized sources (game servers, admins)"
    - "Status transitions follow tournament business rules"
  artifacts:
    - path: "pkg/service/match.go"
      provides: "Match service implementation with TDD-tested business logic"
      exports: ["MatchService", "validateMatchWinner", "advanceWinner", "GetTournamentMatches", "SubmitMatchResult"]
      min_lines: 400
    - path: "pkg/service/match_test.go"
      provides: "Comprehensive test suite for match business logic"
      contains: "TestValidateMatchWinner", "TestAdvanceWinner", "TestSubmitMatchResult"
      min_lines: 200
  key_links:
    - from: "MatchService"
      to: "MatchStorage"
      via: "storage interface calls"
      pattern: "s.matchStorage\.GetMatch.*UpdateMatch"
    - from: "validateMatchWinner"
      to: "TournamentParticipant"
      via: "user ID validation"
      pattern: "participant.*UserId.*winner"
    - from: "advanceWinner"
      to: "existing bracket generation logic"
      via: "round and position calculations"
      pattern: "GenerateBrackets.*round.*position"
---

<objective>
Implement match service core logic with TDD for business rules validation

Purpose: Create match management service with rigorously tested business logic for result validation, winner advancement, and bracket progression
Output: Complete match service implementation with comprehensive test coverage of all tournament competition rules
</objective>

<feature>
  <name>Match Result Validation and Winner Advancement</name>
  <files>[pkg/service/match.go, pkg/service/match_test.go]</files>
  <behavior>
    Match result validation:
    - Must accept winner user ID that is one of the two participants
    - Must reject winner IDs not in participant list
    - Must reject submissions for already completed matches
    - Must accept submissions from game servers (service token) and admins (bearer token)
    
    Winner advancement logic:
    - Winner advances to next round match position based on bracket math
    - First round position (1) advances to position 1 in next round
    - First round position (2) advances to position 1 in next round
    - First round position (3) advances to position 2 in next round
    - First round position (4) advances to position 2 in next round
    - Bye participants automatically advance without result submission
    
    Cases: input -> expected output
    - validateMatchWinner("user1", [participant1:"user1", participant2:"user2"]) -> nil (valid)
    - validateMatchWinner("user3", [participant1:"user1", participant2:"user2"]) -> error (invalid winner)
    - advanceWinner(match, round:1, position:1) -> updates round:2, position:1 match
    - advanceWinner(match, round:1, position:2) -> updates round:2, position:1 match
  </behavior>
  <implementation>
    RED: Write failing tests for validation and advancement logic
    GREEN: Implement validateMatchWinner() with participant checking, advanceWinner() with bracket math
    REFACTOR: Optimize bracket position calculations and error messages
  </implementation>
</feature>

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
@03-competition-02-SUMMARY.md
@pkg/service/tournament.go
@pkg/storage/match.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: RED - Write failing tests for match result validation</name>
  <files>pkg/service/match_test.go</files>
  <action>
    Create comprehensive test suite for match business logic before implementation:
    
    1. Test structure setup:
       - Create mock MatchStorage using existing mock patterns
       - Create MatchService instance with mock dependencies
       - Set up test match data with participants
    
    2. Write failing tests for validateMatchWinner:
       - TestValidWinner: Both participant IDs should be valid winners
       - TestInvalidWinner: Non-participant ID should return error
       - TestEmptyWinner: Empty winner ID should return error
       - TestAlreadyCompleted: Completed match should reject new results
    
    3. Write failing tests for advanceWinner:
       - TestPosition1Advancement: Position 1 advances to position 1 next round
       - TestPosition2Advancement: Position 2 advances to position 1 next round  
       - TestPosition3Advancement: Position 3 advances to position 2 next round
       - TestPosition4Advancement: Position 4 advances to position 2 next round
       - TestByeHandling: Bye matches should auto-advance without result
    
    4. Write failing tests for bracket math:
       - TestRoundCalculations: Verify next round position calculations
       - TestFinalMatch: Final match winner should not advance
       - TestEdgeCases: Handle odd numbers and bracket positions
    
    5. Run tests - ALL MUST FAIL before proceeding to GREEN phase
  </action>
  <verify>go test ./pkg/service/ -run TestMatch fails with all test cases showing red state</verify>
  <done>Failing test suite written covering all match validation and advancement business rules</done>
</task>

<task type="auto">
  <name>Task 2: GREEN - Implement match service business logic</name>
  <files>pkg/service/match.go</files>
  <action>
    Implement match service logic to make all tests pass:
    
    1. Create MatchService struct following TournamentService pattern:
       - Include MatchStorage dependency
       - Include authInterceptor for permission checking
       - Include logger for audit trails
    
    2. Implement validateMatchWinner function:
       - Check if winner_user_id matches either participant1.UserId or participant2.UserId
       - Return nil for valid winners, error for invalid ones
       - Use existing gRPC status error patterns
    
    3. Implement advanceWinner function:
       - Calculate next round position based on current position
       - Position 1 & 2 -> Next round position 1
       - Position 3 & 4 -> Next round position 2
       - General formula: nextPosition = (currentPosition - 1) / 2 + 1
       - Handle final match case (no advancement needed)
    
    4. Implement match service methods:
       - SubmitMatchResult: Validate, update storage, advance winner
       - GetTournamentMatches: Organize by round for bracket display
       - GetMatch: Individual match details retrieval
    
    5. Authentication and authorization:
       - Check game server permissions for SubmitMatchResult
       - Check admin permissions for AdminSubmitMatchResult
       - Use existing authInterceptor patterns
    
    6. Run tests - ALL MUST PASS before proceeding to REFACTOR
  </action>
  <verify>go test ./pkg/service/ -run TestMatch passes with all test cases showing green state</verify>
  <done>Match service implementation completed with all business logic tests passing</done>
</task>

<task type="auto">
  <name>Task 3: REFACTOR - Optimize and clean up implementation</name>
  <files>pkg/service/match.go, pkg/service/match_test.go</files>
  <action>
    Refactor match service implementation while maintaining all tests passing:
    
    1. Code organization improvements:
       - Extract bracket position calculation to separate function
       - Consolidate error messages into constants
       - Add comprehensive function documentation
       - Follow existing Go naming and formatting conventions
    
    2. Performance optimizations:
       - Optimize match retrieval queries in GetTournamentMatches
       - Add caching for frequently accessed bracket data
       - Batch operations where possible
    
    3. Error handling improvements:
       - Use existing gRPC status code patterns from tournament service
       - Add context to error messages for debugging
       - Implement proper error wrapping with structured logging
    
    4. Add integration tests:
       - Test complete match submission workflow
       - Test bracket progression through multiple rounds
       - Test tournament completion scenarios
    
    5. Ensure all tests still pass after refactoring:
       - Run full test suite: go test ./pkg/service/
       - Verify test coverage remains high (>90%)
       - Check that no regressions were introduced
  </action>
  <verify>go test ./pkg/service/ passes and coverage report shows >90% for match service</verify>
  <done>Match service refactored with optimized performance, improved error handling, and comprehensive test coverage</done>
</task>

</tasks>

<verification>
- [ ] RED phase: All tests initially fail for unimplemented business logic
- [ ] GREEN phase: All tests pass with minimal implementation
- [ ] REFACTOR phase: All tests still pass after optimization
- [ ] Match result validation prevents invalid winner submissions
- [ ] Winner advancement algorithm follows correct bracket mathematics
- [ ] Authentication patterns match existing tournament service
- [ ] Error handling follows established gRPC status patterns
- [ ] Test coverage exceeds 90% with comprehensive edge cases
</verification>

<success_criteria>
TDD-tested match service ready for tournament integration
- Match result validation with comprehensive participant checking
- Winner advancement algorithm with correct bracket position math
- Authentication and authorization following existing patterns
- Comprehensive test suite covering all business rules and edge cases
- Optimized performance and clean, maintainable code structure
</success_criteria>

<output>
After completion, create `.planning/phases/03-competition/03-competition-03-SUMMARY.md`
</output>