package cards

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// Handler returns an http.Handler serving the card + image routes:
//
//	GET /cards/{id}        — card metadata (name, image_uri, etc.)
//	GET /cards/{id}/image  — image bytes; downloads on miss
//
// Both endpoints require authentication upstream — main.go wraps
// this handler with auth.Middleware before mounting it.
//
// The handler consults idx for metadata and cache for image bytes.
// A nil cache disables the /image route (returns 503); a nil idx
// returns 404 for everything (useful when the Scryfall dump has not
// been downloaded yet).
func Handler(idx *Index, cache *ImageCache) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cards/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if idx == nil {
			writeErr(w, http.StatusNotFound, "card index not loaded")
			return
		}
		card, ok := idx.Get(id)
		if !ok {
			writeErr(w, http.StatusNotFound, "unknown card")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_ = json.NewEncoder(w).Encode(card)
	})

	mux.HandleFunc("GET /cards/{id}/image", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if idx == nil || cache == nil {
			writeErr(w, http.StatusServiceUnavailable, "card image service unavailable")
			return
		}
		size := r.URL.Query().Get("size")
		// Bound the total time a single image request can hold; a
		// hung CDN shouldn't tie up request goroutines.
		ctx, cancel := contextWithTimeout(r, 15*time.Second)
		defer cancel()

		path, err := cache.Fetch(ctx, idx, id, size)
		if err != nil {
			switch {
			case errors.Is(err, ErrNoImage):
				writeErr(w, http.StatusNotFound, "no image for this card")
			case errors.Is(err, os.ErrNotExist):
				writeErr(w, http.StatusNotFound, "unknown card")
			default:
				writeErr(w, http.StatusBadGateway, err.Error())
			}
			return
		}
		// Browsers cache aggressively; the image bytes for a given
		// Scryfall UUID are immutable (Scryfall re-prints change
		// IDs), so a week is fine.
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		http.ServeFile(w, r, path)
	})
	return mux
}

func parseID(r *http.Request) (uuid.UUID, error) {
	raw := r.PathValue("id")
	if raw == "" {
		return uuid.Nil, errors.New("missing card id")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid card id")
	}
	return id, nil
}

func writeErr(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
