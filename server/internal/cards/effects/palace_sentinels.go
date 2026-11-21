package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Palace Sentinels — Creature — Human Soldier {3}{W}, 2/4 (Commander
// Masters, #5948):
//
//	"When this creature enters, you become the monarch."
//
// The shared monarch ETB (CR 725.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d29d3b44-06dd-45dd-bbb2-506a6fce6a8d",
		Name:         "Palace Sentinels",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Palace Sentinels"),
		},
	})
}
