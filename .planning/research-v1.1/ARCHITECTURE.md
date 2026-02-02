# Architecture Research

**Domain:** Static HTML/CSS/JS UI served from Go/Gin REST API service  
**Researched:** 2025-02-02  
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌────────────────────────────────────────────────────────────────┐
│                      Browser (User Agent)                       │
├────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ List Page    │  │ Detail Page  │  │ Bracket UI   │          │
│  │ (tournaments)│  │ (single)     │  │ (rendering)  │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                 │                 │                   │
│         └─────────────────┴─────────────────┘                   │
│                           │                                     │
│                    ┌──────▼──────┐                              │
│                    │ API Client  │  (fetch JSON)                │
│                    │ (apiClient) │                              │
│                    └──────┬──────┘                              │
├───────────────────────────┼──────────────────────────────────────┤
│                   HTTP(S) │                                     │
├───────────────────────────┼──────────────────────────────────────┤
│                  Go/Gin HTTP Server                             │
├────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐   ┌────────────────────────────────────┐  │
│  │ Static Routes   │   │ API Routes (existing)              │  │
│  │ /               │   │ /v1/public/namespace/...           │  │
│  │ /tournaments/*  │   │ /v1/admin/namespace/...            │  │
│  │ /static/*       │   │                                    │  │
│  └────────┬────────┘   └──────┬─────────────────────────────┘  │
│           │                   │                                 │
│  ┌────────▼────────┐   ┌──────▼─────────────────────────────┐  │
│  │ Static Handler  │   │ gRPC Gateway (existing)            │  │
│  │ gin.Static()    │   │ REST → gRPC                        │  │
│  │ fs.ServeFile()  │   │                                    │  │
│  └─────────────────┘   └────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────┤
│                    gRPC Service (existing)                      │
│              ┌───────────────────────────────────┐              │
│              │ TournamentService, MatchService   │              │
│              └────────────┬──────────────────────┘              │
├──────────────────────────┼───────────────────────────────────────┤
│                   MongoDB Storage                               │
└────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Static Handler** | Serve HTML/CSS/JS files | `gin.Static()` or `http.FileServer()` with proper cache headers |
| **API Client (JS)** | Fetch JSON from REST API | ES6 fetch with async/await, centralized error handling |
| **Page Components** | Render UI from API data | Vanilla JS DOM manipulation, modular page modules |
| **Bracket Renderer** | Visualize tournament tree | Canvas or SVG rendering with recursive layout algorithm |
| **Router** | Handle URL navigation | HTML5 History API or hash-based routing for SPA feel |

## Recommended Project Structure

```
extend-tournament-service/
├── main.go                          # Modified to add static routes
├── web/                             # NEW: Static web files
│   ├── index.html                   # Tournament list page
│   ├── tournament.html              # Tournament detail page
│   ├── static/                      # Static assets
│   │   ├── css/
│   │   │   ├── main.css            # Global styles
│   │   │   ├── tournament-list.css # List page styles
│   │   │   ├── tournament-detail.css
│   │   │   └── bracket.css         # Bracket visualization styles
│   │   ├── js/
│   │   │   ├── main.js             # App initialization
│   │   │   ├── apiClient.js        # Centralized API calls
│   │   │   ├── router.js           # Client-side routing (optional)
│   │   │   ├── pages/
│   │   │   │   ├── tournamentList.js  # List page controller
│   │   │   │   └── tournamentDetail.js # Detail page controller
│   │   │   ├── components/
│   │   │   │   ├── tournamentCard.js   # Tournament card component
│   │   │   │   ├── bracketRenderer.js  # Bracket visualization
│   │   │   │   └── matchCard.js        # Individual match display
│   │   │   └── utils/
│   │   │       ├── formatters.js   # Date/time formatting helpers
│   │   │       └── constants.js    # Shared constants
│   │   └── images/                 # Icons, logos
│   └── robots.txt                  # SEO
├── pkg/
│   ├── common/
│   │   └── gateway.go              # MODIFIED: Add static routes
│   ├── server/                     # Existing gRPC server
│   ├── service/                    # Existing business logic
│   └── storage/                    # Existing MongoDB layer
└── ... (existing files)
```

### Structure Rationale

- **web/:** Root directory for all web assets, separate from Go code but part of binary via embed
- **web/static/:** Static assets served at `/static/*` for cache control and CDN compatibility
- **web/static/js/pages/:** Page-level controllers that orchestrate components and API calls
- **web/static/js/components/:** Reusable UI components for rendering specific elements
- **web/static/js/utils/:** Shared utilities that don't fit in components (formatting, validation)

## Architectural Patterns

### Pattern 1: Embed Static Files in Go Binary

**What:** Use Go 1.16+ `embed` package to bundle web files into the compiled binary

**When to use:** Always for production deployment — simplifies deployment to single binary

**Trade-offs:**
- ✅ Single binary deployment (no need to copy static files separately)
- ✅ No file system access issues in containers
- ✅ Faster cold starts (files in memory)
- ❌ Requires recompilation to update static files
- ❌ Larger binary size

**Example:**
```go
package main

import (
    "embed"
    "net/http"
    "github.com/gin-gonic/gin"
)

//go:embed web/static
var staticFS embed.FS

//go:embed web/*.html
var htmlFS embed.FS

func setupStaticRoutes(router *gin.Engine) {
    // Serve embedded static files
    router.StaticFS("/static", http.FS(staticFS))
    
    // Serve HTML files with proper routing
    router.GET("/", serveHTML("web/index.html"))
    router.GET("/tournaments", serveHTML("web/index.html"))
    router.GET("/tournaments/:id", serveHTML("web/tournament.html"))
}

func serveHTML(path string) gin.HandlerFunc {
    return func(c *gin.Context) {
        data, _ := htmlFS.ReadFile(path)
        c.Data(http.StatusOK, "text/html; charset=utf-8", data)
    }
}
```

### Pattern 2: API Client with Namespace Injection

**What:** Centralized JavaScript module for all API calls with namespace handling

**When to use:** Always — prevents code duplication and ensures consistent error handling

**Trade-offs:**
- ✅ Single source of truth for API endpoints
- ✅ Consistent error handling across all pages
- ✅ Easy to add request/response interceptors
- ✅ Namespace can be injected at runtime
- ❌ Slightly more complex than inline fetch calls

**Example:**
```javascript
// apiClient.js
class TournamentAPIClient {
    constructor(namespace) {
        this.namespace = namespace;
        this.baseURL = `/v1/public/namespace/${namespace}`;
    }

    async listTournaments(filters = {}) {
        const params = new URLSearchParams({
            limit: filters.limit || 20,
            offset: filters.offset || 0,
            ...(filters.status && { status: filters.status })
        });
        
        const response = await fetch(`${this.baseURL}/tournaments?${params}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch tournaments: ${response.statusText}`);
        }
        return response.json();
    }

    async getTournament(tournamentId) {
        const response = await fetch(`${this.baseURL}/tournaments/${tournamentId}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch tournament: ${response.statusText}`);
        }
        return response.json();
    }

    async getTournamentMatches(tournamentId, round = null) {
        const url = round !== null
            ? `${this.baseURL}/tournaments/${tournamentId}/matches?round=${round}`
            : `${this.baseURL}/tournaments/${tournamentId}/matches`;
        
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Failed to fetch matches: ${response.statusText}`);
        }
        return response.json();
    }
}

// Usage: Initialize with namespace from config or URL
const apiClient = new TournamentAPIClient('game-namespace');
```

### Pattern 3: Component-Based Rendering (No Framework)

**What:** Modular JavaScript functions that render DOM elements without a framework

**When to use:** For simple UIs where framework overhead isn't justified

**Trade-offs:**
- ✅ No build step required
- ✅ Fast initial page load (no framework parsing)
- ✅ Full control over DOM updates
- ✅ Easy to understand for Go developers
- ❌ Manual DOM manipulation can get verbose
- ❌ No automatic reactivity

**Example:**
```javascript
// components/tournamentCard.js
function createTournamentCard(tournament) {
    const card = document.createElement('div');
    card.className = 'tournament-card';
    card.dataset.tournamentId = tournament.tournament_id;
    
    card.innerHTML = `
        <div class="tournament-card__header">
            <h3 class="tournament-card__title">${escapeHTML(tournament.name)}</h3>
            <span class="tournament-card__status tournament-card__status--${tournament.status.toLowerCase()}">
                ${formatStatus(tournament.status)}
            </span>
        </div>
        <div class="tournament-card__body">
            <p class="tournament-card__description">${escapeHTML(tournament.description)}</p>
            <div class="tournament-card__meta">
                <span class="tournament-card__participants">
                    ${tournament.current_participants} / ${tournament.max_participants} players
                </span>
                <span class="tournament-card__date">
                    ${formatDate(tournament.start_time)}
                </span>
            </div>
        </div>
    `;
    
    // Add click handler for navigation
    card.addEventListener('click', () => {
        window.location.href = `/tournaments/${tournament.tournament_id}`;
    });
    
    return card;
}

// Helper to prevent XSS
function escapeHTML(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
```

### Pattern 4: SVG-Based Bracket Rendering

**What:** Use SVG for scalable, resolution-independent bracket visualization

**When to use:** For tournament brackets — better than Canvas for static diagrams

**Trade-offs:**
- ✅ Crisp rendering at any zoom level
- ✅ CSS styling support
- ✅ Easy to add hover effects and click handlers
- ✅ Accessibility support with ARIA labels
- ❌ More complex DOM structure than Canvas
- ❌ Can be slower for very large brackets (>128 participants)

**Example:**
```javascript
// components/bracketRenderer.js
class BracketRenderer {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.matchWidth = 200;
        this.matchHeight = 60;
        this.roundSpacing = 240;
        this.matchSpacing = 20;
    }

    render(matches, totalRounds) {
        // Group matches by round
        const rounds = this.groupByRound(matches, totalRounds);
        
        // Calculate SVG dimensions
        const width = totalRounds * this.roundSpacing + 100;
        const height = Math.max(
            ...rounds.map((r, i) => this.calculateRoundHeight(r.length, i))
        );
        
        // Create SVG element
        const svg = this.createSVG(width, height);
        
        // Render each round
        rounds.forEach((roundMatches, roundIndex) => {
            this.renderRound(svg, roundMatches, roundIndex, totalRounds);
        });
        
        this.container.innerHTML = '';
        this.container.appendChild(svg);
    }

    renderRound(svg, matches, roundIndex, totalRounds) {
        const x = roundIndex * this.roundSpacing + 50;
        const verticalSpacing = this.calculateVerticalSpacing(matches.length, roundIndex);
        
        matches.forEach((match, matchIndex) => {
            const y = this.calculateYPosition(matchIndex, verticalSpacing, roundIndex);
            this.renderMatch(svg, match, x, y);
            
            // Draw connector lines to next round (except final round)
            if (roundIndex < totalRounds - 1) {
                this.drawConnectorLines(svg, x, y, roundIndex, matchIndex);
            }
        });
    }

    renderMatch(svg, match, x, y) {
        // Create match box
        const group = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        group.setAttribute('class', 'bracket-match');
        group.setAttribute('data-match-id', match.match_id);
        
        // Background rectangle
        const rect = this.createMatchBox(x, y);
        group.appendChild(rect);
        
        // Participant 1
        const p1Text = this.createParticipantText(
            match.participant1?.username || 'TBD',
            x + 10,
            y + 20,
            match.winner === match.participant1?.user_id
        );
        group.appendChild(p1Text);
        
        // Participant 2
        const p2Text = this.createParticipantText(
            match.participant2?.username || 'TBD',
            x + 10,
            y + 40,
            match.winner === match.participant2?.user_id
        );
        group.appendChild(p2Text);
        
        svg.appendChild(group);
    }

    createSVG(width, height) {
        const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('width', width);
        svg.setAttribute('height', height);
        svg.setAttribute('viewBox', `0 0 ${width} ${height}`);
        svg.setAttribute('class', 'bracket-svg');
        return svg;
    }

    groupByRound(matches, totalRounds) {
        const rounds = Array.from({ length: totalRounds }, () => []);
        matches.forEach(match => {
            rounds[match.round - 1].push(match);
        });
        return rounds;
    }
    
    calculateVerticalSpacing(matchCount, roundIndex) {
        // Double spacing for each subsequent round
        return this.matchHeight + this.matchSpacing * Math.pow(2, roundIndex);
    }
    
    calculateYPosition(matchIndex, spacing, roundIndex) {
        const offset = spacing / 2;
        return matchIndex * spacing + offset + 50;
    }
    
    // ... more helper methods
}
```

## Data Flow

### Request Flow

```
[User visits /tournaments]
    ↓
[Go serves index.html] → [Browser parses HTML]
    ↓
[JS loads and executes]
    ↓
[main.js initializes] → [tournamentList.js runs]
    ↓
[apiClient.listTournaments()] → [fetch /v1/public/namespace/.../tournaments]
    ↓
[Go/Gin routes to gRPC Gateway] → [gRPC Gateway calls TournamentService]
    ↓
[MongoDB query] → [JSON response]
    ↓
[JS receives JSON] → [createTournamentCard() for each]
    ↓
[DOM updated with tournament cards]
```

### Page Navigation Flow

```
[User clicks tournament card]
    ↓
[window.location.href = /tournaments/{id}]
    ↓
[Go serves tournament.html]
    ↓
[JS extracts ID from URL]
    ↓
[Parallel API calls:]
    ├─ [apiClient.getTournament(id)]
    ├─ [apiClient.getTournamentParticipants(id)]
    └─ [apiClient.getTournamentMatches(id)]
    ↓
[Render tournament header]
[Render participant list]
[BracketRenderer.render(matches)]
    ↓
[User sees complete tournament view]
```

### Key Data Flows

1. **Tournament List:** Fetch → Filter/Sort → Render Cards → Handle Pagination
2. **Tournament Detail:** Fetch Metadata + Matches + Participants → Render Header + Bracket → Enable Interactions
3. **Bracket Rendering:** Group by Round → Calculate Layout → Generate SVG → Add Event Listeners
4. **Error Handling:** Catch Fetch Error → Show User-Friendly Message → Log to Console

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 0-1k users | Serve static files directly from Go binary with embed, no CDN needed |
| 1k-10k users | Add HTTP cache headers (Cache-Control, ETag), consider nginx reverse proxy |
| 10k-100k users | Move static files to CDN (CloudFront, Cloudflare), keep HTML served from Go |
| 100k+ users | Separate static file service, add Redis caching for API responses |

### Scaling Priorities

1. **First bottleneck:** API response times — Add MongoDB indexes on frequently queried fields (tournament status, start_time)
2. **Second bottleneck:** Static file serving — Add CDN with long cache times (1 year for versioned assets)
3. **Third bottleneck:** Large bracket rendering — Implement lazy loading for tournaments with >64 participants

## Anti-Patterns

### Anti-Pattern 1: Global State Pollution

**What people do:** Put everything in global variables (`window.tournaments = []`)

**Why it's wrong:**
- Name collisions with other scripts
- Hard to test in isolation
- No clear data ownership
- Memory leaks if not cleaned up

**Do this instead:** Use ES6 modules or IIFE pattern with explicit exports
```javascript
// Good: Module pattern
const TournamentApp = (function() {
    let tournaments = []; // Private state
    
    return {
        init() { /* ... */ },
        loadTournaments() { /* ... */ }
    };
})();

