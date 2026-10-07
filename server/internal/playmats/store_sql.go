package playmats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// SQLStore is the user_playmats table from migration 0011. updated_at
// is Unix milliseconds.
type SQLStore struct {
	db  *sql.DB
	now func() time.Time
}

// NewSQLStore wraps an opened, migrated database. The caller owns d.
func NewSQLStore(d *db.DB) *SQLStore {
	return &SQLStore{db: d.DB, now: func() time.Time { return time.Now().UTC() }}
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func get(ctx context.Context, q queryRower, user uuid.UUID) (Playmat, error) {
	var (
		out     Playmat
		updated int64
	)
	err := q.QueryRowContext(ctx,
		`SELECT file, wash, updated_at FROM user_playmats WHERE user_id = ?`, user.String()).
		Scan(&out.File, &out.Wash, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return Playmat{}, ErrNotFound
	}
	if err != nil {
		return Playmat{}, fmt.Errorf("playmats: get: %w", err)
	}
	out.UpdatedAt = time.UnixMilli(updated).UTC()
	return out, nil
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, user uuid.UUID) (Playmat, error) {
	return get(ctx, s.db, user)
}

// Put implements Store. Reading the old row and writing the new one
// share a transaction, so the file it reports as replaced is the one
// this write replaced.
func (s *SQLStore) Put(ctx context.Context, user uuid.UUID, file string, wash int) (Playmat, error) {
	if user == uuid.Nil {
		return Playmat{}, errors.New("playmats: user is required")
	}
	if !ValidFile(file) {
		return Playmat{}, fmt.Errorf("%w\nnot a stored playmat", ErrInvalid)
	}
	if err := ValidWash(wash); err != nil {
		return Playmat{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Playmat{}, err
	}
	defer func() { _ = tx.Rollback() }()
	old, err := get(ctx, tx, user)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Playmat{}, err
	}
	nowMs := s.now().Truncate(time.Millisecond).UnixMilli()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_playmats (user_id, file, wash, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET file = excluded.file, wash = excluded.wash, updated_at = excluded.updated_at`,
		user.String(), file, wash, nowMs); err != nil {
		return Playmat{}, fmt.Errorf("playmats: put: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Playmat{}, err
	}
	return old, nil
}

// SetWash implements Store.
func (s *SQLStore) SetWash(ctx context.Context, user uuid.UUID, wash int) (Playmat, error) {
	if err := ValidWash(wash); err != nil {
		return Playmat{}, err
	}
	nowMs := s.now().Truncate(time.Millisecond).UnixMilli()
	res, err := s.db.ExecContext(ctx,
		`UPDATE user_playmats SET wash = ?, updated_at = ? WHERE user_id = ?`, wash, nowMs, user.String())
	if err != nil {
		return Playmat{}, fmt.Errorf("playmats: set wash: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Playmat{}, ErrNotFound
	}
	return s.Get(ctx, user)
}

// Delete implements Store.
func (s *SQLStore) Delete(ctx context.Context, user uuid.UUID) (Playmat, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Playmat{}, err
	}
	defer func() { _ = tx.Rollback() }()
	old, err := get(ctx, tx, user)
	if err != nil {
		return Playmat{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_playmats WHERE user_id = ?`, user.String()); err != nil {
		return Playmat{}, fmt.Errorf("playmats: delete: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Playmat{}, err
	}
	return old, nil
}

var _ Store = (*SQLStore)(nil)
