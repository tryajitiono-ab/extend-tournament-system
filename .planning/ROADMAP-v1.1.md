# Roadmap: Tournament Management System - v1.1

**Created:** 2026-02-01  
**Revised:** 2026-02-01  
**Milestone:** v1.1 - Tournament Viewing UI  
**Depth:** Quick (2 phases)  
**Coverage:** 25/25 requirements mapped ✓

## Overview

This roadmap delivers a view-only web UI for the existing tournament management API in 2 phases: (1) Core UI with API integration, and (2) Bracket visualization. This is a straightforward implementation adding static HTML/CSS/JS files to an existing service - no complex backend logic, no build tools, just consuming a working REST API with standard UI patterns.

**Context:** v1.0 delivered complete REST API (tournaments, matches, participants). v1.1 adds web UI for public viewing without user actions or admin controls.

## Phases

### Phase 4 - Core UI & API Integration
**Goal:** Complete tournament viewing UI with list page, detail page, and live data from REST API

**Dependencies:** v1.0 API (complete) - builds on existing REST endpoints

**Requirements:** INFRA-01, INFRA-02, INFRA-03, INFRA-04, LIST-01, LIST-02, LIST-03, LIST-04, DETAIL-01, DETAIL-02, API-01, API-02, API-03, API-04, API-05, API-06, API-07, POLISH-01, POLISH-02, POLISH-03, POLISH-04

**Success Criteria:**
1. User can visit `/tournaments` and see tournament list with real data from REST API
2. User sees tournament cards with name, description, status, participant count, relative timestamps
3. User sees loading states during API calls and error messages when calls fail
4. User can click tournament card to view detail page with metadata and participant list
5. User can manually refresh tournament data
6. UI works on mobile (320px) and desktop (1920px) with responsive layout
7. UI works in modern browsers (Chrome, Firefox, Safari, Edge - last 2 versions)
8. Empty state messages display when no tournaments exist

**Status:** Complete ✓ - Core viewing functionality delivered (16/21 requirements implemented, 5 deferred)
*Completed: 2026-02-02*

**Plans:** 4 plans
- [x] 04-01-PLAN.md — Static file infrastructure with embed.FS and Pico CSS
- [x] 04-02-PLAN.md — Tournament list page with API integration
- [x] 04-03-PLAN.md — Tournament detail page with participant list
- [x] 04-04-PLAN.md — Fix gRPC-Gateway registration (gap closure)

**Deferred Requirements (approved):**
- LIST-03: Color-coded status badges → v1.2 (unimportant)
- API-03: Match data fetching → Phase 5 (bracket visualization scope)
- POLISH-03: Relative timestamps → v1.2 (future improvement)
- INFRA-03: Cache headers → later performance optimization
- POLISH-04: Browser testing → trust modern standards

**Rationale:** This is straightforward static file serving + API consumption. Infrastructure, list UI, detail UI, API client, loading states, and polish can be implemented together since we're adding HTML/CSS/JS files to an existing working service. No complex backend changes, no framework complexity, standard patterns throughout.

**Addresses Pitfalls:**
- Static file serving configuration (proper MIME types, cache headers upfront)
- Tight API-DOM coupling (data transformation layer from start)
- No loading states (skeleton screens built in)
- Mobile layout breakage (mobile-first responsive design)
- Browser compatibility (test across environments early)

---

### Phase 5 - Bracket Visualization
**Goal:** Traditional bracket tree visualization with match status, round labels, and mobile responsiveness

**Dependencies:** Phase 4 (Core UI and API client must be working)

**Requirements:** DETAIL-03, DETAIL-04, DETAIL-05, DETAIL-06, DETAIL-07

**Success Criteria:**
1. User sees single-elimination bracket tree on tournament detail page
2. User can distinguish match status (SCHEDULED/IN_PROGRESS/COMPLETED) with color coding
3. User sees round labels (Round 1, Quarter-Finals, Semi-Finals, Finals)
4. User can identify winners highlighted in completed matches
5. User can view bracket on mobile with horizontal scroll or vertical layout

**Plans:** 3 plans

Plans:
- [ ] 05-01-PLAN.md — Bracket data transformation layer (API client + adapter)
- [ ] 05-02-PLAN.md — Bracket rendering with brackets-viewer.js integration
- [ ] 05-03-PLAN.md — Mobile responsiveness and visual polish

**Rationale:** Bracket visualization is the only complex UI component requiring SVG rendering and layout calculations. Separated from Phase 4 because brackets need the working API client, and the complexity justifies isolated development. Dynamic position calculations and mobile responsiveness need focused attention.

**Addresses Pitfalls:**
- Hardcoded SVG rendering (calculate positions dynamically)
- Deep DOM tree performance (keep structure shallow)
- Mobile bracket layout (horizontal scroll and vertical responsive switching)

---

## Progress

| Phase | Status | Start Date | Complete Date | Notes |
|-------|--------|------------|---------------|-------|
| 4 - Core UI & API | Complete | 2026-02-02 | 2026-02-02 | Core viewing functionality (4 plans, 16/21 requirements, 5 deferred) |
| 5 - Bracket Visualization | Planned | 2026-02-02 | — | Bracket tree + match status + mobile responsiveness (3 plans) |

## Requirement Coverage

