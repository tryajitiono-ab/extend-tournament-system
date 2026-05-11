// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func unsignedJWT(claims map[string]interface{}) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload, _ := json.Marshal(claims)
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".fakesig"
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestStripTrustHeaders_RemovesAllTrustHeaders(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		for _, h := range trustHeaders {
			assert.Empty(t, r.Header.Get(h), "trust header %q should have been stripped", h)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/namespace/ns1/tournaments", nil)
	for _, h := range trustHeaders {
		req.Header.Set(h, "evil")
	}
	w := httptest.NewRecorder()
	StripTrustHeaders(next).ServeHTTP(w, req)
	assert.True(t, called)
}

func TestRequireBearerAuth_RejectsMissingAuthorization(t *testing.T) {
	mw := RequireBearerAuth("/tournament", quietLogger(), nil)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/tournament/v1/namespace/securitytest/tournaments", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "FIND-008: tournament list must reject requests with no Authorization header")
}

func TestRequireBearerAuth_RejectsMalformedBearer(t *testing.T) {
	mw := RequireBearerAuth("/tournament", quietLogger(), nil)(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/tournament/v1/admin/namespace/securitytest/tournaments", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer invalid-token-12345")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "FIND-003: malformed bearer must be rejected with 401")
}

func TestRequireBearerAuth_RejectsNonBearerAuth(t *testing.T) {
	mw := RequireBearerAuth("/tournament", quietLogger(), nil)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/tournament/v1/namespace/securitytest/tournaments", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// A well-formed Bearer JWT is allowed through the middleware regardless of its
// `namespace` claim or whether it looks like a service token. Permission and
// namespace authorisation are delegated to the service-layer IAM check.
func TestRequireBearerAuth_AllowsAnyWellFormedJWT(t *testing.T) {
	mw := RequireBearerAuth("/tournament", quietLogger(), nil)(okHandler())

	cases := []struct {
		name   string
		claims map[string]interface{}
		path   string
	}{
		{
			name:   "player JWT same namespace",
			claims: map[string]interface{}{"sub": "u1", "user_name": "player", "namespace": "securitytest"},
			path:   "/tournament/v1/admin/namespace/securitytest/tournaments",
		},
		{
			name:   "publisher JWT cross-namespace (IAM decides)",
			claims: map[string]interface{}{"sub": "admin", "user_name": "publisher_admin", "namespace": "accelbyte"},
			path:   "/tournament/v1/admin/namespace/accelbytetesting/tournaments",
		},
		{
			name:   "user JWT on match-result endpoint",
			claims: map[string]interface{}{"sub": "u2", "user_name": "player", "namespace": "securitytest"},
			path:   "/tournament/v1/admin/namespace/securitytest/tournaments/t-1/matches/m-1/result",
		},
		{
			name:   "service JWT on match-result endpoint",
			claims: map[string]interface{}{"sub": "client-xyz", "namespace": "securitytest"},
			path:   "/tournament/v1/admin/namespace/securitytest/tournaments/t-1/matches/m-1/result",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, c.path, strings.NewReader(`{}`))
			req.Header.Set("Authorization", "Bearer "+unsignedJWT(c.claims))
			w := httptest.NewRecorder()
			mw.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRequireBearerAuth_PublicPathsAreUnauthenticated(t *testing.T) {
	mw := RequireBearerAuth("/tournament", quietLogger(), nil)(okHandler())

	cases := []string{
		"/static/app.js",
		"/tournaments",
		"/tournament",
		"/metrics",
		"/tournament/apidocs/",
		"/tournament/apidocs/api.json",
	}
	for _, p := range cases {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)
		assert.Equalf(t, http.StatusOK, w.Code, "expected %s to be public", p)
	}
}
