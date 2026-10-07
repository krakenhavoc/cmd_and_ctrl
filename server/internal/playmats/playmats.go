// Package playmats is ADR 0128: the image a signed-in person puts
// behind their part of the board, which everyone at the table sees.
//
// Two halves. Store is the user_playmats row (migration 0011): which
// file is theirs and how strongly it is darkened under the cards.
// Files is the image itself, under $CMDCTRL_DATA_DIR/playmats/, stored
// under a random name so a new upload is a new URL and every copy of
// the old one can be cached for good.
//
// An upload is checked the way a bug report's screenshot is
// (internal/bugstore): its type is sniffed from the bytes, never taken
// from the client, and only PNG, JPEG, GIF and WebP are kept. SVG is
// never one of them: it is a document that can run script, and these
// images are served to every player at the table.
package playmats

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxImageBytes caps one playmat image. A 2560px-wide JPEG or WebP
	// is well under it; a lossless PNG of the same can be over, which
	// the error message says.
	MaxImageBytes = 4 << 20
	// MinWash, MaxWash and DefaultWash bound the dark wash over the
	// image, in percent. The floor keeps cards, labels and counters
	// readable over busy art; the ceiling still shows the image.
	MinWash     = 30
	MaxWash     = 90
	DefaultWash = 60
)

var (
	// ErrNotFound: the user has no playmat, or the file is not one of
	// ours.
	ErrNotFound = errors.New("playmats: not found")
	// ErrInvalid wraps every refusal of an upload or a wash.
	ErrInvalid = errors.New("playmats: invalid playmat")
	// ErrTooLarge: the image is over MaxImageBytes.
	ErrTooLarge = errors.New("playmats: image is larger than 4 MiB")
	// ErrNoStore is what NoStore's writes answer, and what Files
	// answers with no directory.
	ErrNoStore = errors.New("playmats: no store configured")
)

// Playmat is one user_playmats row.
type Playmat struct {
	// File is the image's name under the playmats directory.
	File string
	// Wash is the dark wash over it, MinWash to MaxWash.
	Wash      int
	UpdatedAt time.Time
}

// Path is the URL path the image is served at.
func (p Playmat) Path() string {
	if p.File == "" {
		return ""
	}
	return "/playmats/" + p.File
}

// Store reads and writes one person's playmat. Safe for concurrent use.
type Store interface {
	// Get reads the user's playmat. ErrNotFound if there is none.
	Get(ctx context.Context, user uuid.UUID) (Playmat, error)
	// Put sets the user's playmat and returns the one it replaced (a
	// zero Playmat if there was none), so the caller can remove the
	// old file.
	Put(ctx context.Context, user uuid.UUID, file string, wash int) (old Playmat, err error)
	// SetWash changes the wash of an existing playmat. ErrNotFound if
	// there is none.
	SetWash(ctx context.Context, user uuid.UUID, wash int) (Playmat, error)
	// Delete removes the user's playmat and returns it. ErrNotFound if
	// there was none.
	Delete(ctx context.Context, user uuid.UUID) (Playmat, error)
}

// ValidWash reports whether wash is a wash a playmat may have.
func ValidWash(wash int) error {
	if wash < MinWash || wash > MaxWash {
		return fmt.Errorf("%w\nwash must be between %d and %d", ErrInvalid, MinWash, MaxWash)
	}
	return nil
}

// NoStore is the Store for a deployment with no database. No principal
// carries a UserID there, so nothing calls it except defensively.
type NoStore struct{}

// Get always reports ErrNotFound.
func (NoStore) Get(context.Context, uuid.UUID) (Playmat, error) { return Playmat{}, ErrNotFound }

// Put always refuses.
func (NoStore) Put(context.Context, uuid.UUID, string, int) (Playmat, error) {
	return Playmat{}, ErrNoStore
}

// SetWash always refuses.
func (NoStore) SetWash(context.Context, uuid.UUID, int) (Playmat, error) {
	return Playmat{}, ErrNoStore
}

// Delete always reports ErrNotFound.
func (NoStore) Delete(context.Context, uuid.UUID) (Playmat, error) { return Playmat{}, ErrNotFound }

var _ Store = NoStore{}

