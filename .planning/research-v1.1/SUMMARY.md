# Project Research Summary

**Project:** Tournament Management Service - View-Only Web UI (v1.1)
**Domain:** Tournament bracket visualization and viewing interface
**Researched:** 2025-02-02
**Confidence:** HIGH

## Executive Summary

The v1.1 milestone adds a view-only web UI to an existing tournament management REST API built with Go/Gin. Research shows the optimal approach is serving static HTML/CSS/JS files from the existing Go service with zero build tools, using Pico CSS (10KB) for responsive styling and brackets-viewer.js for production-ready bracket visualization.

The recommended architecture embeds static files in the Go binary (via Go 1.16+ embed), serves them alongside existing gRPC-Gateway routes, and uses vanilla JavaScript with native fetch API to consume the existing REST endpoints. This approach aligns perfectly with the "no build tools" constraint while delivering professional tournament viewing capabilities comparable to Challonge and Toornament.

Critical risks include mobile bracket layout complexity (horizontal tree breaks on small screens), tight API-DOM coupling without a transformation layer, and loading performance for large tournaments (64+ participants). These are mitigated by designing mobile-first with vertical layouts, implementing a centralized API client with data transformation, and lazy-loading tournament data progressively rather than fetching everything at once.

## Key Findings

### Recommended Stack

Plain HTML/CSS/JS served from Go with CDN-delivered libraries provides the optimal balance of simplicity and capability. No build tools, no framework overhead, universal browser support, and seamless integration with the existing Go/Gin service architecture.

**Core technologies:**
- **Plain HTML5 + Vanilla JavaScript (ES6+)**: No build step required, native fetch API for REST integration, runs directly in all modern browsers (2015+)
- **Pico CSS (2.1.1+)**: 10KB semantic CSS framework with built-in dark mode, mobile-first responsive design, classless styling reduces HTML verbosity
- **brackets-viewer.js (1.9.0+)**: Production-ready bracket visualization library (213+ GitHub stars), handles single-elimination layout, CDN available, actively maintained

**Supporting patterns:**
- Go `embed` package for static files bundled in binary
- SVG-based bracket rendering for scalable visualization
- Component-based vanilla JS (no framework) for modular UI
- Namespace-aware API client for centralized REST calls

### Expected Features

**Must have (table stakes):**
- Tournament list/grid view with status badges — every tournament platform has browsable entry point
- Traditional bracket tree visualization — single-elimination format has standard visual language users expect
- Match status indicators (SCHEDULED/IN_PROGRESS/COMPLETED) — visual distinction in bracket is critical for understanding
- Participant count display — "16/32 players registered" is universal tournament UI pattern
- Mobile-responsive layout — 60%+ of gaming community views on mobile, bracket must scale/scroll properly
- Round labels (Round 1, Semi-Finals, Finals) — essential for tournament navigation
- Tournament status badges (DRAFT/ACTIVE/STARTED/COMPLETED/CANCELLED) — users need instant recognition
- Click-through navigation (List → Detail → Bracket) — standard tournament browsing flow

**Should have (competitive):**
- Search/filter tournaments by name and status — reduces cognitive load when browsing many tournaments
- Match detail popups — click match to see detailed info without leaving bracket view
- Bracket zoom/pan controls — large tournaments (32-256 players) need navigation aids
- Round-by-round toggle view — helps mobile users focus on current round
- "Live" tournament indicators — highlight active tournaments to create engagement
- Empty state illustrations — "No tournaments yet" with friendly graphics improves perceived quality

**Defer (v2+):**
- Real-time auto-refresh via WebSocket — adds significant complexity, deferred to v2.0
- User registration UI — requires authentication, session management, deferred to v1.2
- Admin dashboard (create/manage tournaments) — separate milestone, API-only in v1.1
- Match chat/comments — social features out of scope for view-only milestone
- Bracket export to PNG/PDF — nice to have but not essential for viewing

### Architecture Approach

Static HTML/CSS/JS files embedded in Go binary and served alongside existing gRPC-Gateway routes. Vanilla JavaScript components fetch JSON from REST API and render DOM using modular page controllers and reusable UI components. No framework, no build tools, no separate frontend server.

