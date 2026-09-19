package decklibrary

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

// SQLStore is the decks table from migration 0005. Hand-written SQL
// over database/sql, as internal/db asks. Every *_at column is Unix
// milliseconds.
type SQLStore struct {
	db  *sql.DB
	now func() time.Time
}

// NewSQLStore wraps an opened, migrated database. The caller owns d
// and closes it.
func NewSQLStore(d *db.DB) *SQLStore {
	return &SQLStore{db: d.DB, now: func() time.Time { return time.Now().UTC() }}
}

// Upsert implements Store. See the interface for the update rule.
//
// Runs in one transaction: find an existing row for (owner, name),
// then either UPDATE it in place or INSERT a fresh one. Two saves for
// the same owner+name racing each other (a double-click) can both
// find no row and both try to insert; the loser gets a constraint
// failure or SQLITE_BUSY and is retried once, at which point it finds
// the winner's row and takes the update path.
func (s *SQLStore) Upsert(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText string, commanders []string, cardCount int) (Deck, error) {
	if owner == uuid.Nil {
		return Deck{}, errors.New("decklibrary: owner is required")
	}
	if name == "" {
		return Deck{}, errors.New("decklibrary: name is required")
	}
	d, err := s.upsert(ctx, owner, name, sourceFormat, sourceText, commanders, cardCount)
	if err != nil && (isConstraint(err) || isBusy(err)) {
		d, err = s.upsert(ctx, owner, name, sourceFormat, sourceText, commanders, cardCount)
	}
	return d, err
}

func (s *SQLStore) upsert(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText string, commanders []string, cardCount int) (Deck, error) {
	commandersJSON, err := json.Marshal(commanders)
	if err != nil {
		return Deck{}, fmt.Errorf("decklibrary: marshal commanders: %w", err)
	}
	now := s.now().Truncate(time.Millisecond)
	nowMs := now.UnixMilli()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Deck{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var id string
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM decks WHERE owner_id = ? AND name = ?`, owner.String(), name).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		id = uuid.New().String()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, owner.String(), name, sourceFormat, sourceText, string(commandersJSON), cardCount, nowMs, nowMs); err != nil {
			return Deck{}, fmt.Errorf("decklibrary: insert: %w", err)
		}
	case err != nil:
		return Deck{}, fmt.Errorf("decklibrary: find existing: %w", err)
	default:
		if _, err := tx.ExecContext(ctx,
			`UPDATE decks SET source_format = ?, source_text = ?, commanders = ?, card_count = ?, updated_at = ? WHERE id = ?`,
			sourceFormat, sourceText, string(commandersJSON), cardCount, nowMs, id); err != nil {
			return Deck{}, fmt.Errorf("decklibrary: update: %w", err)
		}
	}

	deck, err := getDeck(ctx, tx, id)
	if err != nil {
		return Deck{}, err
	}
	if err := tx.Commit(); err != nil {
		return Deck{}, err
	}
	return deck, nil
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, id uuid.UUID) (Deck, error) {
	return getDeck(ctx, s.db, id.String())
}

// List implements Store.
func (s *SQLStore) List(ctx context.Context, owner uuid.UUID) ([]Deck, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at
		 FROM decks WHERE owner_id = ? ORDER BY updated_at DESC, id`, owner.String())
	if err != nil {
		return nil, fmt.Errorf("decklibrary: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Deck
	for rows.Next() {
		d, err := scanDeck(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("decklibrary: list: %w", err)
	}
	return out, nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getDeck(ctx context.Context, q queryRower, id string) (Deck, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at
		 FROM decks WHERE id = ?`, id)
	d, err := scanDeck(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Deck{}, ErrNotFound
	}
	if err != nil {
		return Deck{}, fmt.Errorf("decklibrary: get %s: %w", id, err)
	}
	return d, nil
}

// rowScanner is satisfied by *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanDeck(row rowScanner) (Deck, error) {
	var (
		d                Deck
		idStr, ownerStr  string
		commandersJSON   string
		created, updated int64
	)
	if err := row.Scan(&idStr, &ownerStr, &d.Name, &d.SourceFormat, &d.SourceText, &commandersJSON, &d.CardCount, &created, &updated); err != nil {
		return Deck{}, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return Deck{}, fmt.Errorf("decklibrary: stored id %q: %w", idStr, err)
	}
	owner, err := uuid.Parse(ownerStr)
	if err != nil {
		return Deck{}, fmt.Errorf("decklibrary: stored owner_id %q: %w", ownerStr, err)
	}
	d.ID, d.OwnerID = id, owner
	if err := json.Unmarshal([]byte(commandersJSON), &d.Commanders); err != nil {
		return Deck{}, fmt.Errorf("decklibrary: stored commanders %q: %w", commandersJSON, err)
	}
	d.CreatedAt = time.UnixMilli(created).UTC()
	d.UpdatedAt = time.UnixMilli(updated).UTC()
	return d, nil
}

// isConstraint reports a SQLite constraint violation. modernc surfaces
// these as an error whose text carries "constraint failed"; there is
// no exported sentinel to test against.
func isConstraint(err error) bool {
	return err != nil && strings.Contains(err.Error(), "constraint failed")
}

// isBusy reports SQLITE_BUSY (including the WAL snapshot variant a
// read transaction gets when it tries to write after another
// connection already has).
func isBusy(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "database is locked"))
}

var _ Store = (*SQLStore)(nil)
