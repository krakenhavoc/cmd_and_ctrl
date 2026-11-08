package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sterling Hound — Artifact Creature — Dog {3}, 3/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"When this creature enters, surveil 2."
//
// A.I.M. Synthoids' enters trigger: the surveil goes on the stack, so an
// opponent can respond to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3178c1f9-f644-4926-ac9a-d40e4db79955",
		Name:         "Sterling Hound",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sterling Hound — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
