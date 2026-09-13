// Package bugstore persists the artifacts that ride along with an
// in-app bug report: the reporter's screenshots and a pinned copy of
// the game's replay log at the moment the report was filed.
//
// Why a store at all, when ADR 0017 deliberately kept game state out
// of issues: GitHub's REST API has no attachment-upload endpoint, so
// an image can only render in an issue if it sits behind a URL the
// GitHub image proxy (Camo) can fetch. That forces one unauthenticated
// media route, which is the single largest new attack surface in the
// feature — hence the narrow posture here: raster magic bytes only,
// hard size caps, uuid-keyed paths that can't be walked, and response
// headers that make a stored file inert even if something non-image
// slipped through (see ADR 0017 §6).
//
// The replay pin is the opposite trade: valuable to an operator,
// unsafe to publish. It stays behind admin auth and is never linked
// from the issue body — only its report ID appears, so the operator
// can pull it deliberately.
//
// Layout, rooted at $CMDCTRL_DATA_DIR/bugreports:
//
//	<report-id>/att-1.png      reporter screenshot, publicly readable
//	<report-id>/replay.jsonl   pinned replay tail, admin-only
//	<report-id>/gamelog.txt    pinned public game log, admin-only
//	<report-id>/manifest.json  what this report holds, for pruning + forensics
//
// The directory name is a v4 uuid, so it doubles as the capability
// token in the public image URL: 122 bits of entropy is far past
// guessable, and the URL only ever appears inside a private-repo
// issue. Everything else about the report lives in the issue itself.
package bugstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxImages bounds how many screenshots one report may carry.
	// Four is "the board, the modal, the console, and one more" —
	// past that a reporter is dumping, not illustrating.
	MaxImages = 4

	// MaxImageBytes caps a single image. A full-screen PNG of the
	// board lands around 1–2 MiB; 4 MiB absorbs a retina screenshot
	// without becoming a bulk-upload primitive.
	MaxImageBytes = 4 << 20

	// MaxTotalImageBytes caps one report's images in aggregate, so
	// four maximum-size files can't add up to 16 MiB.
	MaxTotalImageBytes = 10 << 20

	// MaxGameLogPinBytes caps the pinned public game log. The log is
	// a bounded ring of a couple of hundred rendered lines, so this is
	// a "something is very wrong" ceiling rather than a policy — it
	// exists because the renderer's input is game state, and a runaway
	// there must not fill the disk.
	MaxGameLogPinBytes = 1 << 20

	// MaxReplayPinBytes caps the pinned replay copy. Replay lines are
	// whole unfiltered snapshots, so a long game's log runs to tens of
	// MiB; past this we keep the TAIL, because the bug is at the end.
	MaxReplayPinBytes = 32 << 20

	// DefaultRetention is how long a report's artifacts survive. Long
	// enough that a triaged issue still has its screenshots when
	// someone finally picks it up, short enough that the data dir
	// doesn't grow without bound on a box nobody watches.
	DefaultRetention = 90 * 24 * time.Hour

	// DefaultMaxStoreBytes caps the whole artifact directory,
	// independent of age.
	//
	// Retention alone is not a disk-space guarantee: the rate limiter
	// allows roughly one report per 30 s per IP, so a hijacked session
	// could write tens of GiB inside the 90-day window. This box also
	// shares its data dir with a ~630 MiB Scryfall dump and the replay
	// logs, and filling it takes the game server down — a far worse
	// outcome than losing the screenshots from an old bug. Oldest
	// reports go first when over.
	DefaultMaxStoreBytes = 512 << 20
)

// Errors surfaced to the HTTP layer. The handler maps these to
// statuses; none of them carry a filesystem path.
var (
	// ErrDisabled means the requested store feature is unavailable:
	// either no data dir (nothing to store in) or no public base URL
	// for image hosting. Text-only reports still work — the caller
	// degrades rather than failing.
	ErrDisabled = errors.New("bugstore: disabled")

	// ErrUnsupportedType means the bytes are not one of the raster
	// formats GitHub renders. Notably SVG is rejected: it is a script
	// carrier, and an unauthenticated route that serves attacker-
	// supplied SVG is a stored-XSS hole on our own origin.
	ErrUnsupportedType = errors.New("bugstore: unsupported image type (png, jpeg, gif, webp only)")

	// ErrTooLarge means a single image or the report's running total
	// exceeded the caps above.
	ErrTooLarge = errors.New("bugstore: image exceeds the size limit")

	// ErrTooMany means the report already holds MaxImages.
	ErrTooMany = errors.New("bugstore: too many images")

	// ErrNotFound means the report id or artifact name doesn't
	// resolve. Deliberately identical for "never existed", "pruned",
	// and "malformed id" so the route can't be used to probe which
	// report ids are real.
	ErrNotFound = errors.New("bugstore: not found")
)

