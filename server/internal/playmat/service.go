package playmat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"os"
	"sync"

	"github.com/google/uuid"
)

// URLPrefix is where a stored playmat is served. Same origin as the
// page, so the session cookie rides the request.
const URLPrefix = "/playmats/"

// ErrBusy means too many images are being processed at once.
var ErrBusy = errors.New("the server is busy with other playmats; try again in a moment")

// ErrNoUser means the account row does not exist.
var ErrNoUser = errors.New("playmat: no such user")

// maxConcurrent bounds Normalize and Fetch together. A 40 MP image
// decodes to 160 MB; four people uploading at once on a small VPS is a
// restart, not a feature.
const maxConcurrent = 2

// Info describes a stored playmat.
type Info struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Service is the playmat feature: the file store plus users.playmat_id.
// Safe for concurrent use. A nil *Service is a disabled feature.
type Service struct {
	files   *FileStore
	db      *sql.DB
	fetcher *Fetcher
	sem     chan struct{}

	mu     sync.RWMutex
	cache  map[uuid.UUID]string // user -> playmat id, "" for none
	washes map[uuid.UUID]int    // user -> wash, read through like cache
}

// NewService returns a Service. d may be nil (no database), which
// disables it, as does a FileStore with no directory. f may be nil for
// the production Fetcher.
func NewService(files *FileStore, d *sql.DB, f *Fetcher) *Service {
	if f == nil {
		f = NewFetcher()
	}
	return &Service{
		files: files, db: d, fetcher: f,
		sem:    make(chan struct{}, maxConcurrent),
		cache:  map[uuid.UUID]string{},
		washes: map[uuid.UUID]int{},
	}
}

// Enabled reports whether playmats can be stored: a data directory and
// a database.
func (s *Service) Enabled() bool { return s != nil && s.db != nil && s.files.Enabled() }

func urlOf(id string) string { return URLPrefix + id }

func (s *Service) acquire(ctx context.Context) error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrBusy
	}
}

func (s *Service) release() { <-s.sem }

