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

// ErrNameTaken is returned by Rename when the owner already has a deck
// with that name: the upsert rule keys on (owner, name).
var ErrNameTaken = errors.New("decklibrary: you already have a deck with that name")

// ErrLibraryFull is returned by Upsert when it would insert a deck past
// MaxDecks. Updating an existing deck is never refused.
var ErrLibraryFull = errors.New("decklibrary: your deck library is full")

// MaxDecks is the most decks one person may keep (ADR 0110 section 6).
const MaxDecks = 200

// Deck is one decks row.
type Deck struct {
	ID           uuid.UUID
	OwnerID      uuid.UUID
	Name         string
	SourceFormat string // "moxfield" | "text"
	SourceText   string // what the player pasted, or what was fetched from SourceURL
	// SourceURL is the link a deck was imported from, "" for a pasted
	// one (ADR 0110 owner decision 7). SourceText is the list as it was
	// fetched, so re-seating never calls the network.
	SourceURL  string
	Commanders []string
	CardCount  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Store reads and writes a person's deck library. Implementations are
// safe for concurrent use.
type Store interface {
	// Upsert saves one deck for owner. The update rule (see ADR 0051
	// implementation notes, sub-PR 5): a row already owned by owner
	// with the same name is updated in place — source, commanders and
	// card_count replaced, updated_at bumped, id unchanged. Anything
	// else (a new name, or no existing row) inserts a new deck.
	//
	// A pasted deck: source_url is cleared, so a deck re-saved from
	// text no longer claims to come from a link. An insert that would
	// leave the owner with more than MaxDecks is ErrLibraryFull.
	Upsert(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText string, commanders []string, cardCount int) (Deck, error)
	// UpsertFromLink is Upsert for a deck imported from a link: the
	// list as fetched, with the link beside it.
	UpsertFromLink(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText, sourceURL string, commanders []string, cardCount int) (Deck, error)
	// Save is UpsertFromLink that also reports which way it went:
	// replaced is true when owner already had a deck by that name and
	// it was updated in place, false when a new deck was inserted. An
	// empty sourceURL clears the link, as Upsert does. POST /me/decks
	// answers {deck, replaced} with it (ADR 0112 §3 item 4).
	Save(ctx context.Context, owner uuid.UUID, name, sourceFormat, sourceText, sourceURL string, commanders []string, cardCount int) (d Deck, replaced bool, err error)
	// Count is how many decks owner has.
	Count(ctx context.Context, owner uuid.UUID) (int, error)
	// Delete removes owner's deck. In the same transaction it sets
	// seats.deck_id to NULL wherever it pointed at the deck (the seat
	// keeps its deck_name). ErrNotFound if there is no such deck or it
	// is someone else's, so an id reveals nothing.
	Delete(ctx context.Context, owner, id uuid.UUID) error
	// Rename renames owner's deck without touching its updated_at.
	// ErrNotFound as for Delete; ErrNameTaken if owner has another deck
	// with that name (renaming to its own name is a no-op).
	Rename(ctx context.Context, owner, id uuid.UUID, name string) (Deck, error)
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

// UpsertFromLink always fails, like Upsert.
func (NoStore) UpsertFromLink(context.Context, uuid.UUID, string, string, string, string, []string, int) (Deck, error) {
	return Deck{}, errors.New("decklibrary: no store configured")
}

// Save always fails, like Upsert.
func (NoStore) Save(context.Context, uuid.UUID, string, string, string, string, []string, int) (Deck, bool, error) {
	return Deck{}, false, errors.New("decklibrary: no store configured")
}

// Count is always 0.
func (NoStore) Count(context.Context, uuid.UUID) (int, error) { return 0, nil }

// Delete always reports ErrNotFound: there are no decks.
func (NoStore) Delete(context.Context, uuid.UUID, uuid.UUID) error { return ErrNotFound }

// Rename always reports ErrNotFound: there are no decks.
func (NoStore) Rename(context.Context, uuid.UUID, uuid.UUID, string) (Deck, error) {
	return Deck{}, ErrNotFound
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
