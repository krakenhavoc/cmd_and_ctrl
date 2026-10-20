package cards

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// Handler returns an http.Handler serving the card + image routes:
//
//	GET /cards/{id}        — card metadata (name, image_uri, etc.)
//	GET /cards/{id}/image  — image bytes; downloads on miss
//	GET /cards/{id}/printings — available English printings
//
// Metadata and images require authentication upstream. main.go also
// mounts the printing-list route publicly, like /catalog; it performs
// no external requests.
//
// The handler consults idx for metadata and cache for image bytes.
// A nil cache disables the /image route (returns 503); a nil idx
// returns 404 for everything (useful when the Scryfall dump has not
// been downloaded yet).
func Handler(idx *Index, cache *ImageCache) http.Handler {
	mux := metrics.NewServeMux()
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

	mux.HandleFunc("GET /cards/{id}/printings", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if idx == nil {
			writeErr(w, http.StatusServiceUnavailable, "card index not loaded")
			return
		}
		if _, ok := idx.Get(id); !ok {
			writeErr(w, http.StatusNotFound, "unknown card")
			return
		}
		type printing struct {
			ID         uuid.UUID `json:"id"`
			Name       string    `json:"name"`
			FlavorName string    `json:"flavor_name,omitempty"`
			Set        string    `json:"set"`
			Number     string    `json:"collector_number"`
		}
		out := []printing{}
		for _, c := range idx.Printings(id) {
			out = append(out, printing{c.ID, c.Name, c.FlavorName, c.SetCode, c.CollectorNo})
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_ = json.NewEncoder(w).Encode(map[string]any{"printings": out})
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
		// ADR 0034: ?face=N picks a printed face. Absent — which is
		// every request a pre-0034 client makes, and every request
		// for a single-faced card — is face 0, the front, and serves
		// exactly the bytes it always did. A non-numeric value is a
		// client bug and is refused rather than silently coerced to
		// the front face.
		face := 0
		if raw := r.URL.Query().Get("face"); raw != "" {
			n, convErr := strconv.Atoi(raw)
			if convErr != nil {
				writeErr(w, http.StatusBadRequest, "invalid image face")
				return
			}
			face = n
		}
		// Bound the total time a single image request can hold; a
		// hung CDN shouldn't tie up request goroutines.
		ctx, cancel := contextWithTimeout(r, 15*time.Second)
		defer cancel()

		path, err := cache.FetchFace(ctx, idx, id, size, face)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidSize):
				writeErr(w, http.StatusBadRequest, "invalid image size")
			case errors.Is(err, ErrInvalidFace):
				writeErr(w, http.StatusBadRequest, "invalid image face")
			case errors.Is(err, ErrNoImage):
				writeErr(w, http.StatusNotFound, "no image for this card")
			case errors.Is(err, os.ErrNotExist):
				writeErr(w, http.StatusNotFound, "unknown card")
			default:
				writeErr(w, http.StatusFailedDependency, err.Error())
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
	id, err := ParseCardID(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid card id")
	}
	return id, nil
}

// ParseCardID parses a card id from a request. Only the canonical
// 36-character hyphenated hex form is accepted: uuid.Parse also takes
// "urn:uuid:" prefixes, braces and bare 32-digit hex, and an id that
// reaches a cache path or a lookup should have exactly one spelling.
func ParseCardID(raw string) (uuid.UUID, error) {
	if len(raw) != 36 {
		return uuid.Nil, errors.New("invalid card id")
	}
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case i == 8 || i == 13 || i == 18 || i == 23:
			if c != '-' {
				return uuid.Nil, errors.New("invalid card id")
			}
		case (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F'):
		default:
			return uuid.Nil, errors.New("invalid card id")
		}
	}
	return uuid.Parse(raw)
}

func writeErr(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
