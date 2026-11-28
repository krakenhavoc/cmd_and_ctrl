package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Colossus — Artifact Creature — Phyrexian Golem {7}, 8/8 (Eighth
// Edition, #5607):
//
//	"This creature doesn't untap during your untap step.
//	 Pay 8 life: Untap this creature.
//	 This creature can't be blocked except by three or more creatures."
//
// Battered Golem's untap restriction (CR 502.3) and untap ability with a life
// payment as its whole cost (CR 119.4: a player can't pay more life than they
// have), and Pathrazer of Ulamog's three-blocker rule, refused at declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "caab555c-7852-418a-bdc3-f908af7e0932",
		Name:                  "Phyrexian Colossus",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		BlockRules:            []game.BlockRule{effects.MinBlockers(effects.OnSelf(), 3)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 8 life: Untap this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.PayLife(8),
			Effect:  effects.B31UntapSelf,
		}},
	})
}
