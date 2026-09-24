package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// Handler returns the catalog routes:
//
//	GET /catalog              — every automated card + its completeness
//	GET /catalog/image/{id}   — card art, for catalog cards only
//
// # These routes need a session
//
// main.go mounts both behind auth.Middleware, like /cards/. They were
// built unauthenticated as a public showcase, and are gated because
// AGENTS.md §1 and §8 describe the project as private and personal-use:
// serving card art to anonymous visitors is a different posture from
// the one the repository states. main.go's comment at the mount is the
// record; check there before assuming otherwise. The public roadmap
// (internal/roadmap, ADR 0092) is the page that answers "what does the
// engine do?" without a session, and it publishes card names only.
//
// The handler itself stays free of session state. Everything served
// here is derived from the Scryfall bulk dump and this repository's own
// card files; the handler holds no *Lobby, no *Game and no
// authenticator, so there is no game, seat or session state it could
// leak even by accident.
//
// # Why the image route is not just /cards/{id}/image
//
// That route accepts any Scryfall UUID in the dump, so it is a
// general-purpose image proxy for all ~35k cards, fetching and caching
// arbitrary art on demand.
//
// This route accepts only the representative printing of a card the
// catalog registers — a few hundred UUIDs, fixed at build time and
// already published by GET /catalog. That is exactly the set the page
// renders and nothing beyond it, which still matters behind the session
// gate, since a session is cheap to obtain. It reuses the same
// ImageCache, so the two routes share one disk cache and one
// SSRF-guarded fetch path rather than growing a second.
//
// A nil idx serves an empty catalog and 404s images — the honest
// answer on a deployment whose dump has not been downloaded yet. A
// nil cache disables images alone.
func Handler(idx *cards.Index, cache *cards.ImageCache) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /catalog", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// The body changes only on a server build or a dump reload,
		// and a browser holding a five-minute-old catalog is fine.
		// Short enough that a refresh after a deploy shows new cards.
		w.Header().Set("Cache-Control", "public, max-age=300")
		_ = json.NewEncoder(w).Encode(Build(idx))
	})

	mux.HandleFunc("GET /catalog/image/{id}", func(w http.ResponseWriter, r *http.Request) {
		raw := r.PathValue("id")
		id, err := uuid.Parse(raw)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid card id")
			return
		}
		// The scope check, and the reason this route can be public.
		// A UUID that is not a catalog card's representative printing
		// gets 404 — the same answer an unknown card gets, so the
		// route reveals nothing about which IDs exist in the dump.
		//
		// It runs BEFORE the availability check on purpose: an
		// out-of-scope ID should not learn whether this deployment
		// has an image cache, and the check needs only the index.
		if !IsCatalogPrinting(idx, id) {
			writeErr(w, http.StatusNotFound, "not a catalog card")
			return
		}
		if cache == nil {
			writeErr(w, http.StatusServiceUnavailable, "card image service unavailable")
			return
		}

		face := 0
		if rawFace := r.URL.Query().Get("face"); rawFace != "" {
			n, convErr := strconv.Atoi(rawFace)
			if convErr != nil {
				writeErr(w, http.StatusBadRequest, "invalid image face")
				return
			}
			face = n
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		path, err := cache.FetchFace(ctx, idx, id, r.URL.Query().Get("size"), face)
		if err != nil {
			switch {
			case errors.Is(err, cards.ErrInvalidSize):
				writeErr(w, http.StatusBadRequest, "invalid image size")
			case errors.Is(err, cards.ErrInvalidFace):
				writeErr(w, http.StatusBadRequest, "invalid image face")
			case errors.Is(err, cards.ErrNoImage):
				writeErr(w, http.StatusNotFound, "no image for this card")
			case errors.Is(err, os.ErrNotExist):
				writeErr(w, http.StatusNotFound, "unknown card")
			default:
				writeErr(w, http.StatusBadGateway, err.Error())
			}
			return
		}
		// Same reasoning as /cards/{id}/image: the bytes behind a
		// Scryfall UUID never change, so a week is safe.
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		http.ServeFile(w, r, path)
	})

	return mux
}

func writeErr(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
