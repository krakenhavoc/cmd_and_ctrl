package draftpacks

import (
	"encoding/json"
	"net/http"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// Handler serves the draftable-set list:
//
//	GET /draft/sets   — every set a pack can be made for, whether it can
//	                    be drafted, and the cards that block it
//
// # This route is public
//
// main.go mounts it without auth.Middleware, like /catalog. The body is
// set names and card names from the Scryfall dump joined to this
// repository's own catalog, and card data may not sit behind a sign-in
// or a Discord gate (PLAN.md §3, ADR 0222 §9). The handler holds no
// lobby, game or authenticator, reads nothing from the request and
// stores nothing.
//
// It is read-only: no pack is generated or opened here. Drafting is
// #5438.
//
// A nil idx, or a deployment whose dump has not been downloaded yet,
// serves an empty list.
func Handler(idx *cards.Index) http.Handler {
	lister := NewLister(idx)
	mux := metrics.NewServeMux()
	mux.HandleFunc("GET /draft/sets", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Same reasoning as /catalog: the body changes only with a
		// server build or a dump reload.
		w.Header().Set("Cache-Control", "public, max-age=300")
		_ = json.NewEncoder(w).Encode(lister.Sets())
	})
	return mux
}
