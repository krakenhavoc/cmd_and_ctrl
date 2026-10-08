package playmat

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// URLPrefix is where a stored playmat is served. Same origin as the
// page, so the session cookie rides the request.
const URLPrefix = "/playmats/"

// ErrBusy means too many images are being processed at once.
var ErrBusy = errors.New("the server is busy with other playmats; try again in a moment")

// ErrNoUser means the account row does not exist.
var ErrNoUser = errors.New("playmat: no such user")

// ErrBadSlot means a slot outside 1 to MaxSlots. A 400.
var ErrBadSlot = fmt.Errorf("a playmat slot is a number from 1 to %d", MaxSlots)

// ErrNoSlot means the slot holds no playmat, for an action that needs
// one (use it, fit it). A 404.
var ErrNoSlot = errors.New("there is no playmat in that slot")

// ErrConflict means the slot changed while a fit was being made: the
// crop was for an image that is no longer there. A 409; ask again.
var ErrConflict = errors.New("that playmat changed while it was being fitted; try again")

// maxConcurrent bounds Normalize, Fetch and Fit together. A 40 MP image
// decodes to 160 MB; four people uploading at once on a small VPS is a
// restart, not a feature.
const maxConcurrent = 2

// ValidSlot reports whether n names one of an account's slots.
func ValidSlot(n int) bool { return n >= 1 && n <= MaxSlots }

// Slot is one saved playmat.
type Slot struct {
	Slot   int
	ID     string
	URL    string
	Width  int
	Height int
	// Fits is Fits(Width, Height). Suggestion is what a fit would do,
	// present only when the shape is off (see Suggest).
	Fits       bool
	Suggestion *Suggestion
}

// State is an account's saved playmats and which one the table shows.
type State struct {
	// Slots holds the occupied slots, in slot order.
	Slots []Slot
	// Active is the slot the table shows, or 0 for none. Saved mats and
	// no active one is a legal state.
	Active int
}

// Find returns the slot n, if it is occupied.
func (st State) Find(n int) (Slot, bool) {
	for _, s := range st.Slots {
		if s.Slot == n {
			return s, true
		}
	}
	return Slot{}, false
}

