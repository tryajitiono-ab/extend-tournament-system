# Phase 2: Participation - Context

**Gathered:** 2026-01-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Players can register for tournaments and manage their participation. This phase handles registration flow, capacity enforcement, and participant information display within existing tournaments.

</domain>

<decisions>
## Implementation Decisions

### Registration flow behavior
- Reject registrations when tournament is full (no waitlist functionality)
- Immediate confirmation with tournament details after successful registration
- Show "X/Y spots filled" prominently in tournament information
- FIFO handling for race conditions when multiple players attempt to register for last spot

### Withdrawal handling
- No player withdrawals allowed (for simplicity)
- Admin intervention required for any participant removal from tournament
- No automated withdrawal processing or user-initiated withdrawal options

### Participant list visibility
- Public visibility - anyone can view participant lists
- Show only usernames (no personal information beyond identifiers)
- Display participants in registration order (first registered appears first)
- Show "X/Y participants" count at top of participant list

### Registration state management
- Prevent duplicate registrations for same player in same tournament
- Binary state system: either registered or not registered
- No communication or notification system for registration changes
- Basic validation only: capacity enforcement + authentication verification

### Claude's Discretion
- Error message wording and formatting
- Loading states during registration processing
- UI layout and visual design elements
- Implementation details of FIFO race condition handling

</decisions>

<specifics>
## Specific Ideas

- Simple, straightforward registration process similar to basic event signup
- Clear feedback when tournament is full - "Tournament is full (X/Y participants)"
- Admin-only participant removal keeps system simple and reduces edge cases
- Public participant lists help players gauge tournament popularity and competition

</specifics>

<deferred>
## Deferred Ideas

- Waitlist functionality - could be added in future phase for high-demand tournaments
- Player withdrawal system - automated withdrawals could be valuable but add complexity
- Participant notifications/communication - registration confirmations via email
- Advanced participant information display (stats, history, etc.)

</deferred>

---

*Phase: 02-participation*
*Context gathered: 2026-01-27*