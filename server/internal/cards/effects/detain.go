package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Detain is the keyword action (CR 701.35): "Until your next turn, that
// permanent can't attack or block and its activated abilities can't be
// activated."
//
// It is a restriction record with a CR 611.2 duration, pinned to the
// permanents that matched AS THE ABILITY RESOLVED. The pin is the rule,
// not an optimisation: CR 701.35a names "that permanent", so a creature
// that arrives later is not detained, and one that is flickered in
// response is a new object (CR 400.7) the effect no longer follows.
// That is the opposite of RestrictUntilEOT{Scope}, whose "creatures
// can't block this turn" prints no object and so reads a live set
// (#1650) -- which is why this does not go through ScopedEffectFor's
// Match, and builds the pinned set itself.
//
// "Its activated abilities can't be activated" covers mana abilities
// (CantActivateMana, the Arrest reading) and loyalty abilities (they
// are activated abilities, CR 606.1). Static and triggered abilities
// are untouched, which is what the restriction bits say.
//
// The duration is the DETAINING player's next turn, not the
// permanent's controller's: "your next turn" is read off the ability's
// controller (DurationUntilYourNextTurn).
type Detain struct {
	// Match selects the permanents, evaluated once, now.
	Match CardPredicate

	// Label is attribution for logs and the wire.
	Label string
}

// detainRestrictions is everything a detained permanent loses.
const detainRestrictions = game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana

func (d Detain) Apply(ctx *Context) error {
	set := eotSnapshot(ctx, uuid.Nil, d.Match)
	if set == nil {
		return nil
	}
	ctx.Game.RegisterScopedEffectForEffect(ctx.Source(), set.affectedObjects(),
		[]game.Mod{game.AddRestrictionsMod(detainRestrictions)},
		DurationUntilYourNextTurn(ctx, ctx.Controller()),
		eotLabel(d.Label, "detain"))
	return nil
}
