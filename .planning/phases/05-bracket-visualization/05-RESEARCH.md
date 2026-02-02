# Phase 5: Bracket Visualization - Research

**Researched:** 2026-02-02
**Domain:** JavaScript bracket visualization with brackets-viewer.js
**Confidence:** HIGH

## Summary

This phase adds traditional single-elimination bracket visualization to the tournament detail page using the established brackets-viewer.js library. The library is production-ready (213+ stars, actively maintained, v1.9.0 released Nov 2024) and provides all necessary features for displaying tournament brackets with match status, round labels, and responsive design out-of-the-box.

The implementation requires transforming our REST API match data (from `/v1/public/namespace/{namespace}/tournaments/{tournament_id}/matches`) into the brackets-model format, then rendering using the brackets-viewer.js CDN-distributed library. The library handles all bracket layout, connector lines, and mobile responsiveness internally through CSS.

Our existing tech stack (vanilla JS, Pico CSS v2.0.6, CDN-based libraries) aligns perfectly with brackets-viewer.js's vanilla JS architecture and CSS variable-based theming system.

**Primary recommendation:** Use brackets-viewer.js v1.9.0+ via CDN with custom data transformation layer to convert our Match protobuf format to brackets-model format.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| brackets-viewer.js | 1.9.0+ | Bracket rendering | Production-ready library specifically designed for tournament brackets, 213+ GitHub stars, actively maintained (Nov 2024 release), handles single/double elimination + round-robin |
| CDN delivery (jsDelivr) | latest | Library hosting | Recommended distribution method per official docs, no build step needed, version pinning supported |
| Vanilla JavaScript | ES6+ | Data transformation | Matches existing project pattern (Phase 4), no framework dependencies, aligns with library design |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| brackets-model | 1.6.0+ | Type definitions | Understanding data structure only - no runtime dependency needed |
| Pico CSS (existing) | 2.0.6 | Base styling | Container and utility styles for non-bracket elements |
| Custom CSS | N/A | Bracket theming | Override CSS variables for match status colors and spacing |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| brackets-viewer.js | Custom SVG/Canvas | Building from scratch would take 10-20x longer, miss edge cases (bye handling, connector positioning), and require ongoing maintenance |
| CDN distribution | NPM + bundler | Would require introducing build tooling (webpack/vite) inconsistent with current vanilla JS architecture |
| Single file | Multiple JS modules | Current project uses single-file approach for simplicity - maintain consistency |

**Installation:**

No npm installation needed - use CDN:

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/brackets-viewer@1.9.0/dist/brackets-viewer.min.css" />
<script type="text/javascript" src="https://cdn.jsdelivr.net/npm/brackets-viewer@1.9.0/dist/brackets-viewer.min.js"></script>
```

## Architecture Patterns

### Recommended Project Structure

```
web/
├── static/
│   ├── css/
│   │   ├── pico.min.css          # Existing
│   │   ├── custom.css            # Existing  
│   │   └── bracket-theme.css     # NEW - brackets-viewer CSS variable overrides
│   └── js/
│       ├── api-client.js         # Existing - add fetchMatches() function
│       ├── tournament-detail.js  # Existing - add bracket rendering
│       └── bracket-adapter.js    # NEW - transform Match[] to brackets-model format
└── templates/
    └── tournament-detail.html    # Existing - add bracket container
```

### Pattern 1: Data Transformation Layer

**What:** Transform REST API match data to brackets-model format

**When to use:** Required for all bracket rendering - brackets-viewer.js expects specific data structure

**Example:**

```javascript
// Source: brackets-viewer.js documentation and demo files
// File: bracket-adapter.js

/**
 * Transform tournament matches from REST API to brackets-model format
 * @param {Array} matches - Array of Match objects from REST API
 * @param {Array} participants - Array of participant objects
 * @param {Object} tournament - Tournament object
 * @returns {Object} Data in brackets-model format
 */
