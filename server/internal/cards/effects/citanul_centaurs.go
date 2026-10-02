package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Citanul Centaurs — Creature — Centaur, {3}{G}, 6/3:
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Echo {3}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "831487cf-a377-4c9b-920d-c5bc68f17f24",
		Name:            "Citanul Centaurs",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Triggered: []game.TriggeredAbility{
			Echo("Citanul Centaurs", "{3}{G}"),
		},
	})
}
