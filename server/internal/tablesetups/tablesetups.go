// Package tablesetups remembers the last setup each signed-in person
// started a table with (ADR 0110 section 5, migration 0008's
// table_setups): the table settings, every bot seat and the other
// signed-in humans who sat there.
//
// The body is built by the server from the game it starts, never
// accepted from a client, so there is no client JSON to validate. The
// table settings are carried as raw JSON (a complete SettingsPatch)
// so this package depends on nothing from the game engine.
//
// One row per person. game_id names the table it was captured from
// and is deliberately not a foreign key: deleting a game keeps the
// setup.
package tablesetups

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// ErrNotFound: the user has no remembered setup.
var ErrNotFound = errors.New("tablesetups: not found")

// ErrNoStore is what NoStore's Put answers.
var ErrNoStore = errors.New("tablesetups: no store configured")

// Bot is one bot seat in a setup.
type Bot struct {
	Tier   string `json:"tier"`
	DeckID string `json:"deck_id"`
	Name   string `json:"name"`
}

// Setup is the body of one table_setups row.
type Setup struct {
	// Settings is the table's settings as a complete SettingsPatch.
	Settings json.RawMessage `json:"settings"`
	// Bots lists every bot seat, in seat order.
	Bots []Bot `json:"bots"`
	// Tablemates are the user IDs of the other signed-in humans who
	// sat at the table.
	Tablemates []uuid.UUID `json:"tablemates"`
}

// Record is one stored setup.
type Record struct {
	Setup Setup
	// GameID is the table it was captured from, or uuid.Nil.
	GameID    uuid.UUID
	UpdatedAt time.Time
}

// Store reads and writes a person's remembered setup. Safe for
// concurrent use.
type Store interface {
	// Get reads the user's setup. ErrNotFound if there is none.
	Get(ctx context.Context, user uuid.UUID) (Record, error)
	// Put replaces the user's setup, creating the row if needed.
	// gameID may be uuid.Nil.
	Put(ctx context.Context, user uuid.UUID, gameID uuid.UUID, setup Setup) error
}

// NoStore is the Store for a deployment with no database.
type NoStore struct{}

// Get always reports ErrNotFound.
func (NoStore) Get(context.Context, uuid.UUID) (Record, error) { return Record{}, ErrNotFound }

// Put always fails: there is nowhere to put the row.
func (NoStore) Put(context.Context, uuid.UUID, uuid.UUID, Setup) error { return ErrNoStore }

var _ Store = NoStore{}

// SQLStore is the table_setups table from migration 0008. Every *_at
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
func (s *SQLStore) Get(ctx context.Context, user uuid.UUID) (Record, error) {
	var (
		body    string
		gameID  sql.NullString
		updated int64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT body, game_id, updated_at FROM table_setups WHERE user_id = ?`, user.String()).
		Scan(&body, &gameID, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	if err != nil {
		return Record{}, fmt.Errorf("tablesetups: get: %w", err)
	}
	var rec Record
	if err := json.Unmarshal([]byte(body), &rec.Setup); err != nil {
		return Record{}, fmt.Errorf("tablesetups: stored body: %w", err)
	}
	if gameID.Valid {
		if id, err := uuid.Parse(gameID.String); err == nil {
			rec.GameID = id
		}
	}
	rec.UpdatedAt = time.UnixMilli(updated).UTC()
	return rec, nil
}

// Put implements Store.
func (s *SQLStore) Put(ctx context.Context, user uuid.UUID, gameID uuid.UUID, setup Setup) error {
	if user == uuid.Nil {
		return errors.New("tablesetups: user is required")
	}
	// Normalise nils so the stored JSON is always arrays and an
	// object, never null.
	if setup.Bots == nil {
		setup.Bots = []Bot{}
	}
	if setup.Tablemates == nil {
		setup.Tablemates = []uuid.UUID{}
	}
	if len(setup.Settings) == 0 {
		setup.Settings = json.RawMessage(`{}`)
	}
	body, err := json.Marshal(setup)
	if err != nil {
		return fmt.Errorf("tablesetups: marshal: %w", err)
	}
	var game any
	if gameID != uuid.Nil {
		game = gameID.String()
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO table_setups (user_id, body, game_id, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET body = excluded.body, game_id = excluded.game_id, updated_at = excluded.updated_at`,
		user.String(), string(body), game, s.now().Truncate(time.Millisecond).UnixMilli())
	if err != nil {
		return fmt.Errorf("tablesetups: put: %w", err)
	}
	return nil
}

var _ Store = (*SQLStore)(nil)