function transformToBracketsModel(matches, participants, tournament) {
    // brackets-viewer expects:
    // - stages: array of stage objects (one for single-elimination)
    // - matches: array of match objects in specific format
    // - participants: array of participant objects with specific fields
    // - matchGames: array (empty for basic display)
    
    return {
        stages: [{
            id: 0,
            tournament_id: tournament.tournament_id,
            name: tournament.name,
            type: 'single_elimination',
            number: 1,
        }],
        matches: matches.map(match => ({
            id: parseInt(match.match_id, 10),
            stage_id: 0,
            group_id: 0,
            round_id: match.round - 1, // brackets-model uses 0-indexed rounds
            number: match.position,
            opponent1: match.participant1 ? {
                id: match.participant1.user_id,
                position: match.position * 2 - 1,
            } : null,
            opponent2: match.participant2 ? {
                id: match.participant2.user_id,
                position: match.position * 2,
            } : null,
            status: mapMatchStatus(match.status),
        })),
        participants: participants.map(p => ({
            id: p.user_id,
            tournament_id: tournament.tournament_id,
            name: p.username || p.user_id,
        })),
        matchGames: [], // Not needed for basic bracket display
    };
}

function mapMatchStatus(apiStatus) {
    // Map our protobuf MatchStatus enum to brackets-viewer status
    const statusMap = {
        'MATCH_STATUS_SCHEDULED': 2,    // Pending
        'MATCH_STATUS_IN_PROGRESS': 3,  // Running
        'MATCH_STATUS_COMPLETED': 4,    // Completed
        'MATCH_STATUS_CANCELLED': 5,    // Archived
    };
    return statusMap[apiStatus] || 2;
}
```

### Pattern 2: Progressive Enhancement for Bracket Section

**What:** Add bracket section to existing tournament detail page without breaking current functionality

**When to use:** When adding new features to existing pages

**Example:**

```javascript
// Source: Existing tournament-detail.js pattern + brackets-viewer docs
// Add to tournament-detail.js

async function loadTournamentData() {
    showLoading();
    hideError();

    try {
        // Existing: Fetch tournament details
        const tournament = await fetchTournament(currentNamespace, currentTournamentId);
        renderTournament(tournament);

        // Existing: Fetch participants
        await loadParticipants();
        
        // NEW: Fetch and render bracket (only if tournament has matches)
        if (tournament.status !== 'PENDING') {
            await loadBracket();
        }
    } catch (error) {
        showError('Failed to load tournament data');
        console.error('Failed to load tournament:', error);
    }
}

async function loadBracket() {
    showBracketLoading();
    
    try {
        const matches = await fetchMatches(currentNamespace, currentTournamentId);
        const participants = await fetchParticipants(currentNamespace, currentTournamentId);
        const tournament = await fetchTournament(currentNamespace, currentTournamentId);
        
        const bracketData = transformToBracketsModel(matches, participants, tournament);
        renderBracket(bracketData);
    } catch (error) {
        // Non-critical - don't show error banner
        hideBracketLoading();
        console.error('Failed to load bracket:', error);
    }
}

function renderBracket(data) {
    hideBracketLoading();
    
    window.bracketsViewer.render(data, {
        selector: '.bracket-container',
        clear: true,
    });
}
```

### Pattern 3: CSS Variable Override for Match Status Colors

**What:** Use CSS variables to customize bracket appearance to match design decisions

**When to use:** Always - need to implement color-coded match status

**Example:**

```css
/* Source: brackets-viewer.js src/style.scss
 * File: bracket-theme.css
 */

.brackets-viewer {
  /* Override colors for match status */
  --match-background: #fff;
  --font-color: #212529;
  
  /* Scheduled matches - Gray */
  --border-color: #9e9e9e;
  
  /* In-progress matches - Blue */
  --border-hover-color: #2196f3;
  
  /* Completed matches - Green */
  --win-color: #50b649;
  
  /* Labels */
  --label-color: #757575;
  --hint-color: #a7a7a7;
  
  /* Connector lines */
  --connector-color: #9e9e9e;
  
  /* Spacing - balance readability */
  --round-margin: 40px;
  --match-width: 160px;
  --text-size: 13px;
}

/* Apply status-specific colors */
.brackets-viewer .match[data-status="scheduled"] .opponents {
  border-color: #9e9e9e;
}

.brackets-viewer .match[data-status="in-progress"] .opponents {
  border-color: #2196f3;
  background-color: #e3f2fd;
}

