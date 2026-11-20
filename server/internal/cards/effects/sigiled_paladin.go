package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigiled Paladin — Creature — Human Knight {W}{W}, 2/2 (Shards of Alara,
// #5588):
//
//	"First strike
//	 Exalted"
//
// Both are printed keywords; exalted's trigger is derived by the engine
// (CR 702.83a, game/exalted.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b68e99df-7899-493e-9453-59da14ff7de4",
		Name:            "Sigiled Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", game.KeywordExalted},
	})
}
