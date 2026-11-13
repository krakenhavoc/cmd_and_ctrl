package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hematite Golem — Artifact Creature — Golem {4}, 1/4 (Mirrodin, #5606):
//
//	"{1}{R}: This creature gets +2/+0 until end of turn."
//
// Avarax's pump at +2/+0: a layer 7c boost on the ability's source
// (CR 613.4c) that ends at cleanup (CR 514.2). It can be activated any
// number of times, each activation a separate boost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2e1bb32-de4c-4abc-ac15-6e6deaa8fe5f",
		Name:         "Hematite Golem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(2, 0, "Hematite Golem — +2/+0 until end of turn"),
		}},
	})
}
