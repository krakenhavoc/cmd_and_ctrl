package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_from_hand.go — "if you haven't cast a spell from your hand this
// turn" (Outlaws of Thunder Junction, #5235).
//
// One fact about the player's TURN, read off the cast tally
// (game.CastTally.FromHand): the spells the player cast this turn from
// their hand, the zone each was in as it was cast, before CR 601.2a moved
// it to the stack. A spell cast from a graveyard, exile (a plotted card),
// the command zone or the top of a library does not count, and a copy of
// a spell isn't cast (CR 707.10). It resets with the rest of the turn's
// tallies.
//
// Readers, one per place the printed text puts the condition:
//
//   - CastASpellFromHandThisTurn(g, you) for a resolution-time check or
//     a static's condition.
//   - YouHaventCastASpellFromHand is the When as the trigger fires.
//   - CondNoSpellFromHand is the intervening "if" (CR 603.4) re-checked
//     as the ability resolves, for InterveningIf and If, with a footprint
//     condition (game.FootprintCondNoSpellsCastFromHand): it can only
//     become false, by a cast, and no step casts.
//   - AtYourEndStepIfNoSpellFromHand is the printed trigger shape, "At
//     the beginning of your end step, if you haven't cast a spell from
//     your hand this turn, …".

// CastASpellFromHandThisTurn reports whether `you` cast a spell from your
// hand this turn. Caller must hold g.mu (every effect body, static
// AppliesTo and trigger condition does).
func CastASpellFromHandThisTurn(g *game.Game, you uuid.UUID) bool {
	return you != uuid.Nil && g.CastTallyFor(you).FromHand() > 0
}

// YouHaventCastASpellFromHand is the When for "if you haven't cast a
// spell from your hand this turn" as the trigger fires, for composing
// with AllOf: read for the source's controller.
func YouHaventCastASpellFromHand(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return !CastASpellFromHandThisTurn(g, source.Controller)
}

// CondNoSpellFromHand is the condition as an intervening-if Condition,
// for InterveningIf and If: read for the item's controller as it
// resolves, so a spell cast from the hand in response turns it off.
type CondNoSpellFromHand struct{}

func (CondNoSpellFromHand) Holds(ctx *Context) bool {
	return !CastASpellFromHandThisTurn(ctx.Game, ctx.Controller())
}

func (CondNoSpellFromHand) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondNoSpellsCastFromHand, 0), true
}

// AtYourEndStepIfNoSpellFromHand is "At the beginning of your end step,
// if you haven't cast a spell from your hand this turn, …": checked as
// the end step begins and again as the ability resolves (CR 603.4).
// `also` are further trigger-time conditions the printed "and" adds
// (Inventive Wingsmith's "and this creature doesn't have a flying
// counter on it"); the effect re-checks those itself.
func AtYourEndStepIfNoSpellFromHand(label string, effect Effect, also ...When) game.TriggeredAbility {
	when := AllOf(append([]When{ByYou, YouHaventCastASpellFromHand}, also...)...)
	return InterveningIf(CondNoSpellFromHand{}, On(game.EventBeginEndStep, when, label, effect))
}
