package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avian Changeling — Creature — Shapeshifter {2}{W}, 2/2 (Lorwyn, #5593):
//
//	"Changeling (This card is every creature type.)
//	 Flying"
//
// Two printed keywords: changeling is every creature type in every zone
// (CR 702.73a), and flying is evasion (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "767c1f97-03de-474d-9112-e7521ba3b3f3",
		Name:            "Avian Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "flying"},
	})
}
