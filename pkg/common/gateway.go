// Copyright (c) 2023-2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
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

func NewGateway(ctx context.Context, grpcServerEndpoint string, basePath string) (*Gateway, error) {
	// Configure gateway to forward custom headers to gRPC metadata
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			// Forward all headers to gRPC metadata by default
			// This is important for testing mode where we use custom headers
			// like x-user-id, x-username, namespace, etc.
			return key, true
		}),
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

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Strip the base path, since the base_path configuration in protofile won't actually do the routing
	// Reference: https://github.com/grpc-ecosystem/grpc-gateway/pull/919/commits/1c34df861cfc0d6cb19ea617921d7d9eaa209977
	http.StripPrefix(g.basePath, g.mux).ServeHTTP(w, r)
}