// id reads the user's playmat id, through the cache.
func (s *Service) id(ctx context.Context, user uuid.UUID) (string, error) {
	s.mu.RLock()
	id, ok := s.cache[user]
	s.mu.RUnlock()
	if ok {
		return id, nil
	}
	var v sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT playmat_id FROM users WHERE id = ?`, user.String()).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("playmat: read: %w", err)
	}
	s.mu.Lock()
	s.cache[user] = v.String
	s.mu.Unlock()
	return v.String, nil
}

// URL returns the same-origin URL of user's playmat, or "" for none.
// It is what the table stamps on a seat, so it never fails loudly: a
// database error reads as "no playmat" and is retried next time.
func (s *Service) URL(user uuid.UUID) string {
	if !s.Enabled() || user == uuid.Nil {
		return ""
	}
	id, err := s.id(context.Background(), user)
	if err != nil || id == "" {
		return ""
	}
	return urlOf(id)
}

// Get returns user's playmat, with its dimensions read from the stored
// file's header. ok is false when the user has none.
func (s *Service) Get(ctx context.Context, user uuid.UUID) (info Info, ok bool, err error) {
	if !s.Enabled() {
		return Info{}, false, ErrDisabled
	}
	id, err := s.id(ctx, user)
	if err != nil || id == "" {
		return Info{}, false, err
	}
	info = Info{ID: id, URL: urlOf(id)}
	f, _, oerr := s.files.Open(id)
	if oerr != nil {
		// The pointer outlived the file (a restored database, a wiped
		// volume). Report no playmat rather than a dead link.
		return Info{}, false, nil
	}
	defer func() { _ = f.Close() }()
	if cfg, _, derr := image.DecodeConfig(f); derr == nil {
		info.Width, info.Height = cfg.Width, cfg.Height
	}
	return info, true, nil
}

// SetFromBytes validates data, stores it as user's playmat and deletes
// the one it replaces.
func (s *Service) SetFromBytes(ctx context.Context, user uuid.UUID, data []byte) (Info, error) {
	if !s.Enabled() {
		return Info{}, ErrDisabled
	}
	if err := s.acquire(ctx); err != nil {
		return Info{}, err
	}
	defer s.release()
	return s.store(ctx, user, data)
}

// SetFromURL fetches rawURL once, then stores it exactly as an upload.
func (s *Service) SetFromURL(ctx context.Context, user uuid.UUID, rawURL string) (Info, error) {
	if !s.Enabled() {
		return Info{}, ErrDisabled
	}
	if _, err := ParseURL(rawURL); err != nil {
		return Info{}, err
	}
	if err := s.acquire(ctx); err != nil {
		return Info{}, err
	}
	defer s.release()
	data, err := s.fetcher.Fetch(ctx, rawURL)
	if err != nil {
		return Info{}, err
	}
	return s.store(ctx, user, data)
}

func (s *Service) store(ctx context.Context, user uuid.UUID, data []byte) (Info, error) {
	n, err := Normalize(data)
	if err != nil {
		return Info{}, err
	}
	id, err := s.files.Save(n.JPEG)
	if err != nil {
		return Info{}, err
	}
	old, err := s.swap(ctx, user, id)
	if err != nil {
		_ = s.files.Delete(id)
		return Info{}, err
	}
	if old != "" {
		_ = s.files.Delete(old)
	}
	return Info{ID: id, URL: urlOf(id), Width: n.Width, Height: n.Height}, nil
}

// Remove deletes user's playmat. Removing none is not an error.
func (s *Service) Remove(ctx context.Context, user uuid.UUID) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	old, err := s.swap(ctx, user, "")
	if err != nil {
		return err
	}
	if old != "" {
		return s.files.Delete(old)
	}
	return nil
}

// swap sets users.playmat_id (NULL for "") and returns the previous
// value, in one transaction so two replacements cannot both believe
// they replaced nothing and leak a file.
func (s *Service) swap(ctx context.Context, user uuid.UUID, id string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("playmat: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var old sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT playmat_id FROM users WHERE id = ?`, user.String()).Scan(&old)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNoUser
	}
	if err != nil {
		return "", fmt.Errorf("playmat: read: %w", err)
	}
	var arg any
	if id != "" {
		arg = id
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET playmat_id = ? WHERE id = ?`, arg, user.String()); err != nil {
		return "", fmt.Errorf("playmat: write: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("playmat: commit: %w", err)
	}
	s.mu.Lock()
	s.cache[user] = id
	s.mu.Unlock()
	return old.String, nil
}

// The owner-set wash (ADR 0128 amendment): how strongly the playmat is
// darkened under the cards, in percent. Everyone at the table sees the
// owner's choice. NULL in users.playmat_wash reads as DefaultWash, the
// scrim the feature shipped with.
const (
	MinWash     = 30
	MaxWash     = 90
	DefaultWash = 58
)

// ErrBadWash means a wash outside MinWash..MaxWash.
var ErrBadWash = fmt.Errorf("wash must be between %d and %d", MinWash, MaxWash)

// Wash returns user's wash, DefaultWash when they have not set one. It
// is what the table stamps beside the URL, so like URL it never fails
// loudly: an error reads as the default.
func (s *Service) Wash(user uuid.UUID) int {
	if !s.Enabled() || user == uuid.Nil {
		return DefaultWash
	}
	w, err := s.wash(context.Background(), user)
	if err != nil {
		return DefaultWash
	}
	return w
}

func (s *Service) wash(ctx context.Context, user uuid.UUID) (int, error) {
	s.mu.RLock()
	w, ok := s.washes[user]
	s.mu.RUnlock()
	if ok {
		return w, nil
	}
	var v sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT playmat_wash FROM users WHERE id = ?`, user.String()).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultWash, nil
	}
	if err != nil {
		return DefaultWash, fmt.Errorf("playmat: read wash: %w", err)
	}
	w = DefaultWash
	if v.Valid && v.Int64 >= MinWash && v.Int64 <= MaxWash {
		w = int(v.Int64)
	}
	s.mu.Lock()
	s.washes[user] = w
	s.mu.Unlock()
	return w, nil
}

// SetWash stores user's wash. It may be set with no image yet: it is a
// preference, kept for the next upload.
func (s *Service) SetWash(ctx context.Context, user uuid.UUID, wash int) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	if wash < MinWash || wash > MaxWash {
		return ErrBadWash
	}
	res, err := s.db.ExecContext(ctx, `UPDATE users SET playmat_wash = ? WHERE id = ?`, wash, user.String())
	if err != nil {
		return fmt.Errorf("playmat: write wash: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNoUser
	}
	s.mu.Lock()
	s.washes[user] = wash
	s.mu.Unlock()
	return nil
}

// Open opens a stored image for the serving route.
func (s *Service) Open(id string) (*os.File, os.FileInfo, error) {
	if s == nil {
		return nil, nil, ErrDisabled
	}
	return s.files.Open(id)
}
