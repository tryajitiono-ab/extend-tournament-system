# Pitfalls Research

**Domain:** Tournament Viewing UI with Bracket Visualization
**Researched:** 2025-02-02
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Bracket Layout Breaks on Small Screens

**What goes wrong:**
Traditional tournament bracket trees are designed for wide screens. They break catastrophically on mobile devices - elements overlap, text becomes unreadable, horizontal scrolling becomes unusable beyond 32+ participants, and users can't follow match progression.

**Why it happens:**
Developers design brackets on desktop monitors and test only on wide viewports. The exponential width growth of tournament trees (doubles each round) is manageable on 1920px screens but impossible on 375px mobile screens. CSS overflow-x scrolling feels like a solution but creates terrible UX.

**How to avoid:**
- Design mobile-first: Start with 320px viewport, expand upward
- Use vertical/list layout for mobile (<768px): rounds stack vertically, matches shown as cards
- Switch to horizontal tree only on tablets/desktop (≥768px)
- Test with real tournament sizes: 8, 16, 32, 64 participants minimum
- Consider alternative visualizations: accordion rounds, collapsible sections
- Never rely solely on horizontal scrolling for navigation

**Warning signs:**
- Horizontal scroll appears on devices < 768px width
- Text size becomes < 12px on mobile
- Match connections (lines/brackets) overlap or disappear
- Users need to scroll horizontally more than 2-3 screens
- Bracket doesn't fit iPhone SE (375px) or smaller Android devices

**Phase to address:**
Phase 1 (UI Structure) - Must be core layout decision, not afterthought

---

### Pitfall 2: Hardcoded SVG/Canvas Bracket Rendering

**What goes wrong:**
Drawing bracket connections (lines, curves) using hardcoded pixel coordinates. When tournament size changes (8 vs 64 participants), viewport changes, or styling updates, the entire rendering logic breaks. Emergency fixes require JavaScript rewrites.

**Why it happens:**
Developers look at beautiful bracket examples with SVG/Canvas art and replicate the pixel-perfect approach without considering dynamic data. Math for positioning is complex, so they hardcode "working" values rather than building flexible formulas.

**How to avoid:**
- Use CSS Grid or Flexbox for bracket positioning, not absolute positioning
- Let browser handle layout flow, use pseudo-elements for connecting lines
- If using SVG/Canvas, calculate positions from: round depth, match index, viewport dimensions
- Store dimensions as percentages or viewport units, not pixels
- Test with multiple tournament sizes: 4, 8, 16, 32, 64 participants
- Build helper functions: `getMatchPosition(round, index)`, `getConnectionPath(match1, match2)`

**Warning signs:**
- Bracket code contains magic numbers like `top: 247px` or `left: 583px`
- Adding a new round requires changing 10+ coordinate values
- Different tournament sizes need separate rendering functions
- Changing font size breaks bracket alignment
- Lines don't connect to matches after window resize

**Phase to address:**
Phase 1 (UI Structure) - Must establish flexible rendering from start

---

### Pitfall 3: Loading Entire Tournament Data on Page Load

**What goes wrong:**
Fetching all tournament data (details, participants, all matches, all rounds) in a single API call on page load. For 64-participant tournaments with 63 matches, this creates 500KB+ payloads, 2-3 second load times, and unresponsive UI. Mobile users on slow networks wait 5+ seconds for anything to render.

**Why it happens:**
It's easiest to fetch everything once and avoid loading state management. Developers assume tournaments are "small data" compared to social feeds or images. They don't test on 3G networks or large tournaments.

**How to avoid:**
- Load tournament metadata first (name, status, max participants): ~1KB, < 200ms
- Load bracket structure separately: only current + next round initially
- Lazy load completed rounds: fetch when user expands/scrolls to them
- Paginate participant lists beyond 32 participants
- Use HTTP caching headers for tournament metadata (5min cache)
- Show skeleton UI immediately, load data progressively

**Warning signs:**
- Initial API call returns > 100KB of data
- Page shows blank screen for > 1 second on fast networks
- Browser DevTools waterfall shows single long request blocking page
- Time to First Contentful Paint (FCP) > 1.5 seconds
- Mobile testing reveals 4+ second load times

**Phase to address:**
Phase 2 (Data Fetching) - Critical for performance from start

---

### Pitfall 4: No Loading States Between Data Fetches

