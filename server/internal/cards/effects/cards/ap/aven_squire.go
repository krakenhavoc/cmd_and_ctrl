package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Squire — Creature — Bird Soldier {1}{W}, 1/1 (Conflux, #5587):
//
//	"Flying
//	 Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Both are printed keywords. Exalted is the engine's keyword
// (game/exalted.go), which derives one trigger per instance when a
// creature you control attacks alone (CR 702.83a, 702.83b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01518a87-6b16-4b7b-805d-b652404c6356",
		Name:            "Aven Squire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordExalted},
	})
}
