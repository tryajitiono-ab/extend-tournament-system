# Feature Research

**Domain:** Tournament Viewing Web UI  
**Researched:** 2025-02-02  
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Tournament list/grid view | Every tournament platform (Challonge, Start.gg, Toornament) has browsable tournament list as entry point | LOW | API already exists: `/v1/public/namespace/{ns}/tournaments` with filtering |
| Tournament status badges | Users need instant recognition of DRAFT/ACTIVE/STARTED/COMPLETED/CANCELLED states | LOW | Visual indicator (color-coded badges) for 5 tournament states |
| Participant count display | "16/32 players registered" is universal - shows availability at a glance | LOW | API provides `current_participants` and `max_participants` |
| Traditional bracket tree visualization | Single-elimination format has standard visual language - users expect nested tree layout | HIGH | Most complex feature - requires layout algorithm for rounds/positions |
| Match status indicators | Visual distinction between SCHEDULED, IN_PROGRESS, COMPLETED matches in bracket | MEDIUM | Color coding + icons for 4 match states |
| Winner declaration | Tournament winner must be prominently displayed on completed tournaments | LOW | API provides winner in final match data |
| Basic filtering (status/date) | Users expect to filter "Active tournaments" or "Upcoming" without scrolling everything | MEDIUM | API supports status and date range filtering |
| Mobile-responsive layout | 60%+ of gaming community views on mobile - bracket must scale/scroll properly | HIGH | Horizontal scroll for brackets, vertical stack for tournament list |
| Click-through navigation | List → Detail → Bracket is standard flow, each element should be clickable | LOW | Standard routing pattern |
| Round indicators | "Round 1", "Quarter-Finals", "Semi-Finals", "Finals" labels on bracket columns | MEDIUM | Calculate from total rounds: Round 1, Round 2, ..., Semi-Finals, Finals |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Bracket zoom/pan controls | Large tournaments (32-256 players) are hard to view - zoom makes navigation easier | MEDIUM | CSS transform + mouse wheel handlers or pinch gestures |
| Match detail popups/modals | Click match to see detailed info without leaving bracket view - reduces navigation friction | LOW | Modal overlay with participant names, timestamps, winner |
| Search/filter by tournament name | Reduces cognitive load when browsing 100+ tournaments | LOW | Client-side filtering initially, could be server-side later |
| "Live" tournament indicator | Highlight active tournaments with ongoing matches - creates engagement | LOW | Badge/animation for STARTED status tournaments |
| Round-by-round view toggle | Option to view single round at a time vs full bracket - helps focus on current round | MEDIUM | Toggle between "Full Bracket" and "Round X" view modes |
| Participant list with avatars | Shows all registered players - adds personality and helps identify friends | LOW | API provides participant list, integrate placeholder avatars initially |
| Tournament timeline | Visual timeline showing registration period → start → completion | LOW | Timeline component with dates from API (start_time, end_time) |
| Shareable tournament URLs | Direct links to specific tournaments - enables community sharing | LOW | `/tournaments/{id}` URL pattern with proper meta tags |
| Empty state illustrations | Friendly "No tournaments yet" instead of blank page - improves perceived quality | LOW | SVG illustration + helpful text for zero-state |
| Compact vs detailed list view | Toggle between card view (more info) and compact list (more density) | LOW | Two CSS layouts, user preference toggle |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Real-time auto-refresh | "I want to see updates instantly without refreshing" | Adds WebSocket complexity, server load, connection management - out of scope for view-only v1.1 | Manual refresh button + "Updated X seconds ago" timestamp - sets expectation |
| User registration through UI | "Let me register directly from the web page" | Requires authentication, session management, error handling - explicitly deferred to v1.2+ | Show "Registration available via API" message or link to API docs |
| Admin controls in UI | "I should be able to manage tournaments from web" | Security concerns, permission management, audit logging - explicitly deferred to v1.2+ | Admin operations remain API-only in v1.1 |
| Bracket editing/drag-drop | "Let me rearrange the bracket manually" | Breaks bracket integrity, requires complex validation - not in specification | Display read-only bracket generated by API |
| Tournament creation form | "Add a 'Create Tournament' button" | Requires full admin workflow, validation - deferred to future admin dashboard | API-only creation in v1.1 |
| Comments/chat on matches | "Let players discuss matches" | Requires moderation, storage, real-time sync - adds social features out of scope | Focus on viewing, defer social features to v2.0+ |
| Export bracket to image | "I want to download bracket as PNG" | Adds server-side rendering or canvas complexity - nice to have but not essential | Screenshot functionality sufficient for v1.1 |
| Email notifications | "Notify me when my match starts" | Requires email service, user preferences, opt-in/out - beyond view-only scope | External notification systems can poll API |

