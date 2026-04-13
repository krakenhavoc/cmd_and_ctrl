package cards

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ImageCache is a disk-backed on-demand image fetcher. First request
// for a given card pulls the image from Scryfall's CDN and writes it
// into $root/<aa>/<id>.jpg (sharded by the first two hex characters
// of the UUID to keep any single directory under ~4000 files).
// Subsequent requests serve the file straight from disk.
//
// Concurrency: an in-flight request for a given card holds a
// per-card mutex so two HTTP handlers don't race to download the
// same asset. The per-card mutex is itself held inside a sync.Map so
// parallel fetches of different cards don't serialise.
//
// Safe for concurrent use.
type ImageCache struct {
	root        string
	http        *http.Client
	DefaultSize string
	inflight    sync.Map // uuid.UUID → *sync.Mutex
}

// NewImageCache constructs a cache rooted at dir. Creates dir if it
// doesn't exist.
func NewImageCache(dir string) (*ImageCache, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &ImageCache{
		root:        dir,
		http:        &http.Client{Timeout: 10 * time.Second},
		DefaultSize: "normal",
	}, nil
}

// ErrNoImage is returned when the requested card has no image URL in
// the Scryfall index (rare; some tokens and emblems lack one).
var ErrNoImage = errors.New("cards: no image uri for card")

// Fetch returns the on-disk path of the image for id, downloading
// it from the Scryfall CDN if not already cached. Always returns an
// absolute path to a file on disk; callers can http.ServeFile from
// it. ctx cancels the download.
//
// If size is empty, Cache.DefaultSize is used. Valid Scryfall sizes:
// small, normal, large, png, art_crop, border_crop.
func (c *ImageCache) Fetch(ctx context.Context, idx *Index, id uuid.UUID, size string) (string, error) {
	if size == "" {
		size = c.DefaultSize
	}
	path := c.pathFor(id, size)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	// Look up the card's image URI before taking the per-id lock so
	// that a missing card fails fast without blocking other fetches.
	card, ok := idx.Get(id)
	if !ok {
		return "", fmt.Errorf("cards: unknown card %s", id)
	}
	uri := ImageURI(card, size)
	if uri == "" {
		return "", ErrNoImage
	}

	mu := c.lockFor(id)
	mu.Lock()
	defer mu.Unlock()

	// Re-check after acquiring the lock: another fetch for the same
	// id may have landed while we were waiting.
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if err := c.download(ctx, uri, path); err != nil {
		return "", err
	}
	return path, nil
}

// pathFor returns the disk path used for id + size. Sharded by the
// first two hex characters so ls-ing the root directory remains
// manageable.
func (c *ImageCache) pathFor(id uuid.UUID, size string) string {
	s := id.String()
	return filepath.Join(c.root, s[:2], s+"."+size+".jpg")
}

// lockFor returns the per-id mutex, creating it on first use. Uses
// sync.Map so concurrent creates for different ids don't serialise
// on a single shard-level mutex.
func (c *ImageCache) lockFor(id uuid.UUID) *sync.Mutex {
	if m, ok := c.inflight.Load(id); ok {
		return m.(*sync.Mutex)
	}
	fresh := &sync.Mutex{}
	actual, _ := c.inflight.LoadOrStore(id, fresh)
	return actual.(*sync.Mutex)
}

// download fetches uri and writes it to path atomically via a
// sibling .tmp file. Caller holds the per-id mutex so no other
// goroutine competes for the same tmp path.
func (c *ImageCache) download(ctx context.Context, uri, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return err
	}
	// Scryfall asks API users to identify themselves; be polite.
	req.Header.Set("User-Agent", "cmd_and_ctrl/0.1 (+https://github.com/krakenhavoc/cmd_and_ctrl)")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cards: scryfall returned HTTP %d for %s", resp.StatusCode, uri)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		if tmp != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	tmp = nil
	return os.Rename(tmpPath, path)
}
