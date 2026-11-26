package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hagra Diabolist — Creature — Ogre Shaman Ally {4}{B}, 3/2 (Zendikar, S82 OLD
// b495):
//
//	"Whenever this creature or another Ally you control enters, you may have target
//	 player lose life equal to the number of Allies you control."
//
// A rally trigger (one per entering Ally), targeted and optional; the number is
// the Allies the controller has as it resolves (Old495HagraDiabolistDrain).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e2c1e0e-0a10-416a-9b50-96ee0cbbc24e",
		Name:         "Hagra Diabolist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.S82OldB313Rally("Hagra Diabolist — target player loses life equal to the number of Allies you control", effects.Old495HagraDiabolistDrain),
				effects.TargetPlayer("target player")),
				"Hagra Diabolist — have target player lose life equal to the number of Allies you control?"),
		},
	})
}
