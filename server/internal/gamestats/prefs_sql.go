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

func scanPrefs(adv sql.NullInt64, by sql.NullString, updated int64, linkAt sql.NullInt64, share int64) Prefs {
	p := Prefs{SharePods: share == 1}
	if linkAt.Valid && linkAt.Int64 > 0 {
		p.ShareLinkAt = time.UnixMilli(linkAt.Int64).UTC()
	}
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
		linkAt  sql.NullInt64
		share   int64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT advanced, advanced_set_by, updated_at, share_link_at, share_pods FROM stats_prefs WHERE user_id = ?`, userID).
		Scan(&adv, &by, &updated, &linkAt, &share)
	if errors.Is(err, sql.ErrNoRows) {
		return Prefs{}, nil
	}
	if err != nil {
		return Prefs{}, fmt.Errorf("gamestats: prefs: %w", err)
	}
	return scanPrefs(adv, by, updated, linkAt, share), nil
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
	return s.Prefs(ctx, userID)
}

// SetSharePods implements PrefsStore. It touches only share_pods, so the
// advanced switch (and its NULL "follow the default") and the share link
// survive it.
func (s *SQLPrefsStore) SetSharePods(ctx context.Context, userID string, on bool, now time.Time) (Prefs, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share pods: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var n int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, userID).Scan(&n); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share pods: %w", err)
	}
	if n == 0 {
		return Prefs{}, ErrNoUser
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO stats_prefs (user_id, share_pods, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET share_pods = excluded.share_pods, updated_at = excluded.updated_at`,
		userID, boolInt(on), now.UnixMilli()); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share pods: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share pods: %w", err)
	}
	return s.Prefs(ctx, userID)
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
		`SELECT user_id, advanced, advanced_set_by, updated_at, share_link_at FROM stats_prefs
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
			linkAt  sql.NullInt64
		)
		if err := rows.Scan(&a.UserID, &adv, &by, &updated, &linkAt); err != nil {
			return nil, fmt.Errorf("gamestats: accounts: %w", err)
		}
		a.Prefs = scanPrefs(adv, by, updated, linkAt, 0) // an admin listing never reads share_pods
		out = append(out, a)
	}
	return out, rows.Err()
}

// SetShareLink implements PrefsStore. It touches only the share-link
// columns, so the advanced switch survives a new link.
func (s *SQLPrefsStore) SetShareLink(ctx context.Context, userID, hash string, now time.Time) (Prefs, error) {
	if hash == "" {
		return Prefs{}, errors.New("gamestats: set share link: empty hash")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share link: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var n int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, userID).Scan(&n); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share link: %w", err)
	}
	if n == 0 {
		return Prefs{}, ErrNoUser
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO stats_prefs (user_id, share_link_hash, share_link_at, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET share_link_hash = excluded.share_link_hash,
		   share_link_at = excluded.share_link_at, updated_at = excluded.updated_at`,
		userID, hash, now.UnixMilli(), now.UnixMilli()); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share link: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Prefs{}, fmt.Errorf("gamestats: set share link: %w", err)
	}
	return s.Prefs(ctx, userID)
}

// ClearShareLink implements PrefsStore.
func (s *SQLPrefsStore) ClearShareLink(ctx context.Context, userID string, now time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`UPDATE stats_prefs SET share_link_hash = NULL, share_link_at = NULL, updated_at = ?
		  WHERE user_id = ? AND share_link_hash IS NOT NULL`, now.UnixMilli(), userID); err != nil {
		return fmt.Errorf("gamestats: clear share link: %w", err)
	}
	return nil
}

// ShareLinkOwner implements PrefsStore. A deleted account's row is gone
// with the account (ADR 0149); the deleted_at check covers a tombstone
// all the same.
func (s *SQLPrefsStore) ShareLinkOwner(ctx context.Context, hash string) (string, error) {
	if hash == "" {
		return "", ErrNoShareLink
	}
	var id string
	err := s.db.QueryRowContext(ctx,
		`SELECT p.user_id FROM stats_prefs p JOIN users u ON u.id = p.user_id
		  WHERE p.share_link_hash = ? AND u.deleted_at IS NULL`, hash).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNoShareLink
	}
	if err != nil {
		return "", fmt.Errorf("gamestats: share link owner: %w", err)
	}
	return id, nil
}
