package discord

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// AvatarCache fetches and serves Discord avatar images. Backing
// storage is $CMDCTRL_DATA_DIR/avatars/<discord_id>/<hash>.png so
// that (a) a deploy restart doesn't lose the cache, (b) multiple
// server processes sharing the same DataDir see each other's
// entries, and (c) the file layout is obvious when debugging —
// `ls /var/lib/cmd_and_ctrl/data/avatars/` shows the lineup.
//
// We cache by (discord_id, hash) rather than discord_id alone:
// Discord's avatar URL uses `<id>/<hash>.png` and the hash rolls
// every time the user changes their avatar. Keying on id alone
// would serve a stale portrait for hours after an update; the
// two-key shape matches how the client builds URLs from
// SeatInfo.DiscordAvatarHash.
//
// Concurrent misses for the same key are collapsed via inflight
// so we don't hammer Discord's CDN with N parallel fetches for
// the same avatar on the first request burst of a game.
type AvatarCache struct {
	Dir    string       // root directory; "" disables disk cache (memory-only fallback is not implemented yet — the handler 503s instead)
	Client *http.Client // HTTP client for cdn.discordapp.com fetches; nil → http.DefaultClient

	mu       sync.Mutex
	inflight map[string]chan struct{} // key = id + "/" + hash
}

// NewAvatarCache constructs a cache rooted at dir. Missing or
// uncreatable dir is not an error at construction time — the
// handler probes on each call, so a dir that appears later (e.g.
// systemd mounting a volume after boot) starts working without
// a restart.
func NewAvatarCache(dir string, client *http.Client) *AvatarCache {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &AvatarCache{
		Dir:      dir,
		Client:   client,
		inflight: map[string]chan struct{}{},
	}
}

// avatarKeyPattern constrains ids and hashes to the characters
// Discord actually uses, so a malicious client can't
// path-traverse via `..` or a null byte. Discord IDs are
// snowflakes (digits only) and avatar hashes are 32 hex chars
// (or `a_` + 32 hex for animated avatars); the ^[A-Za-z0-9_]+$
// regex is generous enough to accept both without letting `.` or
// `/` through.
var avatarKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// ErrInvalidAvatarKey is returned when the id or hash fails the
// character check. Handler surfaces as 400.
var ErrInvalidAvatarKey = errors.New("discord avatar: invalid id or hash")

// ErrCacheDisabled is returned when Dir == "". Handler surfaces
// as 503 so a deploy without a data dir surfaces clearly rather
// than transparently proxying every request.
var ErrCacheDisabled = errors.New("discord avatar: disk cache disabled")

// Serve writes the cached avatar PNG for (id, hash) to w. On a
// miss, fetches from Discord's CDN, stores under Dir, and serves
// the newly-stored file. Returns a non-nil error if the upstream
// 4xx/5xx's, the file can't be written, or the key fails
// validation — the HTTP handler maps these to appropriate
// statuses. Pass the inbound *http.Request through so ServeFile
// can honour conditional-GET / Range headers correctly.
func (c *AvatarCache) Serve(w http.ResponseWriter, r *http.Request, id, hash string) error {
	if c.Dir == "" {
		return ErrCacheDisabled
	}
	if !avatarKeyPattern.MatchString(id) || !avatarKeyPattern.MatchString(hash) {
		return ErrInvalidAvatarKey
	}

	path := filepath.Join(c.Dir, id, hash+".png")
	// Happy path: already on disk. os.Stat avoids opening the file
	// twice when http.ServeFile is going to re-open it anyway.
	if _, err := os.Stat(path); err == nil {
		c.writeCacheHeaders(w)
		http.ServeFile(w, r, path)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat cache: %w", err)
	}

	// Miss: fetch from cdn.discordapp.com. Collapse concurrent
	// misses for the same key so the first one does the fetch and
	// the rest wait on its completion.
	key := id + "/" + hash
	c.mu.Lock()
	wait, loading := c.inflight[key]
	if !loading {
		wait = make(chan struct{})
		c.inflight[key] = wait
	}
	c.mu.Unlock()
	if loading {
		<-wait
		// Someone else populated the cache (or failed). Re-check
		// on disk; on a hit we serve it, on a miss the cooperating
		// fetcher's error already surfaced to their request, so we
		// just 502 the waiter rather than hammer the CDN again.
		if _, err := os.Stat(path); err == nil {
			c.writeCacheHeaders(w)
			http.ServeFile(w, r, path)
			return nil
		}
		return fmt.Errorf("discord avatar: concurrent fetch failed")
	}

	// We own the fetch. Always close the inflight channel so
	// waiters wake even on error.
	defer func() {
		c.mu.Lock()
		delete(c.inflight, key)
		c.mu.Unlock()
		close(wait)
	}()

	if err := c.fetchAndStore(r.Context(), id, hash, path); err != nil {
		return err
	}
	c.writeCacheHeaders(w)
	http.ServeFile(w, r, path)
	return nil
}

// cdnURLHost is the Discord CDN origin. Package-level var so
// tests can redirect traffic to an httptest.Server without
// touching every call site. Prod code never mutates it.
var cdnURLHost = "https://cdn.discordapp.com"

// cdnURL builds the Discord CDN URL for (id, hash). Animated
// avatars (hash prefix "a_") would ideally be .gif; we ship .png
// unconditionally for the first pass — static frame is fine,
// animated support is a follow-up.
func cdnURL(id, hash string) string {
	return fmt.Sprintf("%s/avatars/%s/%s.png", cdnURLHost, id, hash)
}

// avatarMaxBytes caps how much we pull from the CDN. Discord
// avatar PNGs are typically under 128 KiB at the default size;
// 1 MiB leaves margin for high-res variants without letting a
// hostile CDN response fill the disk.
const avatarMaxBytes = 1 << 20

func (c *AvatarCache) fetchAndStore(ctx context.Context, id, hash, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cdnURL(id, hash), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "cmd_and_ctrl/0.1 (+https://github.com/krakenhavoc/cmd_and_ctrl)")
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("cdn fetch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("cdn: avatar not found for id=%s hash=%s", id, hash)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cdn: status %d", resp.StatusCode)
	}
	// Cheap pre-check when the CDN declares a length. The post-read
	// check below remains authoritative for chunked / lying responses.
	if resp.ContentLength > avatarMaxBytes {
		return fmt.Errorf("cdn: avatar is %d bytes, over the %d-byte cap", resp.ContentLength, avatarMaxBytes)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	// Write to a .tmp first and rename in place so a killed
	// process doesn't leave a half-written PNG that later reads
	// treat as cached.
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}
	// Read up to cap+1: landing past the cap proves the body is
	// oversized, and the response must be REJECTED, not silently
	// truncated — a truncated PNG cached as immutable would serve a
	// corrupt image for a day per client.
	n, err := io.Copy(f, io.LimitReader(resp.Body, avatarMaxBytes+1))
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write tmp: %w", err)
	}
	if n > avatarMaxBytes {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("cdn: avatar exceeds the %d-byte cap; not caching", avatarMaxBytes)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func (c *AvatarCache) writeCacheHeaders(w http.ResponseWriter) {
	// The URL contains the avatar hash, which changes every time
	// the user updates their Discord avatar, so the response is
	// safely immutable for the entire stable-URL lifetime. One
	// day is plenty for the client to keep it around.
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	w.Header().Set("Content-Type", "image/png")
}
