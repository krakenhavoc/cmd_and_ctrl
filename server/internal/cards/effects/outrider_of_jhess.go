package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outrider of Jhess — Creature — Human Knight {3}{U}, 2/2 (Shards of
// Alara, #5588):
//
//	"Exalted"
//
// Exalted is the canonical keyword (CR 702.83a, game/exalted.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccfe72b2-8a9a-4ad6-ab1d-676bb376b0ac",
		Name:            "Outrider of Jhess",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
	})
}