**What goes wrong:**
When user navigates between tournaments or refreshes bracket data, the UI shows stale data or blank screens without indication of loading. Users click again thinking it didn't work, causing duplicate requests. Race conditions occur when fast-clicking between tournaments.

**Why it happens:**
Plain HTML/JS without framework forces manual loading state management. Developers focus on "happy path" where data loads instantly in development. They forget network delays, slow servers, and user behavior patterns.

**How to avoid:**
- Create global loading state management: `window.app = { loading: false, currentTournament: null }`
- Show loading indicators: skeleton screens for initial load, spinners for refreshes
- Disable navigation during loads: grey out tournament links while fetching
- Cancel in-flight requests on navigation: use AbortController
- Handle race conditions: ignore responses for non-current tournament
- Show error states with retry buttons, not silent failures

**Warning signs:**
- Users can click navigation while data is loading
- Clicking fast between tournaments shows wrong tournament data
- No visual feedback between clicking and data appearing
- Console shows multiple simultaneous API calls for same data
- Stale data briefly appears before new data loads

**Phase to address:**
Phase 2 (Data Fetching) - Must handle concurrent with API integration

---

### Pitfall 5: Tight Coupling Between API Responses and DOM Rendering

**What goes wrong:**
JavaScript directly inserts API response fields into HTML: `element.innerHTML = tournament.name`. When API adds/renames fields, changes date formats, or returns null values, the entire UI breaks with JavaScript errors. Users see "undefined" text, broken brackets, or white screen.

**Why it happens:**
Without frameworks, developers take shortcuts: directly accessing API data in rendering code. They don't build transformation layer because it feels like unnecessary abstraction for "simple" project.

**How to avoid:**
- Create data transformation layer: `transformTournamentData(apiResponse)` returns normalized object
- Define UI data contracts separate from API: `{ id, name, status, participantCount }`
- Validate API responses before use: check required fields, provide defaults
- Handle null/undefined gracefully: `tournament?.name ?? 'Unknown Tournament'`
- Build render functions that expect UI data model, not raw API responses
- Mock API responses in development to catch coupling early

**Warning signs:**
- Rendering code directly accesses `response.data.tournament.participants[0].user.name`
- API field rename breaks UI (cascading failures)
- Console errors about "Cannot read property 'name' of undefined"
- Adding new API field requires touching 5+ rendering functions
- No clear boundary between API layer and UI layer

**Phase to address:**
Phase 2 (Data Fetching) - Must establish before significant UI development

---

### Pitfall 6: Browser Compatibility Assumptions

**What goes wrong:**
Using modern JavaScript features (optional chaining, nullish coalescing, async/await, fetch API) without transpilation or polyfills. Older browsers (Safari < 13.4, Chrome < 80, Android < 7) show blank pages with console errors. Users on corporate devices or older phones can't access the tournament viewer.

**Why it happens:**
Developers work in modern browsers (latest Chrome) and assume everyone has auto-updates. Plain HTML/JS projects skip build tools, so no Babel/transpilation. They don't test in older browsers because "who still uses that?"

**How to avoid:**
- Define browser support policy: last 2 years of major browsers OR specific versions
- Test in: Safari 13+, Chrome 80+, Firefox 75+, Edge 80+
- Avoid newest features: stick to ES2018 or add transpilation
- Use feature detection, not browser sniffing: `if ('fetch' in window)`
- Provide graceful degradation: fallback to XMLHttpRequest if fetch unavailable
- Test on real devices: older iPhone SE, Android 7 devices

**Warning signs:**
- No browser support policy documented
- Code uses features less than 2 years old without polyfills
- Only tested in Chrome DevTools device mode
- No testing on actual mobile devices
- Console errors in Safari/Firefox but not Chrome

**Phase to address:**
Phase 1 (UI Structure) - Decision must be made before writing JavaScript

---

### Pitfall 7: Static File Serving Configuration Missing

**What goes wrong:**
Static HTML/CSS/JS files aren't served correctly from Go service. Wrong MIME types cause JavaScript to not execute (`text/plain` instead of `text/javascript`). Missing cache headers cause re-downloading on every page load. CORS issues prevent API calls from frontend.

**Why it happens:**
Developers focus on building UI, treating static file serving as "just works" detail. Go's file server has different defaults than Node.js/Python servers they're used to. Testing only on localhost masks caching issues.

