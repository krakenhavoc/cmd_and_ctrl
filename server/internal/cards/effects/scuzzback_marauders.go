package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scuzzback Marauders — Creature — Goblin Warrior {4}{R/G}, 5/2
// (Shadowmoor, #5591):
//
//	"Trample
//	 Persist (When this creature dies, if it had no -1/-1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 -1/-1 counter on it.)"
//
// Trample is a printed keyword. Persist is a canonical keyword whose dies
// trigger the engine derives from the departed permanent's last-known
// abilities (CR 702.79a, ADR 0113 §4, #2075), as on Aerie Ouphes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "423a1fc2-f623-4b5f-9a3e-7f63b2f3cf5d",
		Name:            "Scuzzback Marauders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordPersist},
	})
}
