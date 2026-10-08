package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spire Winder — Creature — Snake {3}{U}, 2/3:
//
//	"Flying
//	 Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 This creature gets +1/+1 as long as you have the city's
//	 blessing."
//
// A layer 7c self pump, read live; the blessing is never lost once earned (CR 702.131c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8cfb24a-7e52-4e39-a9e1-969b4a65b62e",
		Name:            "Spire Winder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordAscend},
		Static:          []game.StaticAbility{SelfPumpWhileCitysBlessing(1, 1)},
	})
}
