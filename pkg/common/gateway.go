// Copyright (c) 2023-2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/credentials/insecure"

	pb "extend-tournament-service/pkg/pb"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Gateway struct {
	mux      *runtime.ServeMux
	basePath string
}

// NewGateway creates a new gateway with client-based registration (connects to gRPC server via network)
func NewGateway(ctx context.Context, grpcServerEndpoint string, basePath string) (*Gateway, error) {
	// Custom error handler to log gRPC errors
	errorHandler := func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("gRPC-Gateway error", "error", err, "path", r.URL.Path, "method", r.Method)
		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
	}

	// Configure gateway to forward custom headers to gRPC metadata
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			// Forward all headers to gRPC metadata by default
			// This is important for testing mode where we use custom headers
			// like x-user-id, x-username, namespace, etc.
			return key, true
		}),
		// Enable error details in responses for debugging
		runtime.WithErrorHandler(errorHandler),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterTournamentServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		mux:      mux,
		basePath: basePath,
	}, nil
}

// NewGatewayWithServer creates a new gateway with direct server registration (no network connection needed)
func NewGatewayWithServer(ctx context.Context, server pb.TournamentServiceServer, basePath string) (*Gateway, error) {
	// Custom error handler to log gRPC errors
	errorHandler := func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("gRPC-Gateway error", "error", err, "path", r.URL.Path, "method", r.Method)
		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
	}

	// Configure gateway to forward custom headers to gRPC metadata
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			// Forward all headers to gRPC metadata by default
			// This is important for testing mode where we use custom headers
			// like x-user-id, x-username, namespace, etc.
			return key, true
		}),
		// Enable error details in responses for debugging
		runtime.WithErrorHandler(errorHandler),
	)

	// Register the server directly (no network connection needed)
	err := pb.RegisterTournamentServiceHandlerServer(ctx, mux, server)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		mux:      mux,
		basePath: basePath,
	}, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Strip the base path, since the base_path configuration in protofile won't actually do the routing
	// Reference: https://github.com/grpc-ecosystem/grpc-gateway/pull/919/commits/1c34df861cfc0d6cb19ea617921d7d9eaa209977

	// Debug: Log the incoming request details
	slog.Debug("gateway received request", "method", r.Method, "path", r.URL.Path, "request_uri", r.RequestURI, "base_path", g.basePath)

	http.StripPrefix(g.basePath, g.mux).ServeHTTP(w, r)
}
