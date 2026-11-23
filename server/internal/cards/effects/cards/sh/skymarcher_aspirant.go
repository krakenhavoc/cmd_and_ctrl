package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skymarcher Aspirant — Creature — Vampire Soldier {W}, 2/1:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 This creature has flying as long as you have the city's
//	 blessing."
//
// A layer 6 keyword grant, read live; the blessing is never lost once earned (CR 702.131c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b6db929-1b6a-4372-8146-aa047258b552",
		Name:            "Skymarcher Aspirant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Static:          []game.StaticAbility{effects.SelfKeywordWhileCitysBlessing("flying")},
	})
}
