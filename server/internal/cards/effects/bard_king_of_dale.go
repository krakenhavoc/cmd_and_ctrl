package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bard, King of Dale — Legendary Creature — Human Noble Archer {4}{W}{U},
// 3/5 (The Hobbit, HOB):
//
//	"Reach, vigilance
//	 If you would draw a card except the first one you draw in each of your
//	 draw steps, draw two cards instead.
//	 If one or more tokens would be created under your control, twice that
//	 many of those tokens are created instead."
//
// The two replacements are Alhammarret's Archive's draw half
// (YouDrawTwiceInsteadExceptTheFirst) and Parallel Lives'
// token doubler (TokensDoubled), each its own declared effect, so Bard
// beside either of them is x4 with the CR 616 ordering prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d05db2c1-a19a-4803-8e8a-fa2f9b798181",
		Name:            "Bard, King of Dale",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Replacements: []game.ReplacementEffect{
			YouDrawTwiceInsteadExceptTheFirst("Bard, King of Dale: draw two cards instead"),
			TokensDoubled("Bard, King of Dale: double tokens"),
		},
	})
}