| Category | Requirements | Mapped |
|----------|--------------|--------|
| Static Infrastructure | 4 | 4 ✓ |
| Tournament List Page | 4 | 4 ✓ |
| Tournament Detail Page | 7 | 7 ✓ |
| API Integration | 7 | 7 ✓ |
| Production Quality | 4 | 4 ✓ |

**Total:** 25/25 requirements mapped

**Traceability:**
- Phase 4: INFRA-01, INFRA-02, INFRA-03, INFRA-04, LIST-01, LIST-02, LIST-03, LIST-04, DETAIL-01, DETAIL-02, API-01, API-02, API-03, API-04, API-05, API-06, API-07, POLISH-01, POLISH-02, POLISH-03, POLISH-04
- Phase 5: DETAIL-03, DETAIL-04, DETAIL-05, DETAIL-06, DETAIL-07

---

## Phase Ordering Rationale

**Two-phase structure for straightforward implementation:**

**Phase 4 (Core UI & API Integration)** combines infrastructure, list page, detail page, API client, and polish because:
- Static file serving is configuration, not complex development
- Tournament list/detail are standard CRUD UI patterns
- API client uses native fetch with simple JSON transformation
- Loading states and error handling are standard patterns
- All pieces work together as cohesive "viewing experience"
- No framework complexity - plain HTML/CSS/JS

**Phase 5 (Bracket Visualization)** separated because:
- Bracket rendering is the only truly complex UI component
- SVG layout calculations require focused development
- Mobile responsiveness for brackets needs specific attention
- Can build and test in isolation once API client works
- Natural boundary: "can I see tournaments?" vs "can I see brackets?"

**Why not more phases?**
- This is view-only UI consuming existing REST API
- No backend changes needed (API already complete)
- No build tools or framework complexity
- Standard patterns for file serving, API calls, responsive design
- Infrastructure + basic UI + API integration = adding static files to working service
- Only bracket SVG rendering justifies isolated development

---

## Technical Stack

**Core Technologies:**
- Plain HTML5 + Vanilla JavaScript (ES6+) - No build tools, native fetch API
- Pico CSS (2.1.1+) - 10KB semantic CSS framework with dark mode
- brackets-viewer.js (1.9.0+) - Production-ready bracket visualization (213+ stars)
- Go embed package - Static files bundled in binary
- SVG rendering - Scalable bracket visualization

**Architecture:**
- Static files embedded in Go binary via `embed.FS`
- Served alongside existing gRPC-Gateway routes
- Vanilla JavaScript components fetch JSON from REST API
- No framework, no build tools, no separate frontend server

---

## Key Decisions

| Decision | Rationale | Phase |
|----------|-----------|-------|
| Static file serving from Go | Zero dependencies, single binary deployment, reuses existing server | Phase 4 |
| Pico CSS framework | 10KB minimal CSS, classless styling, built-in dark mode | Phase 4 |
| Vanilla JavaScript (no framework) | No build tools constraint, reduces complexity, universal compatibility | Phase 4 |
| Data transformation layer | Prevents API-DOM coupling, easier to handle API changes | Phase 4 |
| brackets-viewer.js library | Production-ready, actively maintained, handles single-elimination | Phase 5 |
| Mobile-first responsive design | 60%+ mobile traffic, prevents expensive retrofitting | Phase 4 |
| Progressive data loading | Avoids 500KB+ payloads, improves perceived performance | Phase 4 |
| 2-phase structure | Straightforward implementation, static files + existing API | All |

---

## Out of Scope (v1.1)

Explicitly deferred to future milestones:

| Feature | Reason | Future Milestone |
|---------|--------|------------------|
| WebSocket real-time updates | Complexity, v1.1 uses manual refresh | v2.0 |
| User registration UI | Requires authentication flow | v1.2 |
| Admin tournament management UI | View-only in v1.1 | v1.2 |
| Match chat/comments | Social features out of scope | v2.0+ |
| Bracket export (PNG/PDF) | Nice to have, not essential | v1.3+ |
| Double-elimination brackets | API only supports single-elimination | v2.0 |
| Search/filter tournaments | Enhanced list features deferred | v1.2 |
| Match detail popups | Enhanced detail features deferred | v1.2 |
| Bracket zoom/pan controls | Large tournament features deferred | v1.2 |

---

## Research Integration

**Research confidence:** HIGH - All technologies verified as production-ready with active maintenance.

**Critical pitfalls addressed:**
1. Bracket layout breaks on mobile (Phase 4 mobile-first design, Phase 5 bracket responsive handling)
2. Hardcoded SVG rendering (Phase 5 addresses with dynamic calculations)
3. Loading entire tournament at once (Phase 4 addresses with progressive loading)
4. No loading states (Phase 4 addresses with skeleton screens)
5. Tight API-DOM coupling (Phase 4 addresses with transformation layer)
6. Static file serving config issues (Phase 4 addresses upfront)

**Competitor analysis:** Analyzed Challonge, Start.gg, Toornament for table stakes features and UX patterns.

---

*Roadmap created: 2026-02-01*  
*Roadmap revised: 2026-02-01 (consolidated from 4 phases to 2 phases)*  
*Phases numbered 4-5 (continuing from v1.0 which ended at Phase 3)*  
*Ready for Phase 4 planning*
