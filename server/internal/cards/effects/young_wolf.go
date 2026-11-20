package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Young Wolf — Creature — Wolf {G}, 1/1 (Innistrad Remastered, S82 GAP,
// #5948):
//
//	"Undying (When this creature dies, if it had no +1/+1 counters on it,
//	 return it to the battlefield under its owner's control with a +1/+1
//	 counter on it.)"
//
// Undying is the shared keyword (game/undying_persist.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b492764-10b6-4506-be11-22daa9220a91",
		Name:            "Young Wolf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
	})
}
