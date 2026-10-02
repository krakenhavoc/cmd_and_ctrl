package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simian Grunts — Creature — Ape, {2}{G}, 3/4:
//
//	"Flash
//	 Echo {2}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "158f4e87-60f1-43c1-b03d-9bbe4b824b39",
		Name:            "Simian Grunts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Echo("Simian Grunts", "{2}{G}"),
		},
	})
}
