package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dusk Charger — Creature — Horse {3}{B}, 3/3:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 This creature gets +2/+2 as long as you have the city's
//	 blessing."
//
// A layer 7c self pump, read live; the blessing is never lost once earned (CR 702.131c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0190e13-89f6-4b51-a277-bf3eb842963e",
		Name:            "Dusk Charger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Static:          []game.StaticAbility{SelfPumpWhileCitysBlessing(2, 2)},
	})
}
