package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dread Drone — Creature — Eldrazi Drone {4}{B}, 4/1 (Rise of the
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
	Register(Spec{
		OracleID:     "b6821131-9ae1-4a3c-8235-90deae33c05f",
		Name:         "Dread Drone",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{roeB1SpawnsWhenThisEnters("Dread Drone", 2)},
	})
}
