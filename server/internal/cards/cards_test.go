package cards

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

// fixtureJSON is a tiny 2-card Scryfall-shaped bulk file for tests.
// Keeping it in-source makes the unit test hermetic.
const fixtureJSON = `[
{"id":"11111111-1111-1111-1111-111111111111","name":"Sol Ring","set":"c21","collector_number":"1","lang":"en","image_uris":{"normal":"https://example.test/sol-ring-normal.jpg","small":"https://example.test/sol-ring-small.jpg"}},
{"id":"22222222-2222-2222-2222-222222222222","name":"Lightning Bolt","set":"lea","collector_number":"161","lang":"en","image_uris":{"normal":"https://example.test/bolt-normal.jpg"}}
]`

func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "default-cards.json")
	if err := os.WriteFile(path, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestIndexLoadAndGet(t *testing.T) {
	path := writeFixture(t)
	idx := NewIndex()
	n, err := idx.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if n != 2 {
		t.Errorf("Load count: got %d, want 2", n)
	}
	if got := idx.Count(); got != 2 {
		t.Errorf("Count: got %d, want 2", got)
	}

	solRing := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	c, ok := idx.Get(solRing)
	if !ok {
		t.Fatal("Get(sol ring): not found")
	}
	if c.Name != "Sol Ring" {
		t.Errorf("name: got %q, want %q", c.Name, "Sol Ring")
	}
	if idx.LoadedAt().IsZero() {
		t.Error("LoadedAt is zero after successful load")
	}
}

func TestIndexLoadMissingFile(t *testing.T) {
	idx := NewIndex()
	_, err := idx.Load(filepath.Join(t.TempDir(), "nope.json"))
	if !os.IsNotExist(err) {
		t.Errorf("Load missing: got %v, want os.ErrNotExist", err)
	}
}

func TestIndexLoadRejectsNonArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bogus.json")
	_ = os.WriteFile(path, []byte(`{"not":"an array"}`), 0o644)
	idx := NewIndex()
	if _, err := idx.Load(path); err == nil {
		t.Error("Load non-array: expected error, got nil")
	}
}

// TestPutPreservesLegitimateCanonicalAgainstHijack is a regression
// for Scryfall art-series records whose Name is "X // X" with two
// faces both named "X". Prior to the guard, the face loop would
// overwrite the legitimate single-card "X" entry with the art-series
// record, which carries commander: not_legal — valid decks got
// `not_legal_in_format` for perfectly playable cards like Garruk's
// Uprising.
func TestPutPreservesLegitimateCanonicalAgainstHijack(t *testing.T) {
	legitID := uuid.New()
	legit := Card{
		ID:         legitID,
		Name:       "Garruk's Uprising",
		Legalities: map[string]string{"commander": "legal"},
	}
	hijack := Card{
		ID:   uuid.New(),
		Name: "Garruk's Uprising // Garruk's Uprising",
		CardFaces: []CardFace{
			{Name: "Garruk's Uprising"},
			{Name: "Garruk's Uprising"},
		},
		Legalities: map[string]string{"commander": "not_legal"},
	}

	for _, order := range []struct {
		label string
		first Card
		next  Card
	}{
		{"legit first then hijack", legit, hijack},
		{"hijack first then legit", hijack, legit},
	} {
		t.Run(order.label, func(t *testing.T) {
			fresh := NewIndex()
			fresh.Put(order.first)
			fresh.Put(order.next)
			got, ok := fresh.FindByName("Garruk's Uprising")
			if !ok {
				t.Fatal("FindByName: not found")
			}
			if got.ID != legitID {
				t.Errorf("ID: got %v, want legit %v (hijack record won the race)", got.ID, legitID)
			}
			if got.Legalities["commander"] != "legal" {
				t.Errorf("commander legality: got %q, want legal", got.Legalities["commander"])
			}
			// The hijack record is still reachable under its full name —
			// we only block the single-face hijack, not the full-name
			// lookup.
			full, ok := fresh.FindByName("Garruk's Uprising // Garruk's Uprising")
			if !ok {
				t.Error("full-name lookup for hijack record: not found")
			}
			if full.ID == legitID {
				t.Error("full-name should resolve to the art-series record, not the legit printing")
			}
		})
	}
}

