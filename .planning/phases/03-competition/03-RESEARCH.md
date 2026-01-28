# Phase 3: Competition - Research

**Researched:** 2026-01-28
**Domain:** Go gRPC tournament match management and result tracking system
**Confidence:** HIGH

## Summary

This research focused on implementing automated match management and result tracking for single-elimination tournaments in Go with gRPC and MongoDB. The system already has a solid foundation with tournament CRUD operations, participant registration, authentication, and bracket generation algorithms implemented in Phase 1 and 2.

Key findings indicate that the existing codebase already implements sophisticated bracket generation logic with proper bye handling, status transition validation, and participant management. The system uses MongoDB for flexible tournament data storage with transaction support, gRPC with dual authentication (Bearer + Service tokens), and follows established Go patterns for service architecture.

The primary areas needing implementation are match result submission endpoints, winner advancement logic, match visualization APIs, and tournament progression automation. The existing `GenerateBrackets` function provides the mathematical foundation for bracket creation, and the status transition system ensures proper tournament lifecycle management.

**Primary recommendation:** Extend existing gRPC service with match management endpoints, leverage MongoDB transactions for atomic result submissions, and build upon the established authentication and validation patterns.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| MongoDB Driver | v1.17.3 | Tournament data storage | Flexible document schema, transaction support, battle-tested |
| gRPC | v1.72.0 | API communication | Type-safe contracts, streaming support, industry standard |
| Protocol Buffers | v1.36.6 | Schema definitions | Language-agnostic contracts, validation support |
| Go | 1.24.0 | Backend language | Performance, concurrency, AccelByte ecosystem |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| grpc-gateway/v2 | v2.26.3 | REST API endpoints | Web client integration, OpenAPI docs |
| protovalidate | latest | Message validation | Input validation, business rules enforcement |
| go-grpc-middleware/v2 | v2.3.1 | Interceptors | Authentication, logging, metrics |
| logrus | v1.8.1 | Structured logging | Audit trails, debugging |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| PostgreSQL | MongoDB | SQL provides stronger consistency, but MongoDB offers more flexible bracket data modeling |
| REST only | gRPC + REST | gRPC provides better type safety and performance for internal services |

**Installation:**
```bash
go get go.mongodb.org/mongo-driver@v1.17.3
go get google.golang.org/grpc@v1.72.0
go get github.com/grpc-ecosystem/go-grpc-middleware/v2@v2.3.1
go get github.com/grpc-ecosystem/grpc-gateway/v2@v2.26.3
```

## Architecture Patterns

### Recommended Project Structure
```
pkg/
├── proto/
│   └── tournament.proto          # Extend with match messages
├── service/
│   ├── tournament.go             # Existing tournament service
│   ├── match.go                 # NEW: Match management service
│   └── participant.go           # Existing participant service
├── storage/
│   ├── tournament.go             # Existing tournament storage
│   ├── match.go                 # NEW: Match storage layer
│   └── participant.go           # Existing participant storage
└── server/
    └── tournament.go             # Server integration
```

### Pattern 1: Service Composition Pattern
**What:** Extend existing TournamentServiceServer with match management methods following established delegation patterns
**When to use:** Core service functionality requiring database operations
**Example:**
```go
// Source: Existing tournament.go (lines 24-34)
type TournamentServiceServer struct {
	serviceextension.UnimplementedTournamentServiceServer
	tokenRepo          repository.TokenRepository
	tournamentStorage  storage.TournamentStorage
	matchStorage       storage.MatchStorage     // NEW
	authInterceptor    *extendcustomguildservice.TournamentAuthInterceptor
	logger             *slog.Logger
}
```

### Pattern 2: Bracket Data Modeling Pattern
**What:** Use existing Bracket and BracketData structures with MongoDB document storage
**When to use:** Tournament bracket representation and persistence
**Example:**
```go
// Source: Existing tournament.go (lines 141-157)
type Bracket struct {
	MatchId      string                 `json:"matchId"`
	Round        int32                  `json:"round"`
	Position     int32                  `json:"position"`
	Participant1 *TournamentParticipant `json:"participant1,omitempty"`
	Participant2 *TournamentParticipant `json:"participant2,omitempty"`
	Winner       string                 `json:"winner,omitempty"`
	Bye          bool                   `json:"bye"`
}
```

### Anti-Patterns to Avoid
- **Ad-hoc validation**: Don't implement custom validation - use protovalidate for consistency
- **Direct database access**: Always go through storage layer, don't bypass existing patterns
- **Status bypass**: Never update tournament status without using ValidateStatusTransition

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bracket generation algorithm | Custom power-of-2 calculation | Existing GenerateBrackets method | Handles byes, positioning, edge cases already implemented |
| Authentication logic | Custom token validation | Existing TournamentAuthInterceptor | Dual Bearer/Service token support already implemented |
| Database transactions | Manual MongoDB session handling | Existing transaction patterns in participant storage | Ensures consistency, proper rollback handling |
| Status validation | If/else status checks | ValidateStatusTransition method | Centralized rules, audit logging, maintainable |

**Key insight:** The existing codebase has 777 lines of production-tested tournament logic - extending it is far safer than building parallel systems.

## Common Pitfalls

