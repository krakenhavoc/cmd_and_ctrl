package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benalish Infantry — Creature — Human Soldier {2}{W}, 1/3 (Weatherlight, S82 older sets batch 780, #5629):
//
//	"Banding"
//
// Banding rides PrintedKeywords; the engine owns the rest (CR 702.22, as Benalish Hero).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "543af1f4-f93f-4c7f-a067-ad2db599a4ae",
		Name:            "Benalish Infantry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
	})
}
