package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Outrider of Jhess — Creature — Human Knight {3}{U}, 2/2 (Shards of
// Alara, #5588):
//
//	"Exalted"
//
// Exalted is the canonical keyword (CR 702.83a, game/exalted.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ccfe72b2-8a9a-4ad6-ab1d-676bb376b0ac",
		Name:            "Outrider of Jhess",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
	})
}
