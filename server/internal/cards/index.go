// Package cards provides the card-metadata layer: a read-only index
// loaded from the Scryfall default-cards bulk dump, plus a disk-
// backed lazy image cache that serves card images via Scryfall's CDN.
//
// The bulk dump lives at $CMDCTRL_DATA_DIR/scryfall/default-cards.json
// and is refreshed weekly by scripts/scryfall-refresh.sh. The file is
// streamed on load (it's ~450MB uncompressed) and only the fields the
// server cares about are kept in memory.
package cards

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Card is the trimmed subset of Scryfall's card record we keep
// in-memory. Everything the S04 server needs (name, image URIs) is
// here; higher-level fields (mana cost, oracle text, types) come
// later when the rules engine arrives in S13+.
type Card struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	SetCode     string    `json:"set"`
	CollectorNo string    `json:"collector_number"`
	Lang        string    `json:"lang"`
	// ImageURIs is Scryfall's multi-size image map. Keys we care
	// about: "small", "normal", "large", "png", "art_crop",
	// "border_crop". The server picks based on Cache.DefaultSize.
	ImageURIs map[string]string `json:"image_uris"`
	// CardFaces is populated for double-faced / split cards. The
	// front face's image_uris is what we surface for now; handling
	// flip state is a UI concern.
	CardFaces []struct {
		Name      string            `json:"name"`
		ImageURIs map[string]string `json:"image_uris"`
	} `json:"card_faces"`
}

// Index is a read-only map from card UUID → Card, built at server
// start from the Scryfall bulk dump. Safe for concurrent reads.
//
// Zero-value Index is a valid "empty" index — Get always returns
// (Card{}, false). Useful when the dump file hasn't been downloaded
// yet on a fresh deployment; the server can still run, image
// lookups just 404.
type Index struct {
	mu     sync.RWMutex
	byID   map[uuid.UUID]Card
	loaded time.Time
	path   string
}

// NewIndex returns an empty Index. Call Load to populate it from a
// Scryfall bulk dump file.
func NewIndex() *Index {
	return &Index{byID: make(map[uuid.UUID]Card)}
}

// Load replaces the in-memory map with the cards parsed from the
// JSON file at path. Uses a streaming decoder so peak memory stays
// bounded (the dump is a top-level JSON array of ~30k objects).
//
// On success returns the number of cards loaded. If the file does
// not exist, returns os.ErrNotExist unchanged so callers can decide
// whether a missing dump is fatal or merely a "not refreshed yet"
// log warning.
func (i *Index) Load(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	// Top-level must be a JSON array — step past the opening bracket.
	tok, err := dec.Token()
	if err != nil {
		return 0, fmt.Errorf("read opening token: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return 0, fmt.Errorf("expected top-level JSON array, got %v", tok)
	}

	loaded := make(map[uuid.UUID]Card, 40_000) // ballpark for default_cards
	for dec.More() {
		var c Card
		if err := dec.Decode(&c); err != nil {
			return 0, fmt.Errorf("decode card: %w", err)
		}
		if c.ID == uuid.Nil {
			continue // a record without an ID is useless; skip rather than reject
		}
		loaded[c.ID] = c
	}
	// Consume the closing bracket; a decoder that swallows it would
	// also swallow trailing junk, so we check.
	if _, err := dec.Token(); err != nil && !errors.Is(err, io.EOF) {
		return 0, fmt.Errorf("read closing token: %w", err)
	}

	i.mu.Lock()
	i.byID = loaded
	i.loaded = time.Now().UTC()
	i.path = path
	i.mu.Unlock()
	return len(loaded), nil
}

// Get returns the card for id, or (zero, false) if no such card is
// loaded.
func (i *Index) Get(id uuid.UUID) (Card, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	c, ok := i.byID[id]
	return c, ok
}

// Count returns the number of cards currently in the index.
func (i *Index) Count() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return len(i.byID)
}

// LoadedAt returns the timestamp at which Load last succeeded, or
// the zero time if Load has never been called.
func (i *Index) LoadedAt() time.Time {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.loaded
}

// ImageURI returns the best image URL for c at the given size.
// Falls back to the front face's image if c is a double-faced card,
// and to "normal" if the requested size is missing.
func ImageURI(c Card, size string) string {
	// Prefer the top-level map (single-faced cards).
	if uri, ok := c.ImageURIs[size]; ok && uri != "" {
		return uri
	}
	if uri, ok := c.ImageURIs["normal"]; ok && uri != "" {
		return uri
	}
	// Fall back to the front face (double-faced, split, etc).
	if len(c.CardFaces) > 0 {
		face := c.CardFaces[0]
		if uri, ok := face.ImageURIs[size]; ok && uri != "" {
			return uri
		}
		if uri, ok := face.ImageURIs["normal"]; ok && uri != "" {
			return uri
		}
	}
	return ""
}
