package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Palace Sentinels — Creature — Human Soldier {3}{W}, 2/4 (Commander
// Masters, #5948):
//
//	"When this creature enters, you become the monarch."
//
// The shared monarch ETB (CR 725.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d29d3b44-06dd-45dd-bbb2-506a6fce6a8d",
		Name:         "Palace Sentinels",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Palace Sentinels"),
		},
	})
}
