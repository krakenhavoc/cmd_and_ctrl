package draftpacks

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

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
	return ListerHandler(NewLister(idx))
}

// ListerHandler is Handler over a Lister the caller shares with the
// draft and the image route, so the list is computed once for all
// three.
func ListerHandler(lister *Lister) http.Handler {
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

// ImageHandler serves the art of the cards a draft hands out:
//
//	GET /draft/image/{id}   — a printing some draftable set's packs can
//	                          contain, or a cube can; ?size= and ?face=
//	                          as /cards/
//
// # This route is public
//
// A visitor drafts against bots with no sign-in (ADR 0222 §9), so the
// pick screen cannot use /cards/{id}/image, which needs a session, and
// card data may not sit behind a sign-in (PLAN.md §3). /catalog/image
// does not cover it either: that route serves one representative
// printing of each card, and a pack holds the set's own printing, and
// basic lands, which the catalog does not register.
//
// It is scoped the way /catalog/image is, and for the same reason: it
// answers only for the printings in the pools of draftable sets, a set
// of IDs fixed by the dump and the catalog and already published by GET
// /draft/sets, so it is not a general image proxy. A cube draft (ADR
// 0222 §8) hands out the printings its list names, so the route also
// answers for any printing of a card the catalog holds as Full and for
// basic lands (Lister.IsCubePrinting): still a set fixed by the dump
// and the catalog, of cards the public catalogue shows. Anything else
// is a 404, the answer an unknown card gets. It shares the one ImageCache
// and its guarded fetch.
func ImageHandler(lister *Lister, idx *cards.Index, cache *cards.ImageCache) http.Handler {
	mux := metrics.NewServeMux()
	mux.HandleFunc("GET /draft/image/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := cards.ParseCardID(r.PathValue("id"))
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid card id")
			return
		}
		// The scope check runs before the availability check, so an
		// out-of-scope ID does not learn whether there is a cache.
		if !lister.IsDraftPrinting(id) && !lister.IsCubePrinting(id) {
			writeErr(w, http.StatusNotFound, "not a draft card")
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
				writeErr(w, http.StatusFailedDependency, err.Error())
			}
			return
		}
		// The bytes behind a Scryfall UUID never change.
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
