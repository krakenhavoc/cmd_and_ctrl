package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restrictions.go — the card-facing primitives for S24's restriction
// vocabulary: "enchanted creature can't attack or block" (Pacifism,
// Arrest, Faith's Fetters), "equipped creature can't be blocked"
// (Whispersilk Cloak), "this creature can't block" (Carrion Feeder),
// "target creature can't be blocked this turn" (Rogue's Passage).
//
// The bits and the engine gates that read them live in
// server/internal/game/restrictions.go; that file's doc comment is
// the taxonomy. This one is three one-liners over the same
// `game.StaticAbility` every other static uses, in the three
// durations a card can want:
//
//	RestrictAttached(…)   for as long as the Aura / Equipment is on
//	                      its host        (Pacifism, the Cloak)
//	RestrictSelf(…)       for as long as the permanent is on the
//	                      battlefield     (Carrion Feeder)
//	RestrictUntilEOT{…}   until end of turn (Rogue's Passage)
//
// WHICH LAYER. Layer 6, and the choice does not matter. A
// restriction is not applied in any CR 613 layer — it is not a
// characteristic — but the layer pass is the only machinery that
// knows which permanents a continuous effect applies to, so the
// restriction rides through it. Because `Characteristic.Restrictions`
// is only ever OR'd into and nothing clears it, timestamp order
// between two restriction effects is unobservable, and a layer-6
// "loses all abilities" cannot strip a Pacifism. That last part is
// the rules-correct outcome, not a happy accident: the "can't
// attack" belongs to the Aura, not to the creature.

// RestrictAttached is "Enchanted creature can't attack or block" /
// "Equipped creature can't be blocked" — a restriction scoped to
// whatever the source is attached to, by the same AttachedToSource
// predicate every other attachment static uses.
//
// The restriction ends the moment the attachment does, with no
// bookkeeping: the AppliesTo re-reads the relation on every
// recompute, so destroying the Pacifism un-pacifies the creature in
// the same pass.
func RestrictAttached(r game.Restriction) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= r
		},
	}
}

// RestrictSelf is a restriction a permanent prints on itself —
// Carrion Feeder's "This creature can't block". Scoped by selfOnly,
// the same predicate a self-referential pump uses.
func RestrictSelf(r game.Restriction) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= r
		},
	}
}

// RestrictUntilEOT is "target creature can't be blocked this turn"
// (Rogue's Passage) or a mass form ("creatures your opponents
// control can't block this turn"). The turn-scoped sibling of
// RestrictAttached, registered into the S32 until-end-of-turn
// registry and swept at cleanup (CR 514.2).
//
// The affected set is snapshotted at resolution (CR 611.2c), exactly
// as BoostUntilEOT and GrantKeywordUntilEOT do, so a creature that
// enters after the ability resolves is not covered — and a creature
// that leaves and returns is a new object (CR 400.7) and is not
// covered either, which the entry-stamp half of the snapshot key
// enforces.
type RestrictUntilEOT struct {
	// Target pins the effect to one permanent. Ignored when Match
	// is set.
	Target uuid.UUID

	// Match selects the affected permanents, evaluated ONCE at
	// resolution (CR 611.2c).
	Match CardPredicate

	// Restrictions is the bit set to add. game.CantAttackOrBlock is
	// the common pair.
	Restrictions game.Restriction

	// Label is attribution for logs and tests; defaults to a
	// generic string when empty.
	Label string
}

func (r RestrictUntilEOT) Apply(ctx *Context) error {
	if r.Restrictions == 0 {
		return nil
	}
	set := eotSnapshot(ctx, r.Target, r.Match)
	if set == nil {
		return nil
	}
	bits := r.Restrictions
	ctx.Game.RegisterScopedStaticForEffect(game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: set.appliesTo(),
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= bits
		},
	}, ctx.Source(), eotLabel(r.Label, "restriction until end of turn"),
		ctx.Game.UntilEndOfTurnDuration())
	return nil
}