// Or better: ES6 modules
// tournamentList.js
let tournaments = []; // Module-scoped, not global
export function loadTournaments() { /* ... */ }
```

### Anti-Pattern 2: Inline Event Handlers in HTML

**What people do:** `<button onclick="deleteTournament(123)">Delete</button>`

**Why it's wrong:**
- Mixes behavior with markup
- Creates implicit globals
- Can't use event delegation
- CSP (Content Security Policy) violations

**Do this instead:** Use addEventListener in JavaScript
```javascript
// Good: Attach handlers in JS
document.querySelectorAll('.tournament-card').forEach(card => {
    card.addEventListener('click', handleTournamentClick);
});

// Even better: Event delegation
document.getElementById('tournament-list').addEventListener('click', (e) => {
    const card = e.target.closest('.tournament-card');
    if (card) {
        handleTournamentClick(card.dataset.tournamentId);
    }
});
```

### Anti-Pattern 3: Serving HTML Through API Gateway

**What people do:** Route `/` and `/tournaments` through gRPC gateway to serve HTML

**Why it's wrong:**
- gRPC gateway adds unnecessary overhead for static content
- Harder to set proper cache headers
- Can't easily serve different content types
- Complicates debugging

**Do this instead:** Add separate static routes in main.go before API routes
```go
// Good: Separate static handler
func newGRPCGatewayHTTPServer(addr string, handler http.Handler, logger *slog.Logger) *http.Server {
    mux := http.NewServeMux()
    
    // Static routes first (order matters!)
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
    mux.HandleFunc("/", serveIndexHTML)
    mux.HandleFunc("/tournaments/", serveTournamentHTML)
    
    // API routes
    mux.Handle("/v1/", handler) // gRPC gateway
    
    return &http.Server{Addr: addr, Handler: mux}
}
```

### Anti-Pattern 4: No Loading States

**What people do:** Start API call, show nothing until data arrives

**Why it's wrong:**
- Appears broken on slow connections
- No user feedback
- Can't distinguish between loading and empty state

**Do this instead:** Show loading skeleton/spinner immediately
```javascript
// Good: Show loading state
async function loadTournaments() {
    const container = document.getElementById('tournament-list');
    container.innerHTML = '<div class="loading-spinner">Loading tournaments...</div>';
    
    try {
        const data = await apiClient.listTournaments();
        renderTournaments(data.tournaments);
    } catch (error) {
        container.innerHTML = '<div class="error-message">Failed to load tournaments</div>';
    }
}
```

### Anti-Pattern 5: Rendering HTML Strings Without Escaping

**What people do:** `element.innerHTML = '<div>' + tournament.name + '</div>'`

**Why it's wrong:**
- XSS vulnerability if tournament.name contains `<script>` tags
- Security risk from user-generated content

**Do this instead:** Use textContent or escape HTML
```javascript
// Good: Safe rendering
function createTournamentCard(tournament) {
    const card = document.createElement('div');
    
    const title = document.createElement('h3');
    title.textContent = tournament.name; // Auto-escaped
    
    card.appendChild(title);
    return card;
}