// Service is the playmat feature: the file store plus user_playmats and
// users.playmat_id (the active pointer). Safe for concurrent use. A nil
// *Service is a disabled feature.
type Service struct {
	files   *FileStore
	db      *sql.DB
	fetcher *Fetcher
	sem     chan struct{}

	mu     sync.RWMutex
	cache  map[uuid.UUID]string // user -> ACTIVE playmat id, "" for none
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

// id reads the user's ACTIVE playmat id, through the cache.
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

// URL returns the same-origin URL of user's ACTIVE playmat, the one the
// table shows, or "" for none. It is what the table stamps on a seat,
// so it never fails loudly: a database error reads as "no playmat" and
// is retried next time.
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

// State returns user's saved playmats, with each one's dimensions and
// its fit. A user that does not exist has none. A saved row whose file
// is gone (a restored database, a wiped volume) is left out rather than
// listed as a dead link; replacing or removing its slot still works.
func (s *Service) State(ctx context.Context, user uuid.UUID) (State, error) {
	if !s.Enabled() {
		return State{}, ErrDisabled
	}
	// One statement, so the active pointer and the rows agree.
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.playmat_id, p.slot, p.playmat_id, p.width, p.height
		FROM users u LEFT JOIN user_playmats p ON p.user_id = u.id
		WHERE u.id = ? ORDER BY p.slot`, user.String())
	if err != nil {
		return State{}, fmt.Errorf("playmat: list: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var st State
	var active string
	for rows.Next() {
		var (
			act, id sql.NullString
			slot    sql.NullInt64
			wd, ht  sql.NullInt64
		)
		if err := rows.Scan(&act, &slot, &id, &wd, &ht); err != nil {
			return State{}, fmt.Errorf("playmat: list: %w", err)
		}
		active = act.String
		if !slot.Valid || !id.Valid {
			continue // the user has no saved playmats
		}
		w, h := int(wd.Int64), int(ht.Int64)
		if !wd.Valid || !ht.Valid || w <= 0 || h <= 0 {
			var ok bool
			if w, h, ok = s.headerSize(id.String); !ok {
				continue
			}
		} else if !s.files.Exists(id.String) {
			continue
		}
		sl := Slot{Slot: int(slot.Int64), ID: id.String, URL: urlOf(id.String), Width: w, Height: h, Fits: Fits(w, h)}
		if sg, ok := Suggest(w, h); ok {
			sl.Suggestion = &sg
		}
		st.Slots = append(st.Slots, sl)
		if id.String == active {
			st.Active = sl.Slot
		}
	}
	if err := rows.Err(); err != nil {
		return State{}, fmt.Errorf("playmat: list: %w", err)
	}
	return st, nil
}

// headerSize reads a stored image's dimensions from its header. ok is
// false when the file is missing or is not an image.
func (s *Service) headerSize(id string) (w, h int, ok bool) {
	f, _, err := s.files.Open(id)
	if err != nil {
		return 0, 0, false
	}
	defer func() { _ = f.Close() }()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

// SetFromBytes validates data, stores it in user's slot and deletes the
// file it replaces. The first mat saved by someone showing none becomes
// the active one; any other write leaves the active mat alone, unless
// it replaced the active slot, which then shows the new image.
func (s *Service) SetFromBytes(ctx context.Context, user uuid.UUID, slot int, data []byte) (Slot, error) {
	if !s.Enabled() {
		return Slot{}, ErrDisabled
	}
	if !ValidSlot(slot) {
		return Slot{}, ErrBadSlot
	}
	if err := s.acquire(ctx); err != nil {
		return Slot{}, err
	}
	defer s.release()
	return s.store(ctx, user, slot, data)
}

// SetFromURL fetches rawURL once, then stores it exactly as an upload.
func (s *Service) SetFromURL(ctx context.Context, user uuid.UUID, slot int, rawURL string) (Slot, error) {
	if !s.Enabled() {
		return Slot{}, ErrDisabled
	}
	if !ValidSlot(slot) {
		return Slot{}, ErrBadSlot
	}
	if _, err := ParseURL(rawURL); err != nil {
		return Slot{}, err
	}
	if err := s.acquire(ctx); err != nil {
		return Slot{}, err
	}
	defer s.release()
	data, err := s.fetcher.Fetch(ctx, rawURL)
	if err != nil {
		return Slot{}, err
	}
	return s.store(ctx, user, slot, data)
}

func (s *Service) store(ctx context.Context, user uuid.UUID, slot int, data []byte) (Slot, error) {
	n, err := Normalize(data)
	if err != nil {
		return Slot{}, err
	}
	return s.put(ctx, user, slot, n.JPEG, n.Width, n.Height, "")
}

// put saves jpeg as a new file and points the slot at it, deleting the
// file it replaces. expectOld, when set, is the id the caller made the
// new image from: the slot must still hold it, or the new file is
// dropped and ErrConflict returned.
func (s *Service) put(ctx context.Context, user uuid.UUID, slot int, jpg []byte, w, h int, expectOld string) (Slot, error) {
	id, err := s.files.Save(jpg)
	if err != nil {
		return Slot{}, err
	}
	old, err := s.putRow(ctx, user, slot, id, w, h, expectOld)
	if err != nil {
		_ = s.files.Delete(id)
		return Slot{}, err
	}
	if old != "" {
		_ = s.files.Delete(old)
	}
	out := Slot{Slot: slot, ID: id, URL: urlOf(id), Width: w, Height: h, Fits: Fits(w, h)}
	if sg, ok := Suggest(w, h); ok {
		out.Suggestion = &sg
	}
	return out, nil
}

// begin starts a write transaction and takes the user's row first, so
// the read that follows is made under the write lock: a deferred
// transaction that reads, then writes, can lose to another writer with
// a snapshot error that busy_timeout does not retry. It returns the
// active pointer.
func (s *Service) begin(ctx context.Context, user uuid.UUID) (*sql.Tx, string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("playmat: begin: %w", err)
	}
	res, err := tx.ExecContext(ctx, `UPDATE users SET playmat_id = playmat_id WHERE id = ?`, user.String())
	if err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("playmat: lock: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_ = tx.Rollback()
		return nil, "", ErrNoUser
	}
	var active sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT playmat_id FROM users WHERE id = ?`, user.String()).Scan(&active); err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("playmat: read: %w", err)
	}
	return tx, active.String, nil
}

// slotID reads the id saved in a slot, "" for an empty one.
func slotID(ctx context.Context, tx *sql.Tx, user uuid.UUID, slot int) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `SELECT playmat_id FROM user_playmats WHERE user_id = ? AND slot = ?`, user.String(), slot).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("playmat: read slot: %w", err)
	}
	return id, nil
}

// setActive writes the active pointer (NULL for ""). Every caller has
// already made sure id is saved in one of the user's slots, or is "":
// that is the invariant users.playmat_id keeps (migration 0013).
func setActive(ctx context.Context, tx *sql.Tx, user uuid.UUID, id string) error {
	var arg any
	if id != "" {
		arg = id
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET playmat_id = ? WHERE id = ?`, arg, user.String()); err != nil {
		return fmt.Errorf("playmat: write active: %w", err)
	}
	return nil
}

func (s *Service) commit(tx *sql.Tx, user uuid.UUID, active string) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("playmat: commit: %w", err)
	}
	s.mu.Lock()
	s.cache[user] = active
	s.mu.Unlock()
	return nil
}

// putRow saves id into the slot and returns the id it replaced.
func (s *Service) putRow(ctx context.Context, user uuid.UUID, slot int, id string, w, h int, expectOld string) (string, error) {
	tx, active, err := s.begin(ctx, user)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	old, err := slotID(ctx, tx, user, slot)
	if err != nil {
		return "", err
	}
	if expectOld != "" && old != expectOld {
		return "", ErrConflict
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_playmats (user_id, slot, playmat_id, width, height, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id, slot) DO UPDATE SET
			playmat_id = excluded.playmat_id, width = excluded.width,
			height = excluded.height, created_at = excluded.created_at`,
		user.String(), slot, id, w, h, time.Now().UnixMilli())
	if err != nil {
		return "", fmt.Errorf("playmat: write slot: %w", err)
	}
	switch {
	case old != "" && old == active:
		active = id // replacing the mat on show: the table shows the new one
	case old == "" && active == "":
		active = id // the first mat saved by someone showing none
	}
	if err := setActive(ctx, tx, user, active); err != nil {
		return "", err
	}
	if err := s.commit(tx, user, active); err != nil {
		return "", err
	}
	return old, nil
}

