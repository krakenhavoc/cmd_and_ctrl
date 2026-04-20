package discord

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// pngBytes is a minimal valid PNG header + IHDR + IEND. Good
// enough for tests — we never decode it, just stash bytes and
// read them back.
var pngBytes = []byte{
	0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
	0, 0, 0, 0x0D, 'I', 'H', 'D', 'R',
	0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0,
	0x1F, 0x15, 0xC4, 0x89,
	0, 0, 0, 0, 'I', 'E', 'N', 'D', 0xAE, 0x42, 0x60, 0x82,
}

// stubCDN stands up an httptest.Server that fakes Discord's
// cdn.discordapp.com for (id, hash) URLs. Returns the server
// and a call counter so tests can assert on cache behaviour.
func stubCDN(t *testing.T) (*httptest.Server, *int) {
	t.Helper()
	var calls int
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		if strings.Contains(r.URL.Path, "missing") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func newTestAvatarCache(t *testing.T, cdn *httptest.Server) *AvatarCache {
	t.Helper()
	c := NewAvatarCache(t.TempDir(), cdn.Client())
	// Redirect cdnURL to the stub. We can't override cdnURL() but
	// we can swap via a package-level var if it were one. For now
	// the test patches the httptest.Server URL into a helper
	// wrapper by running the test against a client whose transport
	// rewrites cdn.discordapp.com → srv.URL.
	orig := cdnURLHost
	cdnURLHost = cdn.URL
	t.Cleanup(func() { cdnURLHost = orig })
	return c
}

func TestAvatarCacheHitOnSecondCall(t *testing.T) {
	srv, calls := stubCDN(t)
	c := newTestAvatarCache(t, srv)

	req := httptest.NewRequest("GET", "/avatars/1/abc", nil)
	w1 := httptest.NewRecorder()
	if err := c.Serve(w1, req, "1", "abc"); err != nil {
		t.Fatalf("first Serve: %v", err)
	}
	if w1.Code != http.StatusOK {
		t.Errorf("first status: %d", w1.Code)
	}
	if *calls != 1 {
		t.Errorf("first fetch should hit CDN once; got %d calls", *calls)
	}

	// Second request for the same (id, hash) hits disk and should
	// NOT talk to the CDN.
	w2 := httptest.NewRecorder()
	if err := c.Serve(w2, req, "1", "abc"); err != nil {
		t.Fatalf("second Serve: %v", err)
	}
	if *calls != 1 {
		t.Errorf("second fetch should come from cache; CDN calls now %d, want 1", *calls)
	}
}

func TestAvatarCacheNewHashReFetches(t *testing.T) {
	// A hash rotation (user changed their avatar) must re-fetch.
	srv, calls := stubCDN(t)
	c := newTestAvatarCache(t, srv)

	for _, hash := range []string{"hashone", "hashtwo"} {
		req := httptest.NewRequest("GET", "/avatars/1/"+hash, nil)
		w := httptest.NewRecorder()
		if err := c.Serve(w, req, "1", hash); err != nil {
			t.Fatalf("Serve %s: %v", hash, err)
		}
	}
	if *calls != 2 {
		t.Errorf("new-hash fetch: got %d CDN calls, want 2", *calls)
	}
}

func TestAvatarCacheRejectsPathTraversal(t *testing.T) {
	srv, _ := stubCDN(t)
	c := newTestAvatarCache(t, srv)

	cases := [][2]string{
		{"../etc/passwd", "abc"},
		{"1", "../../secret"},
		{"1/../2", "abc"},
		{"", "abc"},
		{"1", ""},
	}
	for _, kv := range cases {
		req := httptest.NewRequest("GET", "/avatars/"+kv[0]+"/"+kv[1], nil)
		w := httptest.NewRecorder()
		err := c.Serve(w, req, kv[0], kv[1])
		if err != ErrInvalidAvatarKey {
			t.Errorf("id=%q hash=%q: got %v, want ErrInvalidAvatarKey", kv[0], kv[1], err)
		}
	}
}

func TestAvatarCacheDisabledWhenNoDir(t *testing.T) {
	c := NewAvatarCache("", nil)
	req := httptest.NewRequest("GET", "/avatars/1/a", nil)
	w := httptest.NewRecorder()
	if err := c.Serve(w, req, "1", "abc"); err != ErrCacheDisabled {
		t.Errorf("got %v, want ErrCacheDisabled", err)
	}
}

func TestAvatarCache404FromCDN(t *testing.T) {
	srv, _ := stubCDN(t)
	c := newTestAvatarCache(t, srv)

	req := httptest.NewRequest("GET", "/avatars/1/missing", nil)
	w := httptest.NewRecorder()
	if err := c.Serve(w, req, "1", "missing"); err == nil {
		t.Fatal("expected error for CDN 404")
	}
	// And no file should have been left behind.
	if _, statErr := os.Stat(filepath.Join(c.Dir, "1", "missing.png")); !os.IsNotExist(statErr) {
		t.Errorf("404 should not leave a cached file; stat err = %v", statErr)
	}
}
