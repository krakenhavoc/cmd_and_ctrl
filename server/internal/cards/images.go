package cards

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ValidSizes is the set of Scryfall image sizes we accept. Anything
// outside this list is rejected before it reaches the filesystem — see
// ErrInvalidSize and the path-traversal note on pathForFace.
var ValidSizes = map[string]struct{}{
	"small":       {},
	"normal":      {},
	"large":       {},
	"png":         {},
	"art_crop":    {},
	"border_crop": {},
}

// ErrInvalidSize is returned when the requested size is not one of the
// Scryfall-blessed variants. The HTTP handler maps this to 400.
var ErrInvalidSize = errors.New("cards: invalid image size")

// MaxFaceIndex bounds the ?face= parameter. No printing in the
// Scryfall dump has more than two faces; the cap is generous and its
// job is not accuracy but keeping a caller-supplied integer from
// growing into an unbounded filename component. See pathForFace.
const MaxFaceIndex = 7

// ErrInvalidFace is returned when ?face= is negative or beyond
// MaxFaceIndex. The HTTP handler maps this to 400. A face index that
// is in range but that this particular card does not have resolves
// to no image URI at all and comes back as ErrNoImage / 404 —
// "that's not a face" and "that card has no such face" are different
// answers and are reported differently.
var ErrInvalidFace = errors.New("cards: invalid image face")

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
// small, normal, large, png, art_crop, border_crop. Any other size
// returns ErrInvalidSize — this guard is what keeps a caller-supplied
// size from becoming a path-traversal vector through pathForFace.
func (c *ImageCache) Fetch(ctx context.Context, idx *Index, id uuid.UUID, size string) (string, error) {
	return c.FetchFace(ctx, idx, id, size, 0)
}

// FetchFace is Fetch for one printed face of a multi-face card (ADR
// 0034). Face 0 is the front and produces exactly the paths Fetch
// always produced, so no cached image is orphaned by this change.
//
// The face index is part of the ON-DISK KEY, not just the URL:
// without that, the first request for either face would poison the
// cache for the other, and they are different pictures. It is bounds
// checked before it reaches pathForFace for the same reason `size`
// is — a caller-supplied component of a filesystem path is a
// traversal vector until it is proven to be a small non-negative
// integer.
func (c *ImageCache) FetchFace(ctx context.Context, idx *Index, id uuid.UUID, size string, face int) (string, error) {
	if size == "" {
		size = c.DefaultSize
	}
	if _, ok := ValidSizes[size]; !ok {
		return "", ErrInvalidSize
	}
	if face < 0 || face > MaxFaceIndex {
		return "", ErrInvalidFace
	}
	path := c.pathForFace(id, size, face)
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
	uri := ImageURIForFace(card, size, face)
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

// pathForFace is pathFor with the printed face folded into the key.
//
// Face 0 keeps the ORIGINAL filename, unsuffixed. That is not
// cosmetic: every image already on disk in every deployment was
// written under that name, and a scheme that suffixed the front face
// too would silently orphan the entire cache and re-download it.
//
// `face` must already have been bounds-checked by the caller
// (FetchFace does it) — it becomes part of a filesystem path.
func (c *ImageCache) pathForFace(id uuid.UUID, size string, face int) string {
	s := id.String()
	name := s + "." + size + ".jpg"
	if face > 0 {
		name = s + ".face" + strconv.Itoa(face) + "." + size + ".jpg"
	}
	return filepath.Join(c.root, s[:2], name)
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
	if err := validateImageURI(uri); err != nil {
		return err
	}
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
	defer func() { _ = resp.Body.Close() }()
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

// validateImageURI enforces that a URI we are about to fetch is one we
// trust. The Scryfall index is loaded from disk at startup; if that
// file were ever swapped or tampered with, a malicious image_uris
// entry could otherwise direct the server to fetch arbitrary internal
// URLs (SSRF). Production traffic must be https to a Scryfall host.
// Loopback is permitted as an escape hatch for tests (httptest) and
// local development.
func validateImageURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("cards: invalid image uri: %w", err)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("cards: image uri missing host: %s", raw)
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return nil
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	if u.Scheme != "https" {
		return fmt.Errorf("cards: image uri must be https: %s", raw)
	}
	if !isScryfallHost(host) {
		return fmt.Errorf("cards: image uri host not allowed: %s", host)
	}
	return nil
}

// isScryfallHost matches Scryfall's CDN hostnames. They currently
// serve images from c*.scryfall.com and cards.scryfall.io; we accept
// any subdomain of either apex to stay robust against CDN shuffles.
func isScryfallHost(host string) bool {
	host = strings.ToLower(host)
	return host == "scryfall.com" ||
		host == "scryfall.io" ||
		strings.HasSuffix(host, ".scryfall.com") ||
		strings.HasSuffix(host, ".scryfall.io")
}
