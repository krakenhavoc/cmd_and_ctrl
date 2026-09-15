package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// SessionCookie is the name of the httpOnly cookie used to carry the
// credential on browser HTTP requests. WS upgrades on browsers
// cannot set Authorization headers; the cookie transports the
// credential for same-origin WS as well, with ?token= as a fallback
// for cross-origin deployments and for CLI tooling that prefers
// query params over cookies.
const SessionCookie = "cmdctrl_session"

// CredentialFromRequest extracts the bearer credential from r,
// preferring (in order) the session cookie, the Authorization
// header, and finally the ?token= query parameter. Returns "" if
// none is present.
//
// All three transports are honoured so that browser HTTP, browser
// WS, and CLI clients share one validation code path.
func CredentialFromRequest(r *http.Request) string {
	if c, err := r.Cookie(SessionCookie); err == nil && c.Value != "" {
		return c.Value
	}
	if h := r.Header.Get("Authorization"); h != "" {
		if v, ok := strings.CutPrefix(h, "Bearer "); ok {
			return v
		}
	}
	return r.URL.Query().Get("token")
}

// ctxKey is unexported so outside packages can't populate the
// principal themselves — the only way to inject a Principal into
// request context is to pass through Middleware.
type ctxKey struct{}

// PrincipalFromContext returns the Principal attached to ctx by
// Middleware, or the zero Principal and false if none is attached.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}

// WithPrincipal returns a new context with p attached. Exposed for
// tests and for the WS authorizer to hand off a validated principal
// to the rest of the server.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// Middleware returns an http middleware that validates the request's
// credential via a, attaches the resulting Principal to the request
// context, and invokes next. Requests with no credential or an
// invalid / expired one are rejected with 401 and a small JSON body.
//
// RequireRoles optionally enforces that the authenticated principal's
// role is in the allowed list. Empty list = any authenticated role.
func Middleware(a Authenticator, requireRoles ...Role) func(http.Handler) http.Handler {
	allowed := make(map[Role]struct{}, len(requireRoles))
	for _, r := range requireRoles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cred := CredentialFromRequest(r)
			if cred == "" {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			p, err := a.Validate(r.Context(), cred)
			if err != nil {
				status := http.StatusUnauthorized
				var msg string
				switch err {
				case ErrExpiredCredential:
					msg = "session expired"
				case ErrInvalidCredential:
					msg = "invalid credential"
				default:
					msg = err.Error()
				}
				writeError(w, status, msg)
				return
			}
			if len(allowed) > 0 {
				if _, ok := allowed[p.Role]; !ok {
					writeError(w, http.StatusForbidden, "insufficient role")
					return
				}
			}
			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
		})
	}
}

// writeError is the uniform JSON error response shape used by auth
// and lobby handlers. Kept compact on the wire since the browser
// doesn't render server error bodies beyond status-line defaults.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