## Feature Dependencies

```
[Bracket Visualization]
    └──requires──> [Match Data API]
                       └──requires──> [Tournament Started]

[Tournament Detail Page]
    └──requires──> [Tournament List Page]

[Match Status Indicators] ──enhances──> [Bracket Visualization]

[Participant List] ──enhances──> [Tournament Detail Page]

[Round Labels] ──enhances──> [Bracket Visualization]

[Mobile Responsive] ──conflicts──> [Full Bracket on Small Screen]
                                    (requires horizontal scroll or round-by-round)

[Real-time Updates] ──conflicts──> [View-Only Scope]
                                    (requires WebSocket, deferred to v2.0)
```

### Dependency Notes

- **Bracket Visualization requires Match Data:** Cannot display bracket until tournament is STARTED and matches exist (API returns matches by round)
- **Tournament Detail requires List Page:** Users need entry point to discover tournaments before viewing details
- **Match Status enhances Bracket:** Without status indicators, bracket is harder to understand (completed vs pending)
- **Participant List enhances Detail Page:** Shows who's playing, adds context to tournament
- **Mobile Responsive conflicts with Full Bracket:** Cannot fit 64-player bracket on mobile without scroll or alternate view
- **Real-time conflicts with View-Only:** Live updates require bi-directional communication, adds complexity beyond v1.1 scope

## MVP Definition

### Launch With (v1.1)

Minimum viable product — what's needed to validate the concept.

- [x] **Tournament List Page** — Primary entry point, browse all tournaments with status/participant count
- [x] **Tournament Detail Page** — View tournament info, dates, status, participant count
- [x] **Participant List Display** — Show registered players for a tournament
- [x] **Single-Elimination Bracket Tree** — Traditional nested bracket visualization with rounds/positions
- [x] **Match Status Indicators** — Visual distinction for SCHEDULED/IN_PROGRESS/COMPLETED/CANCELLED
- [x] **Round Labels** — "Round 1", "Semi-Finals", "Finals" headers on bracket columns
- [x] **Tournament Status Badges** — Color-coded DRAFT/ACTIVE/STARTED/COMPLETED/CANCELLED indicators
- [x] **Mobile-Responsive Layout** — Bracket scrolls horizontally, list stacks vertically on small screens
- [x] **Basic Error States** — "Tournament not found", "No tournaments available"
- [x] **Static File Serving** — HTML/CSS/JS served from Go service at `/` or `/ui/` path

### Add After Validation (v1.2+)

Features to add once core is working.

- [ ] **User Registration UI** — Let users register for tournaments through web interface (requires auth)
- [ ] **Search/Filter Tournaments** — Client-side search by name, filter by multiple criteria
- [ ] **Match Detail Modals** — Click match to view detailed info in popup
- [ ] **Bracket Zoom Controls** — Zoom in/out for large tournaments (32+ players)
- [ ] **Round-by-Round Toggle** — View single round at a time for mobile
- [ ] **Shareable URLs with Meta Tags** — Open Graph tags for social sharing
- [ ] **"Updated X ago" Timestamps** — Show data freshness without auto-refresh
- [ ] **Manual Refresh Button** — Let users explicitly reload tournament data

### Future Consideration (v2.0+)

Features to defer until product-market fit is established.

