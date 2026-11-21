package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Vines — Creature — Plant Wall {G}, 0/3 (Magic 2011, #5579):
//
//	"Defender
//	 Reach"
//
// Both are printed engine keywords, as on Whip Vine.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51ce0158-c9a5-4fa5-9704-46ab02f01b86",
		Name:            "Wall of Vines",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
	})
}
