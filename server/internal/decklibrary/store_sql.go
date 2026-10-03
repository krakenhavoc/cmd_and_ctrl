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
	return s.UpsertFromLink(ctx, owner, name, sourceFormat, sourceText, "", commanders, cardCount)
}

// UpsertFromLink implements Store. See Upsert.
func (s *SQLStore) UpsertFromLink(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText, sourceURL string, commanders []string, cardCount int) (Deck, error) {
	d, _, err := s.Save(ctx, owner, name, sourceFormat, sourceText, sourceURL, commanders, cardCount)
	return d, err
}

// Save implements Store. See Upsert for the update rule and the retry.
func (s *SQLStore) Save(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText, sourceURL string, commanders []string, cardCount int) (Deck, bool, error) {
	if owner == uuid.Nil {
		return Deck{}, false, errors.New("decklibrary: owner is required")
	}
	if name == "" {
		return Deck{}, false, errors.New("decklibrary: name is required")
	}
	d, replaced, err := s.upsert(ctx, owner, name, sourceFormat, sourceText, sourceURL, commanders, cardCount)
	if err != nil && (isConstraint(err) || isBusy(err)) {
		d, replaced, err = s.upsert(ctx, owner, name, sourceFormat, sourceText, sourceURL, commanders, cardCount)
	}
	return d, replaced, err
}

func (s *SQLStore) upsert(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText, sourceURL string, commanders []string, cardCount int) (Deck, bool, error) {
	commandersJSON, err := json.Marshal(commanders)
	if err != nil {
		return Deck{}, false, fmt.Errorf("decklibrary: marshal commanders: %w", err)
	}
	nowMs := s.now().Truncate(time.Millisecond).UnixMilli()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Deck{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	// updated_at is what List orders by, and it is only millisecond
	// precise, so two saves inside one millisecond would tie and fall
	// through to a random id (#1165). Keep one owner's stamps strictly
	// increasing instead: a save never lands at or before the owner's
	// newest existing stamp. The bump is at most a millisecond past the
	// clock, and only when two saves collide.
	var newest int64
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(updated_at), 0) FROM decks WHERE owner_id = ?`, owner.String()).Scan(&newest); err != nil {
		return Deck{}, false, fmt.Errorf("decklibrary: newest stamp: %w", err)
	}
	if nowMs <= newest {
		nowMs = newest + 1
	}

	var (
		id       string
		replaced bool
	)
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM decks WHERE owner_id = ? AND name = ?`, owner.String(), name).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		var have int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM decks WHERE owner_id = ?`, owner.String()).Scan(&have); err != nil {
			return Deck{}, false, fmt.Errorf("decklibrary: count: %w", err)
		}
		if have >= MaxDecks {
			return Deck{}, false, ErrLibraryFull
		}
		id = uuid.New().String()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO decks (id, owner_id, name, source_format, source_text, source_url, commanders, card_count, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, owner.String(), name, sourceFormat, sourceText, nullString(sourceURL), string(commandersJSON), cardCount, nowMs, nowMs); err != nil {
			return Deck{}, false, fmt.Errorf("decklibrary: insert: %w", err)
		}
	case err != nil:
		return Deck{}, false, fmt.Errorf("decklibrary: find existing: %w", err)
	default:
		replaced = true
		if _, err := tx.ExecContext(ctx,
			`UPDATE decks SET source_format = ?, source_text = ?, source_url = ?, commanders = ?, card_count = ?, updated_at = ? WHERE id = ?`,
			sourceFormat, sourceText, nullString(sourceURL), string(commandersJSON), cardCount, nowMs, id); err != nil {
			return Deck{}, false, fmt.Errorf("decklibrary: update: %w", err)
		}
	}

	deck, err := getDeck(ctx, tx, id)
	if err != nil {
		return Deck{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Deck{}, false, err
	}
	return deck, replaced, nil
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, id uuid.UUID) (Deck, error) {
	return getDeck(ctx, s.db, id.String())
}

// List implements Store.
func (s *SQLStore) List(ctx context.Context, owner uuid.UUID) ([]Deck, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, owner_id, name, source_format, source_text, source_url, commanders, card_count, created_at, updated_at
		 FROM decks WHERE owner_id = ? ORDER BY updated_at DESC, rowid DESC`, owner.String())
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

// Count implements Store.
func (s *SQLStore) Count(ctx context.Context, owner uuid.UUID) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM decks WHERE owner_id = ?`, owner.String()).Scan(&n); err != nil {
		return 0, fmt.Errorf("decklibrary: count: %w", err)
	}
	return n, nil
}

// Delete implements Store. The seats update and the delete share one
// transaction. The seats.deck_id foreign key (migration 0008) is
// ON DELETE SET NULL as well, so this is belt and braces: the store
// does not depend on the schema's action to keep the promise.
func (s *SQLStore) Delete(ctx context.Context, owner, id uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var found string
	err = tx.QueryRowContext(ctx, `SELECT id FROM decks WHERE id = ? AND owner_id = ?`, id.String(), owner.String()).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("decklibrary: find %s: %w", id, err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE seats SET deck_id = NULL WHERE deck_id = ?`, id.String()); err != nil {
		return fmt.Errorf("decklibrary: detach seats from %s: %w", id, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM decks WHERE id = ? AND owner_id = ?`, id.String(), owner.String()); err != nil {
		return fmt.Errorf("decklibrary: delete %s: %w", id, err)
	}
	return tx.Commit()
}

// Rename implements Store.
func (s *SQLStore) Rename(ctx context.Context, owner, id uuid.UUID, name string) (Deck, error) {
	if name == "" {
		return Deck{}, errors.New("decklibrary: name is required")
	}
	d, err := s.rename(ctx, owner, id, name)
	if err != nil && (isConstraint(err) || isBusy(err)) {
		d, err = s.rename(ctx, owner, id, name)
	}
	return d, err
}

func (s *SQLStore) rename(ctx context.Context, owner, id uuid.UUID, name string) (Deck, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Deck{}, err
	}
	defer func() { _ = tx.Rollback() }()

	cur, err := getDeck(ctx, tx, id.String())
	if err != nil {
		return Deck{}, err
	}
	if cur.OwnerID != owner {
		return Deck{}, ErrNotFound
	}
	var other string
	err = tx.QueryRowContext(ctx, `SELECT id FROM decks WHERE owner_id = ? AND name = ? AND id <> ?`,
		owner.String(), name, id.String()).Scan(&other)
	switch {
	case err == nil:
		return Deck{}, ErrNameTaken
	case !errors.Is(err, sql.ErrNoRows):
		return Deck{}, fmt.Errorf("decklibrary: check name: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE decks SET name = ? WHERE id = ?`, name, id.String()); err != nil {
		return Deck{}, fmt.Errorf("decklibrary: rename: %w", err)
	}
	cur.Name = name
	if err := tx.Commit(); err != nil {
		return Deck{}, err
	}
	return cur, nil
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getDeck(ctx context.Context, q queryRower, id string) (Deck, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, owner_id, name, source_format, source_text, source_url, commanders, card_count, created_at, updated_at
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
		sourceURL        sql.NullString
		created, updated int64
	)
	if err := row.Scan(&idStr, &ownerStr, &d.Name, &d.SourceFormat, &d.SourceText, &sourceURL, &commandersJSON, &d.CardCount, &created, &updated); err != nil {
		return Deck{}, err
	}
	d.SourceURL = sourceURL.String
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
