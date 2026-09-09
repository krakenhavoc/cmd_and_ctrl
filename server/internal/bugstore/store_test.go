package bugstore

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testBaseURL = "https://cmd.example.test"

// pngBytes is a minimal but real PNG: the 8-byte signature is what
// http.DetectContentType keys on, so this is enough to exercise the
// sniffing allowlist without embedding a fixture file.
func pngBytes(pad int) []byte {
	b := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	return append(b, bytes.Repeat([]byte{0x00}, pad)...)
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return New(t.TempDir(), testBaseURL)
}

func TestDisabledWithoutDir(t *testing.T) {
	for name, s := range map[string]*Store{
		"no dir":  New("", testBaseURL),
		"neither": New("", ""),
	} {
		if s.Enabled() {
			t.Errorf("%s: Enabled() = true, want false", name)
		}
		if _, err := s.NewReport(); !errors.Is(err, ErrDisabled) {
			t.Errorf("%s: NewReport err = %v, want ErrDisabled", name, err)
		}
	}
}

func TestNoBaseURLDisablesAttachmentsOnly(t *testing.T) {
	s := New(t.TempDir(), "")
	if !s.Enabled() {
		t.Fatal("Enabled() = false, want true with data dir present")
	}
	if s.AttachmentsEnabled() {
		t.Fatal("AttachmentsEnabled() = true, want false without base URL")
	}
	r, err := s.NewReport()
	if err != nil {
		t.Fatalf("NewReport: %v", err)
	}
	if _, err := r.AddImage(pngBytes(64)); !errors.Is(err, ErrDisabled) {
		t.Fatalf("AddImage err = %v, want ErrDisabled", err)
	}
}

func TestAddImageMintsCamoReachableURL(t *testing.T) {
	s := newTestStore(t)
	r, err := s.NewReport()
	if err != nil {
		t.Fatalf("NewReport: %v", err)
	}
	img, err := r.AddImage(pngBytes(64))
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	if img.MIME != "image/png" {
		t.Errorf("MIME = %q, want image/png", img.MIME)
	}
	if img.Name != "att-1.png" {
		t.Errorf("Name = %q, want att-1.png", img.Name)
	}
	// Absolute — markdown in a GitHub issue resolves a relative URL
	// against github.com, so a relative path would 404 silently.
	want := testBaseURL + "/bugreport/att/" + r.ID + "/att-1.png"
	if img.URL != want {
		t.Errorf("URL = %q, want %q", img.URL, want)
	}
	if _, err := os.Stat(filepath.Join(s.dir, r.ID, "att-1.png")); err != nil {
		t.Errorf("image not on disk: %v", err)
	}
}

// The client's declared content type is irrelevant; only the bytes
// decide. This is the stored-XSS guard for the unauthenticated route.
func TestAddImageRejectsNonRaster(t *testing.T) {
	s := newTestStore(t)
	cases := map[string][]byte{
		"html":       []byte("<html><script>alert(1)</script></html>"),
		"svg":        []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
		"plain text": []byte("just a description, not an image at all"),
		"pdf":        []byte("%PDF-1.7\n%????\n"),
		"empty":      {},
	}
	for name, data := range cases {
		r, err := s.NewReport()
		if err != nil {
			t.Fatalf("%s: NewReport: %v", name, err)
		}
		if _, err := r.AddImage(data); !errors.Is(err, ErrUnsupportedType) {
			t.Errorf("%s: AddImage err = %v, want ErrUnsupportedType", name, err)
		}
	}
}

func TestAddImageEnforcesCaps(t *testing.T) {
	s := newTestStore(t)

	t.Run("single image too large", func(t *testing.T) {
		r, _ := s.NewReport()
		if _, err := r.AddImage(pngBytes(MaxImageBytes)); !errors.Is(err, ErrTooLarge) {
			t.Errorf("err = %v, want ErrTooLarge", err)
		}
	})

	t.Run("running total too large", func(t *testing.T) {
		r, _ := s.NewReport()
		// Three at 4 MiB is 12 MiB — over the 10 MiB aggregate, so the
		// third must be refused even though each one is legal alone.
		each := MaxImageBytes - 16
		if _, err := r.AddImage(pngBytes(each)); err != nil {
			t.Fatalf("first: %v", err)
		}
		if _, err := r.AddImage(pngBytes(each)); err != nil {
			t.Fatalf("second: %v", err)
		}
		if _, err := r.AddImage(pngBytes(each)); !errors.Is(err, ErrTooLarge) {
			t.Errorf("third err = %v, want ErrTooLarge", err)
		}
	})

	t.Run("too many images", func(t *testing.T) {
		r, _ := s.NewReport()
		for i := 0; i < MaxImages; i++ {
			if _, err := r.AddImage(pngBytes(8)); err != nil {
				t.Fatalf("image %d: %v", i, err)
			}
		}
		if _, err := r.AddImage(pngBytes(8)); !errors.Is(err, ErrTooMany) {
			t.Errorf("err = %v, want ErrTooMany", err)
		}
	})
}

