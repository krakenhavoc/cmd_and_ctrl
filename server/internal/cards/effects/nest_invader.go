package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nest Invader — Creature — Eldrazi Drone {1}{G}, 2/2 (Rise of the
// Eldrazi, #5581):
//
//	"When this creature enters, create a 0/1 colorless Eldrazi Spawn
//	 creature token. It has "Sacrifice this token: Add {C}.""
//
// The token is the catalog's Eldrazi Spawn (tokens.go), made by the enters
// trigger, which uses the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7584a28-0829-4897-abf4-6be1a9232347",
		Name:         "Nest Invader",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{roeB1SpawnsWhenThisEnters("Nest Invader", 1)},
	})
}
