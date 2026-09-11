package lobby

import (
	"net/http"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/appenv"
)

// clientConfigResponse is the body of GET /config: what deployment
// the client is talking to, and which dev-only surfaces it may
// render. Mirrored in client/src/lib/env.ts.
//
// This is the discovery seam, not the security boundary. The client
// uses it to decide what to *draw*; the server independently refuses
// every dev route and dev action when Env is not dev (requireDev,
// and the action registry's own gate). A client that lies to itself
// about the flags gets a 404, not a cheat.
type clientConfigResponse struct {
	Env      string          `json:"env"`
	Features appenv.Features `json:"features"`
}

// clientConfig serves GET /config. Unauthenticated, mirroring
// /auth/discord/config and /bugreport/config: the login screen needs
// the environment name to render its banner before any session
// exists. Nothing here is a secret — in production the response is a
// constant {"env":"prod","features":{...all false}}.
func clientConfig(c Config, w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, clientConfigResponse{
		Env:      c.Env.String(),
		Features: c.Features,
	})
}

// requireDev wraps a handler so it exists only in a dev deployment.
//
// Non-dev responds 404, not 403: production should not confirm that a
// dev route exists at all. Callers still get a JSON body so the
// client's error path doesn't have to special-case an empty response.
//
// Every dev-only HTTP route MUST be wrapped in this. Hiding a control
// in the client is a UI decision; this is the part that stops someone
// from curl-ing the endpoint against the live table.
func requireDev(c Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.Env.IsDev() {
			_ = writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireDevFeature is requireDev plus a check on one specific
// feature flag, for routes that a dev deployment can individually
// switch off (to rehearse how the UI behaves without them).
func requireDevFeature(c Config, enabled func(appenv.Features) bool, next http.Handler) http.Handler {
	return requireDev(c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !enabled(c.Features) {
			_ = writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		next.ServeHTTP(w, r)
	}))
}
