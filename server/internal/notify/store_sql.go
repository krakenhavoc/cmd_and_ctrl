package notify

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// SQLStore is the notification_prefs table from migration 0021. Every
// *_at column is Unix milliseconds.
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

func get(ctx context.Context, q queryRower, user uuid.UUID) (Prefs, error) {
	var (
		out    Prefs
		failed sql.NullInt64
	)
	err := q.QueryRowContext(ctx,
		`SELECT turn, table_ready, table_filled, dm_failed_at FROM notification_prefs WHERE user_id = ?`, user.String()).
		Scan(&out.Turn, &out.TableReady, &out.TableFilled, &failed)
	if errors.Is(err, sql.ErrNoRows) {
		return Prefs{}, nil
	}
	if err != nil {
		return Prefs{}, fmt.Errorf("notify: get: %w", err)
	}
	if failed.Valid {
		out.DMFailedAt = time.UnixMilli(failed.Int64).UTC()
	}
	return out, nil
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, user uuid.UUID) (Prefs, error) {
	return get(ctx, s.db, user)
}

// Put implements Store. The read and the write share one transaction,
// so two tabs saving at once each apply their own switch to the other's
// result.
func (s *SQLStore) Put(ctx context.Context, user uuid.UUID, u Update) (Prefs, error) {
	if user == uuid.Nil {
		return Prefs{}, errors.New("notify: user is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Prefs{}, err
	}
	defer func() { _ = tx.Rollback() }()
	cur, err := get(ctx, tx, user)
	if err != nil {
		return Prefs{}, err
	}
	next := cur
	turnedOn := false
	apply := func(sw *bool, to *bool) {
		if to == nil {
			return
		}
		if *to && !*sw {
			turnedOn = true
		}
		*sw = *to
	}
	apply(&next.Turn, u.Turn)
	apply(&next.TableReady, u.TableReady)
	apply(&next.TableFilled, u.TableFilled)
	var failed any
	if !cur.DMFailedAt.IsZero() && !turnedOn {
		failed = cur.DMFailedAt.UnixMilli()
	} else {
		next.DMFailedAt = time.Time{}
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO notification_prefs (user_id, turn, table_ready, table_filled, dm_failed_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET turn = excluded.turn, table_ready = excluded.table_ready,
		   table_filled = excluded.table_filled, dm_failed_at = excluded.dm_failed_at, updated_at = excluded.updated_at`,
		user.String(), next.Turn, next.TableReady, next.TableFilled, failed, s.now().UnixMilli()); err != nil {
		return Prefs{}, fmt.Errorf("notify: put: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Prefs{}, err
	}
	return next, nil
}

// Active implements Store.
func (s *SQLStore) Active(ctx context.Context) (map[uuid.UUID]Prefs, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_id, turn, table_ready, table_filled FROM notification_prefs
		 WHERE dm_failed_at IS NULL AND (turn = 1 OR table_ready = 1 OR table_filled = 1)`)
	if err != nil {
		return nil, fmt.Errorf("notify: active: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make(map[uuid.UUID]Prefs)
	for rows.Next() {
		var (
			raw string
			p   Prefs
		)
		if err := rows.Scan(&raw, &p.Turn, &p.TableReady, &p.TableFilled); err != nil {
			return nil, fmt.Errorf("notify: active: %w", err)
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			continue
		}
		out[id] = p
	}
	return out, rows.Err()
}

// MarkFailed implements Store. A person with no row has nothing on, so
// nothing was sent to them; there is nothing to mark.
func (s *SQLStore) MarkFailed(ctx context.Context, user uuid.UUID, at time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`UPDATE notification_prefs SET dm_failed_at = ? WHERE user_id = ?`, at.UnixMilli(), user.String()); err != nil {
		return fmt.Errorf("notify: mark failed: %w", err)
	}
	return nil
}

var _ Store = (*SQLStore)(nil)
