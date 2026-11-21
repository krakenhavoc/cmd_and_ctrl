package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f2e1bb32-de4c-4abc-ac15-6e6deaa8fe5f",
		Name:         "Hematite Golem",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Hematite Golem — +2/+0 until end of turn"),
		}},
	})
}
