package roadmap

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Handler serves the roadmap:
//
//	GET /roadmap   — Build(), as JSON
//
// # This route is public
//
// main.go mounts it WITHOUT auth.Middleware, unlike /catalog. That is
// safe because of what the body is, not because of where it is
// mounted: Build publishes card names and caveat sentences this
// repository wrote, and never an image, an image URL, a Scryfall
// printing ID or a card's oracle text (ADR 0092 Decision 1).
// TestBuildIsPlainPublishableData and this package's handler test both
// fail if one of those keys ever reaches the body. Anything that would
// add one belongs on the signed-in catalog instead.
//
// Like catalog.Handler, it holds no lobby, game or authenticator, so
// there is no session state it could leak.
//
// The body is fixed for the life of the binary (the registry and the
// catalog are both fixed at init), so it is encoded once and the same
// bytes are written to every request. Any method other than GET or
// HEAD is answered 405 by the method-aware mux.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /roadmap", func(w http.ResponseWriter, _ *http.Request) {
		body, err := encoded()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "roadmap unavailable"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Same reasoning as /catalog: the body changes only with a
		// server build, and five minutes is short enough that a
		// refresh after a deploy shows the new roadmap.
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(body)
	})
	return mux
}

var (
	encodeOnce sync.Once
	encodedRM  []byte
	encodeErr  error
)

// encoded is Build() marshalled once, with a trailing newline to match
// what json.Encoder writes on the other JSON routes.
func encoded() ([]byte, error) {
	encodeOnce.Do(func() {
		b, err := json.Marshal(Build())
		if err != nil {
			encodeErr = err
			return
		}
		encodedRM = append(b, '\n')
	})
	return encodedRM, encodeErr
}