func TestPinReplayCopiesWholeFile(t *testing.T) {
	s := newTestStore(t)
	src := filepath.Join(t.TempDir(), "game.jsonl")
	body := "{\"seq\":1}\n{\"seq\":2}\n{\"seq\":3}\n"
	if err := os.WriteFile(src, []byte(body), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	r, _ := s.NewReport()
	pin, err := r.PinReplay(src)
	if err != nil {
		t.Fatalf("PinReplay: %v", err)
	}
	if pin.Truncated {
		t.Error("Truncated = true, want false for a small file")
	}
	if pin.Lines != 3 {
		t.Errorf("Lines = %d, want 3", pin.Lines)
	}
	got, err := os.ReadFile(filepath.Join(s.dir, r.ID, replayFileName))
	if err != nil {
		t.Fatalf("read pin: %v", err)
	}
	if string(got) != body {
		t.Errorf("pinned content = %q, want %q", got, body)
	}
}

// The pin is a snapshot, not a link: appending to the source after the
// report is filed must not change what the operator downloads. This is
// the whole reason the feature copies instead of referencing.
func TestPinReplayIsASnapshot(t *testing.T) {
	s := newTestStore(t)
	src := filepath.Join(t.TempDir(), "game.jsonl")
	if err := os.WriteFile(src, []byte("{\"seq\":1}\n"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	r, _ := s.NewReport()
	if _, err := r.PinReplay(src); err != nil {
		t.Fatalf("PinReplay: %v", err)
	}
	// The game plays on, then the room is evicted and the file is gone.
	if err := os.WriteFile(src, []byte("{\"seq\":1}\n{\"seq\":2}\n"), 0o600); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := os.Remove(src); err != nil {
		t.Fatalf("remove src: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(s.dir, r.ID, replayFileName))
	if err != nil {
		t.Fatalf("read pin after source removal: %v", err)
	}
	if string(got) != "{\"seq\":1}\n" {
		t.Errorf("pinned content = %q, want the snapshot taken at report time", got)
	}
}

func TestPinReplayKeepsTailWhenOversized(t *testing.T) {
	s := newTestStore(t)
	src := filepath.Join(t.TempDir(), "big.jsonl")
	// One line just over the cap, then a short distinctive tail.
	head := strings.Repeat("x", MaxReplayPinBytes) + "\n"
	if err := os.WriteFile(src, []byte(head+"{\"seq\":999}\n"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	r, _ := s.NewReport()
	pin, err := r.PinReplay(src)
	if err != nil {
		t.Fatalf("PinReplay: %v", err)
	}
	if !pin.Truncated {
		t.Error("Truncated = false, want true")
	}
	if pin.Bytes != MaxReplayPinBytes {
		t.Errorf("Bytes = %d, want %d", pin.Bytes, MaxReplayPinBytes)
	}
	got, err := os.ReadFile(filepath.Join(s.dir, r.ID, replayFileName))
	if err != nil {
		t.Fatalf("read pin: %v", err)
	}
	// The end of the game — where the bug is — must survive.
	if !bytes.HasSuffix(got, []byte("{\"seq\":999}\n")) {
		t.Error("tail of the source was not kept")
	}
}

func TestPinReplayMissingSource(t *testing.T) {
	s := newTestStore(t)
	r, _ := s.NewReport()
	for name, path := range map[string]string{
		"empty path":   "",
		"missing file": filepath.Join(t.TempDir(), "nope.jsonl"),
	} {
		if _, err := r.PinReplay(path); !errors.Is(err, ErrNotFound) {
			t.Errorf("%s: err = %v, want ErrNotFound", name, err)
		}
	}
}

func TestServeImageSetsInertHeaders(t *testing.T) {
	s := newTestStore(t)
	r, _ := s.NewReport()
	img, err := r.AddImage(pngBytes(32))
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bugreport/att/"+r.ID+"/"+img.Name, nil)
	if err := s.ServeImage(rec, req, r.ID, img.Name); err != nil {
		t.Fatalf("ServeImage: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	// Every one of these is load-bearing on an unauthenticated route
	// that serves reporter-supplied bytes.
	want := map[string]string{
		"Content-Type":            "image/png",
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "default-src 'none'; sandbox",
	}
	for h, v := range want {
		if got := rec.Header().Get(h); got != v {
			t.Errorf("%s = %q, want %q", h, got, v)
		}
	}
}

// The id is a path segment on an unauthenticated route, so traversal
// and malformed-id handling is the security boundary.
func TestServeImageRejectsBadPaths(t *testing.T) {
	s := newTestStore(t)
	r, _ := s.NewReport()
	if _, err := r.AddImage(pngBytes(32)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	// A file the route must never hand out, sitting where a traversal
	// would land.
	secret := filepath.Join(s.dir, "..", "secret.png")
	if err := os.WriteFile(secret, pngBytes(16), 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	cases := []struct{ id, name string }{
		{"../..", "att-1.png"},
		{r.ID, "../../secret.png"},
		{r.ID, "replay.jsonl"},  // the admin artifact is not public
		{r.ID, "manifest.json"}, // nor is the manifest
		{r.ID, "att-1.svg"},     // not an allowlisted extension
		{"not-a-uuid", "att-1.png"},
		{r.ID, "att-99.png"}, // well-formed name, no such file
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/bugreport/att/x/y", nil)
		if err := s.ServeImage(rec, req, c.id, c.name); !errors.Is(err, ErrNotFound) {
			t.Errorf("ServeImage(%q, %q) err = %v, want ErrNotFound", c.id, c.name, err)
		}
	}
}

func TestOpenReplayAndNotFound(t *testing.T) {
	s := newTestStore(t)
	src := filepath.Join(t.TempDir(), "game.jsonl")
	if err := os.WriteFile(src, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	r, _ := s.NewReport()
	if _, err := r.PinReplay(src); err != nil {
		t.Fatalf("PinReplay: %v", err)
	}
	f, info, err := s.OpenReplay(r.ID)
	if err != nil {
		t.Fatalf("OpenReplay: %v", err)
	}
	defer func() { _ = f.Close() }()
	if info.Size() != 3 {
		t.Errorf("size = %d, want 3", info.Size())
	}

	// A report with no pinned replay, and a malformed id.
	empty, _ := s.NewReport()
	for name, id := range map[string]string{"no pin": empty.ID, "bad id": "../../etc"} {
		if _, _, err := s.OpenReplay(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("%s: err = %v, want ErrNotFound", name, err)
		}
	}
}

func TestDiscardRemovesArtifacts(t *testing.T) {
	s := newTestStore(t)
	r, _ := s.NewReport()
	if _, err := r.AddImage(pngBytes(32)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	r.Discard()
	if _, err := os.Stat(filepath.Join(s.dir, r.ID)); !os.IsNotExist(err) {
		t.Errorf("report dir survived Discard: %v", err)
	}
}

func TestCommitWritesManifest(t *testing.T) {
	s := newTestStore(t)
	r, _ := s.NewReport()
	if _, err := r.AddImage(pngBytes(32)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	if err := r.Commit("game-1", "https://github.com/o/r/issues/7"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	buf, err := os.ReadFile(filepath.Join(s.dir, r.ID, manifestFileName))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	for _, want := range []string{r.ID, "game-1", "issues/7", "att-1.png"} {
		if !strings.Contains(string(buf), want) {
			t.Errorf("manifest missing %q:\n%s", want, buf)
		}
	}
}

func TestPruneRemovesOnlyOldReports(t *testing.T) {
	s := newTestStore(t)
	old, _ := s.NewReport()
	fresh, _ := s.NewReport()
	// A stray non-report directory must be left strictly alone — the
	// bugreports dir is inside the shared data dir.
	stray := filepath.Join(s.dir, "not-a-report")
	if err := os.MkdirAll(stray, 0o755); err != nil {
		t.Fatalf("mkdir stray: %v", err)
	}
	past := time.Now().Add(-100 * 24 * time.Hour)
	for _, d := range []string{filepath.Join(s.dir, old.ID), stray} {
		if err := os.Chtimes(d, past, past); err != nil {
			t.Fatalf("chtimes: %v", err)
		}
	}

	removed, err := s.Prune(DefaultRetention, 0)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}
	if _, err := os.Stat(filepath.Join(s.dir, old.ID)); !os.IsNotExist(err) {
		t.Error("old report survived Prune")
	}
	if _, err := os.Stat(filepath.Join(s.dir, fresh.ID)); err != nil {
		t.Errorf("fresh report was pruned: %v", err)
	}
	if _, err := os.Stat(stray); err != nil {
		t.Errorf("Prune deleted a non-report directory: %v", err)
	}
}

// Every key in imageTypes must be exactly what DetectContentType
// returns, or the format silently stops being accepted.
func TestAddImageAcceptsEveryAllowlistedFormat(t *testing.T) {
	s := newTestStore(t)
	samples := map[string][]byte{
		"image/png":  pngBytes(16),
		"image/jpeg": append([]byte{0xff, 0xd8, 0xff, 0xe0}, bytes.Repeat([]byte{0}, 16)...),
		"image/gif":  append([]byte("GIF89a"), bytes.Repeat([]byte{0}, 16)...),
		"image/webp": append(append([]byte("RIFF"), []byte{0x20, 0, 0, 0}...), []byte("WEBPVP8 ")...),
	}
	for want, data := range samples {
		r, err := s.NewReport()
		if err != nil {
			t.Fatalf("%s: NewReport: %v", want, err)
		}
		img, err := r.AddImage(data)
		if err != nil {
			t.Errorf("%s: AddImage: %v", want, err)
			continue
		}
		if img.MIME != want {
			t.Errorf("MIME = %q, want %q", img.MIME, want)
		}
		// And it must survive the round trip out through the route.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		if err := s.ServeImage(rec, req, r.ID, img.Name); err != nil {
			t.Errorf("%s: ServeImage: %v", want, err)
		}
		if got := rec.Header().Get("Content-Type"); got != want {
			t.Errorf("served Content-Type = %q, want %q", got, want)
		}
	}
}

// Retention alone is not a disk guarantee — the rate limiter permits
// tens of GiB inside the 90-day window, on a box whose data dir also
// holds a 630 MiB card dump.
func TestPruneEnforcesTotalSizeCapOldestFirst(t *testing.T) {
	s := newTestStore(t)
	// Three reports, ~1 MiB each, aged so the eviction order is
	// unambiguous.
	var ids []string
	for i := 0; i < 3; i++ {
		r, err := s.NewReport()
		if err != nil {
			t.Fatalf("NewReport: %v", err)
		}
		if _, err := r.AddImage(pngBytes(1 << 20)); err != nil {
			t.Fatalf("AddImage: %v", err)
		}
		ids = append(ids, r.ID)
		when := time.Now().Add(time.Duration(i-3) * time.Hour)
		if err := os.Chtimes(filepath.Join(s.dir, r.ID), when, when); err != nil {
			t.Fatalf("chtimes: %v", err)
		}
	}

	// Cap at ~2.5 MiB: the oldest must go, the two newest must stay.
	removed, err := s.Prune(DefaultRetention, 5<<19)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	if _, err := os.Stat(filepath.Join(s.dir, ids[0])); !os.IsNotExist(err) {
		t.Error("oldest report survived the size cap")
	}
	for _, id := range ids[1:] {
		if _, err := os.Stat(filepath.Join(s.dir, id)); err != nil {
			t.Errorf("newer report %s was evicted: %v", id, err)
		}
	}
}

// A zero cap means "age only": callers use it to ask whether anything
// is left without imposing a size policy.
func TestPruneZeroCapIgnoresSize(t *testing.T) {
	s := newTestStore(t)
	r, err := s.NewReport()
	if err != nil {
		t.Fatalf("NewReport: %v", err)
	}
	if _, err := r.AddImage(pngBytes(1 << 20)); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	removed, err := s.Prune(DefaultRetention, 0)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 0 {
		t.Errorf("removed = %d, want 0", removed)
	}
}
