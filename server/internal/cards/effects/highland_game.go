package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Highland Game — Creature — Elk {1}{G}, 2/1 (Khans of Tarkir, #5554):
//
//	"When this creature dies, you gain 2 life."
//
// Grasping Longneck's dies trigger (CR 603.10a), without the reach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e0d40a6-e0a4-4fb2-be72-314f7d66b7eb",
		Name:         "Highland Game",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Highland Game — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
