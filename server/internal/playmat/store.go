package playmat

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
)

// ErrNotFound means the id does not resolve to a stored playmat. It is
// the same answer for "never existed", "removed" and "not a uuid", so
// the route cannot be used to probe which ids are real.
var ErrNotFound = errors.New("playmat: not found")

// ErrDisabled means there is nowhere to store a playmat: no data
// directory, or no database to hold the pointer.
var ErrDisabled = errors.New("playmat: disabled")

// idPattern is the canonical lowercase v4 uuid shape. It is what stops
// "..", absolute paths and NUL bytes reaching filepath.Join, the same
// belt bugstore wears.
var idPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidID reports whether id has the shape this package mints.
func ValidID(id string) bool { return idPattern.MatchString(id) }

// FileStore keeps playmat images on disk, one file per image, rooted
// at $CMDCTRL_DATA_DIR/playmats:
//
//	<uuid>.jpg
//
// The uuid is a fresh v4 per image, so replacing a playmat changes the
// URL and the old one can be cached forever. The directory is created
// on first write, so a volume that mounts after boot works without a
// restart. dir == "" disables the store.
type FileStore struct{ dir string }

// NewFileStore returns a FileStore rooted at dir.
func NewFileStore(dir string) *FileStore { return &FileStore{dir: dir} }

// Enabled reports whether images can be stored.
func (s *FileStore) Enabled() bool { return s != nil && s.dir != "" }

func (s *FileStore) path(id string) string { return filepath.Join(s.dir, id+".jpg") }

// Save writes jpeg under a new id and returns it. The write goes to a
// temp name and is renamed, so a reader never sees half an image at a
// URL that has been handed out.
func (s *FileStore) Save(jpeg []byte) (string, error) {
	if !s.Enabled() {
		return "", ErrDisabled
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", fmt.Errorf("playmat: create dir: %w", err)
	}
	id := uuid.NewString()
	tmp, err := os.CreateTemp(s.dir, ".playmat-*.tmp")
	if err != nil {
		return "", fmt.Errorf("playmat: create temp: %w", err)
	}
	name := tmp.Name()
	_, err = tmp.Write(jpeg)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(name, 0o644)
	}
	if err == nil {
		err = os.Rename(name, s.path(id))
	}
	if err != nil {
		_ = os.Remove(name)
		return "", fmt.Errorf("playmat: write: %w", err)
	}
	return id, nil
}

// Open opens a stored image for reading.
func (s *FileStore) Open(id string) (*os.File, os.FileInfo, error) {
	if !s.Enabled() {
		return nil, nil, ErrDisabled
	}
	if !ValidID(id) {
		return nil, nil, ErrNotFound
	}
	f, err := os.Open(s.path(id)) //nolint:gosec // id matched the uuid pattern
	if err != nil {
		return nil, nil, ErrNotFound
	}
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		_ = f.Close()
		return nil, nil, ErrNotFound
	}
	return f, info, nil
}

// Exists reports whether a stored image is on disk, without opening it.
func (s *FileStore) Exists(id string) bool {
	if !s.Enabled() || !ValidID(id) {
		return false
	}
	info, err := os.Stat(s.path(id))
	return err == nil && !info.IsDir()
}

// Delete removes a stored image. A missing file is not an error: the
// caller is cleaning up and the goal state already holds.
func (s *FileStore) Delete(id string) error {
	if !s.Enabled() || !ValidID(id) {
		return nil
	}
	if err := os.Remove(s.path(id)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("playmat: delete: %w", err)
	}
	return nil
}