### Pitfall 1: Match Result Race Conditions
**What goes wrong:** Multiple game servers submit results for the same match simultaneously
**Why it happens:** Concurrent access without proper locking/transaction handling
**How to avoid:** Use MongoDB transactions for atomic result submission and winner validation
**Warning signs:** Duplicate match entries, inconsistent winner advancement

### Pitfall 2: Invalid Status Transitions
**What goes wrong:** Tournament status updated directly without validation
**Why it happens:** Bypassing existing ValidateStatusTransition method for "performance"
**How to avoid:** Always use centralized status transition validation
**Warning signs:** Tournament stuck in invalid states, progression logic broken

### Pitfall 3: Bye Handling Inconsistency
**What goes wrong:** Bracket generation treats bye participants as regular players
**Why it happens:** Not properly checking Bye flag when advancing winners
**How to avoid:** Follow existing bracket generation pattern for bye detection
**Warning signs:** Players advancing from non-existent matches, incorrect round counts

## Code Examples

Verified patterns from existing sources:

### Match Result Submission Pattern
```go
// Source: Pattern based on existing storage/transaction approach
func (s *TournamentServiceServer) SubmitMatchResult(ctx context.Context, req *SubmitMatchResultRequest) (*SubmitMatchResultResponse, error) {
	// Start transaction for atomicity
	session, err := s.matchStorage.StartSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	
	// Transaction callback
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Validate match exists and is in correct state
		match, err := s.matchStorage.GetMatch(sessCtx, req.Namespace, req.TournamentId, req.MatchId)
		if err != nil {
			return nil, err
		}
		
		// Validate winner is actual participant
		if err := s.validateMatchWinner(match, req.WinnerUserId); err != nil {
			return nil, err
		}
		
		// Update match with result
		match.Winner = req.WinnerUserId
		match.CompletedAt = timestamppb.Now()
		
		// Store updated match
		if err := s.matchStorage.UpdateMatch(sessCtx, match); err != nil {
			return nil, err
		}
		
		// Advance winner to next round
		if err := s.advanceWinner(sessCtx, match); err != nil {
			return nil, err
		}
		
		return match, nil
	}
	
	// Execute transaction
	result, err := session.WithTransaction(ctx, callback)
	if err != nil {
		return nil, err
	}
	
	return &SubmitMatchResultResponse{
		Match: result.(*Match),
	}, nil
}
```

### Status Transition Validation
```go
// Source: Existing tournament.go (lines 74-91)
func (s *TournamentServiceServer) ValidateStatusTransition(from, to serviceextension.TournamentStatus) error {
	// Get allowed transitions for the current status
	allowedTransitions := s.GetAllowedStatusTransitions()

	// Check if 'to' status is in the allowed transitions list for the 'from' status
	if allowedTo, exists := allowedTransitions[from]; exists {
		for _, status := range allowedTo {
			if status == to {
				return nil // Transition is allowed
			}
		}
	}

	return grpcStatus.Errorf(codes.InvalidArgument,
		"invalid tournament status transition from %v to %v",
		s.GetStatusName(from),
		s.GetStatusName(to))
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual bracket generation | Mathematical power-of-2 algorithm | Phase 1 | Consistent bye handling, scalable brackets |
| Simple auth | Dual Bearer/Service token pattern | Phase 1 | Secure game server integration |
| Basic CRUD | Status transition validation | Phase 1 | Audit trails, state consistency |

**Deprecated/outdated:**
- Direct status field updates: Use ValidateStatusTransition method
- Separate match databases: Use integrated MongoDB transaction approach

## Open Questions

Things that couldn't be fully resolved:

1. **Frontend Bracket Rendering Technology**
   - What we know: Multiple React bracket libraries exist (react-tournament-brackets, @g-loot/react-tournament-brackets)
   - What's unclear: Specific library choice within Claude's discretion
   - Recommendation: Evaluate react-tournament-brackets for成熟度 and active maintenance

2. **Real-time Updates**
   - What we know: Current system uses REST only, no WebSocket support for v1
   - What's unclear: Whether to implement polling or real-time push for match updates
   - Recommendation: Use REST polling for v1, evaluate WebSocket for v2

3. **Match Storage Schema**
   - What we know: Existing Bracket structures work well, supports bye handling
   - What's unclear: Whether to embed brackets in tournament document or separate collection
   - Recommendation: Separate Match collection for scalability, join queries when needed

## Sources

### Primary (HIGH confidence)
- Existing codebase (777 lines of tournament.go) - Current implementation patterns
- Existing tournament.proto - Established message definitions and service contracts
- MongoDB driver v1.17.3 documentation - Transaction patterns and best practices
- gRPC v1.72.0 documentation - Service patterns and interceptor usage

### Secondary (MEDIUM confidence)
- WebSearch results on tournament bracket algorithms - Confirmed mathematical approach matches existing implementation
- React bracket library research - Verified ecosystem options for frontend visualization

### Tertiary (LOW confidence)
- External tournament system schemas - General MongoDB design patterns applicable but not directly used

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Based on existing go.mod and current implementation
- Architecture: HIGH - Direct analysis of existing 777-line service implementation  
- Pitfalls: HIGH - Identified from existing transaction and status validation patterns

**Research date:** 2026-01-28
**Valid until:** 2026-02-27 (30 days - stable Go/MongoDB ecosystem)