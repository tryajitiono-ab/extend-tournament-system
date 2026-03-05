// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// makeTestJWT builds a minimal unsigned JWT with the given claims.
// Signature verification is not performed by our parsers, so a fake signature suffices.
func makeTestJWT(claims map[string]interface{}) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload, _ := json.Marshal(claims)
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".fakesig"
}

func ctxWithMeta(kv ...string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(kv...))
}

// --- GetContextUserID ---

func TestGetContextUserID_FromHeader(t *testing.T) {
	ctx := ctxWithMeta("x-user-id", "header-user-123")
	userID, err := GetContextUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, "header-user-123", userID)
}

func TestGetContextUserID_FromBearerToken(t *testing.T) {
	token := makeTestJWT(map[string]interface{}{
		"sub":       "jwt-user-456",
		"user_name": "jwtuser",
	})
	ctx := ctxWithMeta("authorization", "Bearer "+token)
	userID, err := GetContextUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, "jwt-user-456", userID)
}

func TestGetContextUserID_HeaderTakesPrecedenceOverBearer(t *testing.T) {
	token := makeTestJWT(map[string]interface{}{"sub": "jwt-user-456"})
	ctx := ctxWithMeta("x-user-id", "header-user-123", "authorization", "Bearer "+token)
	userID, err := GetContextUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, "header-user-123", userID)
}

func TestGetContextUserID_MissingMetadata(t *testing.T) {
	_, err := GetContextUserID(context.Background())
	assert.Error(t, err)
}

func TestGetContextUserID_NoUserIDAnywhere(t *testing.T) {
	ctx := ctxWithMeta("some-other-header", "value")
	_, err := GetContextUserID(ctx)
	assert.Error(t, err)
}

func TestGetContextUserID_BearerTokenMissingSubClaim(t *testing.T) {
	token := makeTestJWT(map[string]interface{}{"user_name": "nosubuser"})
	ctx := ctxWithMeta("authorization", "Bearer "+token)
	_, err := GetContextUserID(ctx)
	assert.Error(t, err)
}

func TestGetContextUserID_MalformedBearerToken(t *testing.T) {
	ctx := ctxWithMeta("authorization", "Bearer not.a.valid.jwt.atall")
	_, err := GetContextUserID(ctx)
	assert.Error(t, err)
}

// --- GetContextUsername ---

func TestGetContextUsername_FromHeader(t *testing.T) {
	ctx := ctxWithMeta("x-username", "headeruser")
	username, err := GetContextUsername(ctx)
	require.NoError(t, err)
	assert.Equal(t, "headeruser", username)
}

func TestGetContextUsername_FromBearerToken(t *testing.T) {
	token := makeTestJWT(map[string]interface{}{
		"sub":       "user-abc",
		"user_name": "jwtuser",
	})
	ctx := ctxWithMeta("authorization", "Bearer "+token)
	username, err := GetContextUsername(ctx)
	require.NoError(t, err)
	assert.Equal(t, "jwtuser", username)
}

func TestGetContextUsername_HeaderTakesPrecedenceOverBearer(t *testing.T) {
	token := makeTestJWT(map[string]interface{}{"user_name": "jwtuser"})
	ctx := ctxWithMeta("x-username", "headeruser", "authorization", "Bearer "+token)
	username, err := GetContextUsername(ctx)
	require.NoError(t, err)
	assert.Equal(t, "headeruser", username)
}

func TestGetContextUsername_MissingMetadata(t *testing.T) {
	_, err := GetContextUsername(context.Background())
	assert.Error(t, err)
}

func TestGetContextUsername_BearerTokenMissingUsernameClaim(t *testing.T) {
	token := makeTestJWT(map[string]interface{}{"sub": "user-abc"})
	ctx := ctxWithMeta("authorization", "Bearer "+token)
	_, err := GetContextUsername(ctx)
	assert.Error(t, err)
}
