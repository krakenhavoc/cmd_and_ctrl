package deckrequests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// SQLStore is migration 0006's two tables. Hand-written SQL over
// database/sql, as internal/db asks. Every time column is Unix
// milliseconds.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore wraps an opened, migrated database. The caller owns d
// and closes it.
func NewSQLStore(d *db.DB) *SQLStore {
	return &SQLStore{db: d.DB}
}

// Lookup implements Store.
func (s *SQLStore) Lookup(ctx context.Context, deckKey string) (Request, error) {
	var (
		r  = Request{DeckKey: deckKey}
		ms int64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT issue_number, issue_url, created_at FROM deck_requests WHERE deck_key = ?`, deckKey).
		Scan(&r.IssueNumber, &r.IssueURL, &ms)
	if errors.Is(err, sql.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("deckrequests: lookup: %w", err)
	}
	r.CreatedAt = time.UnixMilli(ms).UTC()
	return r, nil
}

// Put implements Store.
func (s *SQLStore) Put(ctx context.Context, r Request) error {
	if r.DeckKey == "" || r.IssueNumber <= 0 {
		return errors.New("deckrequests: a request needs a deck key and an issue number")
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO deck_requests (deck_key, issue_number, issue_url, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT (deck_key) DO UPDATE SET
		     issue_number = excluded.issue_number,
		     issue_url    = excluded.issue_url,
		     created_at   = excluded.created_at`,
		r.DeckKey, r.IssueNumber, r.IssueURL, r.CreatedAt.UnixMilli()); err != nil {
		return fmt.Errorf("deckrequests: put: %w", err)
	}
	return nil
}

// AsksSince implements Store.
func (s *SQLStore) AsksSince(ctx context.Context, requester string, since time.Time) ([]time.Time, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT at FROM deck_request_asks WHERE requester = ? AND at >= ? ORDER BY at`,
		requester, since.UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("deckrequests: asks: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []time.Time
	for rows.Next() {
		var ms int64
		if err := rows.Scan(&ms); err != nil {
			return nil, fmt.Errorf("deckrequests: scan ask: %w", err)
		}
		out = append(out, time.UnixMilli(ms).UTC())
	}
	return out, rows.Err()
}

// HasAsked implements Store.
func (s *SQLStore) HasAsked(ctx context.Context, deckKey, requester string, since time.Time) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM deck_request_asks WHERE deck_key = ? AND requester = ? AND at >= ? LIMIT 1`,
		deckKey, requester, since.UnixMilli()).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("deckrequests: has asked: %w", err)
	}
	return true, nil
}

// RecordAsk implements Store.
func (s *SQLStore) RecordAsk(ctx context.Context, deckKey, requester string, at time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO deck_request_asks (deck_key, requester, at) VALUES (?, ?, ?)`,
		deckKey, requester, at.UnixMilli()); err != nil {
		return fmt.Errorf("deckrequests: record ask: %w", err)
	}
	return nil
}

var _ Store = (*SQLStore)(nil)
