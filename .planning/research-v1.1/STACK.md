# Stack Research

**Domain:** Tournament viewing UI (plain HTML/CSS/JS)
**Researched:** 2026-02-02
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Plain HTML5 | Latest | Structure tournament pages | No build step required, universal browser support, semantic markup for accessibility |
| Vanilla JavaScript | ES6+ (2015+) | API integration and DOM manipulation | Native fetch API available, no framework overhead, runs directly in all modern browsers |
| Pico CSS | 2.1.1+ | Responsive styling framework | 10KB minified, classless/semantic HTML styling, built-in dark mode, mobile-first responsive design |
| brackets-viewer.js | 1.9.0+ | Tournament bracket visualization | Production-ready library for single/double elimination brackets, 213+ GitHub stars, active maintenance, CDN available |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| brackets-viewer.js | 1.9.0+ | Render bracket tree visualization | Required for traditional bracket display on tournament detail page |
| Pico CSS | 2.1.1+ | Base responsive styling | Use for tournament list and layout structure, provides clean default styles |
| None required for API | Native Fetch API | HTTP requests to REST endpoints | Built into all modern browsers (2015+), no axios/jquery needed |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Go http.ServeMux | Serve static files | Use `http.Handle()` with `http.FileServer()` in main.go |
| Browser DevTools | Testing and debugging | Native browser tools sufficient for vanilla JS debugging |
| No build tools | Direct file editing | HTML/CSS/JS served directly, no webpack/npm needed |

## Installation

### Via CDN (Recommended for v1.1)

```html
<!-- In your HTML <head> -->

<!-- Pico CSS for base styling -->
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">

<!-- brackets-viewer.js for bracket visualization -->
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/brackets-viewer@1.9.0/dist/brackets-viewer.min.css">
<script src="https://cdn.jsdelivr.net/npm/brackets-viewer@1.9.0/dist/brackets-viewer.min.js"></script>
```

### Go Static File Serving

```go
// In main.go, add to existing HTTP server setup
func newGRPCGatewayHTTPServer(...) *http.Server {
    mux := http.NewServeMux()
    
    // Existing gRPC-Gateway handler
    mux.Handle("/", handler)
    
    // NEW: Serve static files from /static directory
    fs := http.FileServer(http.Dir("./static"))
    mux.Handle("/static/", http.StripPrefix("/static/", fs))
    
    // Serve tournament UI at root or /tournaments path
    mux.HandleFunc("/tournaments", serveIndexPage)
    
    return &http.Server{Addr: addr, Handler: mux}
}
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Pico CSS | Simple.css | If you want even lighter weight (~4KB) but less polish |
| Pico CSS | No framework (custom CSS) | For complete design control or brand-specific styling |
| brackets-viewer.js | Custom SVG/Canvas rendering | Only if you need highly customized bracket appearance |
| Plain HTML/CSS/JS | React/Vue with build tools | Future milestone if UI becomes interactive (editing, real-time updates) |
| CDN delivery | Local file copies | If offline/air-gapped deployment required |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| jQuery | 87KB overhead for features now in vanilla JS | Native fetch API, querySelector, classList |
| Bootstrap | 144KB bloat, requires extensive class usage | Pico CSS (10KB, semantic) |
| npm/webpack/build tools | Violates "no build step" constraint | CDN links in HTML files |
| React/Vue/Angular | Framework overhead for read-only views | Plain JavaScript with DOM manipulation |
| Toornament API libraries | Third-party dependency for own API | Native fetch with JSON parsing |
| WebSockets | Out of scope for v1.1 (manual refresh only) | Regular fetch calls on page load |

## Stack Patterns by Component

### Tournament List Page (`/tournaments` or `/static/index.html`)

**Structure:**
- Use Pico CSS grid/table for responsive tournament cards
- Vanilla JS fetch to GET `/v1/public/namespace/{namespace}/tournaments`
- Display tournament status, name, participant count
- Link to detail page for each tournament

**Pattern:**
```html
<main class="container">
  <h1>Tournaments</h1>
  <div id="tournament-list">
    <!-- Dynamically populated via JS -->
  </div>
</main>

<script>
async function loadTournaments() {
  const response = await fetch('/v1/public/namespace/default/tournaments');
  const data = await response.json();
  renderTournaments(data.tournaments);
}
</script>
```

### Tournament Detail Page (`/static/tournament.html?id={tournament_id}`)

**Structure:**
- Header section with tournament info (Pico CSS semantic elements)
- Participants list (Pico CSS table or cards)
- Bracket visualization (brackets-viewer.js)
- Match results below bracket

**Pattern:**
```html
<main class="container">
  <header>
    <h1 id="tournament-name"></h1>
    <p id="tournament-status"></p>
  </header>
  
  <section>
    <h2>Bracket</h2>
    <div class="brackets-viewer"></div>
  </section>
  
  <section>
    <h2>Participants</h2>
    <div id="participants"></div>
  </section>
</main>

