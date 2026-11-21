package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emrakul's Hatcher — Creature — Eldrazi Drone {4}{R}, 3/3 (Rise of the
// Eldrazi, #5581):
//
//	"When this creature enters, create three 0/1 colorless Eldrazi Spawn
//	 creature tokens. They have "Sacrifice this token: Add {C}.""
//
// The tokens are the catalog's Eldrazi Spawn (tokens.go), made by the
// enters trigger, which uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ee811fb-c25d-4cf5-b396-539c36e8f90f",
		Name:         "Emrakul's Hatcher",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.RoeB1SpawnsWhenThisEnters("Emrakul's Hatcher", 3)},
	})
}
