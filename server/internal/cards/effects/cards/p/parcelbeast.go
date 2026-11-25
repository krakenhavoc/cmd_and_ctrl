package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parcelbeast — Creature — Elemental Beast {2}{G}{U}, 2/4 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {G}{U} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 {1}, {T}: Look at the top card of your library. If it's a land card, you may
//	 put it onto the battlefield. If you don't put the card onto the
//	 battlefield, put it into your hand."
//
// Mutate is the engine's (ADR 0223); the ability is Lantern of Revealing's look
// with the printed ending: a land you choose to put onto the battlefield
// enters untapped, and every other outcome, a nonland card or a land you
// decline, puts the card into your hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "de83b309-ac31-4a57-95e6-ccd3659e6c1f",
		Name:             "Parcelbeast",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{G}{U}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Look at the top card of your library. If it's a land card, you may put it onto the battlefield. If you don't put the card onto the battlefield, put it into your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.S82OldB112LookAtTopLandToBattlefieldElseHand,
		}},
	})
}