// imageTypes maps the content types http.DetectContentType reports
// for the four formats GitHub renders inline onto the extension we
// store them under. The map is the allowlist — a type absent here is
// rejected regardless of what the client claimed in its part headers.
var imageTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// reportIDPattern constrains a path segment to the canonical v4 uuid
// shape. This is what stops `..`, absolute paths, and null bytes from
// reaching filepath.Join — the same belt the avatar cache wears.
var reportIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// imageNamePattern constrains the file segment of a public image URL.
// Only names this package itself mints can match.
var imageNamePattern = regexp.MustCompile(`^att-[1-9][0-9]?\.(png|jpg|gif|webp)$`)

// replayFileName is the fixed name of the pinned replay inside a
// report directory. Fixed rather than derived so the admin route
// needs no name parameter at all.
const replayFileName = "replay.jsonl"

// gameLogFileName is the fixed name of the pinned public game log
// (S31 sub-PR 0). Same posture as the replay: written by the server,
// read only by an admin, never linked from the issue body.
const gameLogFileName = "gamelog.txt"

const manifestFileName = "manifest.json"

// Store owns the bug-report artifact directory.
//
// dir == "" leaves the store disabled entirely: there is nowhere to
// keep report artifacts. baseURL == "" leaves only image hosting
// disabled; replay pinning still works because it is admin-only and
// not fetched by GitHub's image proxy.
type Store struct {
	dir     string
	baseURL string
}

// New returns a Store rooted at dir, minting image URLs against
// baseURL (e.g. "https://cmd.labxp.io"). The directory is created
// lazily on first report, so a data volume that mounts after boot
// starts working without a restart.
func New(dir, baseURL string) *Store {
	return &Store{dir: dir, baseURL: strings.TrimRight(baseURL, "/")}
}

// Enabled reports whether report artifacts can be stored on disk.
func (s *Store) Enabled() bool { return s != nil && s.dir != "" }

// AttachmentsEnabled reports whether screenshots can be hosted at
// public URLs for GitHub issue rendering.
func (s *Store) AttachmentsEnabled() bool { return s != nil && s.dir != "" && s.baseURL != "" }

// Image describes one stored screenshot.
type Image struct {
	Name  string `json:"name"`
	MIME  string `json:"mime"`
	Bytes int64  `json:"bytes"`
	// URL is absolute and Camo-reachable — markdown in an issue body
	// needs a fully-qualified URL, and a relative one would resolve
	// against github.com.
	URL string `json:"url"`
}

// ReplayPin describes the pinned replay copy.
type ReplayPin struct {
	Bytes int64 `json:"bytes"`
	Lines int   `json:"lines"`
	// Truncated is true when the source exceeded MaxReplayPinBytes and
	// only the tail was kept. The issue body says so, because a
	// truncated replay can't be replayed from turn one.
	Truncated bool `json:"truncated"`
}

// GameLogPin describes the pinned public game log.
type GameLogPin struct {
	Bytes int64 `json:"bytes"`
	Lines int   `json:"lines"`
}

