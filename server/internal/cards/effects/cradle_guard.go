package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cradle Guard — Creature — Treefolk, {1}{G}{G}, 4/4:
//
//	"Trample
//	 Echo {1}{G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "846c0867-6f0e-4e44-8f45-8f9720b62065",
		Name:            "Cradle Guard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Echo("Cradle Guard", "{1}{G}{G}"),
		},
	})
}
