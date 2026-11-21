package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stromkirk Bloodthief — Creature — Vampire Rogue {2}{B}, 2/2
// (Foundations, #4465):
//
//	"At the beginning of your end step, if an opponent lost life this
//	 turn, put a +1/+1 counter on target Vampire you control."
//
// Lion Vulture's intervening "if" (CR 603.4), read off the per-turn
// life-loss tally of every opponent as the end step begins and again as
// the ability resolves, with a target in place of "this creature". The
// target is any Vampire you control, the Bloodthief included; "Vampire"
// is read through the card's text (CR 612.2). With no Vampire to target
// the ability is removed from the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c85987b4-090e-472e-b9af-1114cce56ffb",
		Name:         "Stromkirk Bloodthief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(condAnOpponentLostLifeThisTurn,
				effects.Targeting(
					effects.AtYourEndStep("Stromkirk Bloodthief — put a +1/+1 counter on target Vampire you control", effects.PlusOneCounterOnTheTarget),
					effects.TargetPermanent("target Vampire you control", effects.YouControl(), effects.Subtype("Vampire")))),
		},
	})
}

// condAnOpponentLostLifeThisTurn is "if an opponent lost life this turn",
// both halves of the intervening "if" (CR 603.4).
var condAnOpponentLostLifeThisTurn = effects.VentureCond(func(g *game.Game, you uuid.UUID) bool {
	return effects.OpponentLifeLostThisTurn(g, you) > 0
})
