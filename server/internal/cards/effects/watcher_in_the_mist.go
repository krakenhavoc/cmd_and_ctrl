package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watcher in the Mist — Creature — Spirit {3}{U}{U}, 3/4 (Ravnica: Clue Edition):
//
//	"Flying
//	 When this creature enters, surveil 2."
//
// A.I.M. Synthoids's ETB trigger: the surveil goes on the stack, so an
// opponent can respond to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8c0fbea-a678-46b0-b385-e36f14c07672",
		Name:            "Watcher in the Mist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Watcher in the Mist — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