**How to avoid:**
- Use `http.FileServer` with explicit MIME type handling
- Set proper cache headers: `Cache-Control: public, max-age=3600` for static assets
- Configure CORS properly if API and UI on different origins
- Test with browser cache disabled AND enabled
- Verify MIME types in Network tab: JS should be `application/javascript`
- Handle SPA routing: serve index.html for non-file paths

**Warning signs:**
- Browser DevTools shows JavaScript files with wrong MIME type
- API calls fail with CORS errors despite being same-origin
- Every page navigation re-downloads all assets (no 304 responses)
- Static files work in development but fail in Docker container
- Paths with extensions work but clean URLs return 404

**Phase to address:**
Phase 3 (Go Integration) - Must be correct from initial integration

---

### Pitfall 8: API Error Handling Shows Stack Traces to Users

**What goes wrong:**
When tournament API returns 404, 500, or network errors, the UI shows raw error messages, stack traces, or generic "undefined". Users see technical jargon instead of helpful messages. Debugging information leaks server details to potential attackers.

**Why it happens:**
Error handling is added last, after "happy path" works. Developers test with working API, miss error cases. Catch blocks do `console.log(error)` or `alert(error.message)` instead of user-friendly messages.

**How to avoid:**
- Define user-facing error messages for common cases: 404 → "Tournament not found", 500 → "Server error, try again"
- Never show raw error messages or stack traces to users
- Create error handling utility: `handleApiError(error)` returns user message
- Show retry buttons for transient errors (network, 503)
- Log detailed errors to console for debugging, show clean messages to users
- Test error paths: disconnect network, return 500 from API, send malformed JSON

**Warning signs:**
- Users see "NetworkError: Failed to fetch" messages
- Error states show stack traces or technical details
- Different error types all show generic "Something went wrong"
- No way for user to recover from errors (no retry)
- Console errors don't provide debugging information

**Phase to address:**
Phase 2 (Data Fetching) - Must handle concurrent with API integration

---

### Pitfall 9: Match State Updates Don't Reflect in Bracket

**What goes wrong:**
When match results are submitted (via API outside the UI), the bracket display shows stale data. Users see outdated winners, incorrect standings, or matches not advancing. Refreshing entire page is the only way to see updates.

**Why it happens:**
Static HTML/JS with manual DOM manipulation doesn't have reactive updates. Developers fetch data once on page load and don't implement refresh mechanism. Without WebSockets, they assume data is static.

**How to avoid:**
- Implement manual refresh button: prominently visible, shows last update time
- Add auto-refresh option: poll API every 30-60 seconds for active tournaments
- Show "Updated X seconds ago" indicator
- Optimistically update UI when user submits result (if allowed)
- Cache API responses briefly: 10-30 seconds to reduce server load
- Clear indicator when data might be stale (tournament status = "in-progress")

**Warning signs:**
- No visible way to refresh tournament data
- Match results submitted via API don't appear in UI
- Users ask "why isn't the bracket updated?"
- No indication of data freshness (when last loaded)
- Polling every second killing server performance

**Phase to address:**
Phase 2 (Data Fetching) - Design refresh strategy early

---

### Pitfall 10: Nested Tournament Brackets Create Deep DOM Trees

**What goes wrong:**
Rendering 64-participant tournament (7 rounds, 63 matches) creates deeply nested DOM structure with 500+ elements. Browser becomes slow, scrolling stutters, CSS selector performance degrades. Mobile browsers crash or freeze on large tournaments.

**Why it happens:**
Developers create wrapper divs for every match, round, connection line. Each match has 10+ nested elements for styling. They don't test with realistic tournament sizes or monitor DOM node count.

**How to avoid:**
- Keep DOM shallow: avoid unnecessary wrapper elements
- Use CSS Grid/Flexbox instead of deeply nested containers
- Virtualize large brackets: only render visible rounds/matches
- Limit maximum tournament size shown without pagination: 64 participants max
- Monitor DOM node count: keep under 1000 nodes per page
- Use event delegation instead of listeners on every match

**Warning signs:**
- DevTools shows 1000+ DOM nodes for bracket
- Scrolling performance degrades (< 60fps)
- Each match element has 5+ wrapper divs
- Mobile devices show lag when navigating bracket
- Browser DevTools Performance tab shows long rendering times

