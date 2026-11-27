package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honorable Scout — Creature — Human Soldier Scout {W}, 1/1 (Planeshift, S82 OLD b642, #5615):
//
//	"When this creature enters, you gain 2 life for each black and/or red creature
//	 target opponent controls."
//
// Kithkin Zealot's count over creatures: one that is black, red or both counts
// once, read as the trigger resolves (CR 608.2c). A target that has left the game
// makes the trigger do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16345f75-f33e-4a31-adb1-d3563358fcec",
		Name:         "Honorable Scout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Honorable Scout — you gain 2 life for each black and/or red creature target opponent controls",
				effects.OldB642GainPerCreatureOfColors(2, "B", "R")),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
