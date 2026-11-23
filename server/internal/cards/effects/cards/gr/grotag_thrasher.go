package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grotag Thrasher — Creature — Lizard {4}{R}, 3/3 (Worldwake, #5582):
//
//	"Whenever this creature attacks, target creature can't block this turn."
//
// Clamor Shaman's attack trigger with any creature as the target
// (targetCreatureCantBlockThisTurn).
//
// No simplification.
const grotagThrasherLabel = "Grotag Thrasher — target creature can't block this turn"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcfe4750-5311-45d2-8d8e-7b71e69368a2",
		Name:         "Grotag Thrasher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks(grotagThrasherLabel, effects.TargetCreatureCantBlockThisTurn(grotagThrasherLabel)),
				effects.TargetCreature("target creature")),
		},
	})
}
