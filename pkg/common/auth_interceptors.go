// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// parseJWTClaims decodes a JWT token's payload without verifying the signature.
// The token must already be validated before calling this function.
func parseJWTClaims(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, status.Error(codes.Unauthenticated, "invalid JWT format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to decode JWT payload")
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to parse JWT claims")
	}
	return claims, nil
}

func getUserIdFromJWTClaims(token string) (string, error) {
	claims, err := parseJWTClaims(token)
	if err != nil {
		return "", err
	}
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub, nil
	}

	return "", status.Error(codes.Unauthenticated, "user ID (sub) not found in JWT claims")
}

const defaultNamespace = "accelbyte"

// TournamentAuthInterceptor provides authentication and authorization for tournament operations
type TournamentAuthInterceptor struct {
	oauthService iam.OAuth20Service
	validator    validator.AuthTokenValidator
	logger       *slog.Logger
}

// NewTournamentAuthInterceptor creates a new tournament auth interceptor
func NewTournamentAuthInterceptor(oauthService iam.OAuth20Service, validator validator.AuthTokenValidator, logger *slog.Logger) *TournamentAuthInterceptor {
	return &TournamentAuthInterceptor{
		oauthService: oauthService,
		validator:    validator,
		logger:       logger,
	}
}

// CheckTournamentPermission validates if a user has the required tournament permission.
// The {namespace} placeholder in requiredPermission.Resource is resolved to namespace before
// the IAM validator is called, so callers should use GetAdminPermission/GetPlayerPermission
// which embed the placeholder rather than a concrete namespace string.
func (t *TournamentAuthInterceptor) CheckTournamentPermission(ctx context.Context, requiredPermission *iam.Permission, namespace string) error {
	if t.validator == nil {
		t.logger.Debug("authentication disabled, skipping permission check")
		return nil
	}

	resolved := resolvePermission(requiredPermission, namespace)

	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return status.Error(codes.Unauthenticated, "metadata is missing")
	}

	if authHeaders, ok := meta["authorization"]; ok && len(authHeaders) > 0 {
		authorization := authHeaders[0]
		if !strings.HasPrefix(authorization, "Bearer ") {
			return status.Error(codes.Unauthenticated, "invalid authorization header format")
		}
		token := strings.TrimPrefix(authorization, "Bearer ")
		return t.validateToken(ctx, token, resolved, namespace)
	}

	if token := extractTokenFromCookieMetadata(meta); token != "" {
		return t.validateToken(ctx, token, resolved, namespace)
	}

	return status.Error(codes.Unauthenticated, "authorization header or cookie is missing")
}

// resolvePermission returns a copy of p with {namespace} replaced by the actual namespace.
func resolvePermission(p *iam.Permission, namespace string) *iam.Permission {
	return &iam.Permission{
		Action:   p.Action,
		Resource: strings.ReplaceAll(p.Resource, "{namespace}", namespace),
	}
}

// validateToken validates a user Bearer token against the required permission and namespace.
func (t *TournamentAuthInterceptor) validateToken(ctx context.Context, token string, requiredPermission *iam.Permission, namespace string) error {
	var userID *string
	if claims, err := parseJWTClaims(token); err == nil {
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			userID = &sub
		}
	}

	if err := t.validator.Validate(token, requiredPermission, &namespace, userID); err != nil {
		t.logger.Warn("token validation failed", "error", err, "namespace", namespace)
		return status.Error(codes.PermissionDenied, err.Error())
	}

	t.logger.Debug("user token validated successfully", "namespace", namespace)
	return nil
}

// GetAdminPermission returns a permission for admin-only tournament operations.
// Resource: ADMIN:NAMESPACE:{namespace}:EXTEND:APPUI
// Actions (bitmask): CREATE=1, READ=2, UPDATE=4, DELETE=8
func (t *TournamentAuthInterceptor) GetAdminPermission(operation string) *iam.Permission {
	return &iam.Permission{
		Action:   actionBitmask(operation),
		Resource: "ADMIN:NAMESPACE:{namespace}:EXTEND:APPUI",
	}
}

// GetPlayerPermission returns a permission for player (non-admin) tournament operations.
// Resource: NAMESPACE:{namespace}:EXTEND:TOURNAMENT — requires user role override in the AccelByte console.
// Actions (bitmask): CREATE=1, READ=2, UPDATE=4, DELETE=8
func (t *TournamentAuthInterceptor) GetPlayerPermission(operation string) *iam.Permission {
	return &iam.Permission{
		Action:   actionBitmask(operation),
		Resource: "NAMESPACE:{namespace}:EXTEND:TOURNAMENT",
	}
}

