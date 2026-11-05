package gamestats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// SQLPrefsStore is migration 0030's stats_prefs and stats_settings.
// Every *_at column is Unix milliseconds.
type SQLPrefsStore struct {
	db *sql.DB
}

// NewSQLPrefsStore wraps an opened, migrated database. The caller owns d.
func NewSQLPrefsStore(d *db.DB) *SQLPrefsStore { return &SQLPrefsStore{db: d.DB} }

var _ PrefsStore = (*SQLPrefsStore)(nil)

func scanPrefs(adv sql.NullInt64, by sql.NullString, updated int64) Prefs {
	var p Prefs
	if adv.Valid {
		on := adv.Int64 == 1
		p.Advanced = &on
		p.SetBy = SetBy(by.String)
	}
	if updated > 0 {
		p.UpdatedAt = time.UnixMilli(updated).UTC()
	}
	return p
}

// Prefs implements PrefsStore.
func (s *SQLPrefsStore) Prefs(ctx context.Context, userID string) (Prefs, error) {
	var (
		adv     sql.NullInt64
		by      sql.NullString
		updated int64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT advanced, advanced_set_by, updated_at FROM stats_prefs WHERE user_id = ?`, userID).
		Scan(&adv, &by, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return Prefs{}, nil
	}
	if err != nil {
		return Prefs{}, fmt.Errorf("gamestats: prefs: %w", err)
	}
	return scanPrefs(adv, by, updated), nil
}

// SetAdvanced implements PrefsStore. It touches only the advanced
// columns, so the sharing columns later slices add survive a switch.
func (s *SQLPrefsStore) SetAdvanced(ctx context.Context, userID string, on bool, by SetBy, now time.Time) (Prefs, error) {
	if by != SetBySelf && by != SetByAdmin {
		return Prefs{}, fmt.Errorf("gamestats: set by %q", by)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set advanced: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	// A deleted account is a tombstone (ADR 0149): it has no setting.
	var n int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, userID).Scan(&n); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set advanced: %w", err)
	}
	if n == 0 {
		return Prefs{}, ErrNoUser
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO stats_prefs (user_id, advanced, advanced_set_by, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET advanced = excluded.advanced,
		   advanced_set_by = excluded.advanced_set_by, updated_at = excluded.updated_at`,
		userID, boolInt(on), string(by), now.UnixMilli()); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set advanced: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set advanced: %w", err)
	}
	return Prefs{Advanced: &on, SetBy: by, UpdatedAt: time.UnixMilli(now.UnixMilli()).UTC()}, nil
}

// Default implements PrefsStore.
func (s *SQLPrefsStore) Default(ctx context.Context) (DefaultSetting, error) {
	var on, updated int64
	err := s.db.QueryRowContext(ctx,
		`SELECT advanced_default, updated_at FROM stats_settings WHERE id = 1`).Scan(&on, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultSetting{Advanced: AdvancedDefault}, nil
	}
	if err != nil {
		return DefaultSetting{}, fmt.Errorf("gamestats: default: %w", err)
	}
	return DefaultSetting{Advanced: on == 1, Stored: true, UpdatedAt: time.UnixMilli(updated).UTC()}, nil
}

// SetDefault implements PrefsStore.
func (s *SQLPrefsStore) SetDefault(ctx context.Context, on bool, now time.Time) (DefaultSetting, error) {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO stats_settings (id, advanced_default, updated_at) VALUES (1, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET advanced_default = excluded.advanced_default, updated_at = excluded.updated_at`,
		boolInt(on), now.UnixMilli())
	if err != nil {
		return DefaultSetting{}, fmt.Errorf("gamestats: set default: %w", err)
	}
	return DefaultSetting{Advanced: on, Stored: true, UpdatedAt: time.UnixMilli(now.UnixMilli()).UTC()}, nil
}

// Accounts implements PrefsStore.
func (s *SQLPrefsStore) Accounts(ctx context.Context, limit int) ([]AccountPrefs, error) {
	if limit <= 0 || limit > MaxPrefsAccounts {
		limit = MaxPrefsAccounts
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_id, advanced, advanced_set_by, updated_at FROM stats_prefs
		  WHERE advanced IS NOT NULL ORDER BY updated_at DESC, user_id LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("gamestats: accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []AccountPrefs{}
	for rows.Next() {
		var (
			a       AccountPrefs
			adv     sql.NullInt64
			by      sql.NullString
			updated int64
		)
		if err := rows.Scan(&a.UserID, &adv, &by, &updated); err != nil {
			return nil, fmt.Errorf("gamestats: accounts: %w", err)
		}
		a.Prefs = scanPrefs(adv, by, updated)
		out = append(out, a)
	}
	return out, rows.Err()
}
