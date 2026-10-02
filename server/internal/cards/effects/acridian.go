package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acridian — Creature — Insect, {1}{G}, 2/4:
//
//	"Echo {1}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6d0f1bc-b82a-4e64-a4bc-11ad3f54c71a",
		Name:         "Acridian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Acridian", "{1}{G}"),
		},
	})
}
