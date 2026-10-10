package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keeper of the Quiet Hour — Artifact Creature — Chimera {3}, 3/2
// (Reality Fracture):
//
//	"When this creature enters, empower Jace 2."
//
// ADR 0139: an enters trigger on the keyword action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecdfee75-15fd-41ea-89d4-4bfc1fa4d032",
		Name:         "Keeper of the Quiet Hour",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Keeper of the Quiet Hour — empower Jace 2", Do(EmpowerJace{N: 2})),
		},
	})
}
