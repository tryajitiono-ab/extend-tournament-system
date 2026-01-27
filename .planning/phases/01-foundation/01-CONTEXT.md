# Phase 1: Foundation - Context

**Gathered:** 2025-01-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Admins can create tournaments and users can authenticate to access the system. Authentication is handled by AccelByte IAM - admins authenticate through AccelByte to access Swagger UI, while game clients act on behalf of users with token validation.

</domain>

<decisions>
## Implementation Decisions

### Authentication patterns
- Admins authenticate through AccelByte IAM to access Swagger UI for tournament management
- Game clients pass user tokens, Tournament System validates user tokens to ensure they come from actual users
- Tournament System then makes changes on behalf of users using service tokens (if needed)
- Authentication sessions should be long-lived (24-48 hours) for gaming convenience
- Claude's Discretion: Error handling for authentication failures, session management implementation

### Tournament creation flow
- Single comprehensive API endpoint (POST /tournaments) for tournament creation
- Basic validation only: required fields, valid dates, positive participant count
- All configuration in creation payload: name, description, max participants, dates in one request
- Claude's Discretion: Specific field validation rules, error message formats

### Admin vs user permissions
- Permission determination based on AccelByte roles/permissions
- Admins have full CRUD + system management permissions (create, read, update, delete, cancel, start tournaments)
- Regular users have read-only access in Phase 1 (view tournaments and details)
- Claude's Discretion: Specific AccelByte permission mappings, permission error responses

### Tournament browsing experience
- Paginated list with filtering support for game clients
- Tournament list includes basic + key details: ID, name, status, participant count, description, start/end dates, max participants
- Filtering options include status + date range filtering
- Claude's Discretion: Pagination size, default sorting order, specific status values

</decisions>

<specifics>
## Specific Ideas

- Admins will use Swagger UI (/apidocs) for tournament management operations
- Game clients (Unreal/Unity) will make API calls on behalf of users
- Tournament System validates user tokens then uses service tokens for backend operations
- Long-lived sessions (24-48 hours) for better gaming experience

</specifics>

<deferred>
## Deferred Ideas

- User registration for tournaments - Phase 2 scope
- Match management and results - Phase 3 scope
- Advanced tournament features - future phases

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2025-01-27*