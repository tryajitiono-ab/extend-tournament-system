---
phase: 03-competition
plan: 04
type: execute
wave: 1
depends_on: [03-competition-02]
files_modified: [pkg/service/match.go, pkg/service/tournament.go]
autonomous: true
gap_closure: true

must_haves:
  truths:
    - "Winner advancement algorithm correctly progresses tournament brackets"
    - "Tournament status transitions from in_progress to completed"
    - "Bye participants automatically advance without result submission"
    - "System declares tournament winner upon final match completion"
  artifacts:
    - path: "pkg/service/match.go"
      provides: "Complete match service with automated tournament progression"
      contains: "advanceWinner implementation with storage updates"
      min_lines: 400
    - path: "pkg/service/tournament.go"
      provides: "Enhanced tournament service with completion detection"
      contains: "CheckTournamentCompletion, CompleteTournament methods"
      min_lines: 20
  key_links:
    - from: "advanceWinner"
      to: "MatchStorage.UpdateMatch"
      via: "find and update next round match with advancing participant"
      pattern: "GetMatchesByRound.*round.*UpdateMatch"
    - from: "SubmitMatchResult"
      to: "CheckTournamentCompletion"
      via: "check if tournament is finished after each match result"
      pattern: "SubmitMatchResult.*CheckTournamentCompletion"
    - from: "CheckTournamentCompletion"
      to: "TournamentStorage.UpdateTournament"
      via: "transition tournament status to COMPLETED and declare winner"
      pattern: "UpdateTournament.*status.*COMPLETED"
---

<objective>
Complete tournament automation logic for winner advancement, completion detection, and bye handling

Purpose: Implement the core automation that makes tournaments "run with automated match management and result tracking" by fixing the TODOs and missing logic that currently prevents bracket progression
Output: Working tournament progression where winners automatically advance, tournaments complete when all matches finish, and bye participants are handled correctly
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
  <name>Implement winner advancement logic in advanceWinner function</name>
  <files>pkg/service/match.go</files>
  <action>
    Replace the TODO comment in advanceWinner with actual implementation:
    
    1. Calculate next round (currentRound + 1) and next position (currentPosition / 2)
    2. Query MatchStorage.GetMatchesByRound for next round matches
    3. Find the match where the advancing participant should be placed
    4. Update that match's participant1 or participant2 field with the winner
    5. Use MatchStorage.UpdateMatch to persist the advancement
    6. Handle edge cases: final match (no next round), odd participant counts
    
    Reference existing code: From 03-competition-02-SUMMARY, MatchStorage has GetMatchesByRound and UpdateMatch methods
    Gap reason: Current advanceWinner only logs advancement, doesn't update storage (from VERIFICATION.md gap 1)
  </action>
  <verify>Review pkg/service/match.go advanceWinner function - should contain storage calls, not just log statements</verify>
  <done>Winner from completed match automatically appears in next round match participant list</done>
</task>

<task type="auto">
  <name>Add tournament completion detection and winner declaration</name>
  <files>pkg/service/match.go, pkg/service/tournament.go</files>
  <action>
    1. In pkg/service/match.go, create CheckTournamentCompletion function:
       - Query all matches for the tournament using MatchStorage.GetTournamentMatches
       - Check if all matches have status COMPLETED or CANCELLED
       - Return completion status and winner (from final match)
    
    2. In pkg/service/tournament.go, create CompleteTournament method:
       - Call CheckTournamentCompletion to verify tournament is finished
       - Update tournament status to COMPLETED using TournamentStorage.UpdateTournament
       - Set tournament winner field
       - Log tournament completion with winner declaration
       - Return appropriate gRPC response
    
    3. Modify SubmitMatchResult in match.go to call CheckTournamentCompletion after each successful result submission
    - If tournament is complete, automatically call CompleteTournament
    
    Reference existing code: From 03-competition-02-SUMMARY, both MatchStorage and TournamentStorage are available
    Gap reason: No tournament completion logic exists when all matches finish (from VERIFICATION.md gap 3)
  </action>
  <verify>Review code for CheckTournamentCompletion and CompleteTournament functions - should handle status transitions and winner declaration</verify>
  <done>Tournament status automatically changes to COMPLETED when all matches finish and winner is declared</done>
</task>

<task type="auto">
  <name>Implement automatic bye participant advancement</name>
  <files>pkg/service/match.go</files>
  <action>
    1. Create HandleByeAdvancement function:
       - Query matches for current round using MatchStorage.GetMatchesByRound
       - Identify matches with only one participant (bye situation)
       - Automatically advance the single participant to next round
       - Mark the bye match as COMPLETED with the solo participant as winner
    
    2. Integrate bye handling into bracket generation:
       - When tournament starts, call HandleByeAdvancement for round 1
       - This ensures participants with byes are automatically placed in round 2
    
    3. Add logic in SubmitMatchResult to check for and handle byes in subsequent rounds
    
    Reference existing code: From 03-competition-02-SUMMARY, bracket generation exists but bye handling is incomplete
    Gap reason: No automatic advancement logic for bye participants (from VERIFICATION.md gap 4)
  </action>
  <verify>Review HandleByeAdvancement function - should detect solo participants and advance them automatically</verify>
  <done>Participants with byes automatically appear in next round without manual result submission</done>
</task>

</tasks>

<verification>
After completing all tasks, verify the tournament automation workflow:

1. Create a test tournament with odd participants (e.g., 5 participants)
2. Start tournament to generate brackets
3. Verify bye participants are automatically advanced to round 2
4. Submit match results for round 1
5. Verify winners automatically appear in round 2 matches
6. Submit final match result
7. Verify tournament status changes to COMPLETED and winner is declared
</verification>

<success_criteria>
- Winner advancement algorithm correctly progresses tournament brackets (addresses gap 1)
- Tournament status transitions from STARTED to COMPLETED when all matches finish (addresses gap 3)
- Bye participants automatically advance without result submission (addresses gap 5)
- Tournament winner is properly declared upon completion
- All automation logic integrates with existing storage and service patterns
</success_criteria>

<output>
After completion, create `.planning/phases/03-competition/03-competition-04-SUMMARY.md`
</output>