// actionBitmask converts an operation name to its IAM action bitmask value.
func actionBitmask(operation string) int {
	switch strings.ToUpper(operation) {
	case "CREATE":
		return 1
	case "READ":
		return 2
	case "UPDATE":
		return 4
	case "DELETE":
		return 8
	default:
		panic("invalid operation")
	}
}

// TournamentUnaryInterceptor returns a unary interceptor for tournament operations
func (t *TournamentAuthInterceptor) TournamentUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		namespace := t.extractNamespaceFromRequest(req)
		if namespace == "" {
			return nil, status.Error(codes.InvalidArgument, "namespace is required")
		}

		permission := t.permissionForMethod(info.FullMethod)
		if permission == nil {
			t.logger.Warn("unknown method, skipping auth", "method", info.FullMethod)
			return handler(ctx, req)
		}

		if err := t.CheckTournamentPermission(ctx, permission, namespace); err != nil {
			return nil, err
		}

		t.logger.Debug("tournament operation authorized", "method", info.FullMethod, "namespace", namespace)
		return handler(ctx, req)
	}
}

// TournamentStreamInterceptor returns a stream interceptor for tournament operations
func (t *TournamentAuthInterceptor) TournamentStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(srv, ss)
		}

		namespace := t.extractNamespaceFromContext(ss.Context())
		if namespace == "" {
			return status.Error(codes.InvalidArgument, "namespace is required")
		}

		permission := t.permissionForMethod(info.FullMethod)
		if permission == nil {
			t.logger.Warn("unknown method, skipping auth", "method", info.FullMethod)
			return handler(srv, ss)
		}

		if err := t.CheckTournamentPermission(ss.Context(), permission, namespace); err != nil {
			return err
		}

		t.logger.Debug("tournament stream operation authorized", "method", info.FullMethod, "namespace", namespace)
		return handler(srv, ss)
	}
}

// permissionForMethod returns the correct admin or player permission for a gRPC method.
// Returns nil for unknown methods (auth is skipped).
func (t *TournamentAuthInterceptor) permissionForMethod(fullMethod string) *iam.Permission {
	parts := strings.Split(fullMethod, "/")
	methodName := parts[len(parts)-1]
	switch methodName {
	// Admin-only operations
	case "AdminCreateTournament":
		return t.GetAdminPermission("CREATE")
	case "AdminStartTournament", "AdminCancelTournament", "AdminActivateTournament", "AdminSubmitMatchResult":
		return t.GetAdminPermission("UPDATE")
	case "AdminRemoveParticipant":
		return t.GetAdminPermission("DELETE")
	case "AdminListTournaments", "AdminGetTournament":
		return t.GetAdminPermission("READ")
	// Player operations
	case "GetTournament", "ListTournaments", "GetTournamentParticipants", "GetTournamentMatches", "GetMatch":
		return t.GetPlayerPermission("READ")
	case "RegisterForTournament":
		return t.GetPlayerPermission("CREATE")
	default:
		return nil
	}
}

// extractNamespaceFromRequest attempts to extract namespace from request
func (t *TournamentAuthInterceptor) extractNamespaceFromRequest(req interface{}) string {
	// Use reflection or type assertion to extract namespace
	// This is a simplified version - in practice, you'd want to handle different request types
	switch v := req.(type) {
	case interface{ GetNamespace() string }:
		return v.GetNamespace()
	case interface{ GetNamespace() *string }:
		if ns := v.GetNamespace(); ns != nil {
			return *ns
		}
	default:
		// Try to extract from context as fallback
		return t.extractNamespaceFromContext(context.Background())
	}
	return ""
}

// extractNamespaceFromContext attempts to extract namespace from context
func (t *TournamentAuthInterceptor) extractNamespaceFromContext(ctx context.Context) string {
	// Try to get namespace from context metadata
	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		if nsHeaders := meta["namespace"]; len(nsHeaders) > 0 {
			return nsHeaders[0]
		}
	}

	// Fallback to environment variable
	return GetEnv("AB_NAMESPACE", defaultNamespace)
}

// extractTokenFromCookieMetadata parses the "cookie" metadata key and returns the access_token value if present.
func extractTokenFromCookieMetadata(meta metadata.MD) string {
	for _, cookieHeader := range meta.Get("cookie") {
		req := &http.Request{Header: http.Header{"Cookie": {cookieHeader}}}
		if cookie, err := req.Cookie("access_token"); err == nil {
			return cookie.Value
		}
	}
	return ""
}

