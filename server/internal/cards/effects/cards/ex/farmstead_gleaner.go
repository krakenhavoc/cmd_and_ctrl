package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farmstead Gleaner — Artifact Creature — Scarecrow {3}, 2/2 (Modern Horizons):
//
//	"This creature doesn't untap during your untap step.
//	 {2}, {Q}: Put a +1/+1 counter on this creature. ({Q} is the untap
//	 symbol.)"
//
// Battered Golem's untap-step restriction. {Q} is AbilityCost.UntapSelf
// (CR 107.6): the Gleaner must be tapped to pay it, and a creature that has
// not been under its controller's control since their turn began cannot (CR
// 302.6). Paying it is how the Gleaner untaps. The counter goes on the
// Gleaner only while it is still the permanent the ability came from (CR
// 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "e26f7255-8b18-47d6-aafa-22ca22c35d2e",
		Name:                  "Farmstead Gleaner",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {Q}: Put a +1/+1 counter on this creature.",
			Cost:    game.AbilityCost{Mana: "{2}", UntapSelf: true},
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  effects.PutAPlusOneCounterOnThisIfStillHere,
		}},
	})
}
