package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oni of Wild Places — Creature — Demon Spirit {5}{R}, 6/5 (Saviors of Kamigawa, #5601):
//
//	"Haste
//	 At the beginning of your upkeep, return a red creature you control to its owner's hand."
//
// The upkeep return is the shared sokB1ReturnACreatureOfColour: not
// targeted and not optional, the red creature picked as it resolves,
// this creature itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f551093d-40b7-43ac-8023-4cfbb6c646da",
		Name:            "Oni of Wild Places",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{sokB1ReturnACreatureOfColour("Oni of Wild Places", "R", "red")},
	})
}
