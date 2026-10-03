package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AdminModeTTL is how long admin mode lasts once switched on (ADR 0112
// §2 item 1, owner answer 1): 12 hours, like sudo, the default length
// of a session with no user (CMDCTRL_SESSION_TTL). Switching on again
// restarts it.
const AdminModeTTL = 12 * time.Hour

// adminModeClockSkew is how far in the future a stored admin_mode_at
// may be and still count. A row from the future means the clock went
// back since the switch; past this much it is not trusted, and the
// person is in player mode.
const adminModeClockSkew = time.Minute

// AdminModeStore is where users.admin_mode_at lives (migration 0009).
// SQLStore implements it. AdminModes is its only caller, apart from
// InvalidateSessions, which zeroes the column in the same statement
// that moves the revocation watermark (ADR 0112 §2 item 7).
type AdminModeStore interface {
	// AdminModesOn returns every user whose admin_mode_at is set
	// (non-zero), lapsed or not.
	AdminModesOn(ctx context.Context) (map[uuid.UUID]time.Time, error)
	// SetAdminMode stores at as the user's admin_mode_at; the zero time
	// stores 0 (off). ErrNotFound if there is no such user.
	SetAdminMode(ctx context.Context, id uuid.UUID, at time.Time) error
	// ClearAdminModeIfAt sets the user's admin_mode_at to 0 only if it
	// still holds at, so a sweep never undoes a switch it raced. It
	// reports whether it cleared anything.
	ClearAdminModeIfAt(ctx context.Context, id uuid.UUID, at time.Time) (bool, error)
}

// AdminModes is ADR 0112 §2's per-person admin mode: which allowlisted
// people have switched admin mode on, and since when. Off is the
// default and the answer on every failure.
//
// It mirrors Revocations. Every authenticated request asks On, so the
// rows are held in memory: loaded once at boot, then written through on
// every switch (the database first, then the map). The server process
// is the database's only writer, and every write to admin_mode_at goes
// through this type or through InvalidateSessions (which tells this
// type through Revocations), so the cache is never stale.
//
// Unlike Revocations, a failed boot load does not stop the server. The
// safe failure here is fewer admins: NewAdminModes hands back an empty
// set and the error, and the caller logs it. Everyone is then in player
// mode until they switch again.
//
// A nil *AdminModes is the empty set: nobody is in admin mode.
type AdminModes struct {
	store AdminModeStore
	now   func() time.Time

	// wmu serialises the writers (Set, Sweep, forget), so the map
	// always ends up holding what the database last stored. The read
	// lock below is held only for map access, never across a write to
	// the database.
	wmu sync.Mutex

	mu sync.RWMutex
	on map[uuid.UUID]time.Time
}

// NewAdminModes loads every set admin_mode_at from store. It always
// returns a usable *AdminModes. On error it is empty, so every
// allowlisted person is in player mode; the caller should log the
// error at ERROR and carry on.
func NewAdminModes(ctx context.Context, store AdminModeStore) (*AdminModes, error) {
	m := &AdminModes{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
		on:    map[uuid.UUID]time.Time{},
	}
	if store == nil {
		return m, errors.New("users: no admin mode store")
	}
	loaded, err := store.AdminModesOn(ctx)
	if err != nil {
		return m, fmt.Errorf("users: load admin modes: %w", err)
	}
	for id, at := range loaded {
		if !at.IsZero() {
			m.on[id] = at
		}
	}
	return m, nil
}

// SetClock replaces the clock On's callers, Set and Sweep read. For
// tests; production keeps time.Now.
func (m *AdminModes) SetClock(now func() time.Time) {
	if m != nil && now != nil {
		m.now = now
	}
}

// Now is the clock this set reads. Nil-safe.
func (m *AdminModes) Now() time.Time {
	if m == nil || m.now == nil {
		return time.Now().UTC()
	}
	return m.now()
}

// On reports whether userID is in admin mode at now: switched on, and
// not yet AdminModeTTL past the switch. False for a nil set, uuid.Nil
// and anyone not in the map. A lapsed entry reads false at once; the
// sweeper removes it later.
func (m *AdminModes) On(userID uuid.UUID, now time.Time) bool {
	_, ok := m.EndsAt(userID, now)
	return ok
}

// EndsAt is when userID's admin mode lapses, and whether it is on at
// now. The time is meaningless when the bool is false.
func (m *AdminModes) EndsAt(userID uuid.UUID, now time.Time) (time.Time, bool) {
	if m == nil || userID == uuid.Nil {
		return time.Time{}, false
	}
	m.mu.RLock()
	at, ok := m.on[userID]
	m.mu.RUnlock()
	if !ok || !activeAt(at, now) {
		return time.Time{}, false
	}
	return at.Add(AdminModeTTL), true
}

func activeAt(at, now time.Time) bool {
	if at.IsZero() || at.After(now.Add(adminModeClockSkew)) {
		return false
	}
	return now.Before(at.Add(AdminModeTTL))
}

// Set switches userID's admin mode on (restarting the 12 hours) or
// off, and returns when it now ends (zero when off).
//
// The database is written first. Switching on that fails to write
// leaves the map as it was and returns the error. Switching off that
// fails to write still drops the person from the map, so they are in
// player mode in this process whatever the database says, and returns
// the error: the safe half of a failed write is the one that removes
// rights. ErrNotFound for an unknown user.
func (m *AdminModes) Set(ctx context.Context, userID uuid.UUID, on bool) (time.Time, error) {
	if m == nil || m.store == nil {
		return time.Time{}, errors.New("users: admin mode needs the user database")
	}
	if userID == uuid.Nil {
		return time.Time{}, ErrNotFound
	}
	m.wmu.Lock()
	defer m.wmu.Unlock()
	if !on {
		err := m.store.SetAdminMode(ctx, userID, time.Time{})
		m.drop(userID)
		return time.Time{}, err
	}
	at := m.Now().Truncate(time.Millisecond)
	if err := m.store.SetAdminMode(ctx, userID, at); err != nil {
		return time.Time{}, err
	}
	m.mu.Lock()
	m.on[userID] = at
	m.mu.Unlock()
	return at.Add(AdminModeTTL), nil
}

// Sweep ends every admin mode that has lapsed at now: the row is set
// back to 0 and the person leaves the map. It returns the users it
// ended, for the caller to log and to rebind their sockets. A row that
// fails to clear is still dropped from the map (On already reads it as
// lapsed) and its error is joined into the one returned.
func (m *AdminModes) Sweep(ctx context.Context, now time.Time) ([]uuid.UUID, error) {
	if m == nil {
		return nil, nil
	}
	m.wmu.Lock()
	defer m.wmu.Unlock()
	m.mu.RLock()
	lapsed := map[uuid.UUID]time.Time{}
	for id, at := range m.on {
		if !activeAt(at, now) {
			lapsed[id] = at
		}
	}
	m.mu.RUnlock()

	var (
		ended []uuid.UUID
		errs  []error
	)
	for id, at := range lapsed {
		if m.store != nil {
			if _, err := m.store.ClearAdminModeIfAt(ctx, id, at); err != nil {
				errs = append(errs, fmt.Errorf("users: end lapsed admin mode for %s: %w", id, err))
			}
		}
		m.drop(id)
		ended = append(ended, id)
	}
	return ended, errors.Join(errs...)
}

// forget drops userID from the map after something else has already
// zeroed the row: Revocations.RevokeAll, whose statement clears
// admin_mode_at with the watermark.
func (m *AdminModes) forget(userID uuid.UUID) {
	if m == nil {
		return
	}
	m.wmu.Lock()
	defer m.wmu.Unlock()
	m.drop(userID)
}

func (m *AdminModes) drop(userID uuid.UUID) {
	m.mu.Lock()
	delete(m.on, userID)
	m.mu.Unlock()
}

// AdminModesOn implements AdminModeStore.
func (s *SQLStore) AdminModesOn(ctx context.Context) (map[uuid.UUID]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, admin_mode_at FROM users WHERE admin_mode_at > 0`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[uuid.UUID]time.Time{}
	for rows.Next() {
		var (
			idStr string
			ms    int64
		)
		if err := rows.Scan(&idStr, &ms); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("users: stored id %q: %w", idStr, err)
		}
		out[id] = time.UnixMilli(ms).UTC()
	}
	return out, rows.Err()
}

// SetAdminMode implements AdminModeStore.
func (s *SQLStore) SetAdminMode(ctx context.Context, id uuid.UUID, at time.Time) error {
	var ms int64
	if !at.IsZero() {
		ms = at.UnixMilli()
	}
	var stored int64
	err := s.db.QueryRowContext(ctx,
		`UPDATE users SET admin_mode_at = ? WHERE id = ? RETURNING admin_mode_at`, ms, id.String()).Scan(&stored)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("users: set admin mode for %s: %w", id, err)
	}
	return nil
}

// ClearAdminModeIfAt implements AdminModeStore.
func (s *SQLStore) ClearAdminModeIfAt(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET admin_mode_at = 0 WHERE id = ? AND admin_mode_at = ?`, id.String(), at.UnixMilli())
	if err != nil {
		return false, fmt.Errorf("users: clear admin mode for %s: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

var _ AdminModeStore = (*SQLStore)(nil)