// Or: Use escaping helper for innerHTML
function escapeHTML(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Go/Gin HTTP Server | Static file middleware + API routes | Use separate handlers, static routes registered first |
| gRPC Gateway (existing) | Pass-through for `/v1/*` routes | No changes to existing gateway code |
| MongoDB (existing) | No direct access from frontend | All data via REST API |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| HTML ↔ JavaScript | Script tags, ES6 modules | Load order: utils → apiClient → components → pages → main |
| JavaScript ↔ Go API | Fetch API with JSON | Use namespace from config, handle 401/403 for auth |
| Static Handler ↔ Embed FS | Go embed.FS interface | Compile-time embedding, runtime serving |

## Implementation Build Order

### Phase 1: Static File Infrastructure
**Goal:** Go can serve HTML/CSS/JS files

1. Create `web/` directory structure
2. Add embed directives to main.go
3. Modify `newGRPCGatewayHTTPServer()` to add static routes
4. Create basic index.html with "Hello World"
5. Verify `/` serves HTML, `/static/` serves assets

**Verification:** Navigate to http://localhost:8000/ and see HTML page

### Phase 2: API Client Foundation
**Goal:** JavaScript can fetch data from existing API

1. Create `apiClient.js` with TournamentAPIClient class
2. Implement `listTournaments()` method
3. Create basic `main.js` to initialize client with namespace
4. Add error handling and logging
5. Test API calls from browser console

**Verification:** Open console, run `apiClient.listTournaments()`, see JSON data

### Phase 3: Tournament List Page
**Goal:** Display all tournaments in a grid/list

1. Create `pages/tournamentList.js` page controller
2. Implement `components/tournamentCard.js` for card rendering
3. Add CSS for tournament cards and grid layout
4. Implement loading states and error handling
5. Add pagination or "Load More" button

**Verification:** Navigate to `/`, see list of tournaments from API

### Phase 4: Tournament Detail Page
**Goal:** Show tournament details, participants, matches

1. Create `tournament.html` template
2. Create `pages/tournamentDetail.js` controller
3. Extract tournament ID from URL path
4. Fetch tournament metadata and participants
5. Render tournament header and participant list

**Verification:** Click tournament card, see detail page with metadata

### Phase 5: Bracket Visualization
**Goal:** Display tournament bracket tree

1. Create `components/bracketRenderer.js` with SVG logic
2. Fetch matches and group by round
3. Implement bracket layout algorithm (vertical spacing)
4. Render matches with participant names
5. Add CSS for bracket styling (lines, boxes, winner highlight)

**Verification:** View started/completed tournament, see bracket tree

### Phase 6: Polish & Mobile Responsive
**Goal:** Production-ready UI

1. Add responsive CSS (media queries for mobile)
2. Implement status badges with colors
3. Add date/time formatting utilities
4. Implement empty states ("No tournaments found")
5. Add meta tags for SEO and social sharing
6. Test on mobile devices

**Verification:** Resize browser, test on phone, verify usability

## Sources

- [Gin Static File Serving](https://gin-gonic.com/docs/examples/serving-static-files/) — Official Gin documentation
- [Go embed package](https://pkg.go.dev/embed) — Go 1.16+ embedded file system
- [HTTP Handler Patterns](https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/) — Go web server best practices
- [SVG Bracket Layouts](https://observablehq.com/@d3/tournament-bracket) — Bracket visualization algorithms
- [Vanilla JS Best Practices](https://github.com/elsewhencode/project-guidelines) — JavaScript project structure

---
*Architecture research for: Tournament Management System v1.1 Static UI*  
*Researched: 2025-02-02*
