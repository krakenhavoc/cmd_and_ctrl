package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stampeding Serow — Creature — Antelope Beast {2}{G}{G}, 5/4 (Saviors of Kamigawa, #5601):
//
//	"Trample
//	 At the beginning of your upkeep, return a green creature you control to its owner's hand."
//
// The upkeep return is the shared sokB1ReturnACreatureOfColour: not
// targeted and not optional, the green creature picked as it resolves,
// this creature itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7e605a8c-d6d4-410a-aca4-52576d08d81c",
		Name:            "Stampeding Serow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{sokB1ReturnACreatureOfColour("Stampeding Serow", "G", "green")},
	})
}
