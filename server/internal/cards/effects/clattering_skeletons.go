package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clattering Skeletons — Creature — Skeleton {3}{B}, 4/3:
//
//	"When this creature dies, venture into the dungeon."
//
// The venture is game.VentureForEffect (ADR 0203 amendment of
// 2026-10-19, #4100).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e107382-f683-44ed-ada1-775d3ac79dd0",
		Name:         "Clattering Skeletons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies(ventureLabel("Clattering Skeletons"), Venture),
		},
	})
}
