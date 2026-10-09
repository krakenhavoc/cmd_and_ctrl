package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snubhorn Sentry — Creature — Dinosaur {W}, 0/3:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 This creature gets +3/+0 as long as you have the city's
//	 blessing."
//
// A layer 7c self pump, read live; the blessing is never lost once earned (CR 702.131c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fc8d3933-f7fe-4849-aae7-737ce6f067de",
		Name:            "Snubhorn Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Static:          []game.StaticAbility{SelfPumpWhileCitysBlessing(3, 0)},
	})
}
