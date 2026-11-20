package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fcfe4750-5311-45d2-8d8e-7b71e69368a2",
		Name:         "Grotag Thrasher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks(grotagThrasherLabel, targetCreatureCantBlockThisTurn(grotagThrasherLabel)),
				TargetCreature("target creature")),
		},
	})
}
