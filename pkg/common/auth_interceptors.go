// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	"log/slog"
	"strings"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

// CheckTournamentPermission validates if a user has the required tournament permission
func (t *TournamentAuthInterceptor) CheckTournamentPermission(ctx context.Context, requiredPermission *iam.Permission, namespace string) error {
	// If validator is nil, authentication is disabled (testing mode)
	if t.validator == nil {
		t.logger.Debug("authentication disabled, skipping permission check")
		return nil
	}

	// Extract token from metadata
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return status.Error(codes.Unauthenticated, "metadata is missing")
	}

	// Check for Bearer token (user authentication)
	if authHeaders, ok := meta["authorization"]; ok && len(authHeaders) > 0 {
		authorization := authHeaders[0]
		if !strings.HasPrefix(authorization, "Bearer ") {
			return status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		token := strings.TrimPrefix(authorization, "Bearer ")
		return t.validateToken(ctx, token, requiredPermission, namespace)
	}

	// Check for Service token (game server authentication)
	if serviceHeaders, ok := meta["x-service-token"]; ok && len(serviceHeaders) > 0 {
		serviceToken := serviceHeaders[0]
		return t.validateServiceToken(ctx, serviceToken, requiredPermission, namespace)
	}

	// Check for token in cookies (browser-based authentication via gRPC-Gateway)
	if token := extractTokenFromCookieMetadata(meta); token != "" {
		return t.validateToken(ctx, token, requiredPermission, namespace)
	}

	return status.Error(codes.Unauthenticated, "authorization header is missing")
}

// validateToken validates user Bearer token and permissions
func (t *TournamentAuthInterceptor) validateToken(ctx context.Context, token string, requiredPermission *iam.Permission, namespace string) error {
	// Validate token with AccelByte IAM
	err := t.validator.Validate(token, requiredPermission, &namespace, nil)
	if err != nil {
		t.logger.Warn("token validation failed", "error", err, "namespace", namespace)
		return status.Error(codes.PermissionDenied, err.Error())
	}

	t.logger.Debug("user token validated successfully", "namespace", namespace)
	return nil
}

// validateServiceToken validates service token for game server access
func (t *TournamentAuthInterceptor) validateServiceToken(ctx context.Context, serviceToken string, requiredPermission *iam.Permission, namespace string) error {
	// For service tokens, we typically validate against a different set of permissions
	// Service tokens usually have broader access for system operations
	// For now, we'll validate service tokens with the same permission structure
	// but this could be extended to have service-specific permissions

	err := t.validator.Validate(serviceToken, requiredPermission, &namespace, nil)
	if err != nil {
		t.logger.Warn("service token validation failed", "error", err, "namespace", namespace)
		return status.Error(codes.PermissionDenied, err.Error())
	}

	t.logger.Debug("service token validated successfully", "namespace", namespace)
	return nil
}

// GetTournamentPermission returns the required permission for a tournament operation
func (t *TournamentAuthInterceptor) GetTournamentPermission(operation string, namespace string) *iam.Permission {
	return &iam.Permission{
		Action:   t.getActionValue(operation),
		Resource: "ADMIN:NAMESPACE:" + namespace + ":EXTEND:ADMINUI",
	}
}

// getActionValue converts operation string to permission action value
func (t *TournamentAuthInterceptor) getActionValue(operation string) int {
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
		return 1
	}
}

// TournamentUnaryInterceptor returns a unary interceptor for tournament operations
func (t *TournamentAuthInterceptor) TournamentUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for health check endpoints
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		// Extract namespace from request if possible
		namespace := t.extractNamespaceFromRequest(req)
		if namespace == "" {
			return nil, status.Error(codes.InvalidArgument, "namespace is required")
		}

		// Determine operation from method name
		operation := t.extractOperationFromMethod(info.FullMethod)
		if operation == "" {
			t.logger.Warn("unknown operation, skipping auth", "method", info.FullMethod)
			return handler(ctx, req)
		}

		// Get required permission
		requiredPermission := t.GetTournamentPermission(operation, namespace)

		// Check permission
		if err := t.CheckTournamentPermission(ctx, requiredPermission, namespace); err != nil {
			return nil, err
		}

		t.logger.Debug("tournament operation authorized",
			"operation", operation,
			"namespace", namespace,
			"method", info.FullMethod)

		return handler(ctx, req)
	}
}