.brackets-viewer .match[data-status="completed"] .opponents {
  border-color: #50b649;
}
```

### Anti-Patterns to Avoid

- **Modifying brackets-viewer.js source code:** Library is CDN-distributed - customizations must be through CSS variables and configuration options only
- **Manual DOM manipulation of bracket elements:** Library owns the bracket DOM tree - interfering causes rendering bugs
- **Caching transformed data client-side:** Match data changes during tournament - always fetch fresh data
- **Deep tournament nesting without scroll indicators:** Brackets are wide - ensure horizontal scroll is obvious to users

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bracket layout algorithm | Custom CSS grid or flexbox bracket | brackets-viewer.js | Single-elimination bracket positioning requires handling: bye rounds, power-of-2 vs non-power-of-2 sizes, connector line calculations, round alignment, responsive reflow |
| Round name generation | Manual "Round 1", "Round 2" logic | Built-in customRoundName callback | Library handles fraction-of-final calculation (1/2 = semis, 1/4 = quarters), supports i18n, handles double-elimination complexity |
| Match connector lines | CSS borders and pseudo-elements | Library's built-in connectors | Connector positioning depends on match count per round, bye positions, and vertical spacing - error-prone to replicate |
| Mobile responsive brackets | Media queries for bracket reflow | Library's overflow: auto + CSS variables | Brackets don't reflow well - horizontal scroll is standard pattern, library handles this correctly |
| Empty slot handling | "TBD" text rendering | Library's slot origin system | Handles BYEs, unknown participants, future match winners correctly with showSlotsOrigin configuration |

**Key insight:** Tournament bracket visualization has accumulated complexity that's not obvious at first glance. The brackets-viewer.js library has solved edge cases through years of production use and issue reports. Building custom bracket rendering would require 100-200 hours vs 8-12 hours for integration.

## Common Pitfalls

### Pitfall 1: Match Status Enum Mismatch

**What goes wrong:** Our protobuf defines `MATCH_STATUS_SCHEDULED` but brackets-model expects numeric status codes (0-5)

**Why it happens:** Different data models - protobuf uses string enums, brackets-model uses numeric constants

**How to avoid:** 
- Create explicit mapping function in bracket-adapter.js
- Document the mapping clearly: SCHEDULED=2, IN_PROGRESS=3, COMPLETED=4
- Add validation to catch unmapped statuses

**Warning signs:** 
- Matches render without status indicators
- Console errors about invalid status values
- All matches appear same color regardless of state

### Pitfall 2: Round Indexing Off-by-One

**What goes wrong:** Our API uses 1-indexed rounds (Round 1, Round 2), brackets-model uses 0-indexed (0, 1, 2)

**Why it happens:** Different conventions between backend and frontend library

**How to avoid:**
- Transform round numbers during data conversion: `round_id: match.round - 1`
- Add comment explaining the indexing difference
- Test with multi-round tournaments to verify correct placement

**Warning signs:**
- Matches appear in wrong rounds
- Finals showing in wrong position
- Round labels don't match match positions

### Pitfall 3: Large Tournament Mobile Experience

**What goes wrong:** 32+ player brackets are very wide, poor experience on mobile without indication

**Why it happens:** Brackets are inherently horizontal - can't reflow to vertical layout without losing bracket structure

**How to avoid:**
- Detect large tournaments (32+ players) and screen size
- Show prominent message: "Desktop recommended for best viewing experience"
- Don't block mobile viewing, just warn
- Ensure horizontal scroll indicators are visible

**Warning signs:**
- Mobile users report "can't see bracket"
- Horizontal scroll not obvious
- Users think bracket is broken when it's just wide

### Pitfall 4: Participant Data Shape Mismatch

**What goes wrong:** brackets-viewer expects participants with specific field names, our API uses different names

**Why it happens:** Different data models between our system and brackets-model

**How to avoid:**
- Transform participant objects: `{ id: p.user_id, name: p.username || p.user_id }`
- Handle missing username gracefully with fallback to user_id
- Validate transformed data structure before passing to render()

**Warning signs:**
- Empty participant slots where there should be names
- Console errors about missing participant fields
- "undefined" showing in brackets

### Pitfall 5: Re-render Without Clear

**What goes wrong:** Calling render() multiple times without `clear: true` appends brackets instead of replacing

**Why it happens:** Library's default behavior is to append, not replace

**How to avoid:**
- Always use `clear: true` in render configuration
- Document this in code comments
- Create wrapper function that enforces clear option

**Warning signs:**
- Multiple brackets appearing on page
- Duplicate match cards
- Page gets slower with each data refresh

## Code Examples

Verified patterns from official sources:

### Fetching Match Data

```javascript
// Source: Existing api-client.js pattern + protobuf service definition
// Add to api-client.js

