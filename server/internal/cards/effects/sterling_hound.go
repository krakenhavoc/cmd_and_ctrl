package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sterling Hound — Artifact Creature — Dog {3}, 3/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"When this creature enters, surveil 2. (Look at the top two cards of
//	 your library, then put any number of them into your graveyard and the
//	 rest on top of your library in any order.)"
//
// A.I.M. Synthoids' enters trigger: surveil 2 (CR 701.25a).
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
