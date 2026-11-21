package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clattering Skeletons — Creature — Skeleton {3}{B}, 4/3:
//
//	"When this creature dies, venture into the dungeon."
//
// The venture is game.VentureForEffect (ADR 0203 amendment of
// 2026-10-19, #4100).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e107382-f683-44ed-ada1-775d3ac79dd0",
		Name:         "Clattering Skeletons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies(effects.VentureLabel("Clattering Skeletons"), effects.Venture),
		},
	})
}
