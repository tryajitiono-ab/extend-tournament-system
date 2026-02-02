# Requirements: Tournament Management System - v1.1

**Defined:** 2026-02-01
**Milestone:** v1.1 - Tournament Viewing UI
**Core Value:** Players can compete in organized tournaments with automated bracket management and real-time result tracking.

## v1.1 Requirements

Requirements for tournament viewing UI milestone. Each maps to roadmap phases.

### Static Infrastructure

- [ ] **INFRA-01**: Go service serves static HTML/CSS/JS files from embedded filesystem
- [ ] **INFRA-02**: Static routes configured (/tournaments, /static/*) alongside existing API routes
- [ ] **INFRA-03**: Proper MIME types and cache headers for static files
- [ ] **INFRA-04**: Mobile-responsive CSS framework integrated (Pico CSS)

### Tournament List Page

- [ ] **LIST-01**: User can view grid/list of all tournaments
- [ ] **LIST-02**: Each tournament displays name, description, status, participant count
- [ ] **LIST-03**: Tournament status badges (DRAFT/ACTIVE/STARTED/COMPLETED/CANCELLED) display with color coding
- [ ] **LIST-04**: User can click tournament card to view details

### Tournament Detail Page

- [ ] **DETAIL-01**: User can view tournament information (name, description, status, participant count)
- [ ] **DETAIL-02**: User can view list of registered participants
- [ ] **DETAIL-03**: User can view traditional bracket tree visualization for single-elimination tournaments
- [ ] **DETAIL-04**: Bracket displays match status indicators (SCHEDULED/IN_PROGRESS/COMPLETED)
- [ ] **DETAIL-05**: Bracket displays round labels (Round 1, Quarter-Finals, Semi-Finals, Finals)
- [ ] **DETAIL-06**: Bracket highlights winners in completed matches
- [ ] **DETAIL-07**: Bracket is mobile-responsive (horizontal scroll or vertical layout)

### API Integration

- [ ] **API-01**: JavaScript API client fetches tournament list from REST endpoint
- [ ] **API-02**: JavaScript API client fetches tournament details from REST endpoint
- [ ] **API-03**: JavaScript API client fetches match data from REST endpoint
- [ ] **API-04**: JavaScript API client fetches participant data from REST endpoint
- [ ] **API-05**: Data transformation layer separates API responses from UI rendering
- [ ] **API-06**: Loading states display during API calls (skeleton screens/spinners)
- [ ] **API-07**: Error messages display when API calls fail

### Production Quality

- [ ] **POLISH-01**: User can manually refresh tournament data
- [ ] **POLISH-02**: Empty state messages display when no tournaments exist
- [ ] **POLISH-03**: Date/time fields format as relative timestamps ("2 hours ago")
- [ ] **POLISH-04**: UI works in modern browsers (Chrome, Firefox, Safari, Edge - last 2 versions)

## v1.2+ Requirements

Deferred to future releases. Tracked but not in v1.1 roadmap.

### Enhanced List Page

- **LIST-05**: User can search tournaments by name
- **LIST-06**: User can filter tournaments by status
- **LIST-07**: "Live" indicator highlights active tournaments

### Enhanced Detail Page

- **DETAIL-08**: User can click match to view detailed popup
- **DETAIL-09**: User can zoom/pan large brackets (32+ participants)
- **DETAIL-10**: User can toggle round-by-round view for mobile

### Enhanced Polish

- **POLISH-05**: "Updated X ago" indicator shows data freshness
- **POLISH-06**: SEO meta tags for tournament pages
- **POLISH-07**: Tournament page URLs are shareable with preview cards

### User Actions (Future Milestone)

- **ACTION-01**: User can register for tournament via UI
- **ACTION-02**: User can withdraw from tournament via UI
- **ACTION-03**: User can login/authenticate via UI
- **ACTION-04**: User profile shows registered tournaments

### Admin UI (Future Milestone)

- **ADMIN-01**: Admin can create tournament via UI
- **ADMIN-02**: Admin can start/cancel tournament via UI
- **ADMIN-03**: Admin can submit match results via UI
- **ADMIN-04**: Admin dashboard shows all tournaments

### Real-time Features (v2.0+)

- **REALTIME-01**: Bracket auto-updates via WebSocket
- **REALTIME-02**: Live match indicators show active games
- **REALTIME-03**: Push notifications for match updates

## Out of Scope

Explicitly excluded from v1.1. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| WebSocket real-time updates | v1.1 uses manual refresh, defer to v2.0 for complexity |
| User registration UI | Requires authentication flow, separate milestone v1.2 |
| Admin tournament management UI | View-only in v1.1, admin UI separate milestone |
| Match chat/comments | Social features out of scope for viewing milestone |
| Bracket export (PNG/PDF) | Nice to have but not essential for viewing |
| Double-elimination brackets | API only supports single-elimination in v1.0 |
| Swiss/round-robin formats | Not supported by backend API yet |
| Build tools (webpack, npm) | Constraint: plain HTML/CSS/JS only |
| Frontend framework (React/Vue) | Constraint: vanilla JavaScript only |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| INFRA-01 | TBD | Pending |
| INFRA-02 | TBD | Pending |
| INFRA-03 | TBD | Pending |
| INFRA-04 | TBD | Pending |
| LIST-01 | TBD | Pending |
| LIST-02 | TBD | Pending |
| LIST-03 | TBD | Pending |
| LIST-04 | TBD | Pending |
| DETAIL-01 | TBD | Pending |
| DETAIL-02 | TBD | Pending |
| DETAIL-03 | TBD | Pending |
| DETAIL-04 | TBD | Pending |
| DETAIL-05 | TBD | Pending |
| DETAIL-06 | TBD | Pending |
| DETAIL-07 | TBD | Pending |
| API-01 | TBD | Pending |
| API-02 | TBD | Pending |
| API-03 | TBD | Pending |
| API-04 | TBD | Pending |
| API-05 | TBD | Pending |
| API-06 | TBD | Pending |
| API-07 | TBD | Pending |
| POLISH-01 | TBD | Pending |
| POLISH-02 | TBD | Pending |
| POLISH-03 | TBD | Pending |
| POLISH-04 | TBD | Pending |

**Coverage:**
- v1.1 requirements: 25 total
- Mapped to phases: 0 (pending roadmap creation)
- Unmapped: 25 ⚠️

---
*Requirements defined: 2026-02-01*
*Last updated: 2026-02-01 after research completion*