// imageTypes is the allowlist: the types http.DetectContentType
// reports for the formats a browser draws, onto the extension the file
// is stored under. A type absent here is refused, whatever the client
// said it sent.
var imageTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// fileNamePattern is every name Files writes: 32 hex characters and an
// allowlisted extension. Serve and Remove refuse anything else, so a
// request can never name a path outside the directory.
var fileNamePattern = regexp.MustCompile(`^[0-9a-f]{32}\.(png|jpg|gif|webp)$`)

// ValidFile reports whether name is a name Files could have written.
func ValidFile(name string) bool { return fileNamePattern.MatchString(name) }

// Files is the playmat images on disk. The zero value (or one with no
// directory) stores nothing: Save answers ErrNoStore.
type Files struct {
	dir string
}

// NewFiles roots the images at dir, creating it. An empty dir disables
// uploads (no CMDCTRL_DATA_DIR).
func NewFiles(dir string) (*Files, error) {
	if dir == "" {
		return &Files{}, nil
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("playmats: create %s: %w", dir, err)
	}
	return &Files{dir: dir}, nil
}

// Enabled reports whether images can be stored.
func (f *Files) Enabled() bool { return f != nil && f.dir != "" }

// sniff is the image type of head, or "" if it is not an allowlisted
// image.
func sniff(head []byte) string {
	mime := http.DetectContentType(head)
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if _, ok := imageTypes[mime]; !ok {
		return ""
	}
	return mime
}

// Save stores the image read from r under a new random name and
// returns the name. It reads at most MaxImageBytes+1 bytes: past the
// cap it answers ErrTooLarge and keeps nothing.
func (f *Files) Save(r io.Reader) (string, error) {
	if !f.Enabled() {
		return "", ErrNoStore
	}
	data, err := io.ReadAll(io.LimitReader(r, MaxImageBytes+1))
	if err != nil {
		return "", fmt.Errorf("playmats: read upload: %w", err)
	}
	if len(data) > MaxImageBytes {
		return "", ErrTooLarge
	}
	if len(data) == 0 {
		return "", fmt.Errorf("%w\nthe image is empty", ErrInvalid)
	}
	mime := sniff(data[:min(len(data), 512)])
	if mime == "" {
		return "", fmt.Errorf("%w\na playmat must be a PNG, JPEG, GIF or WebP image", ErrInvalid)
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("playmats: name: %w", err)
	}
	name := hex.EncodeToString(b[:]) + "." + imageTypes[mime]
	tmp, err := os.CreateTemp(f.dir, ".upload-*")
	if err != nil {
		return "", fmt.Errorf("playmats: temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("playmats: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("playmats: write: %w", err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(f.dir, name)); err != nil {
		return "", fmt.Errorf("playmats: store: %w", err)
	}
	return name, nil
}

// Remove deletes a stored image. A name Files could not have written,
// or one already gone, is not an error: there is nothing to remove.
func (f *Files) Remove(name string) error {
	if !f.Enabled() || !ValidFile(name) {
		return nil
	}
	err := os.Remove(filepath.Join(f.dir, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Serve writes the named image. The name is checked against the
// pattern Save writes, and the bytes are sniffed again on the way out,
// so a file placed in the directory by hand is never served as an
// image it is not. A name is a new random one per upload, so the
// response may be cached for good; nosniff, an explicit type and a
// sandboxing CSP keep the bytes inert if the URL is opened directly.
func (f *Files) Serve(w http.ResponseWriter, r *http.Request, name string) error {
	if !f.Enabled() {
		return ErrNoStore
	}
	if !ValidFile(name) {
		return ErrNotFound
	}
	path := filepath.Join(f.dir, name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ErrNotFound
	}
	file, err := os.Open(path) //nolint:gosec // name matched fileNamePattern
	if err != nil {
		return ErrNotFound
	}
	defer func() { _ = file.Close() }()
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	mime := sniff(head[:n])
	if mime == "" {
		return ErrNotFound
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return ErrNotFound
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Referrer-Policy", "no-referrer")
	// Private: it is behind a session. Immutable: the name changes
	// with every upload.
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	http.ServeContent(w, r, name, info.ModTime(), file)
	return nil
}
