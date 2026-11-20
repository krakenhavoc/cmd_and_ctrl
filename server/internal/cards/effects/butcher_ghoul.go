package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Butcher Ghoul — {1}{B} Creature — Zombie 1/1 (Avacyn Restored, #5570):
//
//	"Undying (When this creature dies, if it had no +1/+1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 +1/+1 counter on it.)"
//
// Undying (CR 702.93), the engine's keyword, as on Young Wolf.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28caeed4-c175-469f-8d8e-f45bc0a51a1c",
		Name:            "Butcher Ghoul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
	})
}
