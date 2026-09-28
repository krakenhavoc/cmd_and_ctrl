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

	// Indefinite is a copy with no stated duration (CR 611.2a): Unstable
	// Shapeshifter, Lazav, Dimir Doppelganger. False is "until end of
	// turn", which is every other printed card of the class.
	Indefinite bool

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
	d := ctx.Game.UntilEndOfTurnDuration()
	if b.Indefinite {
		d = game.IndefiniteDuration()
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
