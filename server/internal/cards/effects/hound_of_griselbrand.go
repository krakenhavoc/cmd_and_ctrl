package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hound of Griselbrand — {2}{R}{R} Creature — Elemental Dog 2/2 (Avacyn Restored, #5570):
//
//	"Double strike
//	 Undying (When this creature dies, if it had no +1/+1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 +1/+1 counter on it.)"
//
// Double strike (CR 702.4) and undying (CR 702.93), the engine's keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "17425a60-7554-4740-9455-976edd63c106",
		Name:            "Hound of Griselbrand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", game.KeywordUndying},
	})
}
