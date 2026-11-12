package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// excess_damage_this_turn.go — "if a creature or planeswalker an
// opponent controlled was dealt excess damage this turn" (Rith,
// Liberated Primeval, #5663).
//
// One fact about the TURN, read off the turn tally
// (game.TurnTally.ExcessDamaged) through
// Game.OpponentPermanentDealtExcessDamageThisTurn. The engine judges
// each damage event as it lands (CR 120.10; game/excess_damage_tally.go):
// combat damage or not, from any source under anyone's control, to a
// creature past lethal damage or to a planeswalker past its loyalty. The
// permanent's controller is read as it was dealt the damage, so the fact
// stands after the permanent has died or changed hands, and whether or
// not the reader was on the battlefield then.
//
// excess_damage.go is the other reading: "whenever … is dealt excess
// noncombat damage", one event's amount inside a trigger.

// CondOpponentPermanentDealtExcessDamage is the condition as an
// intervening-if Condition, for InterveningIf and If: read for the
// item's controller as it resolves. The tally only rises within a turn,
// so it declares a footprint condition the trigger-independence check
// can argue (game.FootprintCondOpponentPermanentDealtExcessDamage).
type CondOpponentPermanentDealtExcessDamage struct{}

func (CondOpponentPermanentDealtExcessDamage) Holds(ctx *Context) bool {
	return ctx.Game.OpponentPermanentDealtExcessDamageThisTurn(ctx.Controller())
}

func (CondOpponentPermanentDealtExcessDamage) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondOpponentPermanentDealtExcessDamage, 0), true
}

// AnOpponentsPermanentWasDealtExcessDamage is the When for the same
// fact, read for the source's controller as the event happens, for
// composing with AllOf.
func AnOpponentsPermanentWasDealtExcessDamage(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.OpponentPermanentDealtExcessDamageThisTurn(source.Controller)
}

// AtYourEndStepIfOpponentPermanentDealtExcessDamage is "At the beginning
// of your end step, if a creature or planeswalker an opponent controlled
// was dealt excess damage this turn, …": the intervening "if" is checked
// as the end step begins and again as the trigger resolves (CR 603.4).
func AtYourEndStepIfOpponentPermanentDealtExcessDamage(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondOpponentPermanentDealtExcessDamage{},
		On(game.EventBeginEndStep, AllOf(ByYou, AnOpponentsPermanentWasDealtExcessDamage), label, effect))
}