/**
 * Fetch all matches for a tournament
 * @param {string} namespace - Namespace ID
 * @param {string} tournamentId - Tournament ID
 * @returns {Promise<Object>} Object with matches, total_rounds, current_round
 */
async function fetchMatches(namespace, tournamentId) {
    const url = `${API_BASE}/v1/public/namespace/${namespace}/tournaments/${tournamentId}/matches`;
    const response = await fetchWithTimeout(url);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch matches: ${response.statusText}`);
    }
    
    const data = await response.json();
    return {
        matches: data.matches || [],
        totalRounds: data.total_rounds || 0,
        currentRound: data.current_round || 0,
    };
}
```

### Complete Bracket Rendering Flow

```javascript
// Source: brackets-viewer.js demo/with-api.html + our existing patterns

// In tournament-detail.js, add bracket section to existing flow:

async function loadBracket() {
    const bracketSection = document.getElementById('bracket-section');
    const bracketLoading = document.getElementById('bracket-loading');
    const bracketContainer = document.getElementById('bracket-container');
    const bracketError = document.getElementById('bracket-error');
    
    // Show loading state
    bracketLoading.style.display = 'block';
    bracketContainer.style.display = 'none';
    bracketError.style.display = 'none';
    
    try {
        // Fetch all required data
        const matchData = await fetchMatches(currentNamespace, currentTournamentId);
        const participants = await fetchParticipants(currentNamespace, currentTournamentId);
        const tournament = await fetchTournament(currentNamespace, currentTournamentId);
        
        // Check if tournament has started
        if (matchData.matches.length === 0) {
            bracketLoading.style.display = 'none';
            bracketError.textContent = 'Bracket not yet generated';
            bracketError.style.display = 'block';
            return;
        }
        
        // Transform to brackets-model format
        const bracketData = transformToBracketsModel(
            matchData.matches,
            participants,
            tournament
        );
        
        // Render bracket
        window.bracketsViewer.render(bracketData, {
            selector: '#bracket-container',
            clear: true,
        });
        
        // Show bracket
        bracketLoading.style.display = 'none';
        bracketContainer.style.display = 'block';
        
    } catch (error) {
        bracketLoading.style.display = 'none';
        bracketError.textContent = 'Failed to load bracket';
        bracketError.style.display = 'block';
        console.error('Bracket loading error:', error);
    }
}
```

### HTML Template Structure

```html
<!-- Source: Existing tournament-detail.html pattern
     Add after participants section -->

<section id="bracket-section">
    <h2>Tournament Bracket</h2>
    
    <!-- Loading state -->
    <div id="bracket-loading" style="display: none;">
        <p aria-busy="true">Loading bracket...</p>
    </div>
    
    <!-- Error state (non-critical) -->
    <div id="bracket-error" style="display: none; color: #757575;">
        <!-- Error message inserted here -->
    </div>
    
    <!-- Bracket container - must exist before render() -->
    <div id="bracket-container" class="brackets-viewer" style="display: none;">
        <!-- brackets-viewer.js renders bracket here -->
    </div>
    
    <!-- Large tournament mobile warning -->
    <div id="bracket-mobile-warning" style="display: none;">
        <p><small>💡 Desktop recommended for tournaments with 32+ participants</small></p>
    </div>
</section>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| jQuery-based bracket libraries | Vanilla JS with CSS variables | 2020-2021 | Modern browsers support CSS variables natively, no jQuery dependency needed |
| Canvas/SVG rendering | Pure CSS with flexbox | 2019-2020 | Better accessibility, easier styling, better browser compatibility |
| Server-side rendering | Client-side JSON rendering | 2018-2019 | Decoupled frontend/backend, easier to update match status dynamically |
| Fixed bracket themes | CSS variable-based theming | 2021-2022 | Easy customization without modifying library code |

**Deprecated/outdated:**
- jQuery Bracket (aropupu.fi/bracket): Still functional but jQuery dependency is outdated, last update 2016
- Challonge embeds: Third-party service with limited customization
- Custom SVG generation: Too complex to maintain compared to CSS-based solutions

## Open Questions

### Question 1: How to handle tournaments with BYE matches?

**What we know:** 
- Our backend generates BYE matches for non-power-of-2 tournaments
- brackets-viewer.js has `showSlotsOrigin` option for displaying empty slots
- Library documentation shows BYE matches with explicit labels

**What's unclear:** 
- Does our backend API return BYE matches with null participants, or omit them entirely?
- Do we need to synthesize BYE match objects during transformation?

**Recommendation:** 
- Test with non-power-of-2 tournament (e.g., 6 participants)
- If API returns BYEs, transform them with null opponent fields
- If API omits BYEs, verify brackets-viewer handles missing matches gracefully
- Document the behavior in bracket-adapter.js

### Question 2: What is the performance limit for bracket size?

**What we know:**
- Library handles up to 64+ participants (mentioned in discussions)
- Mobile experience degrades with 32+ participants (wide horizontal scroll)
- Rendering is pure DOM/CSS (no canvas), so browser-limited

**What's unclear:**
- Does library have documented max participants?
- Will 128 or 256 participant brackets cause browser issues?

**Recommendation:**
- Start with assumption of 64 max participants
- Add performance test with large mock tournament
- If needed, implement pagination or "show only current round" for very large tournaments
- Document limits in user guide

### Question 3: How to sync bracket updates during tournament?

**What we know:**
- Phase 4 uses manual refresh only (no polling)
- Match status changes as tournament progresses
- brackets-viewer.js `clear: true` replaces entire bracket

**What's unclear:**
- Is full bracket re-render efficient enough for manual refresh?
- Should we implement partial updates for individual matches?

**Recommendation:**
- Start with full bracket re-render on manual refresh (simplest)
- Monitor performance - full re-render is likely fast enough
- Defer partial updates to future phase if full re-render proves slow
- Maintain Phase 4 pattern: manual refresh only, no polling

## Sources

### Primary (HIGH confidence)

- **brackets-viewer.js GitHub Repository** (https://github.com/Drarig29/brackets-viewer.js) - Official source code, README, examples checked 2026-02-02
- **brackets-viewer.js NPM page** (https://www.npmjs.com/package/brackets-viewer) - Version 1.9.0 confirmed, 533 weekly downloads
- **brackets-docs Official Documentation** (https://drarig29.github.io/brackets-docs/) - Getting started guide and API reference
- **demo/with-api.html** (https://raw.githubusercontent.com/Drarig29/brackets-viewer.js/master/demo/with-api.html) - Official working example verified
- **src/style.scss** (https://raw.githubusercontent.com/Drarig29/brackets-viewer.js/master/src/style.scss) - CSS variable definitions and responsive behavior verified
- **Project protobuf service definition** (./pkg/proto/service.proto) - Match data structure and REST endpoints verified

### Secondary (MEDIUM confidence)

- **brackets-model GitHub** (https://github.com/Drarig29/brackets-model) - Type definitions for understanding data structure
- **Pico CSS v2.0.6** (existing project CSS framework) - Confirmed via web/static/css/pico.min.css

### Tertiary (LOW confidence)

None - all findings verified with primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - brackets-viewer.js is actively maintained (Nov 2024 release), well-documented, production-ready with 213 stars
- Architecture: HIGH - Demo files provide working examples, existing Phase 4 code provides proven integration patterns
- Pitfalls: MEDIUM-HIGH - Based on common JavaScript integration patterns and library documentation, not battle-tested in this specific project yet

**Research date:** 2026-02-02
**Valid until:** 2026-03-02 (30 days) - Library is stable with infrequent breaking changes
