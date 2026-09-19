package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// WatermarkStore is where users.sessions_invalid_before lives (ADR 0051
// decision 6). SQLStore implements it. Revocations is its only caller:
// every write to the column goes through Revocations.RevokeAll, which
// is what lets Revocations trust its cache.
type WatermarkStore interface {
	// SessionWatermarks returns every user whose watermark is set
	// (non-zero). A user absent from the map has never revoked
	// anything.
	SessionWatermarks(ctx context.Context) (map[uuid.UUID]time.Time, error)
	// InvalidateSessions moves the user's watermark forward to at and
	// returns the value now stored. It never moves the watermark back:
	// an earlier at leaves a later stored value in place.
	// ErrNotFound if there is no such user.
	InvalidateSessions(ctx context.Context, id uuid.UUID, at time.Time) (time.Time, error)
}

// Revocations is decision 6's per-user revocation, and the
// auth.RevocationList the server's authenticator is wrapped with.
//
// # The cache
//
// Every authenticated HTTP request and every WebSocket upgrade asks
// Revoked, so it must never touch SQLite. It doesn't:
//
//   - NewRevocations reads every non-zero sessions_invalid_before
//     once, at boot. The table holds a playgroup, so this is a handful
//     of rows.
//   - Revoked is a map lookup under a read lock. A user who is not in
//     the map has a zero watermark, which is also true of every user
//     created after boot, since a new row defaults to 0.
//   - RevokeAll writes the database first and then replaces the
//     user's entry with the value the database now holds. That write
//     is the only event that can change the answer, so the cache is
//     invalidated exactly when it goes stale. If the write fails, the
//     cache is left alone and the caller gets the error.
//
// This is sound because the server process is the database's only
// writer (decision 1), and this type is the only code that writes the
// column. A second writer (a migration, a hand-run UPDATE on a live
// server) would not be seen until the next restart. Anything that
// needs to revoke sessions must call RevokeAll.
type Revocations struct {
	store WatermarkStore
	now   func() time.Time

	mu     sync.RWMutex
	before map[uuid.UUID]time.Time
}

// NewRevocations loads every watermark from store. An error here should
// stop the boot: a server that cannot read the watermarks would
// otherwise accept sessions a user has revoked.
func NewRevocations(ctx context.Context, store WatermarkStore) (*Revocations, error) {
	m, err := store.SessionWatermarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: load session watermarks: %w", err)
	}
	if m == nil {
		m = map[uuid.UUID]time.Time{}
	}
	return &Revocations{
		store:  store,
		now:    func() time.Time { return time.Now().UTC() },
		before: m,
	}, nil
}

// Revoked implements auth.RevocationList: a session is revoked when it
// was issued at or before its user's watermark. "At" matters: the
// request that revokes carries a token issued before the watermark,
// and a token minted in the same millisecond as the revocation is
// refused rather than slipping through on a tie.
func (r *Revocations) Revoked(userID uuid.UUID, issuedAt time.Time) bool {
	r.mu.RLock()
	w, ok := r.before[userID]
	r.mu.RUnlock()
	return ok && !issuedAt.After(w)
}

// RevokeAll withdraws every session userID holds now: logout-everywhere
// and the admin's revoke-sessions both land here. It returns the
// watermark stored, in milliseconds like every other *_at column.
// Sessions issued after it validate as normal, so the user can sign in
// again straight away. ErrNotFound for an unknown user.
func (r *Revocations) RevokeAll(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	at := r.now().Truncate(time.Millisecond)
	stored, err := r.store.InvalidateSessions(ctx, userID, at)
	if err != nil {
		return time.Time{}, err
	}
	r.mu.Lock()
	r.before[userID] = stored
	r.mu.Unlock()
	return stored, nil
}

var _ auth.RevocationList = (*Revocations)(nil)

// SessionWatermarks implements WatermarkStore.
func (s *SQLStore) SessionWatermarks(ctx context.Context) (map[uuid.UUID]time.Time, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, sessions_invalid_before FROM users WHERE sessions_invalid_before > 0`)
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

// InvalidateSessions implements WatermarkStore. MAX keeps the watermark
// monotonic, so two revocations racing each other cannot un-revoke
// anything.
func (s *SQLStore) InvalidateSessions(ctx context.Context, id uuid.UUID, at time.Time) (time.Time, error) {
	var ms int64
	err := s.db.QueryRowContext(ctx,
		`UPDATE users SET sessions_invalid_before = MAX(sessions_invalid_before, ?) WHERE id = ?
		 RETURNING sessions_invalid_before`,
		at.UnixMilli(), id.String()).Scan(&ms)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrNotFound
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("users: invalidate sessions for %s: %w", id, err)
	}
	return time.UnixMilli(ms).UTC(), nil
}

var _ WatermarkStore = (*SQLStore)(nil)
