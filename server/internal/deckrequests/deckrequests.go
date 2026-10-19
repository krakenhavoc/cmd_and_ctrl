// Package deckrequests is the store behind ADR 0095's deck requests
// (§3): which GitHub issue currently tracks each deck, and who asked
// for which deck when. The filing itself — the issue, the comment, the
// rules for when to do which — lives in the lobby's handler; this
// package only remembers.
//
// Two Store implementations, the decklibrary pattern:
//
//   - SQLStore, over internal/db (migrations 0006 and 0028), whenever
//     CMDCTRL_DATA_DIR is set.
//   - NoStore, for a deployment with no database. The route refuses to
//     file without a store, because the rate limit and the
//     deduplication both live in it.
package deckrequests

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned by Lookup for a deck nobody has requested.
var ErrNotFound = errors.New("deckrequests: not found")

// ErrNoStore is returned by every NoStore write.
var ErrNoStore = errors.New("deckrequests: no store configured")

// Request is one deck_requests row: the issue currently tracking a
// deck.
type Request struct {
	DeckKey     string // "moxfield:<id>" | "archidekt:<id>"
	IssueNumber int
	IssueURL    string
	// CreatedAt is when THIS issue was filed. A repointed row takes
	// the new issue's time, so "asked since CreatedAt" means "asked on
	// the current issue".
	CreatedAt time.Time
}

// Card is one card a deck request listed: one deck_request_cards row.
type Card struct {
	// Key is the card's oracle id, or "name:<lower-case name>" when the
	// card index could not resolve it. It is how two reports of one
	// deck are compared.
	Key string
	// Name is the card's name as the issue printed it.
	Name string
	// IssueNumber is the GitHub issue that listed the card.
	IssueNumber int
}

// Store reads and writes deck requests. Implementations are safe for
// concurrent use.
type Store interface {
	// Lookup returns the deck's row, or ErrNotFound.
	Lookup(ctx context.Context, deckKey string) (Request, error)
	// Put inserts the deck's row, or repoints an existing one at a new
	// issue.
	Put(ctx context.Context, r Request) error
	// AsksSince returns requester's ask times at or after since, oldest
	// first — the rolling-window rate limit.
	AsksSince(ctx context.Context, requester string, since time.Time) ([]time.Time, error)
	// HasAsked reports whether requester has asked for deckKey at or
	// after since.
	HasAsked(ctx context.Context, deckKey, requester string, since time.Time) (bool, error)
	// RecordAsk remembers one ask.
	RecordAsk(ctx context.Context, deckKey, requester string, at time.Time) error
	// RecordCards remembers that issueNumber lists cards for deckKey.
	// A card the deck already has a row for is left as it was, so the
	// issue that first listed it keeps the credit.
	RecordCards(ctx context.Context, deckKey string, issueNumber int, cards []Card) error
	// CardsOf returns the cards recorded for deckKey, in a stable order.
	CardsOf(ctx context.Context, deckKey string) ([]Card, error)
	// CopyCards gives toKey every card fromKey has a row for that toKey
	// has not. A deck whose key changes in an update inherits its
	// history this way.
	CopyCards(ctx context.Context, fromKey, toKey string) error
}

// NoStore is the Store for a deployment with no database.
type NoStore struct{}

// Lookup always reports ErrNotFound.
func (NoStore) Lookup(context.Context, string) (Request, error) { return Request{}, ErrNotFound }

// Put always fails.
func (NoStore) Put(context.Context, Request) error { return ErrNoStore }

// AsksSince always returns no asks.
func (NoStore) AsksSince(context.Context, string, time.Time) ([]time.Time, error) { return nil, nil }

// HasAsked always reports false.
func (NoStore) HasAsked(context.Context, string, string, time.Time) (bool, error) { return false, nil }

// RecordAsk always fails.
func (NoStore) RecordAsk(context.Context, string, string, time.Time) error { return ErrNoStore }

// RecordCards always fails.
func (NoStore) RecordCards(context.Context, string, int, []Card) error { return ErrNoStore }

// CardsOf always returns no cards.
func (NoStore) CardsOf(context.Context, string) ([]Card, error) { return nil, nil }

// CopyCards always fails.
func (NoStore) CopyCards(context.Context, string, string) error { return ErrNoStore }

var _ Store = NoStore{}
