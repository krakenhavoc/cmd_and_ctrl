package cards

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// faces_test.go — the `?face=N` image route and the on-disk cache
// key it implies (ADR 0034).
//
// The two bugs worth guarding against are opposites. Serving the
// WRONG face happens if the disk key ignores the parameter: the
// first request for either half poisons the cache for the other, and
// they are different pictures. Re-downloading EVERY image happens if
// face 0 gets a suffix it never had: every file already cached in
// every deployment was written under the unqualified name.

// dfcPrint is a modal DFC pointing both faces at a spy CDN.
func dfcPrint(id uuid.UUID, base string) Card {
	return Card{
		ID:     id,
		Name:   "Sea Gate Restoration // Sea Gate, Reborn",
		Layout: "modal_dfc",
		CardFaces: []CardFace{
			{Name: "Sea Gate Restoration", ImageURIs: map[string]string{"normal": base + "/front"}},
			{Name: "Sea Gate, Reborn", ImageURIs: map[string]string{"normal": base + "/back"}},
		},
	}
}

func TestImageURIForFace(t *testing.T) {
	id := uuid.New()
	c := dfcPrint(id, "https://cards.scryfall.io")
	if got := ImageURIForFace(c, "normal", 0); !strings.HasSuffix(got, "/front") {
		t.Errorf("face 0 = %q, want the front face's URI", got)
	}
	if got := ImageURIForFace(c, "normal", 1); !strings.HasSuffix(got, "/back") {
		t.Errorf("face 1 = %q, want the back face's URI", got)
	}
	// ImageURI is face 0 — the pre-ADR-0034 behaviour, unchanged.
	if ImageURI(c, "normal") != ImageURIForFace(c, "normal", 0) {
		t.Error("ImageURI no longer agrees with face 0")
	}

	// A face this card does not have returns EMPTY rather than
	// falling back to the front. A 404 is honest; a picker showing
	// the same art twice is not.
	if got := ImageURIForFace(c, "normal", 5); got != "" {
		t.Errorf("out-of-range face = %q, want empty", got)
	}
	single := Card{ID: id, ImageURIs: map[string]string{"normal": "https://cards.scryfall.io/x"}}
	if got := ImageURIForFace(single, "normal", 1); got != "" {
		t.Errorf("back face of a single-faced card = %q, want empty", got)
	}
	// …but face 0 of a single-faced card still reads the top level.
	if got := ImageURIForFace(single, "normal", 0); got == "" {
		t.Error("face 0 of a single-faced card resolved to nothing")
	}
}

// TestImageCacheKeyIncludesFace is the poisoning guard.
func TestImageCacheKeyIncludesFace(t *testing.T) {
	cdn, hits := spyCDN(t, []byte("bytes"))
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = dfcPrint(id, cdn.URL)
	cache, err := NewImageCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewImageCache: %v", err)
	}

	front, err := cache.FetchFace(context.Background(), idx, id, "normal", 0)
	if err != nil {
		t.Fatalf("FetchFace(0): %v", err)
	}
	back, err := cache.FetchFace(context.Background(), idx, id, "normal", 1)
	if err != nil {
		t.Fatalf("FetchFace(1): %v", err)
	}
	if front == back {
		t.Fatalf("both faces cached to the same path %q — the first "+
			"request for either half would serve the wrong picture", front)
	}
	if hits.Load() != 2 {
		t.Errorf("CDN hits = %d, want 2 (one per face)", hits.Load())
	}
	// Face 0 keeps the ORIGINAL, unsuffixed filename. Anything else
	// orphans every image already on disk in every deployment.
	if strings.Contains(filepath.Base(front), "face") {
		t.Errorf("face 0 path %q carries a face suffix; every cached "+
			"image would be re-downloaded", filepath.Base(front))
	}
	if !strings.Contains(filepath.Base(back), "face1") {
		t.Errorf("face 1 path %q does not name the face", filepath.Base(back))
	}
	// Both are real files with the fetched bytes.
	for _, p := range []string{front, back} {
		if b, err := os.ReadFile(p); err != nil || string(b) != "bytes" {
			t.Errorf("%s: read %q err %v", p, b, err)
		}
	}
}

// TestFetchFaceRejectsOutOfBandFaceIndex: `face` becomes part of a
// filesystem path, so it is bounds-checked for the same reason
// `size` is.
func TestFetchFaceRejectsOutOfBandFaceIndex(t *testing.T) {
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": "https://cards.scryfall.io/x"}}
	cache, _ := NewImageCache(t.TempDir())
	for _, face := range []int{-1, MaxFaceIndex + 1, 1 << 30} {
		if _, err := cache.FetchFace(context.Background(), idx, id, "normal", face); err != ErrInvalidFace {
			t.Errorf("face %d: err = %v, want ErrInvalidFace", face, err)
		}
	}
}

func TestHandlerImageFaceParameter(t *testing.T) {
	cdn, _ := spyCDN(t, []byte("img"))
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = dfcPrint(id, cdn.URL)
	cache, _ := NewImageCache(t.TempDir())
	srv := httptest.NewServer(Handler(idx, cache))
	defer srv.Close()

	get := func(q string) int {
		resp, err := http.Get(fmt.Sprintf("%s/cards/%s/image%s", srv.URL, id, q))
		if err != nil {
			t.Fatalf("GET %s: %v", q, err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}

	// Absent — what every pre-0034 client sends — is the front face.
	if got := get(""); got != http.StatusOK {
		t.Errorf("no face parameter: status %d, want 200", got)
	}
	if got := get("?face=1"); got != http.StatusOK {
		t.Errorf("?face=1: status %d, want 200", got)
	}
	// Not a number is a client bug, refused rather than coerced.
	if got := get("?face=abc"); got != http.StatusBadRequest {
		t.Errorf("?face=abc: status %d, want 400", got)
	}
	// In range but this card has no such face: that is a 404, not a
	// 400 — "not a face" and "not a face OF THIS CARD" are different
	// answers.
	if got := get("?face=3"); got != http.StatusNotFound {
		t.Errorf("?face=3 on a two-faced card: status %d, want 404", got)
	}
	// Out of the allowed band is a 400.
	if got := get(fmt.Sprintf("?face=%d", MaxFaceIndex+1)); got != http.StatusBadRequest {
		t.Errorf("?face beyond MaxFaceIndex: status %d, want 400", got)
	}
}

// TestReversibleCardPrintIsNotPlayable pins the isPlayablePrint
// addition from the index side (ADR 0034). All 81 such printings
// carry mana_cost: null and type_line: null.
func TestReversibleCardPrintIsNotPlayable(t *testing.T) {
	if isPlayablePrint(Card{Layout: "reversible_card"}) {
		t.Error("a reversible_card printing counts as playable")
	}
	for _, layout := range []string{"normal", "modal_dfc", "transform", "split", "adventure"} {
		if !isPlayablePrint(Card{Layout: layout, SetType: "expansion"}) {
			t.Errorf("%q printings count as unplayable", layout)
		}
	}
}