**Major components:**
1. **Static File Handler** — Go `embed.FS` serves HTML/CSS/JS from `/static/*` with proper cache headers and MIME types
2. **API Client (JavaScript)** — Centralized `TournamentAPIClient` class wraps native fetch API with namespace injection, error handling, and data transformation layer
3. **Page Controllers** — `tournamentList.js` and `tournamentDetail.js` orchestrate API calls and coordinate component rendering for each page
4. **Bracket Renderer** — SVG-based visualization with flexible layout algorithm, calculates positions from round/match data dynamically
5. **UI Components** — Modular functions (tournamentCard, matchCard) render reusable DOM elements without framework overhead

**Data flow:** User visits `/tournaments` → Go serves `index.html` → JS executes → `apiClient.listTournaments()` → fetch `/v1/public/namespace/.../tournaments` → Go routes to gRPC Gateway → MongoDB query → JSON response → JS renders tournament cards → DOM updated

**Integration:** Static routes registered before API routes in `main.go`, existing gRPC-Gateway unchanged, MongoDB accessed only via existing REST endpoints

### Critical Pitfalls

1. **Bracket layout breaks on mobile (Phase 1)** — Traditional horizontal tree design breaks catastrophically on small screens. Avoid by designing mobile-first with vertical/list layout for <768px, switching to horizontal tree only on tablets/desktop. Test with 320px, 768px, 1920px viewports and 8-64 participant tournaments.

2. **Hardcoded SVG/Canvas bracket rendering (Phase 1)** — Drawing brackets with hardcoded pixel coordinates breaks when tournament size or viewport changes. Avoid by using CSS Grid/Flexbox for positioning (not absolute positioning), calculating positions dynamically from round/match data with helper functions like `getMatchPosition(round, index)`.

3. **Loading entire tournament data on page load (Phase 2)** — Fetching all data in single API call creates 500KB+ payloads and 2-3 second load times for large tournaments. Avoid by loading tournament metadata first (~1KB), then bracket structure separately, lazy-loading completed rounds only when needed.

4. **No loading states between data fetches (Phase 2)** — UI shows stale data or blank screens without indication during navigation, causing duplicate requests and race conditions. Avoid by implementing global loading state management, skeleton screens for initial loads, disabling navigation during fetches, and using AbortController to cancel in-flight requests.

5. **Tight coupling between API responses and DOM rendering (Phase 2)** — Directly inserting API fields into HTML causes cascading failures when API changes. Avoid by creating data transformation layer (`transformTournamentData()`) that returns normalized UI data model separate from raw API responses, with validation and defaults for null/undefined values.

6. **Static file serving configuration missing (Phase 3)** — Wrong MIME types, missing cache headers, CORS issues prevent proper static file delivery. Avoid by using `http.FileServer` with explicit MIME type handling, proper `Cache-Control` headers, and testing with browser cache both enabled and disabled.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Static File Infrastructure & Basic UI
**Rationale:** Foundation must be correct from the start — file serving, routing, basic HTML structure. Mobile-first responsive design decisions can't be retrofitted after building desktop-first. This phase establishes the architecture pattern all subsequent features build on.

**Delivers:** 
- Go serving static HTML/CSS/JS files from embedded filesystem
- Basic tournament list page with placeholder data
- Tournament detail page template
- Mobile-responsive CSS framework (Pico CSS) integrated
- Routing configured (static routes before API routes)

