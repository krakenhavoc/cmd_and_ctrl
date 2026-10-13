package budget

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// SQLStore is the bot_model_budget table from migration 0014: one row
// per counter, keyed "day:<date>", "month:<yyyy-mm>" or "disabled".
// updated_at is Unix milliseconds.
type SQLStore struct {
	db  *sql.DB
	now func() time.Time
}

// NewSQLStore wraps an opened, migrated database. The caller owns d.
func NewSQLStore(d *db.DB) *SQLStore {
	return &SQLStore{db: d.DB, now: func() time.Time { return time.Now().UTC() }}
}

const (
	dayPrefix   = "day:"
	monthPrefix = "month:"
	disabledKey = "disabled"
)

func (s *SQLStore) get(ctx context.Context, key string) (int64, error) {
	var v int64
	err := s.db.QueryRowContext(ctx, `SELECT value FROM bot_model_budget WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("budget: load %s: %w", key, err)
	}
	return v, nil
}

// Load implements Store.
func (s *SQLStore) Load(ctx context.Context, day, month string) (Snapshot, error) {
	var (
		out Snapshot
		err error
		off int64
	)
	if out.DayTokens, err = s.get(ctx, dayPrefix+day); err != nil {
		return Snapshot{}, err
	}
	if out.MonthTokens, err = s.get(ctx, monthPrefix+month); err != nil {
		return Snapshot{}, err
	}
	if off, err = s.get(ctx, disabledKey); err != nil {
		return Snapshot{}, err
	}
	out.Disabled = off != 0
	return out, nil
}

const upsert = `INSERT INTO bot_model_budget (key, value, updated_at) VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`

// SaveTokens implements Store. It writes both counters and deletes the
// rows of every other period in the same transaction, so the table
// holds one day and one month however long the server has run.
func (s *SQLStore) SaveTokens(ctx context.Context, day string, dayTokens int64, month string, monthTokens int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("budget: save: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := s.now().UnixMilli()
	if _, err := tx.ExecContext(ctx, upsert, dayPrefix+day, dayTokens, now); err != nil {
		return fmt.Errorf("budget: save day: %w", err)
	}
	if _, err := tx.ExecContext(ctx, upsert, monthPrefix+month, monthTokens, now); err != nil {
		return fmt.Errorf("budget: save month: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM bot_model_budget WHERE key NOT IN (?, ?, ?)`,
		dayPrefix+day, monthPrefix+month, disabledKey); err != nil {
		return fmt.Errorf("budget: prune: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("budget: save: %w", err)
	}
	return nil
}

// SaveDisabled implements Store.
func (s *SQLStore) SaveDisabled(ctx context.Context, disabled bool) error {
	v := int64(0)
	if disabled {
		v = 1
	}
	if _, err := s.db.ExecContext(ctx, upsert, disabledKey, v, s.now().UnixMilli()); err != nil {
		return fmt.Errorf("budget: save switch: %w", err)
	}
	return nil
}
