package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Putrid Goblin — Creature — Zombie Goblin {1}{B}, 2/2 (Modern Horizons, #5519):
//
//	"Persist (When this creature dies, if it had no -1/-1 counters on it,
//	 return it to the battlefield under its owner's control with a -1/-1
//	 counter on it.)"
//
// Persist is the engine's keyword (CR 702.79): it comes back once as a
// 1/1 and stays dead the second time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "27faa94b-c7a7-45e7-8630-80126fab74ba",
		Name:            "Putrid Goblin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
	})
}
