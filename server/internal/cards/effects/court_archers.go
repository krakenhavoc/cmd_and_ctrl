package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Court Archers — Creature — Human Archer {2}{G}, 1/3 (Shards of Alara,
// #5588):
//
//	"Reach
//	 Exalted"
//
// Both are printed keywords; exalted's trigger is derived by the engine
// (CR 702.83a, game/exalted.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cc1cb63e-5811-44ab-af50-37a96f79fc78",
		Name:            "Court Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", game.KeywordExalted},
	})
}
