# Phase 4: Core UI & API Integration - Context

**Gathered:** 2026-02-01
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete tournament viewing UI with list page, detail page, and live data from REST API. This phase adds static HTML/CSS/JS files to the existing Go service. View-only interface - no user registration, no admin controls, no real-time updates.

</domain>

<decisions>
## Implementation Decisions

### Visual Presentation
- **Card style:** Minimal cards — just essential info (name, status, participant count), very clean
- **Status indicators:** Plain text labels, optionally add color if feeling adventurous (not required)
- **Information hierarchy:** Name first (tournament name most prominent), status secondary
- **Timestamps:** Absolute format only (relative timestamps deferred to future improvement)

### Data Loading Strategy
- **Initial load:** Fetch on page load — tournament list loads immediately when user visits `/tournaments`
- **Refresh behavior:** Manual only — refresh button, no auto-refresh polling
- **Caching:** No caching — fresh fetch every time (simple, always current)
- **Loading states:** Claude's discretion — as barebones as possible

### Error and Empty States
- **API failure:** Error banner at top with "Failed to load. [Retry]" button
- **No tournaments:** Minimal message — "No tournaments" (very clean)
- **Slow network:** Patient waiting — just show loading state until response arrives
- **Retry strategy:** Manual retry only — user clicks "Retry" button

### Mobile vs Desktop Experience
- **Layout approach:** Simple responsive — same layout works on both desktop and mobile, not broken but not obsessing over mobile perfection
- **Card sizing:** Fixed height — all cards same height, content truncates if needed
- **Touch targets:** Tournament name is clickable link (standard link behavior)
- **Breakpoints:** Claude's discretion — keep it stupid and simple

### Claude's Discretion
- Loading state implementation (barebones as possible)
- Breakpoint strategy (keep simple)
- Exact spacing and typography
- Color choices if adding color to status indicators
- Card truncation behavior

</decisions>

<specifics>
## Specific Ideas

- "Minimal cards" — User emphasized clean, essential-only information
- "Barebones as possible" — Loading states should be extremely simple
- "Keep it stupid and simple" — Complexity reduction is a priority
- Desktop-first mindset with basic mobile compatibility

</specifics>

<deferred>
## Deferred Ideas

- Relative timestamps ("2 hours ago") — future improvement
- Mobile-specific optimizations — basic responsive is enough for v1.1
- Auto-refresh polling — manual refresh only for now
- Client-side caching — start simple with no cache

</deferred>

---

*Phase: 04-core-ui-api-integration*
*Context gathered: 2026-02-01*
