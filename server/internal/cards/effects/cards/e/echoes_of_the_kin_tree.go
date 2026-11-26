package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Echoes of the Kin Tree — Enchantment {1}{W} (Dragons of Tarkir, S82 old batch
// 330):
//
//	"{2}{W}: Bolster 1. (Choose a creature with the least toughness among
//	 creatures you control and put a +1/+1 counter on it.)"
//
// An activated ability with no target: the creature is chosen as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9863a7da-6f51-4d3e-9ace-e8ff201faf13",
		Name:         "Echoes of the Kin Tree",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Bolster 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{W}"),
			Effect:  effects.Do(effects.Bolster{N: 1}),
		}},
	})
}