**Addresses:** 
- Static file serving configuration (Pitfall #6)
- Browser compatibility decisions (Pitfall from PITFALLS.md #6)
- Mobile-first layout foundation (prevents Pitfall #1)
- DOM structure efficiency (prevents Pitfall #10)

**Avoids:**
- Wrong MIME types / cache headers by configuring upfront
- Hardcoded bracket rendering by establishing flexible CSS Grid patterns
- Bracket layout mobile breakage by designing responsive from start

### Phase 2: API Integration & Data Layer
**Rationale:** API client patterns and data transformation must be established before building complex UI features. Tight coupling between API and rendering (Pitfall #5) is extremely expensive to fix later. Loading states and error handling are easier to build in from the start than retrofit.

**Delivers:**
- `TournamentAPIClient` class with centralized fetch logic
- Data transformation layer separating API responses from UI data models
- Error handling utilities and user-friendly error messages
- Loading state management (skeleton screens, spinners)
- Tournament list populated with real API data
- Tournament detail page fetching real metadata

**Uses:**
- Native fetch API (from STACK.md)
- Namespace injection pattern (from ARCHITECTURE.md)
- Progressive data loading (from PITFALLS.md)

**Implements:**
- API Client architecture component
- Data transformation layer
- Error boundary patterns

**Addresses:**
- Tight API-DOM coupling (Pitfall #5)
- No loading states (Pitfall #4)
- Loading entire tournament at once (Pitfall #3)
- API error handling (Pitfall #8)

**Avoids:**
- Race conditions with AbortController
- Cascading API changes by transforming data
- Performance issues by lazy-loading data

### Phase 3: Bracket Visualization
**Rationale:** Bracket rendering is the most complex feature and depends on stable API client and data layer from Phase 2. SVG rendering approach must be flexible and dynamic from the start (Pitfall #2) as hardcoded coordinates are extremely expensive to refactor.

**Delivers:**
- `BracketRenderer` component with SVG-based visualization
- Match positioning algorithm (calculates from round/match data)
- Round labels (Round 1, Semi-Finals, Finals)
- Match status indicators (color-coded SCHEDULED/IN_PROGRESS/COMPLETED)
- Connector lines between rounds
- Participant names in match boxes
- Winner highlighting in completed matches

**Uses:**
- brackets-viewer.js library (from STACK.md)
- SVG rendering pattern (from ARCHITECTURE.md)
- Component-based rendering (from ARCHITECTURE.md)

**Implements:**
- Bracket Renderer architecture component
- Match card components
- Round calculation logic

**Addresses:**
- Hardcoded SVG rendering (Pitfall #2)
- Deep DOM tree performance (Pitfall #10)

**Avoids:**
- Absolute positioning by using calculated SVG coordinates
- Performance issues by keeping DOM shallow
- Mobile breakage by inheriting responsive foundation from Phase 1

### Phase 4: Polish & Production Readiness
**Rationale:** User experience refinements (empty states, timestamps, status badges) and performance optimizations are best added after core functionality works. These are the "looks done but isn't" items from PITFALLS.md that often get skipped but are critical for production quality.

**Delivers:**
- Tournament status badges with color coding
- Participant count display ("16/32 players")
- Date/time formatting utilities (relative time "2 hours ago")
- Empty state messages ("No tournaments found")
- Manual refresh mechanism with "Updated X ago" indicator
- SEO meta tags for tournament pages
- Comprehensive browser/device testing
- Performance optimization (caching, request deduplication)

**Uses:**
- Pico CSS for status badge styling
- Progressive enhancement patterns
- HTTP cache headers configuration

**Addresses:**
- Stale match data (Pitfall #9)
- "Looks done but isn't" checklist items
- UX pitfalls (no loading indicators, no error recovery)

**Avoids:**
- Empty state dead ends by providing guidance
- Confusion about data freshness by showing timestamps
- Poor perceived quality by polishing visual details

### Phase Ordering Rationale

- **Infrastructure → Data → Features → Polish** is the optimal dependency chain: you can't fetch data without infrastructure, can't render brackets without data layer, can't polish features that don't exist
- **Mobile-first in Phase 1** prevents expensive retrofitting of responsive design after building desktop-first bracket layouts (Pitfall #1 is HIGH cost to fix)
- **Data transformation in Phase 2** establishes the boundary before complex UI development, preventing tight API coupling (Pitfall #5 is MEDIUM-HIGH cost to fix)
- **Bracket rendering in Phase 3** depends on stable API client and can leverage flexible layout patterns established in Phase 1 (Pitfall #2 is HIGH cost to fix if built with hardcoded coordinates)
- **Polish in Phase 4** adds production quality after core functionality validated, addressing "looks done but isn't" items systematically

### Research Flags

**Phases with standard patterns (skip research-phase):**
- **Phase 1:** Static file serving from Go — well-documented, established embed patterns
- **Phase 2:** REST API consumption with fetch — standard pattern, native browser API
- **Phase 4:** UX polish and error states — design patterns, no technical complexity

**Phases likely needing deeper research during planning:**
- **Phase 3:** Bracket visualization — `brackets-viewer.js` library integration may need API exploration if documentation is sparse, SVG positioning algorithms may need algorithm research for edge cases (byes, odd participant counts)

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All technologies verified as production-ready, CDN available, active maintenance. Pico CSS (14.8K stars), brackets-viewer.js (213 stars), native fetch API (99%+ browser coverage). |
| Features | HIGH | Analyzed 3 major competitors (Challonge, Start.gg, Toornament), mapped features to existing API capabilities, validated against v1.1 view-only scope constraints. |
| Architecture | HIGH | Go embed patterns well-documented (Go 1.16+), Gin static file serving verified, vanilla JS component patterns standard, existing gRPC-Gateway integration point clear from main.go analysis. |
| Pitfalls | HIGH | Derived from real-world tournament UI implementations and mobile-first responsive design literature. Mobile bracket layout issues verified across competitor analysis. Performance traps validated against Core Web Vitals metrics. |

**Overall confidence:** HIGH

### Gaps to Address

**API response format validation:** Research assumed REST endpoints return JSON matching gRPC service definitions, but actual response structure should be verified during Phase 2 implementation. If response format differs significantly from expectations, data transformation layer may need adjustment.

**brackets-viewer.js data format:** Research identified library as production-ready, but mapping between tournament API match data and library's expected format needs validation during Phase 3 planning. May need adapter layer if formats are incompatible.

**Namespace configuration:** Architecture assumes namespace can be injected at runtime (from config or URL), but actual deployment pattern (hardcoded vs dynamic) should be confirmed during Phase 1. Impacts API client initialization.

**Browser support policy:** Research recommends last 2 versions of major browsers, but actual target browsers should be confirmed with stakeholders. If older browser support required, may need to reconsider ES6+ JavaScript features or add transpilation (conflicts with no-build-tools constraint).

**Performance targets:** Research suggests <1.5s Time to First Contentful Paint and <100KB initial load, but actual performance requirements should be validated. If more aggressive targets needed (e.g., <1s, <50KB), may need additional optimization strategies.

## Sources

### Primary (HIGH confidence)
- **Existing codebase analysis** — main.go, service.proto, gateway.go (verified API endpoints, routing patterns, Go server architecture)
- [Go embed package documentation](https://pkg.go.dev/embed) — Official Go 1.16+ embedded filesystem API
- [Gin Static File Serving](https://gin-gonic.com/docs/examples/serving-static-files/) — Official Gin documentation for static handlers
- [MDN Web Docs - Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) — Standard browser API reference
- [Pico CSS Official](https://picocss.com) — 14.8K GitHub stars, official documentation for semantic CSS framework
- [brackets-viewer.js GitHub](https://github.com/Drarig29/brackets-viewer.js) — 213 stars, production-ready bracket library with examples

### Secondary (MEDIUM confidence)
- **Competitor analysis** — Challonge.com, Start.gg, Toornament.com (tournament platform feature analysis, UX patterns)
- **Web performance best practices** — web.dev Core Web Vitals (performance targets, loading optimization)
- **Mobile-first responsive design** — Smashing Magazine, A List Apart (responsive bracket layout patterns)
- **Vanilla JavaScript patterns** — MDN guides, You Don't Know JS (component patterns without frameworks)

### Tertiary (LOW confidence)
- **Tournament bracket algorithms** — Observable D3 examples (SVG layout algorithms, needs validation)
- **AccelByte Extend constraints** — Inferred from project structure, deployment requirements should be verified

---
*Research completed: 2025-02-02*
*Ready for roadmap: yes*