// TestFindByNameSlashFallback covers the deck-import quirk where
// exports vary between "Fire / Ice" (single slash) and Scryfall's
// canonical "Fire // Ice" (double slash). Direct lookup on the
// single-slash form misses; the fallback on the front-face name
// should resolve both forms to the same Card.
func TestFindByNameSlashFallback(t *testing.T) {
	idx := NewIndex()
	id := uuid.New()
	idx.Put(Card{
		ID:   id,
		Name: "Stump Stomp // Burnwillow Clearing",
		CardFaces: []CardFace{
			{Name: "Stump Stomp"},
			{Name: "Burnwillow Clearing"},
		},
	})

	cases := []struct {
		name  string
		query string
	}{
		{"canonical double slash", "Stump Stomp // Burnwillow Clearing"},
		{"single slash variant", "Stump Stomp / Burnwillow Clearing"},
		{"extra whitespace around single slash", "Stump Stomp   /   Burnwillow Clearing"},
		{"front face only", "Stump Stomp"},
		{"back face only", "Burnwillow Clearing"},
		{"case-insensitive single slash", "stump stomp / burnwillow clearing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, ok := idx.FindByName(tc.query)
			if !ok {
				t.Fatalf("FindByName(%q): not found", tc.query)
			}
			if c.ID != id {
				t.Errorf("FindByName(%q): got ID %v, want %v", tc.query, c.ID, id)
			}
		})
	}
}

func TestImageURIFallbacks(t *testing.T) {
	// Single-faced card, asked for "large" when only "normal" exists:
	// falls back to normal.
	c := Card{ImageURIs: map[string]string{"normal": "n.jpg"}}
	if got := ImageURI(c, "large"); got != "n.jpg" {
		t.Errorf("fallback to normal: got %q, want %q", got, "n.jpg")
	}
	// Double-faced card: falls back to front face.
	c = Card{CardFaces: []CardFace{
		{Name: "Front", ImageURIs: map[string]string{"normal": "front.jpg"}},
	}}
	if got := ImageURI(c, "normal"); got != "front.jpg" {
		t.Errorf("double-faced front: got %q, want %q", got, "front.jpg")
	}
	// No image at all.
	if got := ImageURI(Card{}, "normal"); got != "" {
		t.Errorf("empty card: got %q, want empty", got)
	}
}

// spyCDN is a test double for Scryfall's image CDN. Counts how many
// GETs it sees so the dedup test can assert the per-id lock actually
// collapses concurrent requests into a single network fetch.
func spyCDN(t *testing.T, payload []byte) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func TestImageCacheFetchMissHit(t *testing.T) {
	cdn, hits := spyCDN(t, []byte("fake-jpeg-bytes"))
	// Build an index that points at the spy CDN.
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, Name: "Test", ImageURIs: map[string]string{"normal": cdn.URL + "/a"}}

	cache, err := NewImageCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}

	// First fetch — expect a CDN hit.
	path, err := cache.Fetch(context.Background(), idx, id, "normal")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if hits.Load() != 1 {
		t.Errorf("hits after first fetch: got %d, want 1", hits.Load())
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if string(body) != "fake-jpeg-bytes" {
		t.Errorf("cached bytes: got %q, want fake-jpeg-bytes", body)
	}

	// Second fetch — file is on disk, no CDN hit.
	if _, err := cache.Fetch(context.Background(), idx, id, "normal"); err != nil {
		t.Fatalf("Fetch (cached): %v", err)
	}
	if hits.Load() != 1 {
		t.Errorf("hits after cached fetch: got %d, want 1 (no re-download)", hits.Load())
	}
}

func TestImageCacheConcurrentDedup(t *testing.T) {
	cdn, hits := spyCDN(t, []byte("payload"))
	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": cdn.URL + "/a"}}

	cache, err := NewImageCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := cache.Fetch(context.Background(), idx, id, "normal"); err != nil {
				t.Errorf("Fetch: %v", err)
			}
		}()
	}
	wg.Wait()
	// Per-id lock should fold concurrent requests into 1 network hit.
	if got := hits.Load(); got != 1 {
		t.Errorf("concurrent dedup: got %d CDN hits, want 1", got)
	}
}

func TestImageCacheFetchUnknownCard(t *testing.T) {
	idx := NewIndex()
	cache, _ := NewImageCache(t.TempDir())
	_, err := cache.Fetch(context.Background(), idx, uuid.New(), "normal")
	if err == nil || !strings.Contains(err.Error(), "unknown card") {
		t.Errorf("fetch unknown: got %v, want 'unknown card'", err)
	}
}

