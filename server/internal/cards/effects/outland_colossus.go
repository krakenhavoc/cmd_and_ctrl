package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outland Colossus — Creature — Giant {3}{G}{G}, 6/6:
//
//	"Renown 6 (When this creature deals combat damage to a player, if
//	 it isn't renowned, put six +1/+1 counters on it and it becomes
//	 renowned.)
//	 This creature can't be blocked by more than one creature."
//
// #2049: renown is the engine's keyword trigger (game/renown.go); the
// block limit is MaxBlockers on itself (Hungering Hydra's rule).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5b034e04-70e5-4783-8ac8-04d8e5a9f961",
		Name:            "Outland Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 6"},
		BlockRules:      []game.BlockRule{MaxBlockers(OnSelf(), 1)},
	})
}
