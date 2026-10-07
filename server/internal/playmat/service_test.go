package playmat

import (
	"context"
	"errors"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func newService(t *testing.T, f *Fetcher) (*Service, string, uuid.UUID) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	user := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'A', 1, 1)`, user.String()); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "playmats")
	return NewService(NewFileStore(dir), d.DB, f), dir, user
}

func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range ents {
		out = append(out, e.Name())
	}
	return out
}

func TestSetStoresAJPEGAndPointsTheUserAtIt(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	if got := s.URL(user); got != "" {
		t.Fatalf("a new user has URL %q", got)
	}
	info, err := s.SetFromBytes(ctx, user, pngBytes(t, solid(64, 48, color.White)))
	if err != nil {
		t.Fatal(err)
	}
	if info.Width != 64 || info.Height != 48 || !strings.HasPrefix(info.URL, "/playmats/") {
		t.Errorf("info = %+v", info)
	}
	if s.URL(user) != info.URL {
		t.Errorf("URL() = %q, want %q", s.URL(user), info.URL)
	}
	got, ok, err := s.Get(ctx, user)
	if err != nil || !ok || got != info {
		t.Errorf("Get = %+v, %v, %v; want %+v", got, ok, err, info)
	}
	if fs := filesIn(t, dir); len(fs) != 1 || fs[0] != info.ID+".jpg" {
		t.Errorf("files = %v", fs)
	}
}

func TestReplacingAPlaymatDeletesTheOldFile(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	first, err := s.SetFromBytes(ctx, user, pngBytes(t, solid(20, 20, color.White)))
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.SetFromBytes(ctx, user, pngBytes(t, solid(30, 30, color.Black)))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("a replacement kept the id")
	}
	fs := filesIn(t, dir)
	if len(fs) != 1 || fs[0] != second.ID+".jpg" {
		t.Errorf("files = %v, want only the new one", fs)
	}
	if _, _, err := s.Open(first.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("the old id still opens: %v", err)
	}
	if s.URL(user) != second.URL {
		t.Errorf("URL() = %q, want %q", s.URL(user), second.URL)
	}
}

func TestARejectedReplacementKeepsTheOldPlaymat(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	first, err := s.SetFromBytes(ctx, user, pngBytes(t, solid(20, 20, color.White)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetFromBytes(ctx, user, []byte("not an image")); !errors.Is(err, ErrNotImage) {
		t.Fatalf("err = %v", err)
	}
	if s.URL(user) != first.URL {
		t.Error("a refused upload changed the playmat")
	}
	if fs := filesIn(t, dir); len(fs) != 1 {
		t.Errorf("files = %v", fs)
	}
}

func TestRemoveDeletesTheFileAndClearsThePointer(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	if _, err := s.SetFromBytes(ctx, user, pngBytes(t, solid(20, 20, color.White))); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(ctx, user); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != "" {
		t.Error("URL after Remove is not empty")
	}
	if _, ok, _ := s.Get(ctx, user); ok {
		t.Error("Get reports a playmat after Remove")
	}
	if fs := filesIn(t, dir); len(fs) != 0 {
		t.Errorf("files = %v, want none", fs)
	}
	if err := s.Remove(ctx, user); err != nil {
		t.Errorf("removing none: %v", err)
	}
}

func TestAnUnknownUserHasNoPlaymatAndCannotSetOne(t *testing.T) {
	s, dir, _ := newService(t, nil)
	ghost := uuid.New()
	if _, err := s.SetFromBytes(context.Background(), ghost, pngBytes(t, solid(8, 8, color.White))); !errors.Is(err, ErrNoUser) {
		t.Errorf("err = %v, want ErrNoUser", err)
	}
	if fs := filesIn(t, dir); len(fs) != 0 {
		t.Errorf("a failed set left %v behind", fs)
	}
}

func TestDisabledServices(t *testing.T) {
	var nilSvc *Service
	if nilSvc.Enabled() || nilSvc.URL(uuid.New()) != "" {
		t.Error("a nil service is enabled")
	}
	noDir := NewService(NewFileStore(""), nil, nil)
	if noDir.Enabled() {
		t.Error("a service with no directory and no database is enabled")
	}
	if _, err := noDir.SetFromBytes(context.Background(), uuid.New(), nil); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
	if _, err := noDir.SetFromURL(context.Background(), uuid.New(), "https://example.com/"); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
	if _, _, err := noDir.Get(context.Background(), uuid.New()); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
}

func TestSetFromURLStoresTheBytesNotTheLink(t *testing.T) {
	img := pngBytes(t, solid(24, 16, color.RGBA{9, 9, 200, 255}))
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(img) }))
	defer srv.Close()
	s, dir, user := newService(t, testFetcher(srv, fetcherHooks{allow: guard}))
	info, err := s.SetFromURL(context.Background(), user, srv.URL+"/mat.png")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(info.URL, "127.0.0.1") || !strings.HasPrefix(info.URL, "/playmats/") {
		t.Errorf("the stored playmat points at %q; it must be our own route", info.URL)
	}
	if fs := filesIn(t, dir); len(fs) != 1 {
		t.Errorf("files = %v", fs)
	}
	// And a link that fails stores nothing and keeps what was there.
	if _, err := s.SetFromURL(context.Background(), user, "https://10.0.0.1/x.png"); !errors.Is(err, ErrFetch) {
		t.Errorf("err = %v, want ErrFetch", err)
	}
	if s.URL(user) != info.URL {
		t.Error("a refused link changed the playmat")
	}
}

func TestOpenRefusesAnythingThatIsNotAMintedID(t *testing.T) {
	s, _, _ := newService(t, nil)
	for _, id := range []string{"", "..", "../x", "/etc/passwd", "a/b", "A0000000-0000-4000-8000-000000000000", "x.jpg", strings.Repeat("a", 300)} {
		if _, _, err := s.Open(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("Open(%q) = %v, want ErrNotFound", id, err)
		}
	}
}