// Remove deletes the playmat in a slot and its file; if it was the
// active one, the user shows none. Removing an empty slot is not an
// error.
func (s *Service) Remove(ctx context.Context, user uuid.UUID, slot int) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	if !ValidSlot(slot) {
		return ErrBadSlot
	}
	tx, active, err := s.begin(ctx, user)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	old, err := slotID(ctx, tx, user, slot)
	if err != nil {
		return err
	}
	if old == "" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_playmats WHERE user_id = ? AND slot = ?`, user.String(), slot); err != nil {
		return fmt.Errorf("playmat: delete slot: %w", err)
	}
	if old == active {
		active = ""
		if err := setActive(ctx, tx, user, ""); err != nil {
			return err
		}
	}
	if err := s.commit(tx, user, active); err != nil {
		return err
	}
	return s.files.Delete(old)
}

// RemoveAll deletes every saved playmat and clears the active pointer:
// the admin's "take them all away" (ADR 0128 §9, §11).
func (s *Service) RemoveAll(ctx context.Context, user uuid.UUID) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	tx, _, err := s.begin(ctx, user)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT playmat_id FROM user_playmats WHERE user_id = ?`, user.String())
	if err != nil {
		return fmt.Errorf("playmat: read slots: %w", err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return fmt.Errorf("playmat: read slots: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("playmat: read slots: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_playmats WHERE user_id = ?`, user.String()); err != nil {
		return fmt.Errorf("playmat: delete slots: %w", err)
	}
	if err := setActive(ctx, tx, user, ""); err != nil {
		return err
	}
	if err := s.commit(tx, user, ""); err != nil {
		return err
	}
	var first error
	for _, id := range ids {
		if err := s.files.Delete(id); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// Activate makes slot the playmat the table shows. Slot 0 shows none,
// and keeps every saved mat. Switching never touches a file.
func (s *Service) Activate(ctx context.Context, user uuid.UUID, slot int) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	if slot != 0 && !ValidSlot(slot) {
		return ErrBadSlot
	}
	tx, _, err := s.begin(ctx, user)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	id := ""
	if slot != 0 {
		if id, err = slotID(ctx, tx, user, slot); err != nil {
			return err
		}
		if id == "" {
			return ErrNoSlot
		}
	}
	if err := setActive(ctx, tx, user, id); err != nil {
		return err
	}
	return s.commit(tx, user, id)
}

// Fit crops the playmat in a slot to the ideal shape with its top-left
// corner at (x, y) of the stored image, scales the crop down to the
// ideal size (never up), and stores the result as a new file. The old
// file is deleted and the slot, and the active pointer if it was the
// active slot, move to the new one. The rectangle must be one Suggest
// would offer, anywhere along the axis being cropped.
func (s *Service) Fit(ctx context.Context, user uuid.UUID, slot, x, y int) (Slot, error) {
	if !s.Enabled() {
		return Slot{}, ErrDisabled
	}
	if !ValidSlot(slot) {
		return Slot{}, ErrBadSlot
	}
	st, err := s.State(ctx, user)
	if err != nil {
		return Slot{}, err
	}
	cur, ok := st.Find(slot)
	if !ok {
		return Slot{}, ErrNoSlot
	}
	sg, err := validCrop(cur.Width, cur.Height, x, y)
	if err != nil {
		return Slot{}, err
	}
	if err := s.acquire(ctx); err != nil {
		return Slot{}, err
	}
	defer s.release()
	f, _, err := s.files.Open(cur.ID)
	if err != nil {
		return Slot{}, ErrNoSlot
	}
	src, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		return Slot{}, fmt.Errorf("playmat: decode stored image: %w", err)
	}
	if b := src.Bounds(); b.Dx() != cur.Width || b.Dy() != cur.Height {
		return Slot{}, ErrBadCrop
	}
	out := cropAndScale(src, sg)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return Slot{}, fmt.Errorf("playmat: encode: %w", err)
	}
	return s.put(ctx, user, slot, buf.Bytes(), sg.TargetWidth, sg.TargetHeight, cur.ID)
}

// Open opens a stored image for the serving route.
func (s *Service) Open(id string) (*os.File, os.FileInfo, error) {
	if s == nil {
		return nil, nil, ErrDisabled
	}
	return s.files.Open(id)
}