- [ ] **Real-time Updates via WebSocket** — Auto-refresh match results without polling (requires backend changes)
- [ ] **Admin Dashboard UI** — Web-based tournament management (create, start, cancel)
- [ ] **Player Dashboard** — View "My Tournaments", match history, statistics
- [ ] **Match Chat/Comments** — Social features for match discussion
- [ ] **Tournament Templates** — Save and reuse tournament configurations
- [ ] **Multi-format Support** — Double-elimination, round-robin visualization (requires API changes)
- [ ] **Analytics Dashboard** — Tournament participation trends, popular times
- [ ] **Bracket Export** — Download as PNG/PDF for sharing

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Tournament List Page | HIGH | LOW | P1 |
| Tournament Detail Page | HIGH | LOW | P1 |
| Bracket Tree Visualization | HIGH | HIGH | P1 |
| Match Status Indicators | HIGH | MEDIUM | P1 |
| Mobile-Responsive Layout | HIGH | HIGH | P1 |
| Tournament Status Badges | HIGH | LOW | P1 |
| Round Labels | HIGH | MEDIUM | P1 |
| Participant List Display | MEDIUM | LOW | P1 |
| Basic Error States | HIGH | LOW | P1 |
| Static File Serving | HIGH | MEDIUM | P1 |
| Search/Filter Tournaments | MEDIUM | LOW | P2 |
| Match Detail Modals | MEDIUM | LOW | P2 |
| Manual Refresh Button | MEDIUM | LOW | P2 |
| Bracket Zoom Controls | MEDIUM | MEDIUM | P2 |
| Round-by-Round Toggle | MEDIUM | MEDIUM | P2 |
| Shareable URLs | LOW | LOW | P2 |
| "Updated X ago" | LOW | LOW | P2 |
| User Registration UI | HIGH | HIGH | P2 (deferred to v1.2) |
| Real-time Updates | HIGH | HIGH | P3 (requires WebSocket) |
| Admin Dashboard | HIGH | HIGH | P3 (separate milestone) |
| Player Dashboard | MEDIUM | HIGH | P3 |
| Match Chat | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for v1.1 launch (view-only capabilities)
- P2: Should have, add when possible (enhances view-only experience)
- P3: Nice to have, future consideration (requires auth/admin/backend changes)

## Competitor Feature Analysis

| Feature | Challonge | Start.gg | Toornament | Our Approach (v1.1) |
|---------|-----------|----------|------------|---------------------|
| Tournament List | Grid cards with search | Game-specific browse | Filter by game/status | Simple list/cards with status filter |
| Bracket Display | Horizontal single-elim tree | Horizontal + pool groups | Multiple formats | Horizontal single-elim only |
| Mobile View | Responsive with horiz scroll | App + responsive web | Touch-optimized | Horizontal scroll + vertical stack |
| Match Details | Inline expansion | Modal popups | Dedicated match page | P2 - Modal popups planned |
| Real-time Updates | Auto-refresh every 30s | WebSocket live | Polling | Manual refresh (P2 button) |
| User Actions | Register, report, admin | Full tournament mgmt | Full platform | View-only (actions in v1.2+) |
| Participant Display | List with avatars | Profile integration | Team rosters | Simple list with names |
| Tournament Status | Icon badges | Status labels | Colored indicators | Color-coded badges |
| Round Navigation | Click round headers | Scroll bracket | Round selector | Labels only (round toggle P2) |
| Empty States | Generic messages | Game-themed graphics | Clear CTAs | Friendly illustrations |

**Key Observations:**
- **Challonge**: Simple, focused, fast - good model for v1.1 MVP
- **Start.gg**: Feature-rich but complex - many features out of scope for view-only
- **Toornament**: Professional but requires account - we're more accessible with view-only
- **Our Differentiation**: View-only first, API-powered, focused on single-elimination, mobile-friendly

## Bracket Visualization Specifics

### Layout Requirements

**Horizontal Tree Structure:**
- Rounds progress left-to-right (Round 1 → Finals)
- Each round is a vertical column of matches
- Matches connect with lines showing advancement path
- Winner flows to right into next round match

