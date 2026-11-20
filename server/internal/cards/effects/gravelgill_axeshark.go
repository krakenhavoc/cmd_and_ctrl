package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravelgill Axeshark — Creature — Merfolk Soldier {4}{U/B}, 3/3
// (Shadowmoor, #5591):
//
//	"Persist (When this creature dies, if it had no -1/-1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 -1/-1 counter on it.)"
//
// Persist is a canonical keyword whose dies trigger the engine derives
// from the departed permanent's last-known abilities (CR 702.79a, ADR 0113
// §4, #2075), as on Aerie Ouphes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2141b48d-5a04-40dd-a76b-457ac06b56c2",
		Name:            "Gravelgill Axeshark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
	})
}