**Phase to address:**
Phase 1 (UI Structure) - DOM structure must be efficient from start

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Global variables for state | Quick to implement, no architecture needed | Impossible to debug, state conflicts, race conditions | Never - use namespaced objects at minimum |
| Inline event handlers in HTML | Faster initial development | Impossible to test, security issues, maintenance nightmare | Never - use addEventListener |
| jQuery for DOM manipulation | Familiar API, lots of examples | 30KB dependency for simple tasks, outdated patterns | Never - vanilla JS is sufficient |
| Polling API every 5 seconds | Real-time updates without WebSockets | Massive server load, wasted bandwidth | Only for active tournaments with long polling (30-60s) |
| Copy-paste bracket rendering per size | Works immediately, no abstraction needed | Unmaintainable, bugs multiply, can't add features | Never - build flexible renderer |
| No data transformation layer | Direct API-to-DOM is fastest path | API changes break everything, testing impossible | Never - always transform API data |
| Absolute positioning for bracket layout | Pixel-perfect control | Breaks on resize, different data, accessibility issues | Only for print/PDF export, never for web UI |
| Single CSS file with no organization | Simple, no build process | Impossibly to maintain beyond 500 lines | Only for < 300 lines total CSS |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Tournament API | Assuming data is always present (no null checks) | Validate every field, provide defaults, handle null/undefined gracefully |
| Go file server | Not setting MIME types explicitly | Configure content type headers for .js, .css, .html files |
| REST endpoints | Fetching all data in one request | Split into multiple endpoints: metadata, bracket, matches, participants |
| Browser fetch() | No timeout handling (fetch hangs forever) | Wrap in Promise.race() with timeout, use AbortController |
| CORS configuration | Allowing all origins (*) in production | Explicit allowed origins, proper credentials handling |
| Static asset paths | Hardcoded paths like `/static/app.js` | Use relative paths or base URL config for different deployments |
| Authentication headers | Forgetting to include auth on every API request | Centralize API client with default headers |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Loading all matches at once | Slow page loads, large payloads | Paginate or lazy-load completed rounds | 32+ participant tournaments (31+ matches) |
| Re-rendering entire bracket on update | UI freezes, sluggish interactions | Update only changed match elements | Any update with 64+ participants |
| No request deduplication | Multiple identical API calls | Cache responses, debounce requests | Users clicking through tournaments quickly |
| Rendering all rounds simultaneously | Memory issues, slow DOM | Virtualize or collapse completed rounds | 64+ participants (7 rounds, 63 matches) |
| No image optimization | Slow loads, wasted bandwidth | Serve appropriately sized images, use WebP | Mobile users on slow networks |
| Synchronous DOM updates | Blocking UI thread | Use requestAnimationFrame, batch updates | Updating > 10 elements at once |
| Deep CSS selectors | Slow style calculation | Use classes, shallow selectors | > 500 DOM nodes |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Trusting API data without validation | XSS via tournament names, participant names | Sanitize all user-generated content before inserting in DOM |
| Showing admin controls based on client-side checks | Users can manipulate DOM to show admin buttons | Always validate permissions on server, hide buttons as UX only |
| No rate limiting on API client side | Accidental DoS from rapid requests | Implement client-side debouncing and request throttling |
| Exposing sensitive errors to users | Information disclosure about backend | Generic error messages, log details server-side only |
| Direct innerHTML with API data | XSS injection through tournament/participant names | Use textContent or createElement, never innerHTML with user data |
| No CSRF protection on form submissions | Forged requests from other sites | Use CSRF tokens if allowing form submissions |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No loading indicators | Users don't know if click worked, click multiple times | Skeleton screens, spinners, disable buttons during load |
| Tournament list with no filters | Overwhelming, can't find relevant tournaments | Filter by status (upcoming/active/completed), search by name |
| Bracket with no zoom/pan controls | Can't see details on large tournaments | Pinch-to-zoom on mobile, zoom buttons on desktop |
| No indication of where user is in bracket | Lost in complex brackets, can't find their matches | Highlight user's matches, scroll-to-participant feature |
| Match results with no timestamps | Can't tell when match completed | Show relative time "2 hours ago" and absolute "Jan 15, 2pm" |
| Empty states with no guidance | Dead end, users don't know what to do | "No tournaments yet" with clear next action |
| Error messages with no recovery | User is stuck, must leave page | Retry buttons, back navigation, clear instructions |
| Mobile horizontal scrolling brackets | Terrible UX, gets lost in bracket | Vertical layout on mobile, collapsible rounds |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Bracket Display:** Often missing responsive behavior — verify works on 320px mobile, 768px tablet, 1920px desktop
- [ ] **Data Fetching:** Often missing error handling — verify shows user-friendly errors for 404, 500, network failure
- [ ] **Loading States:** Often missing for fast networks — verify works on throttled 3G connection
- [ ] **Match Updates:** Often missing refresh mechanism — verify can see newly completed matches
- [ ] **Tournament List:** Often missing empty state — verify shows helpful message when no tournaments
- [ ] **API Integration:** Often missing validation — verify handles null values, missing fields gracefully
- [ ] **Static Files:** Often missing cache headers — verify browser caches assets properly
- [ ] **Browser Support:** Often missing Safari testing — verify works in Safari, not just Chrome
- [ ] **Large Tournaments:** Often missing performance optimization — verify works smoothly with 64 participants
- [ ] **Error Recovery:** Often missing retry mechanisms — verify users can retry after errors
- [ ] **Accessibility:** Often missing keyboard navigation — verify can navigate bracket with keyboard only
- [ ] **CORS:** Often missing proper configuration — verify API calls work from deployed domain

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Hardcoded bracket rendering | HIGH | Rebuild with flexible layout system (Grid/Flexbox), extract positioning logic to functions |
| No responsive design | MEDIUM | Add mobile-first CSS, create vertical layout variant, test incrementally |
| Tight API coupling | MEDIUM | Create data transformation layer, define UI data contracts, refactor rendering |
| No loading states | LOW | Add loading flags, insert skeleton/spinner components, disable navigation during loads |
| Missing error handling | LOW | Wrap API calls in try-catch, create error display component, add retry buttons |
| Performance issues (DOM depth) | HIGH | Flatten DOM structure, remove unnecessary wrappers, may require significant refactor |
| Wrong MIME types | LOW | Configure file server middleware, set explicit Content-Type headers |
| Browser compatibility | MEDIUM | Add polyfills or transpile to older ES version, test in target browsers |
| No refresh mechanism | LOW | Add refresh button, implement polling logic, show last updated time |
| XSS vulnerabilities | MEDIUM | Audit all user data insertion points, replace innerHTML with textContent, sanitize inputs |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Bracket layout breaks on mobile | Phase 1: UI Structure | Test on 320px, 768px, 1920px viewports with 8, 16, 32, 64 participant tournaments |
| Hardcoded SVG/Canvas rendering | Phase 1: UI Structure | Change tournament size, verify bracket adapts without code changes |
| Loading entire tournament at once | Phase 2: Data Fetching | Monitor Network tab, verify initial load < 100KB, < 1.5s |
| No loading states | Phase 2: Data Fetching | Throttle network to 3G, verify loading indicators appear |
| API-DOM tight coupling | Phase 2: Data Fetching | Mock API response changes, verify UI doesn't break |
| Browser compatibility issues | Phase 1: UI Structure | Test in Safari, older Chrome/Firefox, verify no console errors |
| Static file serving issues | Phase 3: Go Integration | Check MIME types in Network tab, verify caching works |
| API error handling | Phase 2: Data Fetching | Simulate 404/500 errors, verify user-friendly messages |
| Stale match data | Phase 2: Data Fetching | Submit match result, verify refresh shows update |
| Deep DOM tree performance | Phase 1: UI Structure | Test 64-participant tournament, verify < 1000 DOM nodes |

## Sources

- Real-world tournament bracket UI implementations (Challonge, Battlefy, Toornament)
- Web performance best practices (web.dev, Core Web Vitals)
- Mobile-first responsive design patterns (Smashing Magazine, A List Apart)
- Plain JavaScript pitfalls without frameworks (MDN, You Don't Know JS)
- Go static file serving documentation (golang.org/pkg/net/http)
- REST API integration patterns (Richardson Maturity Model)
- Tournament management system requirements (from PROJECT.md, REQUIREMENTS.md)
- AccelByte Extend service architecture (from main.go, README.md)

---
*Pitfalls research for: Tournament Viewing UI (v1.1)*
*Researched: 2025-02-02*