// Manifest is the on-disk record of a report's artifacts. Written
// last, so a directory without one is an abandoned report that Prune
// can reclaim early.
type Manifest struct {
	ID        string      `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	GameID    string      `json:"game_id,omitempty"`
	IssueURL  string      `json:"issue_url,omitempty"`
	Images    []Image     `json:"images,omitempty"`
	Replay    *ReplayPin  `json:"replay,omitempty"`
	GameLog   *GameLogPin `json:"game_log,omitempty"`
}

// Report accumulates one report's artifacts. Not safe for concurrent
// use by multiple goroutines; a Report belongs to one request.
type Report struct {
	ID      string
	dir     string
	store   *Store
	images  []Image
	total   int64
	replay  *ReplayPin
	gameLog *GameLogPin
}

// NewReport mints a report id and creates its directory.
func (s *Store) NewReport() (*Report, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	id := uuid.NewString()
	dir := filepath.Join(s.dir, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("bugstore: create report dir: %w", err)
	}
	return &Report{ID: id, dir: dir, store: s}, nil
}

// Images returns the images stored so far.
func (r *Report) Images() []Image { return r.images }

// Replay returns the pinned-replay record, or nil if none was pinned.
func (r *Report) Replay() *ReplayPin { return r.replay }

// GameLog returns the pinned game-log record, or nil if none was
// pinned.
func (r *Report) GameLog() *GameLogPin { return r.gameLog }

// PinGameLog stores the rendered public game log for this report.
//
// Unlike the replay, the caller hands over bytes rather than a path:
// the log is not a file on disk anywhere, it is a projection of live
// game state that the lobby handler renders at report time.
//
// Admin-only to read, for consistency with the replay rather than
// because the contents are dangerous — the log is public information
// by construction (see protocol/log.go). What is NOT guaranteed is
// that it was filtered for the reporter: the handler pins the
// unfiltered projection, because an operator triaging "the game
// thought my creature was still there" needs to see what the server
// believed, not what one seat was shown.
func (r *Report) PinGameLog(text []byte) (*GameLogPin, error) {
	if r == nil || r.store == nil || !r.store.Enabled() {
		return nil, ErrDisabled
	}
	if len(text) == 0 {
		return nil, ErrNotFound
	}
	if len(text) > MaxGameLogPinBytes {
		return nil, ErrTooLarge
	}
	tmp, err := os.CreateTemp(r.dir, ".gamelog-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("bugstore: create game log pin: %w", err)
	}
	name := tmp.Name()
	_, err = tmp.Write(text)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(name)
		return nil, fmt.Errorf("bugstore: write game log pin: %w", err)
	}
	final := filepath.Join(r.dir, gameLogFileName)
	if err := os.Rename(name, final); err != nil {
		_ = os.Remove(name)
		return nil, fmt.Errorf("bugstore: commit game log pin: %w", err)
	}
	pin := &GameLogPin{Bytes: int64(len(text)), Lines: countLines(final)}
	r.gameLog = pin
	return pin, nil
}

// OpenGameLog opens a report's pinned game log for reading. Mirrors
// OpenReplay, including its deliberately identical ErrNotFound for
// "no such report" and "no such artifact".
func (s *Store) OpenGameLog(id string) (*os.File, os.FileInfo, error) {
	if !s.Enabled() {
		return nil, nil, ErrDisabled
	}
	if !reportIDPattern.MatchString(id) {
		return nil, nil, ErrNotFound
	}
	path := filepath.Join(s.dir, id, gameLogFileName)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return nil, nil, ErrNotFound
	}
	f, err := os.Open(path) //nolint:gosec // path is uuid + fixed name
	if err != nil {
		return nil, nil, ErrNotFound
	}
	return f, info, nil
}

// AddImage validates and stores one screenshot.
//
// The content type is sniffed from the bytes, never taken from the
// client's part header: a `Content-Type: image/png` on a payload of
// HTML is exactly the trick this route has to refuse, since the
// response is served from our own origin.
func (r *Report) AddImage(data []byte) (Image, error) {
	if r == nil || r.store == nil || !r.store.AttachmentsEnabled() {
		return Image{}, ErrDisabled
	}
	if len(r.images) >= MaxImages {
		return Image{}, ErrTooMany
	}
	if int64(len(data)) > MaxImageBytes || r.total+int64(len(data)) > MaxTotalImageBytes {
		return Image{}, ErrTooLarge
	}
	// DetectContentType reads at most the first 512 bytes and always
	// returns a type; anything not in the allowlist is refused.
	mime := http.DetectContentType(data)
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	ext, ok := imageTypes[mime]
	if !ok {
		return Image{}, ErrUnsupportedType
	}
	name := fmt.Sprintf("att-%d.%s", len(r.images)+1, ext)
	// Write to a temp name and rename, so a reader can never observe
	// a half-written image at a URL we have already published.
	tmp := filepath.Join(r.dir, "."+name+".tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return Image{}, fmt.Errorf("bugstore: write image: %w", err)
	}
	if err := os.Rename(tmp, filepath.Join(r.dir, name)); err != nil {
		_ = os.Remove(tmp)
		return Image{}, fmt.Errorf("bugstore: commit image: %w", err)
	}
	img := Image{
		Name:  name,
		MIME:  mime,
		Bytes: int64(len(data)),
		URL:   fmt.Sprintf("%s/bugreport/att/%s/%s", r.store.baseURL, r.ID, name),
	}
	r.images = append(r.images, img)
	r.total += img.Bytes
	return img, nil
}

// PinReplay copies the game's replay JSONL into the report directory
// so the record survives what happens next: the replay file keeps
// being appended to while the game runs, and game eviction deletes it
// outright. Without a pin, the issue's "pull the replay" instruction
// goes stale the moment the table packs up — which is most of why
// ADR 0017's replay reference never actually got used.
//
// Over MaxReplayPinBytes we keep the tail and mark it truncated: the
// interesting end of a long game beats a complete but unusable head.
func (r *Report) PinReplay(srcPath string) (*ReplayPin, error) {
	if srcPath == "" {
		return nil, ErrNotFound
	}
	src, err := os.Open(srcPath) //nolint:gosec // path comes from ws.Room, not the client
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("bugstore: open replay: %w", err)
	}
	defer func() { _ = src.Close() }()

	info, err := src.Stat()
	if err != nil {
		return nil, fmt.Errorf("bugstore: stat replay: %w", err)
	}
	pin := &ReplayPin{}
	if info.Size() > MaxReplayPinBytes {
		if _, err := src.Seek(info.Size()-MaxReplayPinBytes, io.SeekStart); err != nil {
			return nil, fmt.Errorf("bugstore: seek replay tail: %w", err)
		}
		pin.Truncated = true
	}

	dst, err := os.CreateTemp(r.dir, ".replay-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("bugstore: create replay pin: %w", err)
	}
	tmp := dst.Name()
	n, err := io.Copy(dst, src)
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(tmp)
		return nil, fmt.Errorf("bugstore: write replay pin: %w", err)
	}
	if err := os.Rename(tmp, filepath.Join(r.dir, replayFileName)); err != nil {
		_ = os.Remove(tmp)
		return nil, fmt.Errorf("bugstore: commit replay pin: %w", err)
	}
	pin.Bytes = n
	pin.Lines = countLines(filepath.Join(r.dir, replayFileName))
	r.replay = pin
	return pin, nil
}

// Commit writes the manifest, marking the report complete.
func (r *Report) Commit(gameID, issueURL string) error {
	m := Manifest{
		ID:        r.ID,
		CreatedAt: time.Now().UTC(),
		GameID:    gameID,
		IssueURL:  issueURL,
		Images:    r.images,
		Replay:    r.replay,
		GameLog:   r.gameLog,
	}
	buf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("bugstore: encode manifest: %w", err)
	}
	return os.WriteFile(filepath.Join(r.dir, manifestFileName), buf, 0o644)
}

// Discard removes the report directory. Called when filing the issue
// fails: without it, every GitHub outage would leave orphaned images
// on disk reachable at URLs nothing references.
func (r *Report) Discard() {
	if r == nil || r.dir == "" {
		return
	}
	_ = os.RemoveAll(r.dir)
}

// ServeImage writes a stored screenshot to w.
//
// This is the one unauthenticated route in the feature — Camo cannot
// present a session — so the headers do the work that auth normally
// would. nosniff plus an explicit type from our own allowlist stops
// the browser from re-interpreting the bytes; the CSP and sandbox
// make the response inert if it somehow gets navigated to directly
// rather than loaded as an <img>.
func (s *Store) ServeImage(w http.ResponseWriter, r *http.Request, id, name string) error {
	if !s.AttachmentsEnabled() {
		return ErrDisabled
	}
	if !reportIDPattern.MatchString(id) || !imageNamePattern.MatchString(name) {
		return ErrNotFound
	}
	path := filepath.Join(s.dir, id, name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ErrNotFound
	}
	// Re-sniff on the way out rather than trusting the extension we
	// wrote: cheap, and it means a hand-placed file in the data dir
	// can't be served as an image it isn't.
	f, err := os.Open(path) //nolint:gosec // path is uuid + allowlisted name
	if err != nil {
		return ErrNotFound
	}
	defer func() { _ = f.Close() }()
	head := make([]byte, 512)
	n, _ := io.ReadFull(f, head)
	mime := http.DetectContentType(head[:n])
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if _, ok := imageTypes[mime]; !ok {
		return ErrNotFound
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return ErrNotFound
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "inline; filename=\""+name+"\"")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Referrer-Policy", "no-referrer")
	// The URL is capability-keyed and the bytes never change, so this
	// is safely immutable. Camo caches independently; this is for the
	// operator's browser.
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	http.ServeContent(w, r, name, info.ModTime(), f)
	return nil
}

// OpenReplay opens a report's pinned replay for the admin route. The
// caller closes the file.
func (s *Store) OpenReplay(id string) (*os.File, os.FileInfo, error) {
	if !s.Enabled() {
		return nil, nil, ErrDisabled
	}
	if !reportIDPattern.MatchString(id) {
		return nil, nil, ErrNotFound
	}
	path := filepath.Join(s.dir, id, replayFileName)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return nil, nil, ErrNotFound
	}
	f, err := os.Open(path) //nolint:gosec // path is uuid + fixed name
	if err != nil {
		return nil, nil, ErrNotFound
	}
	return f, info, nil
}

// Prune enforces both retention limits: it deletes report directories
// older than maxAge, then deletes the oldest surviving ones until the
// directory fits under maxBytes. Called at boot and after each filed
// report — there are only ever a handful of directories at this scale,
// so a full scan is cheaper than tracking state. Returns how many it
// removed.
//
// Age comes from the directory mtime rather than the manifest, so a
// report abandoned mid-write (no manifest at all) is reclaimed on the
// same schedule as a complete one.
//
// A maxBytes of 0 means "age only" — useful for asking "is anything
// here?" without a size policy.
func (s *Store) Prune(maxAge time.Duration, maxBytes int64) (int, error) {
	if !s.Enabled() {
		return 0, ErrDisabled
	}
	reports, err := s.scan()
	if err != nil {
		return 0, err
	}
	cutoff := time.Now().Add(-maxAge)
	removed := 0
	var live []reportDir
	var total int64
	for _, r := range reports {
		if r.modTime.Before(cutoff) {
			if err := os.RemoveAll(filepath.Join(s.dir, r.name)); err == nil {
				removed++
				continue
			}
		}
		live = append(live, r)
		total += r.bytes
	}
	// Oldest-first eviction down to the ceiling. The newest reports are
	// the ones with an open issue somebody might still be reading.
	for i := 0; maxBytes > 0 && total > maxBytes && i < len(live); i++ {
		if err := os.RemoveAll(filepath.Join(s.dir, live[i].name)); err != nil {
			continue
		}
		total -= live[i].bytes
		removed++
	}
	return removed, nil
}

// reportDir is one scanned report directory: its name, mtime, and
// total size on disk.
type reportDir struct {
	name    string
	modTime time.Time
	bytes   int64
}

// scan lists report directories oldest-first. Non-report entries are
// skipped entirely — the artifact directory lives inside the shared
// data dir and must never delete a neighbour.
func (s *Store) scan() ([]reportDir, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("bugstore: read dir: %w", err)
	}
	var out []reportDir
	for _, e := range entries {
		if !e.IsDir() || !reportIDPattern.MatchString(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, reportDir{name: e.Name(), modTime: info.ModTime(), bytes: dirBytes(filepath.Join(s.dir, e.Name()))})
	}
	// Oldest first, name as a tiebreaker so tests are deterministic.
	sort.Slice(out, func(i, j int) bool {
		if out[i].modTime.Equal(out[j].modTime) {
			return out[i].name < out[j].name
		}
		return out[i].modTime.Before(out[j].modTime)
	})
	return out, nil
}

// dirBytes sums the regular files directly inside a report directory.
// Not recursive: this package only ever writes flat directories, and a
// nested one would be somebody else's doing.
func dirBytes(dir string) int64 {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	var total int64
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || e.IsDir() {
			continue
		}
		total += info.Size()
	}
	return total
}

// countLines counts newline-terminated records in a file. Used only
// to put a "N lines" figure in the issue body, so a read error is
// reported as zero rather than failing the whole report.
func countLines(path string) int {
	f, err := os.Open(path) //nolint:gosec // path is internal
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, 64<<10)
	lines := 0
	for {
		n, err := f.Read(buf)
		for _, b := range buf[:n] {
			if b == '\n' {
				lines++
			}
		}
		if err != nil {
			return lines
		}
	}
}
