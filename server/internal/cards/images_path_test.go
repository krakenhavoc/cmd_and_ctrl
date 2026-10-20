package cards

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Regression tests for the go/path-injection alerts on the image cache
// (#4204): a size, face or id taken from a request must never name a
// file outside the cache directory.

func TestCanonicalSizeMatchesValidSizes(t *testing.T) {
	for size := range ValidSizes {
		got, ok := canonicalSize(size)
		if !ok || got != size {
			t.Errorf("canonicalSize(%q) = %q, %v; want itself, true", size, got, ok)
		}
	}
	for _, bad := range []string{"", "NORMAL", "normal ", "../normal", "normal/../..", "%2e%2e", "small\x00", "a/b", `..\large`} {
		if got, ok := canonicalSize(bad); ok || got != "" {
			t.Errorf("canonicalSize(%q) = %q, %v; want refusal", bad, got, ok)
		}
	}
}

func TestPathForFaceStaysUnderRoot(t *testing.T) {
	root := t.TempDir()
	c, err := NewImageCache(root)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	for face := 0; face <= MaxFaceIndex; face++ {
		for size := range ValidSizes {
			p, err := c.pathForFace(id, size, face)
			if err != nil {
				t.Fatalf("pathForFace(%s,%s,%d): %v", id, size, face, err)
			}
			rel, err := filepath.Rel(root, p)
			if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
				t.Errorf("path %q escapes root %q (rel %q)", p, root, rel)
			}
		}
	}
	// Face 0 keeps the original, unsuffixed filename.
	p, _ := c.pathForFace(id, "normal", 0)
	if want := filepath.Join(root, id.String()[:2], id.String()+".normal.jpg"); p != want {
		t.Errorf("face 0 path = %q, want %q", p, want)
	}
}

// canonicalSize is the first guard; this drives the confinement check
// behind it directly, with sizes it would never let through.
func TestPathForFaceRefusesAPathOutsideTheRoot(t *testing.T) {
	c, _ := NewImageCache(t.TempDir())
	id := uuid.New()
	for _, size := range []string{
		"../../../../etc/passwd",
		"../../../../../../../../../../tmp/x",
	} {
		if p, err := c.pathForFace(id, size, 0); !errors.Is(err, ErrInvalidPath) {
			t.Errorf("pathForFace(%q) = %q, %v; want ErrInvalidPath", size, p, err)
		}
	}
}

func TestImageCacheRefusesHostileSizesAndFaces(t *testing.T) {
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": "https://example.test/x.jpg"}}
	c, _ := NewImageCache(t.TempDir())
	for _, size := range []string{"../../etc/passwd", "..%2f..%2fetc%2fpasswd", `..\..\x`, "normal/../../x", "NORMAL", "small\x00../x"} {
		if _, err := c.Fetch(context.Background(), idx, id, size); !errors.Is(err, ErrInvalidSize) {
			t.Errorf("size %q: got %v, want ErrInvalidSize", size, err)
		}
	}
	for _, face := range []int{-1, MaxFaceIndex + 1, 1 << 30, -1 << 30} {
		if _, err := c.FetchFace(context.Background(), idx, id, "normal", face); !errors.Is(err, ErrInvalidFace) {
			t.Errorf("face %d: got %v, want ErrInvalidFace", face, err)
		}
	}
}

func TestHandlerImageRefusesTraversalInIDSizeAndFace(t *testing.T) {
	id := uuid.New()
	idx := NewIndex()
	idx.byID[id] = Card{ID: id, ImageURIs: map[string]string{"normal": "https://example.test/x.jpg"}}
	c, _ := NewImageCache(t.TempDir())
	srv := httptest.NewServer(Handler(idx, c))
	defer srv.Close()

	cases := []struct{ name, path string }{
		{"dotdot id", "/cards/..%2f..%2fetc%2fpasswd/image"},
		{"encoded dotdot id", "/cards/%2e%2e%2f%2e%2e/image"},
		{"backslash id", "/cards/..%5c..%5cx/image"},
		{"nul id", "/cards/" + id.String() + "%00/image"},
		{"size traversal", fmt.Sprintf("/cards/%s/image?size=../../x", id)},
		{"size encoded", fmt.Sprintf("/cards/%s/image?size=%%2e%%2e%%2f%%2e%%2e%%2fx", id)},
		{"face traversal", fmt.Sprintf("/cards/%s/image?face=../1", id)},
		{"face huge", fmt.Sprintf("/cards/%s/image?face=99999999999", id)},
		{"face negative", fmt.Sprintf("/cards/%s/image?face=-1", id)},
	}
	for _, tc := range cases {
		resp, err := http.Get(srv.URL + tc.path)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", tc.name, resp.StatusCode)
		}
	}
}
