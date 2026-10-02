package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Patrol — Creature — Goblin, {R}, 2/1:
//
//	"Echo {R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edf25cbe-fa28-44ca-b4be-1d2312e4f6ac",
		Name:         "Goblin Patrol",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Goblin Patrol", "{R}"),
		},
	})
}
