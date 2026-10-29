package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// A.I.M. Synthoids — Artifact Creature — Robot Villain {2}, 1/3
// (Marvel Super Heroes):
//
//	"When this creature enters, surveil 2."
//
// Conduit Pylons's ETB trigger with a count of two: the surveil goes on the
// stack, so an opponent can respond to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10916f33-aaa6-43a1-9835-b60fd061161f",
		Name:         "A.I.M. Synthoids",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("A.I.M. Synthoids — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