<script>
window.bracketsViewer.render({
  stages: data.stage,
  matches: data.match,
  matchGames: data.match_game,
  participants: data.participant,
}, {
  selector: '.brackets-viewer'
});
</script>
```

### Mobile Responsiveness

**Pico CSS provides:**
- Automatic font scaling based on viewport
- Responsive containers (max-width constraints)
- Stack columns on mobile automatically
- Touch-friendly button/link sizing

**Custom additions needed:**
- Horizontal scroll container for wide brackets on mobile
- Hide/show sections with `<details>` elements for long pages
- Larger touch targets for bracket matches

## API Integration Pattern

### REST Endpoints to Consume

Based on existing gRPC-Gateway REST API:

| Endpoint | Method | Purpose | Page |
|----------|--------|---------|------|
| `/v1/public/namespace/{ns}/tournaments` | GET | List all tournaments | List page |
| `/v1/public/namespace/{ns}/tournaments/{id}` | GET | Tournament details | Detail page |
| `/v1/public/namespace/{ns}/tournaments/{id}/participants` | GET | Participant list | Detail page |
| `/v1/public/namespace/{ns}/tournaments/{id}/matches` | GET | All matches for bracket | Detail page |

### JavaScript Fetch Pattern

```javascript
// Centralized API configuration
const API_BASE = '/v1/public/namespace/default';

// Generic fetch wrapper with error handling
async function apiGet(path) {
  try {
    const response = await fetch(`${API_BASE}${path}`);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    return await response.json();
  } catch (error) {
    console.error('API Error:', error);
    showError(`Failed to load data: ${error.message}`);
    return null;
  }
}

// Usage
const tournament = await apiGet(`/tournaments/${tournamentId}`);
```

## Static File Structure

```
/workspace/
├── static/
│   ├── index.html              # Tournament list page
│   ├── tournament.html         # Tournament detail page
│   ├── css/
│   │   └── custom.css          # Custom overrides (optional)
│   └── js/
│       ├── tournaments.js      # List page logic
│       ├── tournament-detail.js # Detail page logic
│       └── api.js              # Shared API utilities
└── main.go                     # Updated to serve /static
```

**Notes:**
- CDN libraries loaded from HTML files (not copied locally)
- No `/static/lib` or `/static/vendor` directories needed
- All custom JS in separate files, not inline
- CSS customizations minimal (Pico CSS provides 95% of needed styles)

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| brackets-viewer.js@1.9.0 | All modern browsers (2015+) | Uses ES6 features, no IE11 support needed |
| Pico CSS@2.1.1 | All modern browsers | CSS Grid and Custom Properties required |
| Native Fetch API | Chrome 42+, Firefox 39+, Safari 10.1+, Edge 14+ | 99%+ browser coverage |
| Go http.FileServer | Go 1.16+ | Embedding support available if needed |

## Browser Support Target

**Minimum supported browsers:**
- Chrome/Edge: Last 2 versions
- Firefox: Last 2 versions  
- Safari: Last 2 versions
- Mobile Safari/Chrome: Last 2 versions

**NOT supporting:**
- Internet Explorer (any version)
- Legacy Edge (<79)
- Opera Mini

**Justification:** AccelByte gaming audience uses modern browsers; tournament viewers expect modern web experience.

## Progressive Enhancement Strategy

**If JavaScript disabled:**
- Show message: "JavaScript required for tournament viewing"
- Static HTML structure still renders with Pico CSS
- Tournament data requires JS to fetch from API

**If CDN unavailable:**
- brackets-viewer.js: Show error message, bracket won't render
- Pico CSS: Browser default styles as fallback (readable but plain)
- **Future:** Consider embedding critical CSS inline

**If API unavailable:**
- Show friendly error message with retry button
- Don't break page layout
- Log errors to console for debugging

## Security Considerations

**Static files:**
- No user-generated content rendered without sanitization
- Use `textContent` not `innerHTML` for tournament names
- Validate tournament IDs from URL params (alphanumeric only)

**API calls:**
- Public read-only endpoints only
- No authentication tokens in frontend code
- CORS headers must be configured on Go server

**Example secure rendering:**
```javascript
// SAFE: Using textContent
element.textContent = tournament.name;

// UNSAFE: Using innerHTML with user data
element.innerHTML = tournament.name; // DON'T DO THIS
```

## Sources

- [brackets-viewer.js GitHub](https://github.com/Drarig29/brackets-viewer.js) — Production-ready bracket library, 213 stars, active maintenance
- [Pico CSS Official](https://picocss.com) — Modern semantic CSS framework, 14.8K GitHub stars
- [MDN Web Docs - Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) — Standard browser API documentation
- [Go net/http Documentation](https://pkg.go.dev/net/http) — Static file serving patterns
- Existing codebase analysis — API endpoints from service.proto and main.go
- AccelByte Extend environment — Deployment constraints and requirements

---
*Stack research for: Tournament Viewing UI v1.1*
*Researched: 2026-02-02*
*Confidence: HIGH - All technologies validated as production-ready and constraint-compliant*
