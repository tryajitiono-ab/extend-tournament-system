# Plan 04-01 Summary: Static File Infrastructure

**Executed:** 2026-02-02  
**Phase:** 04-core-ui-api-integration  
**Plan:** 01  
**Status:** ✓ Complete

## Objective

Set up static file infrastructure for tournament viewing UI. Add Go embed filesystem, configure HTTP routes for static assets and HTML templates, integrate Pico CSS framework for minimal styling.

## What Was Built

### Task 1: Web Directory Structure and Pico CSS
**Commit:** `a613865` - feat(ui): add web directory structure and Pico CSS framework

Created complete web asset directory structure:
- `web/static/css/` - CSS files (Pico CSS v2.0.6, 81KB minified)
- `web/static/js/` - JavaScript files (.gitkeep for now)
- `web/templates/` - HTML templates (base.html)

**Base HTML Template:**
- Minimal HTML5 structure with DOCTYPE, meta charset, viewport
- Links to `/static/css/pico.min.css` stylesheet
- Body with `<main class="container">` wrapping `{{.Content}}` placeholder
- Mobile-responsive viewport meta tag

### Task 2: Static File Serving Routes
**Commit:** `4dfda83` - feat(ui): add embed.FS and static file serving routes

Added static file infrastructure to `main.go`:
- **Imports:** embed, html/template, io/fs
- **Embed directives:** `//go:embed web/static` and `//go:embed web/templates`
- **Static route:** `/static/*` serving from embedded filesystem with http.FileServer
- **Tournaments route:** `/tournaments` rendering base.html template with Go templates
- **Routing order:** Static routes placed before catch-all gRPC-Gateway handler

**Implementation details:**
- Static files extracted from embed.FS using fs.Sub()
- CSS served with correct `text/css; charset=utf-8` MIME type
- Template parsing inline (single template, no complexity)
- Placeholder content: `<h1>Tournaments</h1><p>Loading...</p>`

## Verification Results

**Infrastructure checks (all passing):**
```bash
✓ http://localhost:8000/tournaments - receives HTML page (200 OK)
✓ View source - Pico CSS stylesheet linked at /static/css/pico.min.css
✓ http://localhost:8000/static/css/pico.min.css - loads with 200 OK and text/css MIME type
✓ Page displays with Pico CSS default styling (typography, spacing)
```

**File checks (all passing):**
```bash
✓ web/static/css/pico.min.css exists (81KB)
✓ web/templates/base.html exists (minimal HTML5 structure)
✓ main.go contains embed.FS directives
✓ main.go contains /static/ route
✓ main.go contains /tournaments route
```

**Build verification:**
```bash
✓ Service compiles without errors: go build -o tournament-service .
✓ Service runs successfully with MongoDB connection
✓ Static files embedded in binary (no external file dependencies)
```

## Requirements Satisfied

- [x] **INFRA-01**: Go service serves static HTML/CSS/JS files from embedded filesystem
- [x] **INFRA-02**: Static routes configured (/tournaments, /static/*) alongside existing API routes
- [x] **INFRA-03**: Proper MIME types for static files (text/css verified)
- [x] **INFRA-04**: Mobile-responsive CSS framework integrated (Pico CSS v2.0.6)

## Technical Notes

### MongoDB Connection Issue Resolved
During execution, discovered MongoDB replica set configuration causing connection issues. Solution: added `?directConnection=true` to MONGODB_URI connection string to bypass replica set discovery.

**Working configuration:**
```bash
export MONGODB_URI="mongodb://192.168.97.2:27017/?directConnection=true"
export PLUGIN_GRPC_SERVER_AUTH_ENABLED=false
export BASE_PATH="/"
```

### Routing Order
Critical implementation detail: static routes MUST be registered before the catch-all `mux.Handle("/", handler)` for gRPC-Gateway. Order:
1. `/static/*` - static files
2. `/tournaments` - HTML template
3. `/apidocs/` - Swagger UI (existing)
4. `/apidocs/api.json` - Swagger JSON (existing)
5. `/` - gRPC-Gateway catch-all (existing)

### Embed.FS Benefits
- Zero external dependencies - all web assets bundled in binary
- Single binary deployment unchanged
- No file system access required at runtime
- Works seamlessly with existing gRPC-Gateway HTTP server

## Next Steps

**Plan 04-02** will add:
- JavaScript API client for fetching tournament data
- Tournament list page with cards
- Loading states and error handling
- Actual tournament data from REST API (replacing placeholder content)

## Files Modified

- `web/static/css/pico.min.css` (new) - 81KB minified CSS framework
- `web/static/js/.gitkeep` (new) - preserve directory structure
- `web/templates/base.html` (new) - minimal HTML5 template
- `main.go` (modified) - add embed.FS, static routes, tournaments route

## Commits

1. `a613865` - feat(ui): add web directory structure and Pico CSS framework
2. `4dfda83` - feat(ui): add embed.FS and static file serving routes

---

*Plan 04-01 completed: 2026-02-02*  
*All success criteria met ✓*  
*Ready for Plan 04-02: Tournament list page with API integration*