func TestImageCacheFetchNoImage(t *testing.T) {
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, Name: "no-image"}
	cache, _ := NewImageCache(t.TempDir())
	_, err := cache.Fetch(context.Background(), idx, id, "normal")
	if err != ErrNoImage {
		t.Errorf("fetch no-image: got %v, want ErrNoImage", err)
	}
}

func TestHandlerCardMetadata(t *testing.T) {
	path := writeFixture(t)
	idx := NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	cache, _ := NewImageCache(t.TempDir())

	h := Handler(idx, cache)
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/cards/11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got Card
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Sol Ring" {
		t.Errorf("name: got %q, want %q", got.Name, "Sol Ring")
	}
}

func TestHandlerImageServesAndCaches(t *testing.T) {
	cdn, hits := spyCDN(t, []byte("img"))
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": cdn.URL + "/a"}}
	cache, _ := NewImageCache(t.TempDir())

	h := Handler(idx, cache)
	srv := httptest.NewServer(h)
	defer srv.Close()

	u := fmt.Sprintf("%s/cards/%s/image", srv.URL, id)
	resp, err := http.Get(u)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want %d (body=%s)", resp.StatusCode, http.StatusOK, body)
	}
	if string(body) != "img" {
		t.Errorf("body: got %q, want img", body)
	}
	if got := resp.Header.Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("Cache-Control: got %q, want immutable", got)
	}

	// Second hit is cached on disk.
	resp, _ = http.Get(u)
	resp.Body.Close()
	if hits.Load() != 1 {
		t.Errorf("second request re-downloaded: got %d, want 1 CDN hit", hits.Load())
	}
}

func TestImageCacheRejectsInvalidSize(t *testing.T) {
	// A caller-supplied size outside the allow-list must never reach
	// pathFor: if it did, `size="../../etc/passwd"` would escape the
	// cache root.
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": "https://example.test/x.jpg"}}
	cache, _ := NewImageCache(t.TempDir())
	_, err := cache.Fetch(context.Background(), idx, id, "../../etc/passwd")
	if err != ErrInvalidSize {
		t.Errorf("invalid size: got %v, want ErrInvalidSize", err)
	}
}

func TestHandlerImageInvalidSize400(t *testing.T) {
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": "https://example.test/x.jpg"}}
	cache, _ := NewImageCache(t.TempDir())
	h := Handler(idx, cache)
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(fmt.Sprintf("%s/cards/%s/image?size=..%%2Fescape", srv.URL, id))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", resp.StatusCode)
	}
}

func TestImageCacheRejectsNonScryfallHost(t *testing.T) {
	// A tampered index that points at an internal URL must not cause
	// an outbound fetch (SSRF). Production traffic requires https to
	// a scryfall host; loopback is the only escape hatch.
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": "https://internal.example.com/secret"}}
	cache, _ := NewImageCache(t.TempDir())
	_, err := cache.Fetch(context.Background(), idx, id, "normal")
	if err == nil || !strings.Contains(err.Error(), "host not allowed") {
		t.Errorf("non-scryfall host: got %v, want host-not-allowed", err)
	}
}

func TestImageCacheRejectsNonHTTPS(t *testing.T) {
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": "http://cards.scryfall.io/x.jpg"}}
	cache, _ := NewImageCache(t.TempDir())
	_, err := cache.Fetch(context.Background(), idx, id, "normal")
	if err == nil || !strings.Contains(err.Error(), "must be https") {
		t.Errorf("plain-http scryfall: got %v, want https-required", err)
	}
}

func TestImageCacheAllowsScryfallHost(t *testing.T) {
	// The validator is URI-shape only — an https://*.scryfall.io URI
	// should pass the allow-check even if the DNS doesn't resolve in
	// the test environment. We don't make a network call here; we
	// just confirm validateImageURI accepts a Scryfall host.
	if err := validateImageURI("https://cards.scryfall.io/normal/front/x/y.jpg"); err != nil {
		t.Errorf("scryfall.io: got %v, want nil", err)
	}
	if err := validateImageURI("https://c1.scryfall.com/front/a/b.jpg"); err != nil {
		t.Errorf("scryfall.com subdomain: got %v, want nil", err)
	}
}

func TestHandlerUnknownCard404(t *testing.T) {
	path := writeFixture(t)
	idx := NewIndex()
	_, _ = idx.Load(path)
	cache, _ := NewImageCache(t.TempDir())
	h := Handler(idx, cache)
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/cards/99999999-9999-9999-9999-999999999999")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
