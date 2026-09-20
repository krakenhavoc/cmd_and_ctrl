// Package decklibrary is the deck half of the persistent store (ADR
// 0051 decision 7, S34 sub-PR 5): a decks row is the text a signed-in
// player pasted, kept so it can be re-seated without re-pasting.
//
// Source text, not the parsed card list. The catalog changes weekly
// and a deck stored as resolved cards would drift silently as oracle
// ids move between Scryfall dumps; a deck stored as the pasted text is
// re-parsed through the same parse -> resolve -> validate pipeline an
// upload takes, at the moment it is seated, and a card that no longer
// resolves surfaces as an ordinary validation error rather than a
// silent substitution.
//
// Two Store implementations:
//
//   - SQLStore, over internal/db, whenever CMDCTRL_DATA_DIR is set.
//   - NoStore, when it is not (or a deployment has no database). A
//     principal can never carry a non-zero UserID under NoStore
//     either — see package users — so nothing in the HTTP layer calls
//     these except defensively; NoStore exists so that defensiveness
//     compiles and behaves honestly rather than panicking.
package decklibrary

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned by Get for an unknown deck.
var ErrNotFound = errors.New("decklibrary: not found")

// Deck is one decks row.
type Deck struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	Name         string
	SourceFormat string // "moxfield" | "text"
	SourceText   string // what the player pasted
	Commanders   []string
	CardCount    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Store reads and writes a person's deck library. Implementations are
// safe for concurrent use.
type Store interface {
	// Upsert saves one deck for owner. The update rule (see ADR 0051
	// implementation notes, sub-PR 5): a row already owned by owner
	// with the same name is updated in place — source, commanders and
	// card_count replaced, updated_at bumped, id unchanged. Anything
	// else (a new name, or no existing row) inserts a new deck.
	Upsert(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText string, commanders []string, cardCount int) (Deck, error)
	// Get reads one deck by id. ErrNotFound if there is none.
	Get(ctx context.Context, id uuid.UUID) (Deck, error)
	// List returns owner's decks, most recently updated first.
	List(ctx context.Context, owner uuid.UUID) ([]Deck, error)
}

// NoStore is the Store for a deployment with no database. Every write
// is refused and every read comes back empty, exactly like a person
// with no decks — because under NoStore no principal ever carries a
// non-zero UserID to own one.
type NoStore struct{}

// Upsert always fails: there is nowhere to put the row. Not called in
// practice — see the package doc comment — but a definite error is
// safer than silently discarding a deck the player thinks was saved.
func (NoStore) Upsert(context.Context, uuid.UUID, string, string, string, []string, int) (Deck, error) {
	return Deck{}, errors.New("decklibrary: no store configured")
}

// Get always reports ErrNotFound.
func (NoStore) Get(context.Context, uuid.UUID) (Deck, error) {
	return Deck{}, ErrNotFound
}

// List always returns no decks.
func (NoStore) List(context.Context, uuid.UUID) ([]Deck, error) {
	return nil, nil
}

var _ Store = NoStore{}