// TournamentStreamInterceptor returns a stream interceptor for tournament operations
func (t *TournamentAuthInterceptor) TournamentStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip auth for health check endpoints
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(srv, ss)
		}

		// Extract namespace from context if possible
		namespace := t.extractNamespaceFromContext(ss.Context())
		if namespace == "" {
			return status.Error(codes.InvalidArgument, "namespace is required")
		}

		// Determine operation from method name
		operation := t.extractOperationFromMethod(info.FullMethod)
		if operation == "" {
			t.logger.Warn("unknown operation, skipping auth", "method", info.FullMethod)
			return handler(srv, ss)
		}

		// Get required permission
		requiredPermission := t.GetTournamentPermission(operation, namespace)

		// Check permission
		if err := t.CheckTournamentPermission(ss.Context(), requiredPermission, namespace); err != nil {
			return err
		}

		t.logger.Debug("tournament stream operation authorized",
			"operation", operation,
			"namespace", namespace,
			"method", info.FullMethod)

		return handler(srv, ss)
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

// extractOperationFromMethod extracts operation type from gRPC method name
func (t *TournamentAuthInterceptor) extractOperationFromMethod(fullMethod string) string {
	// Extract method name from full method path (e.g., "/tournament.TournamentService/CreateTournament")
	parts := strings.Split(fullMethod, "/")
	if len(parts) < 2 {
		return ""
	}

	methodName := parts[len(parts)-1]

	// Map method names to operations
	switch methodName {
	case "CreateTournament":
		return "CREATE"
	case "GetTournament":
		return "READ"
	case "ListTournaments":
		return "LIST"
	case "UpdateTournament":
		return "UPDATE"
	case "StartTournament":
		return "START"
	case "CancelTournament":
		return "CANCEL"
	case "DeleteTournament":
		return "DELETE"
	default:
		return ""
	}
}

// extractTokenFromCookieMetadata parses the "cookie" metadata key and returns the access_token value if present.
func extractTokenFromCookieMetadata(meta metadata.MD) string {
	cookieHeaders := meta.Get("cookie")
	for _, cookieHeader := range cookieHeaders {
		for _, part := range strings.Split(cookieHeader, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "access_token=") {
				return strings.TrimPrefix(part, "access_token=")
			}
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

// GetContextUserID extracts user ID from request context
func GetContextUserID(ctx context.Context) (string, error) {
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return "", status.Error(codes.Unauthenticated, "metadata is missing")
	}

	// Extract user ID from various possible metadata sources
	if userIDHeaders := meta["x-user-id"]; len(userIDHeaders) > 0 {
		return userIDHeaders[0], nil
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
		// For now, return a placeholder since token parsing would require additional IAM integration
		// In a full implementation, you'd parse the JWT token to extract the user ID
		return "placeholder-user-id", nil
	}

	return "", status.Error(codes.Unauthenticated, "user ID not found in context")
}

// GetContextUsername extracts username from request context
func GetContextUsername(ctx context.Context) (string, error) {
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return "", status.Error(codes.Unauthenticated, "metadata is missing")
	}

	// Extract username from various possible metadata sources
	if usernameHeaders := meta["x-username"]; len(usernameHeaders) > 0 {
		return usernameHeaders[0], nil
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
		// For now, return a placeholder since token parsing would require additional IAM integration
		// In a full implementation, you'd parse the JWT token to extract the username
		return "placeholder-username", nil
	}

	return "", status.Error(codes.Unauthenticated, "username not found in context")
}

// IsAdminUser checks if the current user has admin privileges
func IsAdminUser(ctx context.Context) (bool, error) {
	meta, found := metadata.FromIncomingContext(ctx)
	if !found {
		return false, status.Error(codes.Unauthenticated, "metadata is missing")
	}

	// Check admin status from various possible metadata sources
	if adminHeaders := meta["x-is-admin"]; len(adminHeaders) > 0 {
		return adminHeaders[0] == "true", nil
	}

	// For now, return false since proper role checking would require additional IAM integration
	// In a full implementation, you'd parse the JWT token to check user roles and permissions
	return false, nil
}