**Visual Elements:**
- **Match Box**: Rectangle containing participant1 vs participant2
- **Connector Lines**: Lines from match to next round (winner advancement)
- **Round Labels**: Headers above each column ("Round 1", "Quarter-Finals", etc.)
- **Status Indicators**: Border color or badge on match box (SCHEDULED=gray, IN_PROGRESS=yellow, COMPLETED=green)
- **Winner Highlight**: Bold or highlighted participant name in completed matches

**Spacing & Sizing:**
- Match box: ~200px wide, ~60px tall (accommodate 2 participant names)
- Vertical spacing: ~20px between matches in same round
- Horizontal spacing: ~100px between rounds (room for connector lines)
- Round header: ~30px tall

**Responsive Behavior:**
- Desktop (>1024px): Full bracket visible, may scroll horizontally for 32+ players
- Tablet (768-1024px): Horizontal scroll required for 16+ players
- Mobile (<768px): Definitely horizontal scroll, reduce match box width to ~150px

### Algorithm Notes

**Round Calculation:**
- Total rounds = log2(participants) rounded up
- Round names: First round is "Round 1", last is "Finals", second-to-last is "Semi-Finals", third-to-last is "Quarter-Finals"
- Example: 16 players = 4 rounds (Round 1, Quarter-Finals, Semi-Finals, Finals)

**Match Positioning:**
- API provides `round` (1-indexed) and `position` (0-indexed within round)
- Round 1 has most matches, each subsequent round has half as many
- Position determines vertical placement within column

**Connector Lines:**
- Each match connects to match in next round at position = floor(current_position / 2)
- Lines drawn with SVG or CSS borders
- Only draw line if winner exists or match is scheduled

**Bye Handling:**
- Byes appear as single-participant matches (participant2 = null)
- Show "BYE" text in empty slot
- Automatically mark as completed with single participant as winner

### Mobile Considerations

**Primary Approach: Horizontal Scroll**
- Container with `overflow-x: auto` and `overflow-y: hidden`
- Fixed height viewport, bracket scrolls left-right
- Touch-friendly scroll (native mobile behavior)

**Alternative Approach (P2): Round-by-Round View**
- Show single round at a time
- "Previous Round" / "Next Round" buttons
- Good for very small screens (< 375px width)

**Touch Interactions:**
- Tap match to view details (modal or inline expansion)
- Swipe left/right to scroll bracket
- Pinch to zoom (P2 feature with zoom controls)

## Sources

- **Tournament Platforms Analyzed:**
  - Challonge.com (industry standard for simple brackets)
  - Start.gg (formerly Smash.gg - gaming tournaments)
  - Toornament.com (professional esports)
  
- **API Capabilities (from service.proto):**
  - `/v1/public/namespace/{ns}/tournaments` - List tournaments (status, date filtering)
  - `/v1/public/namespace/{ns}/tournaments/{id}` - Tournament details
  - `/v1/public/namespace/{ns}/tournaments/{id}/matches` - Match list by round
  - `/v1/public/namespace/{ns}/tournaments/{id}/participants` - Participant list

- **Existing System Constraints:**
  - Single-elimination format only (no double-elimination, round-robin, Swiss)
  - 5 tournament states: DRAFT, ACTIVE, STARTED, COMPLETED, CANCELLED
  - 4 match states: SCHEDULED, IN_PROGRESS, COMPLETED, CANCELLED
  - REST API only (no WebSocket, polling required)

- **Mobile Usage Patterns:**
  - Gaming community has high mobile usage (60%+ based on industry data)
  - Horizontal scroll is standard for tournament brackets on mobile
  - Start.gg uses horizontal scroll + round selector for mobile optimization

- **UX Best Practices:**
  - Status badges with color coding (green=active, blue=started, gray=draft, red=cancelled)
  - Empty states with illustrations improve perceived quality
  - Manual refresh over auto-refresh sets user expectations for view-only
  - Clickable elements should be >44px for touch targets on mobile

---
*Feature research for: Tournament Viewing Web UI (v1.1)*  
*Researched: 2025-02-02*  
*Scope: VIEW-ONLY capabilities, no user actions (registration, admin) until v1.2+*
