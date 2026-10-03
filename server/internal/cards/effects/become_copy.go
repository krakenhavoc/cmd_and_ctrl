package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// become_copy.go — "<permanent> becomes a copy of <object> [until end of
// turn]" (#1593, ADR 0043's amendment of 2026-09-28): a CR 707.2 copy
// effect with a duration, created by a resolving spell or ability
// (CR 611.2) on a permanent that is already on the battlefield.
//
// The engine half is server/internal/game/duration_copy.go. This file is
// the catalog-facing primitive, and the difference from EntersAsCopyOf
// (copy_effects.go) is the whole point: an entry copy is settled once as
// the permanent enters and lasts as long as it does, while this one
// lands on a permanent already in play, has a timestamp of its own, and
// ENDS — Mirage Mirror is Mirage Mirror again at cleanup, and a Clone
// that Cytoshape turned into a Hill Giant is the creature it cloned.
//
// # Amendment (2026-09-28, #1723): a third duration
//
// Every card that shipped before Shapesharer prints one of two
// durations — "until end of turn" or none at all (CR 611.2a,
// "indefinite" below) — so a bool sufficed. Shapesharer's "until your
// next turn" (CR 611.2b) is a third, and the two are easy to confuse:
// "until your next turn" ends as that turn BEGINS, "until end of
// turn" ends at the cleanup that turn already has. `CopyDuration` is
// the enum the bool would have needed a second one anyway; the zero
// value keeps every existing card's `BecomeCopy{...}` literal
// unchanged.

// CopyDuration selects how long a BecomeCopy effect lasts (CR 611.2).
// The zero value, CopyUntilEndOfTurn, is every printed card that
// states no duration of its own other than "until end of turn" —
// Cytoshape, Mirrorweave, Lazav Familiar Stranger, Mizzium
// Transreliquat's {3} line.
type CopyDuration int

const (
	// CopyUntilEndOfTurn is CR 611.2's plain "until end of turn" — the
	// zero value, so every card registered before this type existed
	// reads exactly as it did.
	CopyUntilEndOfTurn CopyDuration = iota

	// CopyIndefinite is a copy with no stated duration (CR 611.2a):
	// Unstable Shapeshifter, Lazav Dimir Mastermind, Dimir
	// Doppelganger, Mizzium Transreliquat's {1}{U}{R} line.
	CopyIndefinite

	// CopyUntilYourNextTurn is CR 611.2b's "until your next turn" —
	// Shapesharer. "Your" is the ability's controller, so Apply reads
	// it off ctx.Controller() rather than a field on this struct.
	CopyUntilYourNextTurn

	// CopyWhileOfRemainsTapped is "for as long as that creature remains
	// tapped" (ADR 0109 §3, Zygon Infiltrator): the copy lasts while Of,
	// the object copied, stays tapped on the battlefield as the same
	// object. The duration is pinned to Of, not to the permanent that
	// becomes the copy, and never starts if Of is untapped or gone.
	CopyWhileOfRemainsTapped
)

// BecomeCopy makes each of Targets a copy of Of.
//
// Of may be in ANY zone — Shifting Woodland and Lazav copy a card in a
// graveyard — and its copiable values are read now, as the effect
// begins (CR 707.2), and stored. The copy does not end or change when Of
// later moves.
type BecomeCopy struct {
	// Targets are the permanents that become the copy. Anything not on
	// the battlefield is skipped.
	Targets []uuid.UUID

	// Of is the object whose copiable values are taken.
	Of uuid.UUID

	// Values, when set, are used instead of reading Of — for a card
	// whose text says which values ("that card", read as last-known
	// information once it has moved; see Lazav).
	Values *game.PrintedValues

	// Duration is how long the copy lasts (CR 611.2). The zero value,
	// CopyUntilEndOfTurn, is every printed card of the class but the
	// three that state otherwise.
	Duration CopyDuration

	// Except is the card's "except" clause, applied to a private copy of
	// the values on their way in: SetName, AddSupertype, AddKeyword,
	// GrantAbility — the same edits EntersAsCopyOf's clause makes.
	Except func(v *game.PrintedValues)

	// Label names the effect in logs and tests.
	Label string
}

// Apply registers the copy. A copy of nothing — Of gone from every zone
// and no Values given — does nothing, which is CR 608.2b's "as much as
// possible" for an effect that has nothing to copy.
func (b BecomeCopy) Apply(ctx *Context) error {
	var v game.PrintedValues
	switch {
	case b.Values != nil:
		v = b.Values.Clone()
	default:
		got, ok := ctx.Game.CopiableValuesForEffect(b.Of)
		if !ok {
			return nil
		}
		v = got
	}
	// #1432: "this creature becomes a copy" after a flicker in response
	// names an object that no longer exists; the new one is left alone.
	targets := make([]uuid.UUID, 0, len(b.Targets))
	for _, t := range b.Targets {
		if !ctx.isNewSourceObject(t) {
			targets = append(targets, t)
		}
	}
	if len(targets) == 0 {
		return nil
	}
	if b.Except != nil {
		b.Except(&v)
	}
	var d game.Duration
	switch b.Duration {
	case CopyIndefinite:
		d = game.IndefiniteDuration()
	case CopyUntilYourNextTurn:
		d = ctx.Game.UntilYourNextTurnDuration(ctx.Controller())
	case CopyWhileOfRemainsTapped:
		var ok bool
		if d, ok = ctx.Game.ForAsLongAsPinnedTappedDuration(b.Of); !ok {
			return nil // CR 611.2b: it never starts
		}
	default:
		d = ctx.Game.UntilEndOfTurnDuration()
	}
	label := b.Label
	if label == "" {
		label = "becomes a copy"
	}
	ctx.Game.BecomeCopyForEffect(ctx.Source(), b.Of, targets, v, d, label)
	return nil
}

// selfBecomesCopyOfTarget is the Effect of every "{cost}: this
// permanent becomes a copy of target <X> until end of turn" ability —
// Mirage Mirror, Shifting Woodland. The target clause is the ability's
// only one, and the CR 608.2b re-check has already countered the
// ability if it is illegal.
func selfBecomesCopyOfTarget(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		// "This permanent" is the source; BecomeCopy leaves it alone if
		// it has left, or come back as a new object (#1432).
		return BecomeCopy{Targets: []uuid.UUID{ctx.Source()}, Of: targets[0].ID, Label: label}.Apply(ctx)
	}
}
