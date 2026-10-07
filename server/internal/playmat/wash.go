package playmat

// wash.go is the owner-set wash (ADR 0128 §10): one number per account,
// not per slot (§11 notes a per-mat wash as a possible follow-up).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

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
