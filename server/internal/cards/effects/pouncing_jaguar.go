package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pouncing Jaguar — Creature — Cat, {G}, 2/2:
//
//	"Echo {G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70a2b5ea-dae4-46e8-87eb-8c8d8dfb242b",
		Name:         "Pouncing Jaguar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Pouncing Jaguar", "{G}"),
		},
	})
}
