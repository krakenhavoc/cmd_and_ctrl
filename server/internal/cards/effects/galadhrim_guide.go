package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galadhrim Guide — Creature — Elf Scout {3}{G}, 3/4 (The Hobbit Eternal,
// HOC):
//
//	"When this creature enters, scry 2."
//
// Faerie Seer's trigger without the flying (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9204de14-cbb6-40e5-9dcc-44cce3723e8e",
		Name:         "Galadhrim Guide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Galadhrim Guide — scry 2", Do(Scry{N: 2})),
		},
	})
}