// GetContextNamespace extracts namespace from request context
func GetContextNamespace(ctx context.Context) (string, error) {
	// Extract namespace from request metadata
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		// When metadata is missing and auth is disabled, return default namespace
		// This commonly happens during REST API calls through gRPC-Gateway without auth headers
		return GetEnv("AB_NAMESPACE", defaultNamespace), nil
	}

	// Try to extract namespace from various possible metadata sources
	if nsHeaders := meta["namespace"]; len(nsHeaders) > 0 {
		return nsHeaders[0], nil
	}

	// Try from authorization token or cookie if available
	token := ""
	if authHeaders := meta["authorization"]; len(authHeaders) > 0 {
		authorization := authHeaders[0]
		if strings.HasPrefix(authorization, "Bearer ") {
			token = strings.TrimPrefix(authorization, "Bearer ")
		}
	}
	if token == "" {
		token = extractTokenFromCookieMetadata(meta)
	}
	if token != "" {
		// For now, return default namespace since token parsing would require additional IAM integration
		// In a full implementation, you'd parse the JWT token to extract the namespace
		return GetEnv("AB_NAMESPACE", defaultNamespace), nil
	}

	// When no namespace found in metadata, return default namespace
	// This allows unauthenticated REST API access when auth is disabled
	return GetEnv("AB_NAMESPACE", defaultNamespace), nil
}

// GetContextUserID extracts the user ID from the Bearer JWT sub claim or session cookie.
// Client-supplied identity headers (x-user-id) are never trusted.
func GetContextUserID(ctx context.Context) (string, error) {
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return "", status.Error(codes.Unauthenticated, "metadata is missing")
	}

	// Prefer the Bearer token's sub claim.
	if authHeaders := meta["authorization"]; len(authHeaders) > 0 {
		authorization := authHeaders[0]
		if strings.HasPrefix(authorization, "Bearer ") {
			token := strings.TrimPrefix(authorization, "Bearer ")
			return getUserIdFromJWTClaims(token)
		}
	}

	// Fall back to the session cookie.
	if token := extractTokenFromCookieMetadata(meta); token != "" {
		return getUserIdFromJWTClaims(token)
	}

	return "", status.Error(codes.Unauthenticated, "user ID not found in context")
}

// GetContextUsername extracts the username from the Bearer JWT user_name claim or session cookie.
// Client-supplied identity headers (x-username) are never trusted.
func GetContextUsername(ctx context.Context) (string, error) {
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return "", status.Error(codes.Unauthenticated, "metadata is missing")
	}

	// Prefer the Bearer token's user_name claim.
	if authHeaders := meta["authorization"]; len(authHeaders) > 0 {
		authorization := authHeaders[0]
		if strings.HasPrefix(authorization, "Bearer ") {
			token := strings.TrimPrefix(authorization, "Bearer ")
			claims, err := parseJWTClaims(token)
			if err != nil {
				return "", err
			}
			if username, ok := claims["user_name"].(string); ok && username != "" {
				return username, nil
			}
			return "", status.Error(codes.Unauthenticated, "username (user_name) not found in JWT claims")
		}
	}

	// Fall back to the session cookie.
	if token := extractTokenFromCookieMetadata(meta); token != "" {
		claims, err := parseJWTClaims(token)
		if err != nil {
			return "", err
		}
		if username, ok := claims["user_name"].(string); ok && username != "" {
			return username, nil
		}
	}

	return "", status.Error(codes.Unauthenticated, "username not found in context")
}

// IsAdminUser checks if the caller holds an admin role by inspecting the JWT Bearer token
// or session cookie claims. Client-supplied privilege headers (x-is-admin) are never trusted.
// The token must have been validated by the IAM interceptor before this function is called.
func IsAdminUser(ctx context.Context) (bool, error) {
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return false, status.Error(codes.Unauthenticated, "metadata is missing")
	}

	token := ""
	if authHeaders := meta["authorization"]; len(authHeaders) > 0 {
		authorization := authHeaders[0]
		if strings.HasPrefix(authorization, "Bearer ") {
			token = strings.TrimPrefix(authorization, "Bearer ")
		}
	}
	if token == "" {
		token = extractTokenFromCookieMetadata(meta)
	}
	if token == "" {
		return false, nil
	}

	claims, err := parseJWTClaims(token)
	if err != nil {
		return false, err
	}

	// Check the roles array for any entry that indicates admin.
	if roles, ok := claims["roles"].([]interface{}); ok {
		for _, r := range roles {
			if role, ok := r.(string); ok && strings.EqualFold(role, "admin") {
				return true, nil
			}
		}
	}

	return false, nil
}
