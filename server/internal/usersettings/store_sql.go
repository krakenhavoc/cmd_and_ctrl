package usersettings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// SQLStore is the user_settings table from migration 0008. Every *_at
// column is Unix milliseconds.
type SQLStore struct {
	db  *sql.DB
	now func() time.Time
}

// NewSQLStore wraps an opened, migrated database. The caller owns d.
func NewSQLStore(d *db.DB) *SQLStore {
	return &SQLStore{db: d.DB, now: func() time.Time { return time.Now().UTC() }}
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, user uuid.UUID) (Settings, error) {
	return get(ctx, s.db, user)
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func get(ctx context.Context, q queryRower, user uuid.UUID) (Settings, error) {
	var (
		out     Settings
		body    string
		updated int64
	)
	err := q.QueryRowContext(ctx,
		`SELECT version, revision, body, updated_at FROM user_settings WHERE user_id = ?`, user.String()).
		Scan(&out.Version, &out.Revision, &body, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return Settings{}, ErrNotFound
	}
	if err != nil {
		return Settings{}, fmt.Errorf("usersettings: get: %w", err)
	}
	out.Body = json.RawMessage(body)
	out.UpdatedAt = time.UnixMilli(updated).UTC()
	return out, nil
}

// Put implements Store. The compare and the write share one
// transaction. Two writers racing on the same revision cannot both win:
// the UPDATE names the revision it read, and the INSERT for a first copy
// hits the primary key, so the loser gets a conflict, never a lost
// write.
func (s *SQLStore) Put(ctx context.Context, user uuid.UUID, version int, body json.RawMessage, ifMatch int64) (Settings, error) {
	if user == uuid.Nil {
		return Settings{}, errors.New("usersettings: user is required")
	}
	if err := Validate(version, body); err != nil {
		return Settings{}, err
	}
	out, err := s.put(ctx, user, version, body, ifMatch)
	if err != nil && isBusy(err) {
		out, err = s.put(ctx, user, version, body, ifMatch)
	}
	return out, err
}

func (s *SQLStore) put(ctx context.Context, user uuid.UUID, version int, body json.RawMessage, ifMatch int64) (Settings, error) {
	nowMs := s.now().Truncate(time.Millisecond).UnixMilli()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Settings{}, err
	}
	defer func() { _ = tx.Rollback() }()

	cur, err := get(ctx, tx, user)
	switch {
	case errors.Is(err, ErrNotFound):
		if ifMatch != 0 {
			return Settings{}, ErrRevisionConflict
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_settings (user_id, version, revision, body, updated_at) VALUES (?, ?, 1, ?, ?)`,
			user.String(), version, string(body), nowMs); err != nil {
			if isConstraint(err) && strings.Contains(err.Error(), "PRIMARY KEY") {
				// Lost a race with another first write.
				return Settings{}, ErrRevisionConflict
			}
			return Settings{}, fmt.Errorf("usersettings: insert: %w", err)
		}
	case err != nil:
		return Settings{}, err
	default:
		if cur.Revision != ifMatch {
			return cur, ErrRevisionConflict
		}
		if version < cur.Version {
			return cur, ErrVersionTooOld
		}
		res, err := tx.ExecContext(ctx,
			`UPDATE user_settings SET version = ?, revision = revision + 1, body = ?, updated_at = ?
			 WHERE user_id = ? AND revision = ?`,
			version, string(body), nowMs, user.String(), ifMatch)
		if err != nil {
			return Settings{}, fmt.Errorf("usersettings: update: %w", err)
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return Settings{}, ErrRevisionConflict
		}
	}
	out, err := get(ctx, tx, user)
	if err != nil {
		return Settings{}, err
	}
	if err := tx.Commit(); err != nil {
		return Settings{}, err
	}
	return out, nil
}

func isConstraint(err error) bool {
	return err != nil && strings.Contains(err.Error(), "constraint failed")
}

func isBusy(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "database is locked"))
}

var _ Store = (*SQLStore)(nil)
