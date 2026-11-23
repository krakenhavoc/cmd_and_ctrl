package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek's Predator — Creature — Eldrazi Drone {3}{G}, 3/3 (Rise of the
// Eldrazi, #5581):
//
//	"When this creature enters, create two 0/1 colorless Eldrazi Spawn
//	 creature tokens. They have "Sacrifice this token: Add {C}.""
//
// The tokens are the catalog's Eldrazi Spawn (tokens.go), made by the
// enters trigger, which uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "452ef7d0-1351-40e6-9379-59cd8bb31da5",
		Name:         "Kozilek's Predator",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.RoeB1SpawnsWhenThisEnters("Kozilek's Predator", 2)},
	})
}
