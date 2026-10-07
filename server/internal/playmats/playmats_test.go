package playmats

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openStore(t *testing.T) (*SQLStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d), d
}

func mustUser(t *testing.T, d *db.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'A', 1, 1)`, id.String()); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

// A 1×1 PNG: the smallest real image http.DetectContentType names.
var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82")

const fileA = "0123456789abcdef0123456789abcdef.png"
const fileB = "fedcba9876543210fedcba9876543210.jpg"

func TestStorePutReplacesAndReportsTheOldFile(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d)
	if _, err := s.Get(ctx, u); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get before Put = %v, want ErrNotFound", err)
	}
	old, err := s.Put(ctx, u, fileA, 60)
	if err != nil || old.File != "" {
		t.Fatalf("first Put = %+v, %v; want no old playmat", old, err)
	}
	old, err = s.Put(ctx, u, fileB, 45)
	if err != nil || old.File != fileA || old.Wash != 60 {
		t.Fatalf("second Put = %+v, %v; want the first as old", old, err)
	}
	got, err := s.Get(ctx, u)
	if err != nil || got.File != fileB || got.Wash != 45 || got.Path() != "/playmats/"+fileB {
		t.Fatalf("Get = %+v, %v", got, err)
	}
}

func TestStoreRefusesABadWashOrFile(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d)
	for _, wash := range []int{MinWash - 1, MaxWash + 1} {
		if _, err := s.Put(ctx, u, fileA, wash); !errors.Is(err, ErrInvalid) {
			t.Errorf("Put wash %d = %v, want ErrInvalid", wash, err)
		}
	}
	for _, file := range []string{"", "../x.png", "abc.png", strings.Repeat("a", 32) + ".svg"} {
		if _, err := s.Put(ctx, u, file, 60); !errors.Is(err, ErrInvalid) {
			t.Errorf("Put file %q = %v, want ErrInvalid", file, err)
		}
	}
	if _, err := s.SetWash(ctx, u, 50); !errors.Is(err, ErrNotFound) {
		t.Errorf("SetWash with no playmat = %v, want ErrNotFound", err)
	}
}

func TestStoreSetWashAndDelete(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d)
	if _, err := s.Put(ctx, u, fileA, 60); err != nil {
		t.Fatal(err)
	}
	got, err := s.SetWash(ctx, u, 80)
	if err != nil || got.Wash != 80 || got.File != fileA {
		t.Fatalf("SetWash = %+v, %v", got, err)
	}
	old, err := s.Delete(ctx, u)
	if err != nil || old.File != fileA {
		t.Fatalf("Delete = %+v, %v", old, err)
	}
	if _, err := s.Delete(ctx, u); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Delete = %v, want ErrNotFound", err)
	}
}

func TestFilesSaveSniffsTheBytesNotTheName(t *testing.T) {
	f, err := NewFiles(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	name, err := f.Save(bytes.NewReader(pngBytes))
	if err != nil || !ValidFile(name) || !strings.HasSuffix(name, ".png") {
		t.Fatalf("Save png = %q, %v", name, err)
	}
	for label, body := range map[string][]byte{
		"svg":  []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
		"html": []byte("<!doctype html><script>alert(1)</script>"),
		"text": []byte("not an image"),
	} {
		if _, err := f.Save(bytes.NewReader(body)); !errors.Is(err, ErrInvalid) {
			t.Errorf("Save %s = %v, want ErrInvalid", label, err)
		}
	}
	if _, err := f.Save(bytes.NewReader(nil)); !errors.Is(err, ErrInvalid) {
		t.Errorf("Save empty = %v, want ErrInvalid", err)
	}
}

func TestFilesSaveRefusesPastTheCapAndKeepsNothing(t *testing.T) {
	dir := t.TempDir()
	f, _ := NewFiles(dir)
	big := append(append([]byte{}, pngBytes...), make([]byte, MaxImageBytes)...)
	if _, err := f.Save(bytes.NewReader(big)); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Save over the cap = %v, want ErrTooLarge", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("left %d files behind", len(entries))
	}
}

func TestFilesServeChecksTheNameAndSetsInertHeaders(t *testing.T) {
	dir := t.TempDir()
	f, _ := NewFiles(dir)
	name, err := f.Save(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	if err := f.Serve(rec, httptest.NewRequest(http.MethodGet, "/playmats/"+name, nil), name); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	h := rec.Header()
	if h.Get("Content-Type") != "image/png" || h.Get("X-Content-Type-Options") != "nosniff" ||
		!strings.Contains(h.Get("Content-Security-Policy"), "sandbox") || !bytes.Equal(rec.Body.Bytes(), pngBytes) {
		t.Errorf("headers %v", h)
	}
	for _, bad := range []string{"../../etc/passwd", "x.png", name + "x"} {
		if err := f.Serve(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), bad); !errors.Is(err, ErrNotFound) {
			t.Errorf("Serve %q = %v, want ErrNotFound", bad, err)
		}
	}
	// A hand-placed file under a valid name that is not an image is
	// never served as one.
	fake := strings.Repeat("a", 32) + ".png"
	if err := os.WriteFile(filepath.Join(dir, fake), []byte("<script>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := f.Serve(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), fake); !errors.Is(err, ErrNotFound) {
		t.Errorf("Serve of a non-image = %v, want ErrNotFound", err)
	}
	if err := f.Remove(name); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
		t.Errorf("Remove left the file: %v", err)
	}
}

func TestFilesWithNoDirectoryStoreNothing(t *testing.T) {
	f, _ := NewFiles("")
	if _, err := f.Save(bytes.NewReader(pngBytes)); !errors.Is(err, ErrNoStore) {
		t.Errorf("Save = %v, want ErrNoStore", err)
	}
}
