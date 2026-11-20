package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eiganjo Free-Riders — Creature — Human Soldier {3}{W}, 3/4 (Saviors of Kamigawa, #5601):
//
//	"Flying
//	 At the beginning of your upkeep, return a white creature you control to its owner's hand."
//
// The upkeep return is the shared sokB1ReturnACreatureOfColour: not
// targeted and not optional, the white creature picked as it resolves,
// this creature itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0e381b7e-2d32-4a9a-a327-e9dd1236dc1b",
		Name:            "Eiganjo Free-Riders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{sokB1ReturnACreatureOfColour("Eiganjo Free-Riders", "W", "white")},
	})
}
