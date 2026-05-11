// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth/validator"
)

// trustHeaders are client-supplied identity/privilege headers that must never be
// trusted for authorisation decisions. They are removed at the gateway entry
// point so they cannot reach any handler — closes FIND-001 and FIND-009.
var trustHeaders = []string{
	"X-Is-Admin",
	"X-User-Id",
	"X-User-Role",
	"X-Admin",
	"X-Auth-User",
	"X-Forwarded-User",
	"X-Remote-User",
	"X-Username",
}

// StripTrustHeaders removes any client-supplied identity/privilege headers from
// every incoming request before passing it down the chain.
func StripTrustHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range trustHeaders {
			r.Header.Del(h)
		}
		next.ServeHTTP(w, r)
	})
}

// publicHTTPPrefixes are URL prefixes served outside the authenticated API surface
// (static assets, HTML templates, metrics). Auth enforcement is skipped for these.
var publicHTTPPrefixes = []string{
	"/static/",
	"/metrics",
}

// publicHTTPExacts are exact HTTP paths served outside the authenticated API surface.
var publicHTTPExacts = map[string]struct{}{
	"/tournaments": {},
	"/tournament":  {},
}

func isPublicPath(p, basePath string) bool {
	if _, ok := publicHTTPExacts[p]; ok {
		return true
	}
	for _, prefix := range publicHTTPPrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	// Swagger UI and OpenAPI JSON served under <basePath>/apidocs.
	if basePath != "" && strings.HasPrefix(p, basePath+"/apidocs") {
		return true
	}
	return false
}

// httpAuthDecision captures the outcome of parsing the Authorization header.
type httpAuthDecision struct {
	hasToken        bool
	tokenWellFormed bool
	rawToken        string
}

func parseRequestToken(r *http.Request) httpAuthDecision {
	d := httpAuthDecision{}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return d
		}
		d.hasToken = true
		d.rawToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
		d.hasToken = true
		d.rawToken = cookie.Value
	} else {
		return d
	}

	if _, err := parseJWTClaims(d.rawToken); err != nil {
		return d
	}
	d.tokenWellFormed = true
	return d
}

// RequireBearerAuth is the HTTP-layer authentication gate for the gRPC-Gateway.
//
// It closes the controls the penetration report flagged as missing on the backend:
//   - 401 for missing/malformed Authorization (FIND-003, FIND-006, FIND-007, FIND-008)
//   - JWT signature/expiry validation via the IAM validator when configured
//
// Namespace and per-method permission authorisation is intentionally NOT enforced
// here. It is handled by the service-layer `CheckTournamentPermission` call which
// passes the URL path namespace into the IAM validator together with the required
// resource — IAM walks the namespace hierarchy and grants publisher-scoped tokens
// access to child game namespaces in Private/Shared Cloud. Duplicating that check
// in this middleware would over-block legitimate cross-namespace flows.
//
// Public paths (static assets, HTML templates, metrics, swagger UI) are skipped.
func RequireBearerAuth(basePath string, logger *slog.Logger, tokenValidator validator.AuthTokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path, basePath) {
				next.ServeHTTP(w, r)
				return
			}

			decision := parseRequestToken(r)
			if !decision.hasToken {
				logger.Warn("auth: missing bearer token", "path", r.URL.Path, "method", r.Method)
				writeJSONStatus(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}
			if !decision.tokenWellFormed {
				logger.Warn("auth: malformed bearer token", "path", r.URL.Path, "method", r.Method)
				writeJSONStatus(w, http.StatusUnauthorized, "invalid Authorization header")
				return
			}

			if tokenValidator != nil {
				if err := tokenValidator.Validate(decision.rawToken, nil, nil, nil); err != nil {
					logger.Warn("auth: token validation failed", "path", r.URL.Path, "error", err)
					writeJSONStatus(w, http.StatusUnauthorized, "invalid token")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONStatus(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + http.StatusText(status) + `","message":"` + escapeJSONString(msg) + `"}`))
}

func escapeJSONString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	return r.Replace(s)
}
