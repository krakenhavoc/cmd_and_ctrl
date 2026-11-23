package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Game-Trail Changeling — Creature — Shapeshifter {3}{G}{G}, 4/4
// (Morningtide, #5592):
//
//	"Changeling (This card is every creature type.)
//	 Trample"
//
// Two printed keywords: changeling is every creature type in every zone
// (CR 702.73a), and trample assigns excess combat damage to the player
// (CR 702.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "384611fb-aca2-4e2d-9125-44ea034de08b",
		Name:            "Game-Trail Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "trample"},
	})
}
