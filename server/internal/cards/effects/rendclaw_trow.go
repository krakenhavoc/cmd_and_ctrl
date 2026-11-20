package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rendclaw Trow — Creature — Troll {2}{B/G}, 2/2 (Eventide, #5590):
//
//	"Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	 Persist (When this creature dies, if it had no -1/-1 counters on it,
//	 return it to the battlefield under its owner's control with a -1/-1
//	 counter on it.)"
//
// Wither (CR 702.80) and persist (CR 702.79).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86177c7e-a6eb-448b-baf8-aedd06f0a2f3",
		Name:            "Rendclaw Trow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither, game.KeywordPersist},
	})
}
