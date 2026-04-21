// Package effects is the S14 card-effect catalog: declarative Forge-
// style specifications for the ~30 starter Commander cards that
// resolve automatically on the stack. Entries are keyed by Scryfall
// ID, registered at package init() time in one-file-per-card under
// this directory.
//
// Design pillars (pinned in docs/decisions/0010-card-effect-catalog.md):
//   - Declarative DSL in Go structs, no oracle-text parsing.
//   - Opt-in per Scryfall ID — non-catalog cards keep today's manual
//     sandbox posture.
//   - Effects run under the resolution write lock; primitives call
//     *Locked helpers on game.Game only. A primitive that calls a
//     public locking mutator deadlocks — enforce via review, Go can't
//     encode "locked context" at the type level.
//   - Events emitted per primitive application (see game.EmitEvent).
//
// This file declares the Spec type. Registry, Context, and primitive
// types live in neighbouring files.
package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spec is the canonical declaration for one card in the catalog.
// Populated once at package init() time via Register; looked up at
// resolution time via Lookup. Absence in the registry ⇒ non-catalog
// ⇒ today's manual sandbox behaviour.
type Spec struct {
	// OracleID is the Scryfall oracle-level card identifier — stable
	// across printings (every printing of Lightning Bolt shares one
	// oracle_id). Required. The registry panics if two Specs collide
	// on this key so a copy-paste duplicate fails at server boot
	// (the earliest, most visible failure mode).
	OracleID string

	// Name is the human-readable card name. Present in the Spec so
	// logs / panics / tests have a nice handle without bouncing
	// through the Scryfall index.
	Name string

	// OnResolve runs when the spell's stack item resolves (after
	// target re-check, before zone routing). Nil means "no effect
	// on resolve" — the card still routes to the battlefield /
	// graveyard per type, but no automation fires. Typical usage
	// for vanilla creatures with no ETB triggers (Birds of Paradise
	// in S14).
	//
	// Returning an error emits an EffectError event and allows
	// resolution to continue — partial failure does not wedge the
	// stack.
	OnResolve func(item *game.StackItem, ctx *Context) error

	// OnETB fires after a permanent moves from the stack (or
	// battlefield source) into the battlefield. S14 direct-call
	// path; S19 re-routes through the listener registry without
	// per-card changes. Nil is common (most catalog cards have no
	// ETB effect).
	OnETB func(card *game.Card, ctx *Context) error

	// StartingLoyalty is the loyalty counter count a planeswalker
	// enters the battlefield with. 0 means "not a planeswalker" or
	// "planeswalker with 0 starting loyalty" (which the SBA would
	// immediately kill — never the real case). The resolution path
	// stamps this via AddCounter when the card crosses into the
	// battlefield. Used by The Wandering Emperor in S14.
	StartingLoyalty int
}

// ZeroUUID is an alias for uuid.Nil. Mostly used in tests to
// distinguish "no target / self" fields from "uninitialised" —
// callers don't need this but having it available keeps the
// effects package self-contained for tests that don't import uuid.
var ZeroUUID = uuid.Nil
