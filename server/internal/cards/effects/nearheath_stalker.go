package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nearheath Stalker — Creature — Vampire Rogue {4}{R}, 4/1 (Dark Ascension, #5571):
//
//	"Undying (When this creature dies, if it had no +1/+1 counters on
//	 it, return it to the battlefield under its owner's control with
//	 a +1/+1 counter on it.)"
//
// Undying is PrintedKeywords (#2075), as on Young Wolf.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c804157e-2911-4973-83bf-f5befc09065d",
		Name:            "Nearheath Stalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
	})
}
