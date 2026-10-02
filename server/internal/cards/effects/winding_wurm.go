package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winding Wurm — Creature — Wurm, {4}{G}, 6/6:
//
//	"Echo {4}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f788188-d901-44f3-9fec-e2e5d0263140",
		Name:         "Winding Wurm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Winding Wurm", "{4}{G}"),
		},
	})
}
