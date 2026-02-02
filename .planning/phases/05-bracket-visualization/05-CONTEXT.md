# Phase 5: Bracket Visualization - Context

**Gathered:** 2026-02-02
**Status:** Ready for planning

<domain>
## Phase Boundary

Traditional bracket tree visualization showing single-elimination tournament progression. Users can view match status, round organization, and winner paths. This is a view-only component displaying match data from the existing REST API.

</domain>

<decisions>
## Implementation Decisions

### Bracket Layout Structure
- **Orientation:** Left-to-right flow (Round 1 on left, Finals on right) — traditional tournament bracket
- **Round labels:** Numeric format (R1, R2, R3...) for simplicity
- **Connector lines:** Claude's discretion — choose cleanest visual approach
- **Spacing/density:** Claude's discretion — balance readability with screen real estate

### Match Status Visualization
- **Color coding:** Gray for scheduled, Blue for in-progress, Green for completed
- **Winner highlighting:** Bold the winner's name in completed matches
- **Status labels:** Show explicit status text on each match card ('Scheduled', 'Live', 'Completed')
- **Bye matches:** Show as explicit bye cards with 'BYE' label (don't hide them)

### Match Card Content
- **Player identification:** Username with fallback to user_id if username not available
- **Match information:** Names only — minimal info, no timestamps or position numbers
- **Empty slots:** Display 'TBD' label for scheduled matches with unknown players
- **Scores:** Don't show scores, just winner indication through bold name

### Mobile Responsiveness
- **Mobile strategy:** Use whatever brackets-viewer.js library provides (delegate to library capabilities)
- **Match card sizing:** Claude's discretion — balance tap targets with screen space
- **Touch interactions:** View-only, no tap interactions or zoom
- **Large tournaments:** Show desktop recommendation message for 32+ player tournaments on mobile

### Claude's Discretion
- Connector line style (straight vs bracket-style curves)
- Bracket density and spacing calculations
- Match card sizing on different screen widths
- Exact implementation of brackets-viewer.js integration
- Error handling when bracket data is incomplete

</decisions>

<specifics>
## Specific Ideas

- Use brackets-viewer.js (1.9.0+) library as mentioned in ROADMAP-v1.1.md — production-ready, 213+ stars
- Mobile experience leverages library's built-in responsive handling
- Desktop recommendation for large tournaments acknowledges UI complexity without blocking functionality

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope. Enhanced features like zoom/pan controls, match detail popups, and bracket export already deferred to v1.2+ in roadmap.

</deferred>

---

*Phase: 05-bracket-visualization*
*Context gathered: 2026-02-02*
